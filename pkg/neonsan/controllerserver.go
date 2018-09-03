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
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type controllerServer struct {
	*csicommon.DefaultControllerServer
}

// This operation MUST be idempotent
// csi.CreateVolumeRequest: name 				+Required
//							capability			+Required
func (cs *controllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	glog.Info("*************** Start CreateVolume ***************")
	defer glog.Info("=============== End CreateVolume ===============")

	glog.Info("Validate input arguments.")
	// Valid controller service capability
	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		glog.Errorf("Invalid create volume req: %v.", req)
		return nil, err
	}

	// Required volume capability
	if req.VolumeCapabilities == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capabilities missing in request.")
	} else if !ContainsVolumeCapabilities(cs.Driver.GetVolumeCapabilityAccessModes(), req.GetVolumeCapabilities()) {
		return nil, status.Error(codes.InvalidArgument, "Volume capabilities not match.")
	}

	// Required volume name
	if len(req.Name) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume name missing in request.")
	}
	volumeName := req.GetName()

	// Create StorageClass object
	glog.Info("Create StorageClass object.")
	sc, err := NewNeonsanStorageClassFromMap(req.GetParameters())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Get request volume size range
	glog.Info("Get request volume size.")
	requiredByte := req.GetCapacityRange().GetRequiredBytes()
	limitByte := req.GetCapacityRange().GetLimitBytes()
	requiredFormatByte := FormatVolumeSize(requiredByte, gib*int64(sc.StepSize))
	if limitByte == 0 {
		limitByte = Int64Max
	}

	// check volume range
	if requiredFormatByte > limitByte {
		glog.Errorf("Request capacity range [%d, %d] bytes, format required size: [%d] gb.",
			requiredByte, limitByte, requiredFormatByte)
		return nil, status.Error(codes.OutOfRange, "Unsupported capacity range.")
	}

	// Find exist volume name
	glog.Infof("Find duplicate volume name [%s].", volumeName)
	exVol, err := FindVolume(volumeName, sc.Pool)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if exVol != nil {
		glog.Infof("Request volume name: [%s], size: [%d], capacity range [%d,%d] Bytes, pool: [%s], replicas: [%d].",
			volumeName, requiredFormatByte, requiredByte, limitByte, sc.Pool, sc.Replicas)
		glog.Infof("Exist volume name: [%s], id: [%s], capacity: [%d] Bytes, pool: [%s], replicas: [%d].",
			exVol.name, exVol.id, exVol.size, exVol.pool, exVol.replicas)
		if exVol.size >= requiredByte && exVol.size <= limitByte && exVol.replicas == sc.Replicas {
			// exisiting volume is compatible with new request and should be
			// reused.
			return &csi.CreateVolumeResponse{
				Volume: &csi.Volume{
					Id:            exVol.name,
					CapacityBytes: exVol.size,
					Attributes:    req.GetParameters(),
				},
			}, nil
		}
		return nil, status.Errorf(codes.AlreadyExists, "Volume [%s] already exists but is incompatible.", volumeName)
	}
	glog.Infof("Not Found duplicate volume name [%s].", volumeName)

	// do create volume
	glog.Infof("Creating volume [%s] with [%d] bytes in pool [%s]...", volumeName, requiredFormatByte, sc.Pool)
	volumeInfo, err := CreateVolume(volumeName, sc.Pool, requiredFormatByte, sc.Replicas)
	if err != nil {
		glog.Errorf("Failed to create volume [%s] with [%d] bytes in pool [%s] with error [%v].", volumeName,
			requiredFormatByte, sc.Pool, err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	glog.Infof("Succeed to create volume [%s] with [%d] bytes in pool [%s].", volumeName, requiredFormatByte,
		sc.Pool)
	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			Id:            volumeInfo.name,
			CapacityBytes: volumeInfo.size,
			Attributes:    req.GetParameters(),
		},
	}, nil
}

// This operation MUST be idempotent
// volume id is REQUIRED in csi.DeleteVolumeRequest
func (cs *controllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	glog.Info("*************** Start DeleteVolume ***************")
	defer glog.Info("=============== End DeleteVolume ===============")

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
	volInfo, err := FindVolumeWithoutPool(volumeId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if volInfo == nil {
		glog.Warningf("Not found volume [%s].", volumeId)
		return &csi.DeleteVolumeResponse{}, nil
	}
	glog.Infof("Found volume [%s].", volumeId)

	// Do delete volume
	glog.Infof("Deleting volume [%s] in pool [%s]...", volumeId, volInfo.pool)
	err = DeleteVolume(volumeId, volInfo.pool)
	if err != nil {
		glog.Errorf("Failed to delete volume: [%s] in pool [%s] with error: [%v].", volumeId, volInfo.pool, err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	glog.Infof("Succeed to delete volume: [%s] in pool [%s]", volumeId, volInfo.pool)
	return &csi.DeleteVolumeResponse{}, nil
}

// This operation MUST be idempotent
// csi.ValidateVolumeCapabilitiesRequest: 	volume id 			+ Required
// 											volume capability 	+ Required
func (cs *controllerServer) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	glog.Info("*************** Start ValidateVolumeCapabilities ***************")
	defer glog.Info("=============== End ValidateVolumeCapabilities ===============")

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
	sc, err := NewNeonsanStorageClassFromMap(req.GetVolumeAttributes())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// check volume exist
	volumeId := req.GetVolumeId()
	glog.Infof("Find volume [%s] in pool [%s].", volumeId, sc.Pool)
	outVol, err := FindVolume(volumeId, sc.Pool)
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
	// Create StorageClass object
	glog.Info("Create StorageClass object.")
	sc, err := NewNeonsanStorageClassFromMap(req.GetParameters())
	if err != nil {
		glog.Info("Failed to create StorageClass object.")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	glog.Info("Succeed to create StorageClass object.")

	// Find pool information
	poolInfo, err := FindPool(sc.Pool)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if poolInfo == nil {
		glog.Infof("Cannot find pool [%s].", sc.Pool)
		return nil, status.Error(codes.NotFound, err.Error())
	}

	glog.Infof("Succeed to find pool name [%s], id [%s], total [%d] bytes, free [%d] bytes, used [%d] bytes.",
		poolInfo.name, poolInfo.id, poolInfo.total, poolInfo.free, poolInfo.used)
	return &csi.GetCapacityResponse{
		AvailableCapacity: poolInfo.free,
	}, nil
}

// CreateSnapshot
// Idempotent: If a snapshot corresponding to the specified snapshot name
// is already successfully cut and uploaded and is compatible with the
// specified source volume id and parameters in the CreateSnapshotRequest.
func (cs *controllerServer) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	glog.Info("*************** Start CreateSnapshot ***************")
	defer glog.Info("=============== End CreateSnapshot ===============")

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
	sc, err := NewNeonsanStorageClassFromMap(req.GetParameters())
	if err != nil {
		glog.Info("Failed to create StorageClass object.")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// for idempotency
	exSnap, err := FindSnapshot(req.GetName(), sc.Pool, req.GetSourceVolumeId())
	if err != nil {
		glog.Errorf("Failed to find snapshot [%s], [%s], error: [%s].", req.GetName(), sc.Pool, err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Volume id in Kubernetes is equal to NeonSAN's volume name.
	if req.GetSourceVolumeId() == exSnap.sourceVolumeName {
		// return snapshot already exists.
		glog.Warningf("Snapshot [%v] already exist. return this.", exSnap)
		return &csi.CreateSnapshotResponse{
			Snapshot: &csi.Snapshot{
				SizeBytes:      exSnap.sizeByte,
				Id:             exSnap.snapName,
				SourceVolumeId: exSnap.sourceVolumeName,
				CreatedAt:      exSnap.createdTime,
				Status: &csi.SnapshotStatus{
					Type: csi.SnapshotStatus_READY,
				},
			},
		}, nil
	} else {
		// snapshot already exists but is incompatible.
		glog.Errorf("Snapshot [%v] already exist. but not compatible with request [%v].", exSnap, req)
		return nil, status.Errorf(codes.AlreadyExists,
			"Snapshot [%s] already exists but is incompatible with the specified volume id [%v].", req.GetName(),
			req.GetSourceVolumeId())
	}

	// 3. do create snapshot
	glog.Infof("Create snapshot [%s] in pool [%s] from volume [%s]...", req.GetName(), sc.Pool, req.GetSourceVolumeId())
	snapInfo, err := CreateSnapshot(req.GetName(), sc.Pool, req.GetSourceVolumeId())
	if err != nil {
		glog.Errorf("Failed to create snapshot with error [%s].", err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	glog.Infof("Succeed to create snapshot [%v].", snapInfo)
	return &csi.CreateSnapshotResponse{
		Snapshot: &csi.Snapshot{
			SizeBytes:      snapInfo.sizeByte,
			Id:             snapInfo.snapName,
			SourceVolumeId: snapInfo.sourceVolumeName,
			CreatedAt:      snapInfo.createdTime,
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
	glog.Info("*************** Start DeleteSnapshot ***************")
	defer glog.Info("=============== End DeleteSnapshot ===============")

	// 1. Check input arguments
	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT); err != nil {
		glog.Warningf("Invalid delete snapshot req: %v.", req)
		return nil, err
	}
	if len(req.GetSnapshotId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Snapshot ID cannot be empty.")
	}

	// 2. Find snapshot by id
	exSnap, err := FindSnapshotWithoutPool(req.GetSnapshotId())
	if err != nil {
		glog.Errorf("Failed to find snapshot [%s], error [%s].", req.GetSnapshotId(), err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	if exSnap == nil {
		glog.Warningf("Snapshot [%s] does not exist.", req.GetSnapshotId())
		return &csi.DeleteSnapshotResponse{}, nil
	}

	// 3. Do delete snapshot
	glog.Infof("Delete snapshot [%v]...", exSnap)
	err = DeleteSnapshot(exSnap.snapName, exSnap.pool, exSnap.sourceVolumeName)
	if err != nil {
		glog.Errorf("Failed to delete snapshot [%v].", exSnap)
		return nil, status.Error(codes.Internal, err.Error())
	}
	glog.Infof("Succeed to delete snapshot [%v].", exSnap)
	return &csi.DeleteSnapshotResponse{}, nil
}

// Source Volume ID:
// Snapshot ID:
func (cs *controllerServer) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
