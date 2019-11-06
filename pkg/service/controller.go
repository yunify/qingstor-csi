/*
Copyright (C) 2018 Yunify, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this work except in compliance with the License.
You may obtain a copy of the License in the LICENSE file, or at:

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/yunify/qingstor-csi/pkg/common"
	"github.com/yunify/qingstor-csi/pkg/service/neonsan"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog"
)

// This operation MUST be idempotent
// This operation MAY create three types of volumes:
// 1. Empty volumes: CREATE_DELETE_VOLUME
// 2. Restore volume from snapshot: CREATE_DELETE_VOLUME and CREATE_DELETE_SNAPSHOT
// 3. Clone volume: CREATE_DELETE_VOLUME and CLONE_VOLUME
// csi.CreateVolumeRequest: name 				+Required
//							capability			+Required
func (s *service) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse,
	error) {
	hash := common.GetContextHash(ctx)
	volName := req.GetName()
	// create StorageClass object
	sc, err := neonsan.NewStorageClassFromMap(req.GetParameters())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	klog.Infof("%s: Create storage class %v", hash, sc)

	// get request volume capacity range
	requiredSizeByte, err := GetRequiredVolumeSizeByte(req.GetCapacityRange())
	if err != nil {
		return nil, status.Errorf(codes.OutOfRange, "unsupported capacity range, error: %s", err.Error())
	}
	klog.Infof("%s: Get required creating volume size in bytes %d", hash, requiredSizeByte)

	// should not fail when requesting to create a volume with already existing name and same capacity
	// should fail when requesting to create a volume with already existing name and different capacity.
	exVol, err := s.storageProvider.FindVolumeByName(volName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "find volume by name error: %s, %s", volName, err.Error())
	}
	if exVol != nil {
		klog.Infof("%s: Request volume name: %s, request size %d bytes", hash, volName, requiredSizeByte)
		klog.Infof("%s: Exist volume name: %s, id: %s, capacity: %d bytes",
			hash, *exVol.VolumeName, *exVol.VolumeID, common.GibToByte(*exVol.Size))
		exVolSizeByte := common.GibToByte(*exVol.Size)
		if common.IsValidCapacityBytes(exVolSizeByte, req.GetCapacityRange()) {
			// existing volume is compatible with new request and should be reused.
			klog.Infof("Volume %s already exists and compatible with %s", volName, *exVol.VolumeID)
			return &csi.CreateVolumeResponse{
				Volume: &csi.Volume{
					VolumeId:      *exVol.VolumeID,
					CapacityBytes: exVolSizeByte,
					VolumeContext: req.GetParameters(),
				},
			}, nil
		} else {
			klog.Errorf("%s: volume %s/%s already exist but is incompatible", hash, volName, *exVol.VolumeID)
			return nil, status.Errorf(codes.AlreadyExists, "volume %s already exist but is incompatible", volName)
		}
	}

	if req.GetVolumeContentSource() != nil {
		switch req.GetVolumeContentSource().GetType().(type) {
		case *csi.VolumeContentSource_Snapshot:
			return nil, status.Errorf(codes.Unimplemented, "create from snapshot")
		case *csi.VolumeContentSource_Volume:
			return nil, status.Errorf(codes.Unimplemented, "create from clone")
		}
	} else {
		// create an empty volume
		requiredSizeGib := common.ByteCeilToGib(requiredSizeByte)
		klog.Infof("%s: Creating empty volume %s with %d Gib ", hash, volName, requiredSizeGib)
		newVolId, err := s.storageProvider.CreateVolume(volName, requiredSizeGib, sc.Replica)
		if err != nil {
			klog.Errorf("%s: Failed to create volume %s, error: %v", hash, volName, err)
			return nil, status.Error(codes.Internal, err.Error())
		}
		newVolInfo, err := s.storageProvider.FindVolume(newVolId)
		if err != nil {
			klog.Errorf("%s: Failed to find volume %s, error: %v", hash, newVolId, err)
			return nil, status.Error(codes.Internal, err.Error())
		}
		if newVolInfo == nil {
			klog.Infof("%s: Cannot find just created volume [%s/%s], please retrying later.", hash, volName, newVolId)
			return nil, status.Errorf(codes.Aborted, "cannot find volume %s", newVolId)
		}
		klog.Infof("%s: Succeed to create empty volume [%s/%s].", hash, volName, newVolId)
		return &csi.CreateVolumeResponse{
			Volume: &csi.Volume{
				VolumeId:      newVolId,
				CapacityBytes: requiredSizeByte,
				VolumeContext: req.GetParameters(),
			},
		}, nil
	}

	return nil, status.Error(codes.Internal, "The plugin SHOULD NOT run here, "+
		"please report at https://github.com/yunify/qingstor-csi.")
}

// This operation MUST be idempotent
// volume id is REQUIRED in csi.DeleteVolumeRequest
func (s *service) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	// For now the image get unconditionally deleted, but here retention policy can be checked
	volumeId := req.GetVolumeId()
	// ensure on call in-flight
	klog.Infof("try to lock resource %s", volumeId)
	if acquired := s.locks.TryAcquire(volumeId); !acquired {
		return nil, status.Errorf(codes.Aborted, common.OperationPendingFmt, volumeId)
	}
	defer s.locks.Release(volumeId)

	// For idempotent:
	// MUST reply OK when volume does not exist
	volInfo, err := s.storageProvider.FindVolume(volumeId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if volInfo == nil {
		return &csi.DeleteVolumeResponse{}, nil
	}
	// Do delete volume
	err = retry.OnError(s.retryTime, func(e error) bool { return true }, func() error {
		klog.Infof("Try to delete volume %s", volumeId)
		if err = s.storageProvider.DeleteVolume(volumeId); err != nil {
			klog.Errorf("Failed to delete volume %s, error: %v", volumeId, err)
			return err
		} else {
			klog.Infof("Succeed to delete volume %s", volumeId)
			return nil
		}
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "Exceed retry times: "+err.Error())
	} else {
		return &csi.DeleteVolumeResponse{}, nil
	}
}

// csi.ControllerPublishVolumeRequest: 	volume id 			+ Required
//										node id				+ Required
//										volume capability 	+ Required
//										readonly			+ Required (This field is NOT provided when requesting in Kubernetes)
func (s *service) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.
	ControllerPublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// This operation MUST be idempotent
// csi.ControllerUnpublishVolumeRequest: 	volume id	+Required
func (s *service) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.
	ControllerUnpublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// This operation MUST be idempotent
// csi.ValidateVolumeCapabilitiesRequest: 	volume id 			+ Required
// 											volume capability 	+ Required
func (s *service) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.
	ValidateVolumeCapabilitiesResponse, error) {
	// check volume exist
	volumeId := req.GetVolumeId()
	vol, err := s.storageProvider.FindVolume(volumeId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if vol == nil {
		return nil, status.Errorf(codes.NotFound, "volume %s does not exist", volumeId)
	}

	// check capability
	for _, c := range req.GetVolumeCapabilities() {
		found := false
		for _, c1 := range s.option.GetVolumeCapability() {
			if c1.GetMode() == c.GetAccessMode().GetMode() {
				found = true
			}
		}
		if !found {
			return &csi.ValidateVolumeCapabilitiesResponse{
				Message: "Driver does not support mode:" + c.GetAccessMode().GetMode().String(),
			}, status.Error(codes.InvalidArgument, "Driver does not support mode:"+c.GetAccessMode().GetMode().String())
		}
	}
	return &csi.ValidateVolumeCapabilitiesResponse{}, nil
}

// ControllerExpandVolume allows the CO to expand the size of a volume
// volume id is REQUIRED in csi.ControllerExpandVolumeRequest
// capacity range is REQUIRED in csi.ControllerExpandVolumeRequest
func (s *service) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest,
) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "expand volume not implement")
}

func (s *service) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *service) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// CreateSnapshot allows the CO to create a snapshot.
// This operation MUST be idempotent.
// 1. If snapshot successfully cut and ready to use, the plugin MUST reply 0 OK.
// 2. If an error occurs before a snapshot is cut, the plugin SHOULD reply a corresponding error code.
// 3. If snapshot successfully cut but still being precessed,
// the plugin SHOULD return 0 OK and ready_to_use SHOULD be set to false.
// Source volume id is REQUIRED
// Snapshot name is REQUIRED
func (s *service) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.
	CreateSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")

}

// CreateSnapshot allows the CO to delete a snapshot.
// This operation MUST be idempotent.
// Snapshot id is REQUIRED
func (s *service) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse,
	error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *service) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *service) ControllerGetCapabilities(ctx context.Context,
	req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: s.option.GetControllerCapability(),
	}, nil
}
