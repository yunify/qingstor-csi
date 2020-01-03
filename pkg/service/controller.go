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
	volumeName := req.GetName()
	// get request volume capacity range
	requiredSizeByte, err := GetRequiredVolumeSizeByte(req.GetCapacityRange())
	if err != nil {
		return nil, status.Errorf(codes.OutOfRange, "unsupported capacity range, error: %s", err.Error())
	}
	klog.Infof("Get required creating volume size %d(%dGi)", requiredSizeByte, requiredSizeByte>>30)
	// check if volume exist for idempotent
	existVolume, err := s.storageProvider.FindVolumeByName(volumeName, req.GetParameters())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "find volume by name error: %s, %s", volumeName, err.Error())
	}
	if existVolume != nil {
		if common.IsValidCapacityBytes(existVolume.CapacityBytes, req.GetCapacityRange()) {
			existVolume.VolumeContext = req.GetParameters()
			return &csi.CreateVolumeResponse{Volume: existVolume}, nil
		} else {
			klog.Errorf("volume %s/%s already exist but is incompatible",  volumeName, existVolume.VolumeId)
			return nil, status.Errorf(codes.AlreadyExists, "volume %s already exist but is incompatible", volumeName)
		}
	}
	var createError error
	var createVolumeID string
	if req.GetVolumeContentSource() == nil {
		// create an empty volume
		createVolumeID, createError = s.storageProvider.CreateVolume(volumeName, requiredSizeByte, req.GetParameters())
	} else {
		switch req.GetVolumeContentSource().GetType().(type) {
		case *csi.VolumeContentSource_Snapshot:
			// create from snapshot
			snapshotID := req.GetVolumeContentSource().GetSnapshot().GetSnapshotId()
			snapshot, err := s.storageProvider.FindSnapshot(snapshotID)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "list snapshot error:%v", err)
			}
			if snapshot == nil {
				return nil, status.Errorf(codes.NotFound, "Requested source snapshot %s not found", snapshotID)
			}
			createVolumeID, createError = s.storageProvider.CreateVolumeFromSnapshot(volumeName, snapshotID, req.GetParameters())
		case *csi.VolumeContentSource_Volume:
			// create from clone
			sourceVolumeId := req.GetVolumeContentSource().GetVolume().GetVolumeId()
			klog.Infof("Clone volume %s from %s", volumeName, sourceVolumeId)
			sourceVolume, err := s.storageProvider.FindVolume(sourceVolumeId)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			if sourceVolume == nil {
				return nil, status.Errorf(codes.NotFound, "cannot find content source volume id [%s]", sourceVolumeId)
			}
			createVolumeID, createError = s.storageProvider.CreateVolumeByClone(volumeName, sourceVolumeId, req.GetParameters())
		}
	}
	if createError != nil {
		klog.Errorf("Failed to create volume %s, contentSource %s, error: %v", req.GetVolumeContentSource(), volumeName, err)
		return nil, status.Error(codes.Internal, createError.Error())
	}
	csiVolume := &csi.Volume{
		VolumeId:      createVolumeID,
		CapacityBytes: requiredSizeByte,
		VolumeContext: req.GetParameters(),
	}
	return &csi.CreateVolumeResponse{Volume: csiVolume}, nil
}

// This operation MUST be idempotent
// volume id is REQUIRED in csi.DeleteVolumeRequest
func (s *service) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	// For now the image get unconditionally deleted, but here retention policy can be checked
	volumeID := req.GetVolumeId()
	// For idempotent:
	// MUST reply OK when volume does not exist
	volume, err := s.storageProvider.FindVolume(volumeID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if volume == nil {
		return &csi.DeleteVolumeResponse{}, nil
	}
	// Do delete volume
	err = s.storageProvider.DeleteVolume(volumeID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
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
	volumeID := req.GetVolumeId()
	volume, err := s.storageProvider.FindVolume(volumeID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if volume == nil {
		return nil, status.Errorf(codes.NotFound, "volume %s does not exist", volumeID)
	}
	// check volume capabilities
	if !s.option.ValidateVolumeCapabilities(req.GetVolumeCapabilities()) {
		return nil, status.Errorf(codes.InvalidArgument, "Driver does not support volume capabilities:%v", req.GetVolumeCapabilities())
	}
	return &csi.ValidateVolumeCapabilitiesResponse{}, nil
}

// ControllerExpandVolume allows the CO to expand the size of a volume
// volume id is REQUIRED in csi.ControllerExpandVolumeRequest
// capacity range is REQUIRED in csi.ControllerExpandVolumeRequest
func (s *service) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	volumeID := req.GetVolumeId()
	// get capacity
	requiredSizeBytes, err := GetRequiredVolumeSizeByte(req.GetCapacityRange())
	if err != nil {
		return nil, status.Errorf(codes.OutOfRange, err.Error())
	}
	// resize volume
	err = s.storageProvider.ResizeVolume(volumeID, requiredSizeBytes)
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
	sourceVolumeID, snapName := req.GetSourceVolumeId(), req.GetName()
	// for idempotent
	snapshot, err := s.storageProvider.FindSnapshotByName(sourceVolumeID, snapName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "find snapshot by name error: %s, %s", snapName, err.Error())
	}
	if snapshot != nil {
		return &csi.CreateSnapshotResponse{Snapshot: snapshot}, nil
	}
	// create snapshot
	err = s.storageProvider.CreateSnapshot(sourceVolumeID, snapName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create snapshot [%s] from source volume [%s] error: %s", snapName, sourceVolumeID, err.Error())
	}
	// query snapshot info
	snapshot, err = s.storageProvider.FindSnapshotByName(sourceVolumeID, snapName)
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
	snapshotID := req.GetSnapshotId()
	// 1. For idempotent:
	// MUST reply OK when snapshot does not exist
	exSnap, err := s.storageProvider.FindSnapshot(snapshotID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if exSnap == nil {
		return &csi.DeleteSnapshotResponse{}, nil
	}
	err = s.storageProvider.DeleteSnapshot(snapshotID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	} else {
		return &csi.DeleteSnapshotResponse{}, nil
	}
}

func (s *service) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *service) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	return &csi.ControllerGetCapabilitiesResponse{Capabilities: s.option.ControllerCap}, nil
}
