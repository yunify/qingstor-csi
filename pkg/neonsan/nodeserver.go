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
	"github.com/yunify/qingstor-csi/pkg/neonsan/manager"
	"github.com/yunify/qingstor-csi/pkg/neonsan/util"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/kubernetes/pkg/util/mount"
	"os"
	"time"
)

type nodeServer struct {
	*csicommon.DefaultNodeServer
}

// This operation MUST be idempotent
// If the volume corresponding to the volume id has already been published at the specified target path,
// and is compatible with the specified volume capability and readonly flag, the plugin MUST reply 0 OK.
// csi.NodePublishVolumeRequest:	volume id			+ Required
//									target path			+ Required
//									volume capability	+ Required
//									read only			+ Required (This field is NOT provided when requesting in Kubernetes)
func (ns *nodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	defer util.EntryFunction("NodePublishVolume")()

	glog.Info("Validate input arguments.")
	// 0. Preflight
	// check volume id
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request.")
	}
	// check target path
	if len(req.GetStagingTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request.")
	}
	// Check volume capability
	if req.GetVolumeCapability() == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capabilities missing in request.")
	} else if !util.ContainsVolumeCapability(ns.Driver.GetVolumeCapabilityAccessModes(), req.GetVolumeCapability()) {
		return nil, status.Error(codes.FailedPrecondition, "Exceed capabilities.")
	}
	// check stage path
	if len(req.GetStagingTargetPath()) == 0 {
		return nil, status.Error(codes.FailedPrecondition, "Staging target path not set.")
	}
	// set parameter
	targetPath := req.GetTargetPath()
	stagePath := req.GetStagingTargetPath()
	volumeId := req.GetVolumeId()

	// set fsType
	glog.Info("Create StorageClass.")
	sc, err := manager.NewNeonsanStorageClassFromMap(req.GetVolumeAttributes())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	fsType := sc.VolumeFsType

	// Check volume exist
	glog.Infof("Find volume [%s] in pool [%s].", volumeId, sc.Pool)
	pool := sc.Pool
	volInfo, err := manager.FindVolume(volumeId, pool)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if volInfo == nil {
		glog.Errorf("Cannot find volume [%s].", volumeId)
		return nil, status.Errorf(codes.NotFound, "Volume [%s] does not exist.", volumeId)
	}
	glog.Infof("Found volume [%s].", volumeId)

	// 1. Mount
	// Make dir if dir not presents
	glog.Infof("Find target path [%s].", targetPath)
	_, err = os.Stat(targetPath)
	if os.IsNotExist(err) {
		glog.Infof("Create target path [%s].", targetPath)
		if err = os.MkdirAll(targetPath, 0750); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	glog.Infof("Succeed to find target path [%s].", targetPath)

	// check targetPath is mounted
	glog.Infof("Check target path [%s] mounted status.", targetPath)
	mounter := mount.New("")
	notMnt, err := mounter.IsNotMountPoint(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(targetPath, 0750); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			notMnt = true
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	// For idempotent:
	// If the volume corresponding to the volume id has already been published at the specified target path,
	// and is compatible with the specified volume capability and readonly flag, the plugin MUST reply 0 OK.
	glog.Infof("Is target path [%s] mounted: [%t].", targetPath, !notMnt)
	if !notMnt {
		glog.Warningf("Volume [%s] has been mounted at [%s].", volumeId, targetPath)
		return &csi.NodePublishVolumeResponse{}, nil
	}

	// set bind mount options
	options := []string{"bind"}
	if req.GetReadonly() == true {
		options = append(options, "ro")
	}
	glog.Infof("Bind mount [%s] at [%s], fsType [%s], options [%v] ...", stagePath, targetPath, fsType, options)
	if err := mounter.Mount(stagePath, targetPath, fsType, options); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	glog.Infof("Mount bind [%s] at [%s] succeed.", stagePath, targetPath)
	return &csi.NodePublishVolumeResponse{}, nil
}

// csi.NodeUnpublishVolumeRequest:	volume id	+ Required
//									target path	+ Required
func (ns *nodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	defer util.EntryFunction("NodeUnpublishVolume")()

	// 0. Preflight
	glog.Info("Validate input arguments.")
	// check arguments
	if len(req.GetTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request.")
	}
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request.")
	}
	// set parameter
	volumeId := req.GetVolumeId()
	targetPath := req.GetTargetPath()

	// Check volume exist
	glog.Infof("Find volume [%s].", volumeId)
	volInfo, err := manager.FindVolumeWithoutPool(volumeId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	glog.Infof("Found volume [%s] info [%v].", volumeId, volInfo)
	if volInfo == nil {
		return nil, status.Errorf(codes.NotFound, "Volume [%s] does not exist.", volumeId)
	}

	// 1. Unmount
	// check targetPath is mounted
	glog.Infof("Check target path [%s] mounted status.", targetPath)
	mounter := mount.New("")
	notMnt, err := mounter.IsNotMountPoint(targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	glog.Infof("Is target path [%s] mounted: [%t].", targetPath, !notMnt)
	if notMnt {
		glog.Warningf("Volume [%s] does not has not mount point.", volumeId)
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}
	// do unmount
	glog.Infof("Unbind mounted volume [%s]/[%s].", targetPath, volumeId)
	if err = mounter.Unmount(targetPath); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	glog.Infof("Unbind mounted volume succeed.")

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// This operation MUST be idempotent
// csi.NodeStageVolumeRequest: 	volume id			+ Required
//								stage target path	+ Required
//								volume capability	+ Required
func (ns *nodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	defer util.EntryFunction("NodeStageVolume")()

	capRsp, _ := ns.NodeGetCapabilities(context.Background(), nil)
	if flag := util.ContainsNodeServiceCapability(capRsp.GetCapabilities(),
		csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME); flag == false {
		glog.Errorf("driver capability %v", capRsp.GetCapabilities())
		return nil, status.Error(codes.Unimplemented, "Node has not unstage capability.")
	}

	// Check arguments
	glog.Info("Check input arguments.")
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request.")
	}
	if len(req.GetStagingTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request.")
	}
	if req.GetVolumeCapability() == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability missing in request.")
	}
	// Set parameter
	volumeId := req.GetVolumeId()
	targetPath := req.GetStagingTargetPath()

	// create sc object
	glog.Info("Get storage class from map.")
	sc, err := manager.NewNeonsanStorageClassFromMap(req.VolumeAttributes)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Check volume exist
	glog.Infof("Find volume [%s] in pool [%s].", volumeId, sc.Pool)
	volInfo, err := manager.FindVolume(volumeId, sc.Pool)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if volInfo == nil {
		glog.Errorf("Not found volume [%s] in pool [%s].", volumeId, sc.Pool)
		return nil, status.Errorf(codes.NotFound, "Volume [%s] does not exist", volumeId)
	}
	glog.Infof("Found volume [%s] in pool [%s] info [%v]", volumeId, sc.Pool, volInfo)

	// Attach volume
	// map volume
	glog.Infof("Map volume [%s] in pool [%s].", volumeId, sc.Pool)
	err = manager.AttachVolume(volumeId, sc.Pool)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	glog.Infof("Succeed to map volume [%s].", volumeId)

	// find volume device path
	glog.Infof("Find volume [%s] mapping info.", volumeId)
	var attInfo *manager.AttachInfo = nil
	for i := 1; i < 6; i++ {
		attInfo, err = manager.FindAttachedVolumeWithoutPool(volumeId)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if attInfo != nil {
			break
		}
		glog.Warningf("Cannot find attached volume [%s], [%d] remaining retries...", volumeId, i)
		time.Sleep(time.Duration(i) * time.Second)
	}

	glog.Infof("Found volume [%s] attached info [%v]", volumeId, attInfo)
	if attInfo == nil {
		return nil, status.Errorf(codes.Internal, "Cannot find attached volume [%s].", volumeId)
	} else if attInfo.Pool != sc.Pool {
		return nil, status.Errorf(codes.Internal, "Volume [%s] pool mismatch: expect pool [%s], "+
			"but actually [%s].", volumeId, sc.Pool, attInfo.Pool)
	}

	// if volume already mounted
	glog.Infof("Check targetPath [%s] mount info.", targetPath)
	notMnt, err := mount.New("").IsLikelyNotMountPoint(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(targetPath, 0750); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			notMnt = true
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	// already mount
	glog.Infof("Is target path [%s] mounted: [%t].", targetPath, !notMnt)
	if !notMnt {
		glog.Warningf("Target path [%s] has been mounted.", targetPath)
		return &csi.NodeStageVolumeResponse{}, nil
	}
	glog.Infof("Target path [%s] has not been mounted yet.", targetPath)

	// do mount
	devicePath := attInfo.Device
	fsType := sc.VolumeFsType
	glog.Infof("Mounting [%s] to [%s] with format [%s]...", volumeId, targetPath, fsType)
	diskMounter := &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: mount.NewOsExec()}
	if err := diskMounter.FormatAndMount(devicePath, targetPath, fsType, []string{}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	glog.Infof("Mount [%s] to [%s] succeed.", volumeId, targetPath)

	return &csi.NodeStageVolumeResponse{}, nil
}

// This operation MUST be idempotent
// csi.NodeUnstageVolumeRequest:	volume id	+ Required
//									target path	+ Required
func (ns *nodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	defer util.EntryFunction("NodeUnstageVolume")()

	capRsp, _ := ns.NodeGetCapabilities(context.Background(), nil)
	if flag := util.ContainsNodeServiceCapability(capRsp.GetCapabilities(),
		csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME); flag == false {
		glog.Errorf("Driver capability %v", capRsp.GetCapabilities())
		return nil, status.Error(codes.Unimplemented, "Node has not unstage capability")
	}

	// check argument
	glog.Info("Check input arguments.")
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if len(req.GetStagingTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}

	// set parameter
	volumeId := req.GetVolumeId()
	targetPath := req.GetStagingTargetPath()

	// check volume exist
	glog.Info("Get volume info.")
	volInfo, err := manager.FindVolumeWithoutPool(volumeId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if volInfo == nil {
		return nil, status.Errorf(codes.NotFound, "Volume [%s] does not exist.", volumeId)
	}

	// Unmount
	// get target path mount status
	glog.Infof("Check target path [%s] mount info.", targetPath)
	mounter := mount.New("")
	notMnt, err := mounter.IsLikelyNotMountPoint(targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	glog.Infof("Is target path [%s] mounted: [%t].", targetPath, !notMnt)
	if !notMnt {
		glog.Infof("Target path [%s] has been mounted.", targetPath)

		// count mount point
		_, cnt, err := mount.GetDeviceNameFromMount(mounter, targetPath)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		// do unmount
		glog.Infof("Un-mount target path [%s].", targetPath)
		err = mounter.Unmount(targetPath)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		glog.Infof("Volume [%s] has been unmounted.", volumeId)
		cnt--
		glog.Infof("Mount count: [%d].", cnt)
		if cnt > 0 {
			glog.Errorf("Volume [%s] still mounted current node.", volumeId)
			return nil, status.Error(codes.Internal, "Un-mount failed")
		}
	}
	glog.Warningf("Target path [%s] has not been mounted yet.", targetPath)

	// Unmap
	// check map status
	glog.Infof("Get attached volume [%s] info.", volumeId)
	attInfo, err := manager.FindAttachedVolumeWithoutPool(volumeId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if attInfo == nil {
		glog.Warningf("Volume [%s] has been unmapped.", volumeId)
		return &csi.NodeUnstageVolumeResponse{}, nil
	} else if volInfo.Pool != attInfo.Pool {
		glog.Errorf("Find duplicate volume [%s] in pool [%s] and [%s].", volumeId, volInfo.Pool, attInfo.Pool)
		return nil, status.Errorf(codes.Internal, "Find duplicate volume [%s] in pool [%s] and [%s].", volumeId,
			volInfo.Pool, attInfo.Pool)
	}
	glog.Infof("Attached volume [%s] info [%v]", volumeId, attInfo)

	// do unmap
	pool := volInfo.Pool
	glog.Infof("Un-mapping volume [%s] in pool [%s]...", volumeId, pool)
	err = manager.DetachVolume(volumeId, pool)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	glog.Info("Un-map volume succeed.")

	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (ns *nodeServer) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	defer util.EntryFunction("NodeGetCapabilities")()
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
					},
				},
			},
		},
	}, nil
}
