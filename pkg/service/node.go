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
	"k8s.io/kubernetes/pkg/util/mount"
	"k8s.io/kubernetes/pkg/volume"
	"os"
)

// This operation MUST be idempotent
// csi.NodeStageVolumeRequest: 	volume id			+ Required
//								stage target path	+ Required
//								volume capability	+ Required
func (s *service) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	// set parameter
	volumeId := req.GetVolumeId()
	targetPath := req.GetStagingTargetPath()
	// ensure one call in-flight
	if acquired := s.locks.TryAcquire(volumeId); !acquired {
		return nil, status.Errorf(codes.Aborted, common.OperationPendingFmt, volumeId)
	}
	defer s.locks.Release(volumeId)
	// set fsType
	sc, err := NewStorageClass(req.GetPublishContext())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Check volume exist
	volInfo, err := s.storageProvider.FindVolume(volumeId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if volInfo == nil {
		return nil, status.Errorf(codes.NotFound, "Volume %s does not exist", volumeId)
	}
	// Attach if need
	err = s.storageProvider.NodeAttachVolume(volumeId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// 1. Mount
	// if volume already mounted
	notMnt, err := s.mounter.Interface.IsLikelyNotMountPoint(targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// already mount
	if !notMnt {
		return &csi.NodeStageVolumeResponse{}, nil
	}

	devicePath, err := s.storageProvider.NodeGetDevice(volumeId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// do mount
	klog.Infof("Mounting %s to %s format ...", volumeId, targetPath)
	if err := s.mounter.FormatAndMount(devicePath, targetPath, sc.FsType, []string{}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	klog.Infof("Mount %s to %s succeed", volumeId, targetPath)
	return &csi.NodeStageVolumeResponse{}, nil
}

// This operation MUST be idempotent
// csi.NodeUnstageVolumeRequest:	volume id	+ Required
//									target path	+ Required
func (s *service) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.
	NodeUnstageVolumeResponse, error) {

	// set parameter
	volumeId := req.GetVolumeId()
	targetPath := req.GetStagingTargetPath()
	// ensure one call in-flight
	klog.Infof("Try to lock resource %s", volumeId)
	if acquired := s.locks.TryAcquire(volumeId); !acquired {
		return nil, status.Errorf(codes.Aborted, common.OperationPendingFmt, volumeId)
	}
	defer s.locks.Release(volumeId)
	// Check volume exist
	volInfo, err := s.storageProvider.FindVolume(volumeId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if volInfo == nil {
		return nil, status.Errorf(codes.NotFound, "Volume %s does not exist", volumeId)
	}

	// 1. Unmount
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
	klog.Infof("Disk volume %s has been unmounted.", volumeId)
	cnt--
	klog.Infof("Disk volume mount count: %d", cnt)
	if cnt > 0 {
		klog.Errorf("Volume %s still mounted in instance %s", volumeId, s.option.NodeId)
		return nil, status.Error(codes.Internal, "unmount failed")
	}

	// node detach volume
	err = s.storageProvider.NodeDetachVolume(volumeId)
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
func (s *service) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.
	NodePublishVolumeResponse, error) {
	// set parameter
	targetPath := req.GetTargetPath()
	stagePath := req.GetStagingTargetPath()
	volumeId := req.GetVolumeId()

	// ensure one call in-flight
	if acquired := s.locks.TryAcquire(volumeId); !acquired {
		return nil, status.Errorf(codes.Aborted, common.OperationPendingFmt, volumeId)
	}
	defer s.locks.Release(volumeId)

	// set fsType
	sc, err := NewStorageClass(req.GetVolumeContext())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Check volume exist
	volInfo, err := s.storageProvider.FindVolume(volumeId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if volInfo == nil {
		return nil, status.Errorf(codes.NotFound, "Volume %s does not exist", volumeId)
	}

	// 1. Mount
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
	klog.Infof("Bind mount %s at %s, fsType %s, options %v ...", stagePath, targetPath, sc.FsType, options)
	if err := mounter.Mount(stagePath, targetPath, sc.FsType, options); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	klog.Infof("Mount bind %s at %s succeed", stagePath, targetPath)
	return &csi.NodePublishVolumeResponse{}, nil
}

// csi.NodeUnpublishVolumeRequest:	volume id	+ Required
//									target path	+ Required
func (s *service) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.
	NodeUnpublishVolumeResponse, error) {
	// set parameter
	volumeId := req.GetVolumeId()
	targetPath := req.GetTargetPath()
	// ensure one call in-flight
	if acquired := s.locks.TryAcquire(volumeId); !acquired {
		return nil, status.Errorf(codes.Aborted, common.OperationPendingFmt, volumeId)
	}
	defer s.locks.Release(volumeId)
	// Check volume exist
	volInfo, err := s.storageProvider.FindVolume(volumeId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if volInfo == nil {
		return nil, status.Errorf(codes.NotFound, "Volume %s does not exist", volumeId)
	}

	// do unmount
	klog.Infof("Unbind mount volume %s/%s", targetPath, volumeId)
	if err = mount.CleanupMountPoint(targetPath, s.mounter.Interface, true); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	klog.Infof("Unbound mount volume succeed")
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (s *service) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.
	NodeGetCapabilitiesResponse, error) {
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
	return nil, status.Errorf(codes.Unimplemented, "expand volume not implement")
}

// NodeGetVolumeStats
// Input Arguments:
//  volume id: REQUIRED
//  volume path: REQUIRED
func (s *service) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	hash := common.GetContextHash(ctx)
	volumeId := req.GetVolumeId()
	volumePath := req.GetVolumePath()

	// Checkout device
	klog.Infof("%s: Get device name from mount point %s", hash, volumePath)
	volume.NewMetricsDu(volumePath)
	devicePath, _, err := mount.GetDeviceNameFromMount(s.mounter, volumePath)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "cannot get device name from mount point %s", volumePath)
	}
	klog.Infof("%s: Succeed to get device name %s", hash, devicePath)

	volumeDevicePath, err := s.storageProvider.NodeGetDevice(volumeId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "volume device not found %s,", volumeId)
	}

	if devicePath == "" || volumeDevicePath != devicePath {
		return nil, status.Errorf(codes.NotFound, "device path mismatch, from mount point %s, "+
			"from storage provider %s", devicePath, volumeDevicePath)
	}

	// Get metrics
	metricsStatFs := volume.NewMetricsStatFS(volumePath)
	metrics, err := metricsStatFs.GetMetrics()
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	klog.Infof("%s: Succeed to get metrics", hash)
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
