package neonsan

import (
	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/golang/glog"
	"github.com/kubernetes-csi/drivers/pkg/csi-common"
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
	glog.Info("*************** Start NodePublishVolume ***************")
	defer glog.Info("=============== End NodePublishVolume ===============")
	// 0. Preflight
	// check volume id
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume id missing in request")
	}
	// check target path
	if len(req.GetStagingTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}
	// Check volume capability
	if req.GetVolumeCapability() == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capabilities missing in request")
	} else if !ContainsVolumeCapability(ns.Driver.GetVolumeCapabilityAccessModes(), req.GetVolumeCapability()) {
		return nil, status.Error(codes.FailedPrecondition, "Exceed capabilities")
	}
	// check stage path
	if len(req.GetStagingTargetPath()) == 0 {
		return nil, status.Error(codes.FailedPrecondition, "Staging target path not set")
	}
	// set parameter
	targetPath := req.GetTargetPath()
	stagePath := req.GetStagingTargetPath()
	volumeId := req.GetVolumeId()

	// set fsType
	sc, err := NewNeonsanStorageClassFromMap(req.GetVolumeAttributes())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	fsType := sc.VolumeFsType

	// Check volume exist
	pool := sc.Pool
	volInfo, err := FindVolume(volumeId, pool)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if volInfo == nil {
		return nil, status.Errorf(codes.NotFound, "Volume [%s] does not exist", volumeId)
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
	if !notMnt {
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
	glog.Infof("Mount bind [%s] at [%s] succeed", stagePath, targetPath)
	return &csi.NodePublishVolumeResponse{}, nil
}

// csi.NodeUnpublishVolumeRequest:	volume id	+ Required
//									target path	+ Required
func (ns *nodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	glog.Info("*************** Start NodeUnpublishVolume ***************")
	defer glog.Info("=============== End NodeUnpublishVolume ===============")
	// 0. Preflight
	// check arguments
	if len(req.GetTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume id missing in request")
	}
	// set parameter
	volumeId := req.GetVolumeId()
	targetPath := req.GetTargetPath()

	// Check volume exist
	volInfo, err := FindVolumeWithoutPool(volumeId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if volInfo == nil {
		return nil, status.Errorf(codes.NotFound, "Volume [%s] does not exist", volumeId)
	}

	// 1. Unmount
	// check targetPath is mounted
	mounter := mount.New("")
	notMnt, err := mounter.IsNotMountPoint(targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if notMnt {
		glog.Warningf("Volume [%s] has not mount point", volumeId)
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}
	// do unmount
	glog.Infof("Unbind mounted volume [%s]/[%s]", targetPath, volumeId)
	if err = mounter.Unmount(targetPath); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	glog.Infof("Unbound mounted volume succeed")

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// This operation MUST be idempotent
// csi.NodeStageVolumeRequest: 	volume id			+ Required
//								stage target path	+ Required
//								volume capability	+ Required
func (ns *nodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	glog.Info("*************** Start NodeStageVolume ***************")
	defer glog.Info("=============== End NodeStageVolume ===============")
	capRsp, _ := ns.NodeGetCapabilities(context.Background(), nil)
	if flag := ContainsNodeServiceCapability(capRsp.GetCapabilities(), csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME); flag == false {
		glog.Errorf("driver capability %v", capRsp.GetCapabilities())
		return nil, status.Error(codes.Unimplemented, "Node has not unstage capability")
	}
	// Check arguments
	glog.Info("Check input arguments")
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if len(req.GetStagingTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}
	if req.GetVolumeCapability() == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability missing in request")
	}
	// Set parameter
	volumeId := req.GetVolumeId()
	targetPath := req.GetStagingTargetPath()

	// create sc object
	glog.Info("Get storage class from map")
	sc, err := NewNeonsanStorageClassFromMap(req.VolumeAttributes)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Check volume exist
	glog.Infof("Find volume [%s] in pool [%s]", volumeId, sc.Pool)
	volInfo, err := FindVolume(volumeId, sc.Pool)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if volInfo == nil {
		return nil, status.Errorf(codes.NotFound, "Volume [%s] does not exist", volumeId)
	}

	// Attach volume
	// map volume
	glog.Infof("Map volume [%s] in pool [%s]", volumeId, sc.Pool)
	err = AttachVolume(volumeId, sc.Pool)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// find volume device path
	glog.Infof("Find volume [%s] map info", volumeId)
	var attInfo *attachInfo = nil
	for i := 1; i < 6; i++ {
		attInfo, err = FindAttachedVolumeWithoutPool(volumeId)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if attInfo == nil {
			glog.Warningf("Cannot find attached volume [%s], retry for [%d] times...", volumeId, i)
			time.Sleep(time.Duration(i) * time.Second)
		} else if attInfo.pool != sc.Pool {
			return nil, status.Errorf(codes.Internal, "Volume [%s] pool mismatch: expect pool [%s], but actually [%s]", volumeId, sc.Pool, attInfo.pool)
		}
	}
	if attInfo == nil {
		return nil, status.Errorf(codes.Internal, "Cannot find attached volume [%s]", volumeId)
	}

	// if volume already mounted
	glog.Infof("Check targetPath [%s] mount info", targetPath)
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
	if !notMnt {
		glog.Infof("Target path [%s] has been mounted", targetPath)
		return &csi.NodeStageVolumeResponse{}, nil
	}

	// do mount
	devicePath := attInfo.device
	fsType := sc.VolumeFsType
	glog.Infof("Mounting [%s] to [%s] with format [%s]...", volumeId, targetPath, fsType)
	diskMounter := &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: mount.NewOsExec()}
	if err := diskMounter.FormatAndMount(devicePath, targetPath, fsType, []string{}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	glog.Infof("Mount [%s] to [%s] succeed", volumeId, targetPath)

	return &csi.NodeStageVolumeResponse{}, nil
}

// This operation MUST be idempotent
// csi.NodeUnstageVolumeRequest:	volume id	+ Required
//									target path	+ Required
func (ns *nodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	glog.Info("*************** Start NodeUnstageVolume ***************")
	defer glog.Info("=============== End NodeUnstageVolume ===============")
	capRsp, _ := ns.NodeGetCapabilities(context.Background(), nil)
	if flag := ContainsNodeServiceCapability(capRsp.GetCapabilities(), csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME); flag == false {
		glog.Errorf("Driver capability %v", capRsp.GetCapabilities())
		return nil, status.Error(codes.Unimplemented, "Node has not unstage capability")
	}

	// check argument
	glog.Info("Check input arguments")
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
	glog.Info("Get volume info")
	volInfo, err := FindVolumeWithoutPool(volumeId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if volInfo == nil {
		return nil, status.Errorf(codes.NotFound, "Volume [%s] does not exist", volumeId)
	}

	// Unmount
	// get target path mount status
	mounter := mount.New("")
	notMnt, err := mounter.IsLikelyNotMountPoint(targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !notMnt {
		glog.Warningf("Not mount point at target path [%s]", targetPath)

		// count mount point
		_, cnt, err := mount.GetDeviceNameFromMount(mounter, targetPath)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		// do unmount
		glog.Infof("Unmount target path [%s]", targetPath)
		err = mounter.Unmount(targetPath)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		glog.Infof("Volume [%s] has been unmounted.", volumeId)
		cnt--
		glog.Infof("Mount count: [%d]", cnt)
		if cnt > 0 {
			glog.Errorf("Volume [%s] still mounted current node", volumeId)
			return nil, status.Error(codes.Internal, "Unmount failed")
		}
	}

	// Unmap
	// check map status
	glog.Infof("Get attached volume [%s] info.", volumeId)
	attInfo, err := FindAttachedVolumeWithoutPool(volumeId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if attInfo == nil {
		glog.Warningf("Volume [%s] has been unmap", volumeId)
		return &csi.NodeUnstageVolumeResponse{}, nil
	} else if volInfo.pool != attInfo.pool {
		return nil, status.Errorf(codes.Internal, "Find duplicate volume [%s] in pool [%s] and [%s]", volumeId, volInfo.pool, attInfo.pool)
	}

	// do unmap
	pool := volInfo.pool
	glog.Infof("Unmap volume [%s] in pool [%s]", volumeId, pool)
	err = DetachVolume(volumeId, pool)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (ns *nodeServer) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	glog.Info("*************** Start NodeGetCapabilities ***************")
	defer glog.Info("=============== End NodeGetCapabilities ===============")
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
