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
func (s *service) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	hash := common.GetContextHash(ctx)
	volumeName := req.GetName()
	// create StorageClass object
	sc, err := NewStorageClass(req.GetParameters())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	// get request volume capacity range
	requiredSizeByte, err := GetRequiredVolumeSizeByte(req.GetCapacityRange())
	if err != nil {
		return nil, status.Errorf(codes.OutOfRange, "unsupported capacity range, error: %s", err.Error())
	}
	klog.Infof("%s: Get required creating volume size %d(%dGi)", hash, requiredSizeByte, requiredSizeByte>>30)
	// check if volume exist for idempotent
	existVolume, err := s.storageProvider.ListVolume(volumeName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "find volume by name error: %s, %s", volumeName, err.Error())
	}
	if existVolume != nil {
		if common.IsValidCapacityBytes(existVolume.CapacityBytes, req.GetCapacityRange()) {
			existVolume.VolumeContext = req.GetParameters()
			return &csi.CreateVolumeResponse{Volume: existVolume}, nil
		} else {
			klog.Errorf("%s: volume %s/%s already exist but is incompatible", hash, volumeName, existVolume.VolumeId)
			return nil, status.Errorf(codes.AlreadyExists, "volume %s already exist but is incompatible", volumeName)
		}
	}
	var createError error
	if req.GetVolumeContentSource() == nil {
		// create an empty volume
		createError = s.storageProvider.CreateVolume(volumeName, requiredSizeByte, sc.Replica)
	} else {
		switch req.GetVolumeContentSource().GetType().(type) {
		// create from snapshot
		case *csi.VolumeContentSource_Snapshot:
			sourceVolumeName, snapshotName := common.SplitSnapshotName(req.GetVolumeContentSource().GetSnapshot().GetSnapshotId())
			snapshot, err := s.storageProvider.ListSnapshot(sourceVolumeName, snapshotName)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "list snapshot error:%v", err)
			}
			if snapshot == nil {
				return nil, status.Errorf(codes.NotFound, "Requested source snapshot %s@%s not found", sourceVolumeName, snapshotName)
			}
			createError = s.storageProvider.CloneVolume(sourceVolumeName, snapshotName, volumeName)
			// create from clone
		case *csi.VolumeContentSource_Volume:
			sourceVolumeName := req.GetVolumeContentSource().GetVolume().GetVolumeId()
			klog.Infof("%s: Clone volume %s from %s", hash, volumeName, sourceVolumeName)
			sourceVolume, err := s.storageProvider.ListVolume(sourceVolumeName)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			if sourceVolume == nil {
				return nil, status.Errorf(codes.NotFound, "cannot find content source volume id [%s]", sourceVolumeName)
			}
			createError = s.storageProvider.CloneVolume(sourceVolumeName, "", volumeName)
		}
	}
	if createError != nil {
		klog.Errorf("%s: Failed to create volume %s, contentSource %s, error: %v", hash, req.GetVolumeContentSource(), volumeName, err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	csiVolume := &csi.Volume{
		VolumeId:      volumeName,
		CapacityBytes: requiredSizeByte,
		VolumeContext: req.GetParameters(),
	}
	return &csi.CreateVolumeResponse{Volume: csiVolume}, nil
}

// This operation MUST be idempotent
// volume id is REQUIRED in csi.DeleteVolumeRequest
func (s *service) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	// For now the image get unconditionally deleted, but here retention policy can be checked
	volumeId := req.GetVolumeId()
	// ensure on call in-flight
	if acquired := s.locks.TryAcquire(volumeId); !acquired {
		return nil, status.Errorf(codes.Aborted, common.OperationPendingFmt, volumeId)
	}
	defer s.locks.Release(volumeId)
	// For idempotent:
	// MUST reply OK when volume does not exist
	volInfo, err := s.storageProvider.ListVolume(volumeId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if volInfo == nil {
		return &csi.DeleteVolumeResponse{}, nil
	}
	// Do delete volume
	err = retry.OnError(s.option.RetryTime, common.DefaultRetryErrorFunc, func() error {
		klog.Infof("Try to delete volume %s", volumeId)
		return s.storageProvider.DeleteVolume(volumeId)
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "Exceed retry times: "+err.Error())
	}
	return &csi.DeleteVolumeResponse{}, nil
}

// csi.ControllerPublishVolumeRequest: 	volume id 			+ Required
//										node id				+ Required
//										volume capability 	+ Required
//										readonly			+ Required (This field is NOT provided when requesting in Kubernetes)
func (s *service) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// This operation MUST be idempotent
// csi.ControllerUnpublishVolumeRequest: 	volume id	+Required
func (s *service) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// This operation MUST be idempotent
// csi.ValidateVolumeCapabilitiesRequest: 	volume id 			+ Required
// 											volume capability 	+ Required
func (s *service) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	// check volume exist
	volumeId := req.GetVolumeId()
	vol, err := s.storageProvider.ListVolume(volumeId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if vol == nil {
		return nil, status.Errorf(codes.NotFound, "volume %s does not exist", volumeId)
	}
	// check volume capabilities
	valid := s.option.ValidateVolumeCapabilities(req.GetVolumeCapabilities())
	if !valid {
		return nil, status.Errorf(codes.InvalidArgument, "Driver does not support volume capabilities:%v", req.GetVolumeCapabilities())
	}
	return &csi.ValidateVolumeCapabilitiesResponse{}, nil
}

// ControllerExpandVolume allows the CO to expand the size of a volume
// volume id is REQUIRED in csi.ControllerExpandVolumeRequest
// capacity range is REQUIRED in csi.ControllerExpandVolumeRequest
func (s *service) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	volumeId := req.GetVolumeId()
	if acquired := s.locks.TryAcquire(volumeId); !acquired {
		return nil, status.Errorf(codes.Aborted, common.OperationPendingFmt, volumeId)
	}
	defer s.locks.Release(volumeId)
	// get capacity
	requiredSizeBytes, err := GetRequiredVolumeSizeByte(req.GetCapacityRange())
	if err != nil {
		return nil, status.Errorf(codes.OutOfRange, err.Error())
	}

	// resize volume
	err = retry.OnError(s.option.RetryTime, common.DefaultRetryErrorFunc, func() error {
		klog.Infof("Try to expand volume %s", volumeId)
		return s.storageProvider.ResizeVolume(volumeId, requiredSizeBytes)
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &csi.ControllerExpandVolumeResponse{
		CapacityBytes:         requiredSizeBytes,
		NodeExpansionRequired: true,
	}, nil
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
func (s *service) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	srcVolId, snapName := req.GetSourceVolumeId(), req.GetName()
	// ensure one call in-flight
	if acquired := s.locks.TryAcquire(srcVolId); !acquired {
		return nil, status.Errorf(codes.Aborted, common.OperationPendingFmt, srcVolId)
	}
	defer s.locks.Release(srcVolId)
	// for idempotent
	snapshot, err := s.storageProvider.ListSnapshot(srcVolId, snapName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "find snapshot by name error: %s, %s", snapName, err.Error())
	}
	if snapshot != nil {
		return &csi.CreateSnapshotResponse{Snapshot: snapshot}, nil
	}
	// create snapshot
	err = s.storageProvider.CreateSnapshot(srcVolId, snapName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create snapshot [%s] from source volume [%s] error: %s", snapName, srcVolId, err.Error())
	}
	// query snapshot info
	snapshot, err = s.storageProvider.ListSnapshot(srcVolId, snapName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "find snapshot by name error: %s, %s", snapName, err.Error())
	}
	if snapshot != nil {
		return &csi.CreateSnapshotResponse{Snapshot: snapshot}, nil
	}
	return nil, status.Errorf(codes.Internal, "not find after create snapshot : %s", snapName, )
}

// CreateSnapshot allows the CO to delete a snapshot.
// This operation MUST be idempotent.
// Snapshot id is REQUIRED
func (s *service) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	fullSnapshotName := req.GetSnapshotId()
	if acquired := s.locks.TryAcquire(fullSnapshotName); !acquired {
		return nil, status.Errorf(codes.Aborted, common.OperationPendingFmt, fullSnapshotName)
	}
	defer s.locks.Release(fullSnapshotName)
	// 1. For idempotent:
	// MUST reply OK when snapshot does not exist
	volumeName, snapshotName := common.SplitSnapshotName(fullSnapshotName)
	exSnap, err := s.storageProvider.ListSnapshot(volumeName, snapshotName)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if exSnap == nil {
		klog.Infof("Cannot find snapshot id [%s].", fullSnapshotName)
		return &csi.DeleteSnapshotResponse{}, nil
	}
	// 2. Retry to delete snapshot
	err = retry.OnError(s.option.RetryTime, common.DefaultRetryErrorFunc, func() error {
		return s.storageProvider.DeleteSnapshot(volumeName, snapshotName)
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "Exceed retry times: "+err.Error())
	} else {
		return &csi.DeleteSnapshotResponse{}, nil
	}
}

func (s *service) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *service) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: s.option.ControllerCap,
	}, nil
}
