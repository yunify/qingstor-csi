/*
Copyright 2018 Yunify, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package neonsan

import (
	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/golang/glog"
	"github.com/kubernetes-csi/drivers/pkg/csi-common"
	"github.com/yunify/qingstor-csi/pkg/neonsan/cache"
	"github.com/yunify/qingstor-csi/pkg/neonsan/manager"
	"github.com/yunify/qingstor-csi/pkg/neonsan/util"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"os"
	"path"
	"strconv"
)

type controllerServer struct {
	*csicommon.DefaultControllerServer
	cache          cache.SnapshotCache
	creatingVolume map[string]bool
}

// This operation MUST be idempotent
// csi.CreateVolumeRequest: name 				+Required
//							capability			+Required
func (cs *controllerServer) CreateVolume(ctx context.Context,
	req *csi.CreateVolumeRequest) (resp *csi.CreateVolumeResponse, err error) {
	defer util.EntryFunction("CreateVolume")()

	glog.Info("Validate input arguments.")
	// Valid controller service capability
	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		glog.Errorf("Invalid create volume req: %v.", req)
		return nil, err
	}

	// Required volume capability
	if req.GetVolumeCapabilities() == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capabilities missing in request.")
	} else if !util.ContainsVolumeCapabilities(cs.Driver.GetVolumeCapabilityAccessModes(), req.GetVolumeCapabilities()) {
		return nil, status.Error(codes.InvalidArgument, "Volume capabilities not match.")
	}

	// Required volume name
	if len(req.GetName()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume name missing in request.")
	}
	volumeName := req.GetName()

	// check is creating
	if _, found := cs.creatingVolume[volumeName]; found {
		return nil, status.Errorf(codes.Aborted, "no more than one call in-flight for volume [%v]", req)
	} else {
		cs.creatingVolume[volumeName] = true
		defer delete(cs.creatingVolume, volumeName)
	}

	// Create StorageClass object
	glog.Info("Create StorageClass object.")
	sc, err := manager.NewNeonsanStorageClassFromMap(req.GetParameters())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Get request volume size range
	glog.Info("Get request volume size.")
	requiredByte := req.GetCapacityRange().GetRequiredBytes()
	limitByte := req.GetCapacityRange().GetLimitBytes()
	requiredFormatByte := util.FormatVolumeSize(requiredByte, util.Gib*int64(sc.StepSize))
	if limitByte == 0 {
		limitByte = util.Int64Max
	}

	// check volume range
	if requiredFormatByte > limitByte {
		glog.Errorf("Request capacity range [%d, %d] bytes, format required size: [%d] gb.",
			requiredByte, limitByte, requiredFormatByte)
		return nil, status.Error(codes.OutOfRange, "Unsupported capacity range.")
	}

	// Find exist volume name
	glog.Infof("Find duplicate volume name [%s].", volumeName)
	exVol, err := manager.FindVolume(volumeName, sc.Pool)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if exVol != nil {
		glog.Infof("Request volume name: [%s], size: [%d], capacity range [%d,%d] Bytes, pool: [%s], replicas: [%d].",
			volumeName, requiredFormatByte, requiredByte, limitByte, sc.Pool, sc.Replicas)
		glog.Infof("Exist volume name: [%s], id: [%s], capacity: [%d] Bytes, pool: [%s], replicas: [%d].",
			exVol.Name, exVol.Id, exVol.SizeByte, exVol.Pool, exVol.Replicas)
		if exVol.SizeByte >= requiredByte && exVol.SizeByte <= limitByte && exVol.Replicas == sc.Replicas {
			// exisiting volume is compatible with new request and should be
			// reused.
			return &csi.CreateVolumeResponse{
				Volume: &csi.Volume{
					Id:            exVol.Name,
					CapacityBytes: exVol.SizeByte,
					Attributes:    req.GetParameters(),
				},
			}, nil
		}
		return nil, status.Errorf(codes.AlreadyExists, "Volume [%s] already exists but is incompatible.", volumeName)
	}
	glog.Infof("Not Found duplicate volume name [%s].", volumeName)

	// Create volume from snapshot
	// Restore volume from snapshot
	if req.GetVolumeContentSource() != nil && req.GetVolumeContentSource().GetSnapshot() != nil {
		// create new volume
		volumeInfo, err := manager.CreateVolume(volumeName, sc.Pool, requiredFormatByte, sc.Replicas)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		deleteVolume := func() {
			if err != nil {
				glog.Errorf("")
				manager.DeleteVolume(volumeInfo.Name, volumeInfo.Pool)
			}
		}
		defer deleteVolume()

		// get snapshot id
		snapId := req.GetVolumeContentSource().GetSnapshot().GetId()
		glog.Infof("Restore volume [%s] from snapshot [%s]", volumeInfo.Name, snapId)
		snapInfo := cs.cache.Find(snapId)
		if snapInfo == nil {
			return nil, status.Errorf(codes.NotFound, "cannot find snapshot %v", snapId)
		}
		if snapInfo.Status != manager.SnapshotStatusOk {
			return nil, status.Errorf(codes.Internal, "status of snapshot %v is not ready", snapId)
		}
		// restore volume
		tempSnapshotFilepath := path.Join(util.TempSnapshotDir, snapInfo.Name)
		deleteTempFile := func() {
			err := os.Remove(tempSnapshotFilepath)
			if err != nil {
				if os.IsNotExist(err) {
					glog.Warningf("File [%s] does not exist", tempSnapshotFilepath)
				}
				glog.Errorf("Failed to remote [%s], error: [%s]", tempSnapshotFilepath, err)
			}
			glog.Infof("Succeed to remove [%s]", tempSnapshotFilepath)
		}
		defer deleteTempFile()

		// 1. export snapshot
		glog.Info("Export snapshot")
		err = manager.ExportSnapshot(manager.ExportSnapshotRequest{
			SnapName:   snapInfo.Name,
			SrcVolName: snapInfo.SrcVolName,
			PoolName:   snapInfo.Pool,
			FilePath:   path.Join(util.TempSnapshotDir, snapInfo.Name),
			Protocol:   manager.Protocol,
		})
		if err != nil {
			glog.Errorf("Export snapshot error: %s", err.Error())
			return nil, status.Errorf(codes.Internal, "export snapshot error: %v", err)
		}
		// 2. import snapshot
		glog.Info("Import snapshot")
		err = manager.ImportSnapshot(manager.ImportSnapshotRequest{
			VolName:  volumeInfo.Name,
			PoolName: volumeInfo.Pool,
			FilePath: path.Join(util.TempSnapshotDir, snapInfo.Name),
			Protocol: manager.Protocol,
		})
		if err != nil {
			glog.Errorf("Import snapshot error: %s", err.Error())
			return nil, status.Errorf(codes.Internal, "import snapshot error: %v", err)
		}
		// 3. restore volume
		glog.Info("Rollback snapshot")
		err = manager.RollbackSnapshot(manager.RollbackSnapshotRequest{
			VolumeName: volumeInfo.Name,
			Pool:       volumeInfo.Pool,
			SnapName:   snapInfo.Name,
		})
		if err != nil {
			glog.Errorf("Rollback snapshot error: %s", err.Error())
			return nil, status.Errorf(codes.Internal, "rollback snapshot error: %v", err)
		}
		return &csi.CreateVolumeResponse{
			Volume: &csi.Volume{
				Id:            volumeInfo.Name,
				CapacityBytes: volumeInfo.SizeByte,
				Attributes:    req.GetParameters(),
			},
		}, nil
	}

	// do create volume
	glog.Infof("Creating volume [%s] with [%d] bytes in pool [%s]...", volumeName, requiredFormatByte, sc.Pool)
	volumeInfo, err := manager.CreateVolume(volumeName, sc.Pool, requiredFormatByte, sc.Replicas)
	if err != nil {
		glog.Errorf("Failed to create volume [%s] with [%d] bytes in pool [%s] with error [%v].", volumeName,
			requiredFormatByte, sc.Pool, err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	glog.Infof("Succeed to create volume [%s] with [%d] bytes in pool [%s].", volumeName, requiredFormatByte,
		sc.Pool)

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			Id:            volumeInfo.Name,
			CapacityBytes: volumeInfo.SizeByte,
			Attributes:    req.GetParameters(),
		},
	}, nil
}

// This operation MUST be idempotent
// volume id is REQUIRED in csi.DeleteVolumeRequest
func (cs *controllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	defer util.EntryFunction("DeleteVolume")()

	glog.Info("Validate input arguments.")
	// Valid controller service capability
	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		glog.Errorf("Invalid delete volume req: %v.", req)
		return nil, err
	}

	// Check sanity of request Name, Volume Capabilities
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume id missing in request.")
	}
	// For now the image get unconditionally deleted, but here retention
	// policy can be checked
	volumeId := req.GetVolumeId()

	// For idempotent:
	// MUST reply OK when volume does not exist
	glog.Infof("Find volume [%s].", volumeId)
	volInfo, err := manager.FindVolumeWithoutPool(volumeId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if volInfo == nil {
		glog.Warningf("Not found volume [%s].", volumeId)
		return &csi.DeleteVolumeResponse{}, nil
	}
	glog.Infof("Found volume [%s].", volumeId)

	// Do delete volume
	glog.Infof("Deleting volume [%s] in pool [%s]...", volumeId, volInfo.Pool)
	err = manager.DeleteVolume(volumeId, volInfo.Pool)
	if err != nil {
		glog.Errorf("Failed to delete volume: [%s] in pool [%s] with error: [%v].", volumeId, volInfo.Pool, err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	cs.cache.Delete(volumeId)
	glog.Infof("Succeed to delete volume: [%s] in pool [%s]", volumeId, volInfo.Pool)
	return &csi.DeleteVolumeResponse{}, nil
}

// This operation MUST be idempotent
// csi.ValidateVolumeCapabilitiesRequest: 	volume id 			+ Required
// 											volume capability 	+ Required
func (cs *controllerServer) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	defer util.EntryFunction("ValidateVolumeCapabilities")()

	glog.Info("Validate input arguments.")
	// require volume id parameter
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "No volume id is provided")
	}

	// require capability parameter
	if len(req.GetVolumeCapabilities()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "No volume capabilities are provided")
	}

	// Create StorageClass object
	glog.Info("Create StorageClass object.")
	sc, err := manager.NewNeonsanStorageClassFromMap(req.GetVolumeAttributes())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// check volume exist
	volumeId := req.GetVolumeId()
	glog.Infof("Find volume [%s] in pool [%s].", volumeId, sc.Pool)
	outVol, err := manager.FindVolume(volumeId, sc.Pool)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if outVol == nil {
		glog.Errorf("Not found volume [%s] in pool [%s].", volumeId, sc.Pool)
		return nil, status.Errorf(codes.NotFound, "Volume %s does not exist.", volumeId)
	}
	glog.Infof("Found volume [%s] in pool [%s].", volumeId, sc.Pool)

	// check capability
	glog.Info("Check capability.")
	for _, c := range req.GetVolumeCapabilities() {
		found := false
		for _, c1 := range cs.Driver.GetVolumeCapabilityAccessModes() {
			if c1.GetMode() == c.GetAccessMode().GetMode() {
				found = true
			}
		}
		if !found {
			return &csi.ValidateVolumeCapabilitiesResponse{
				Supported: false,
				Message:   "Driver does not support mode:" + c.GetAccessMode().GetMode().String(),
			}, nil
		}
	}

	return &csi.ValidateVolumeCapabilitiesResponse{
		Supported: true,
	}, nil
}

// GetCapacity: allow the CO to query the capacity of the storage pool from
// which the controller provisions volumes.
func (cs *controllerServer) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	defer util.EntryFunction("GetCapacity")()

	// Create StorageClass object
	glog.Info("Create StorageClass object.")
	sc, err := manager.NewNeonsanStorageClassFromMap(req.GetParameters())
	if err != nil {
		glog.Info("Failed to create StorageClass object.")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	glog.Info("Succeed to create StorageClass object.")

	// Find pool information
	poolInfo, err := manager.FindPool(sc.Pool)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if poolInfo == nil {
		glog.Infof("Cannot find pool [%s].", sc.Pool)
		return nil, status.Error(codes.NotFound, err.Error())
	}

	glog.Infof("Succeed to find pool name [%s], id [%s], total [%d] bytes, free [%d] bytes, used [%d] bytes.",
		poolInfo.Name, poolInfo.Id, poolInfo.TotalByte, poolInfo.FreeByte, poolInfo.UsedByte)
	return &csi.GetCapacityResponse{
		AvailableCapacity: poolInfo.FreeByte,
	}, nil
}

// CreateSnapshot
// Idempotent: If a snapshot corresponding to the specified snapshot name
// is already successfully cut and uploaded and is compatible with the
// specified source volume id and parameters in the CreateSnapshotRequest.
func (cs *controllerServer) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	defer util.EntryFunction("CreateSnapshot")()

	// 1. Check input arguments
	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT); err != nil {
		glog.Warningf("Invalid create snapshot req: [%v].", req)
		return nil, err
	}
	if len(req.GetName()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Snapshot Name cannot be empty.")
	}
	if len(req.GetSourceVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Source Volume ID cannot be empty.")
	}

	// 2. Check if the snapshot already exists.
	// get storage class
	sc, err := manager.NewNeonsanStorageClassFromMap(req.GetParameters())
	if err != nil {
		glog.Info("Failed to create StorageClass object.")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// 3. Snapshot already exist
	if exSnap := cs.cache.Find(req.GetName()); exSnap != nil {
		if exSnap.SrcVolName == req.GetSourceVolumeId() {
			// compatible with source volume id, return OK
			glog.Warningf("Snapshot [%v] already exist and compatible with srcVolumeId in request, return this.",
				exSnap)
			return &csi.CreateSnapshotResponse{
				Snapshot: &csi.Snapshot{
					SizeBytes:      exSnap.SizeByte,
					Id:             exSnap.Name,
					SourceVolumeId: exSnap.SrcVolName,
					CreatedAt:      exSnap.CreatedTime,
					Status: &csi.SnapshotStatus{
						Type: csi.SnapshotStatus_READY,
					},
				},
			}, nil
		} else {
			// incompatible with source volume id, return ALREADY_EXIST
			glog.Errorf("Snapshot [%v] already exist but incompatible with srcVolumeId in request [%v].", exSnap, req)
			return nil, status.Errorf(codes.AlreadyExists,
				"Snapshot [%s] already exists but is incompatible with the specified volume id [%v].", req.GetName(),
				req.GetSourceVolumeId())
		}
	}

	// 3. do create snapshot
	glog.Infof("Create snapshot [%s] in pool [%s] from volume [%s]...", req.GetName(), sc.Pool, req.GetSourceVolumeId())
	snapInfo, err := manager.CreateSnapshot(req.GetName(), req.GetSourceVolumeId(), sc.Pool)
	if err != nil {
		glog.Errorf("Failed to create snapshot with error [%s].", err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	glog.Infof("Add snapshot [%v] into cache", snapInfo)
	if !cs.cache.Add(snapInfo) {
		glog.Warningf("Snapshot [%s] already exist in cache", snapInfo.Name)
	}

	glog.Infof("Succeed to create snapshot [%v].", snapInfo)
	return &csi.CreateSnapshotResponse{
		Snapshot: &csi.Snapshot{
			SizeBytes:      snapInfo.SizeByte,
			Id:             snapInfo.Name,
			SourceVolumeId: snapInfo.SrcVolName,
			CreatedAt:      snapInfo.CreatedTime,
			Status: &csi.SnapshotStatus{
				Type: csi.SnapshotStatus_READY,
			},
		},
	}, nil
}

// DeleteSnapshot must release the storage space associated with the snapshot
// Idempotent: If a snapshot corresponding to the specified snapshot id does
// not exist, the plugin MUST reply OK.
// csi.DeleteSnapshotRequest: snapshot id is required.
func (cs *controllerServer) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	defer util.EntryFunction("DeleteSnapshot")()

	// 1. Check input arguments
	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT); err != nil {
		glog.Warningf("Invalid delete snapshot req: %v.", req)
		return nil, err
	}
	if len(req.GetSnapshotId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Snapshot ID cannot be empty.")
	}

	// 2. Find snapshot by id
	exSnap := cs.cache.Find(req.GetSnapshotId())
	if exSnap == nil {
		glog.Warningf("Snapshot [%s] does not exist.", req.GetSnapshotId())
		return &csi.DeleteSnapshotResponse{}, nil
	}

	// 3. Do delete snapshot
	glog.Infof("Delete snapshot [%v]...", exSnap)
	err := manager.DeleteSnapshot(exSnap.Name, exSnap.SrcVolName, exSnap.Pool)
	if err != nil {
		glog.Errorf("Failed to delete snapshot [%v].", exSnap)
		return nil, status.Error(codes.Internal, err.Error())
	}
	cs.cache.Delete(exSnap.Name)
	glog.Infof("Succeed to delete snapshot [%v].", exSnap)
	return &csi.DeleteSnapshotResponse{}, nil
}

// ListSnapshots return the information about all snapshots on the storage
// system within the given parameters regardless of how they were created.
// ListSnapshots SHALL NOT list a snapshot that is being created but has not
// been cut successfully yet.
// Use token get next page when only max entries in request.
func (cs *controllerServer) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	defer util.EntryFunction("ListSnapshots")()
	// check input args
	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_LIST_SNAPSHOTS); err != nil {
		glog.Warningf("Invalid list snapshot req: %v", req)
		return nil, err
	}
	snapId := req.GetSnapshotId()
	srcVolId := req.GetSourceVolumeId()

	// case 1: snapshot id
	if len(snapId) != 0 {
		snapInfo := cs.cache.Find(req.GetSnapshotId())
		if snapInfo == nil {
			return &csi.ListSnapshotsResponse{}, nil
		}

		if len(srcVolId) != 0 && srcVolId != snapInfo.SrcVolName {
			return nil, status.Error(codes.Internal, "mismatch snapshot and volume name")
		}
		return &csi.ListSnapshotsResponse{
			Entries: []*csi.ListSnapshotsResponse_Entry{
				{
					Snapshot: manager.ConvertNeonToCsiSnap(snapInfo),
				},
			},
		}, nil
	}

	// case 2: volume id
	// According to csi test, we need not support pageable.
	if len(srcVolId) != 0 {
		// consult
		volInfo, err := manager.FindVolumeWithoutPool(srcVolId)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		// should return empty when the specify source volume id is not exist
		if volInfo == nil {
			return &csi.ListSnapshotsResponse{}, nil
		}

		snapInfoList, err := manager.ListSnapshotByVolume(srcVolId, volInfo.Pool)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		// non-pageable
		return &csi.ListSnapshotsResponse{
			Entries: manager.ConvertNeonSnapToListSnapResp(snapInfoList),
		}, nil
	}

	// case 3: non volume id provided
	// must consider pageable
	// get full snapshot list
	snapList := cs.cache.List()
	// get max entries
	maxEntries := int(req.GetMaxEntries())
	if maxEntries == 0 || maxEntries >= len(snapList) {
		maxEntries = len(snapList)
	}
	// get starting token
	startToken := 0
	if len(req.GetStartingToken()) != 0 {
		var err error
		startToken, err = strconv.Atoi(req.GetStartingToken())
		// token must be an integer
		if err != nil {
			return nil, status.Error(codes.Aborted, err.Error())
		}
		// should fail when the starting_token is greater than total number of snapshots
		if startToken > len(snapList) || startToken < 0 {
			return nil, status.Errorf(codes.Aborted, "invalid starting token")
		}
	}
	endToken := startToken + maxEntries
	glog.Infof("list snapshot [%d, %d) len [%d] with max entries [%d]", startToken, endToken, len(snapList), maxEntries)
	if endToken >= len(snapList) {
		return &csi.ListSnapshotsResponse{
			Entries: manager.ConvertNeonSnapToListSnapResp(snapList[startToken:]),
		}, nil
	} else {
		return &csi.ListSnapshotsResponse{
			Entries:   manager.ConvertNeonSnapToListSnapResp(snapList[startToken:endToken]),
			NextToken: strconv.Itoa(endToken),
		}, nil
	}
}
