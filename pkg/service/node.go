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
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/util/mount"
	"k8s.io/kubernetes/pkg/util/resizefs"
	k8sVolume "k8s.io/kubernetes/pkg/volume"
	"os"
)

// This operation MUST be idempotent
// csi.NodeStageVolumeRequest: 	volume id			+ Required
//								stage target path	+ Required
//								volume capability	+ Required
func (s *service) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	volumeID, targetPath := req.GetVolumeId(), req.GetStagingTargetPath()
	fsType := req.VolumeCapability.GetMount().GetFsType()
	// Check volume exist
	volInfo, err := s.storageProvider.FindVolume(volumeID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if volInfo == nil {
		return nil, status.Errorf(codes.NotFound, "Volume %s does not exist", volumeID)
	}
	// if volume already mounted
	notMnt, err := s.mounter.Interface.IsLikelyNotMountPoint(targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// already mount
	if !notMnt {
		return &csi.NodeStageVolumeResponse{}, nil
	}
	// Attach if need
	err = s.storageProvider.NodeAttachVolume(volumeID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	devicePath, err := s.storageProvider.NodeGetDevice(volumeID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// do mount
	if err := s.mounter.FormatAndMount(devicePath, targetPath, fsType, []string{}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &csi.NodeStageVolumeResponse{}, nil
}

// This operation MUST be idempotent
// csi.NodeUnstageVolumeRequest:	volume id	+ Required
//									target path	+ Required
func (s *service) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	// set parameter
	volumeID, targetPath := req.GetVolumeId(), req.GetStagingTargetPath()
	// Check volume exist
	volume, err := s.storageProvider.FindVolume(volumeID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if volume == nil {
		return nil, status.Errorf(codes.NotFound, "Volume %s does not exist", volumeID)
	}
	// check targetPath is mounted
	// For idempotent:
	// If the volume corresponding to the volume id is not staged to the staging target path,
	// the plugin MUST reply 0 OK.
	mounter := s.mounter.Interface
	notMnt, err := mounter.IsLikelyNotMountPoint(targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if notMnt {
		return &csi.NodeUnstageVolumeResponse{}, nil
	}
	// count mount point
	_, cnt, err := mount.GetDeviceNameFromMount(mounter, targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// do unmount
	err = mounter.Unmount(targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	cnt--
	if cnt > 0 {
		klog.Errorf("Volume %s still mounted in instance %s", volumeID, s.option.NodeId)
		return nil, status.Error(codes.Internal, "unmount failed")
	}
	// node detach volume
	err = s.storageProvider.NodeDetachVolume(volumeID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &csi.NodeUnstageVolumeResponse{}, nil
}

// This operation MUST be idempotent
// If the volume corresponding to the volume id has already been published at the specified target path,
// and is compatible with the specified volume capability and readonly flag, the plugin MUST reply 0 OK.
// csi.NodePublishVolumeRequest:	volume id			+ Required
//									target path			+ Required
//									volume capability	+ Required
//									read only			+ Required (This field is NOT provided when requesting in Kubernetes)
func (s *service) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	// set parameter
	volumeID, targetPath, stagePath := req.GetVolumeId(), req.GetTargetPath(), req.GetStagingTargetPath()
	// set fsType
	fsType := req.GetVolumeCapability().GetMount().GetFsType()
	// Check volume exist
	volInfo, err := s.storageProvider.FindVolume(volumeID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if volInfo == nil {
		return nil, status.Errorf(codes.NotFound, "Volume %s does not exist", volumeID)
	}
	// Make dir if dir not presents
	_, err = os.Stat(targetPath)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(targetPath, 0750); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	// check targetPath is mounted
	mounter := s.mounter.Interface
	notMnt, err := mounter.IsNotMountPoint(targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// For idempotent:
	// If the volume corresponding to the volume id has already been published at the specified target path,
	// and is compatible with the specified volume capability and readonly flag, the plugin MUST reply 0 OK.
	if !notMnt {
		return &csi.NodePublishVolumeResponse{}, nil
	}
	// set bind mount options
	options := []string{"bind"}
	if req.GetReadonly() == true {
		options = append(options, "ro")
	}
	if err := mounter.Mount(stagePath, targetPath, fsType, options); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &csi.NodePublishVolumeResponse{}, nil
}

// csi.NodeUnpublishVolumeRequest:	volume id	+ Required
//									target path	+ Required
func (s *service) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	// set parameter
	volumeID := req.GetVolumeId()
	targetPath := req.GetTargetPath()
	// Check volume exist
	volInfo, err := s.storageProvider.FindVolume(volumeID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if volInfo == nil {
		return nil, status.Errorf(codes.NotFound, "Volume %s does not exist", volumeID)
	}
	// do unmount
	if err = mount.CleanupMountPoint(targetPath, s.mounter.Interface, true); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (s *service) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: s.option.NodeCap,
	}, nil
}

func (s *service) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	return &csi.NodeGetInfoResponse{
		NodeId:            s.option.NodeId,
		MaxVolumesPerNode: s.option.MaxVolume,
	}, nil
}

// NodeExpandVolume will expand filesystem of volume.
// Input Parameters:
//  volume id: REQUIRED
//  volume path: REQUIRED
func (s *service) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	requestSizeBytes, err := GetRequiredVolumeSizeByte(req.GetCapacityRange())
	if err != nil {
		return nil, status.Error(codes.OutOfRange, err.Error())
	}
	volumeID, volumePath := req.GetVolumeId(), req.GetVolumePath()
	devicePath, err := s.storageProvider.NodeGetDevice(volumeID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot find device path of volume %s, error:%s", volumeID, err.Error())
	}
	resizeFs := resizefs.NewResizeFs(s.mounter)
	ok, err := resizeFs.Resize(devicePath, volumePath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !ok {
		return nil, status.Error(codes.Internal, "failed to expand volume filesystem")
	}
	return &csi.NodeExpandVolumeResponse{CapacityBytes: requestSizeBytes}, nil
}

// NodeGetVolumeStats
// Input Arguments:
//  volume id: REQUIRED
//  volume path: REQUIRED
func (s *service) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	volumeId, volumePath := req.GetVolumeId(), req.GetVolumePath()
	// Checkout device
	devicePath, _, err := mount.GetDeviceNameFromMount(s.mounter, volumePath)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "cannot get device name from mount point %s", volumePath)
	}
	// get device
	volumeDevicePath, err := s.storageProvider.NodeGetDevice(volumeId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "volume device not found %s,", volumeId)
	}
	if devicePath == "" || volumeDevicePath != devicePath {
		return nil, status.Errorf(codes.NotFound, "device path mismatch, from mount point %s, "+
			"from storage provider %s", devicePath, volumeDevicePath)
	}
	// Get metrics
	metricsStatFs := k8sVolume.NewMetricsStatFS(volumePath)
	metrics, err := metricsStatFs.GetMetrics()
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	return &csi.NodeGetVolumeStatsResponse{
		Usage: []*csi.VolumeUsage{
			{
				Available: metrics.Available.Value(),
				Total:     metrics.Capacity.Value(),
				Used:      metrics.Used.Value(),
				Unit:      csi.VolumeUsage_BYTES,
			},
			{
				Available: metrics.InodesFree.Value(),
				Total:     metrics.Inodes.Value(),
				Used:      metrics.InodesUsed.Value(),
				Unit:      csi.VolumeUsage_INODES,
			},
		},
	}, nil
}
