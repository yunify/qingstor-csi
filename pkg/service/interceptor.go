package service

import (
	"context"
	"errors"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"github.com/yunify/qingstor-csi/pkg/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
	"time"
)

var (
	errorRequestNil = errors.New("request is nil")
)

func (s *service) Interceptor() grpc.UnaryServerInterceptor {
	return s.interceptor
}

func (s *service) interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	hash := common.GenerateHashInEightBytes(time.Now().UTC().String())
	klog.Infof("GRPC %s request: %s %s", info.FullMethod, protosanitizer.StripSecrets(req), hash)
	err := s.validateRequest(req)
	if err != nil {
		klog.Errorf("GRPC %s fail validate: %v req:%s %s", info.FullMethod, err, protosanitizer.StripSecrets(req), hash)
		return nil, err
	}
	resp, err := handler(common.ContextWithHash(ctx, hash), req)
	if err != nil {
		klog.Errorf("GRPC %s error: %v req:%s %s", info.FullMethod, err, protosanitizer.StripSecrets(req), hash)
	} else if resp != nil {
		klog.Infof("GRPC %s response: %s  req:%s %s", info.FullMethod, protosanitizer.StripSecrets(resp), protosanitizer.StripSecrets(req), hash)
	}
	return resp, err
}

func (s *service) validateRequest(request interface{}) error {
	if request == nil {
		return errorRequestNil
	}
	switch req := request.(type) {
	case *csi.CreateVolumeRequest:
		if isValid := s.option.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); isValid != true {
			return status.Error(codes.Unimplemented, "unsupported controller server capability")
		}
		// Required volume capability
		if req.GetVolumeCapabilities() == nil {
			return status.Error(codes.InvalidArgument, "volume capabilities missing in request")
		} else if !s.option.ValidateVolumeCapabilities(req.GetVolumeCapabilities()) {
			return status.Error(codes.InvalidArgument, "volume capabilities not match")
		}
		// Check sanity of request Name, Volume Capabilities
		if len(req.GetName()) == 0 {
			return status.Error(codes.InvalidArgument, "volume name missing in request")
		}
	case *csi.DeleteVolumeRequest:
		if isValid := s.option.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); isValid != true {
			klog.Errorf("invalid delete volume req: %v", req)
			return status.Error(codes.Unimplemented, "")
		}
		// Check sanity of request Name, Volume Capabilities
		if len(req.GetVolumeId()) == 0 {
			return status.Error(codes.InvalidArgument, "Volume id missing in request")
		}
	case *csi.ControllerPublishVolumeRequest:
		if isValid := s.option.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); isValid != true {
			return status.Error(codes.Unimplemented, "")
		}
		// check volume id arguments
		if len(req.GetVolumeId()) == 0 {
			return status.Error(codes.InvalidArgument, "Volume ID missing in request")
		}
		// check nodeId arguments
		if len(req.GetNodeId()) == 0 {
			return status.Error(codes.InvalidArgument, "Node ID missing in request")
		}
		// check volume capability
		if req.GetVolumeCapability() == nil {
			return status.Error(codes.InvalidArgument, "No volume capability is provided ")
		}

	case *csi.ControllerUnpublishVolumeRequest:
		if isValid := s.option.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME); isValid != true {
			klog.Errorf("Invalid unpublish volume req: %v", req)
			return status.Error(codes.Unimplemented, "")
		}
		if len(req.GetVolumeId()) == 0 {
			return status.Error(codes.InvalidArgument, "Volume ID missing in request")
		}
	case *csi.ValidateVolumeCapabilitiesRequest:
		// require volume id parameter
		if len(req.GetVolumeId()) == 0 {
			return status.Error(codes.InvalidArgument, "No volume id is provided")
		}
		// require capability parameter
		if len(req.GetVolumeCapabilities()) == 0 {
			return status.Error(codes.InvalidArgument, "No volume capabilities are provided")
		}

	case *csi.ControllerExpandVolumeRequest:
		// require volume id parameter
		if len(req.GetVolumeId()) == 0 {
			return status.Error(codes.InvalidArgument, "No volume id is provided")
		}
	case *csi.CreateSnapshotRequest:
		// 0. Prepare
		if isValid := s.option.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT); isValid != true {
			klog.Errorf("Invalid create snapshot request: %v", req)
			return status.Error(codes.Unimplemented, "")
		}
		// Check source volume id
		if len(req.GetSourceVolumeId()) == 0 {
			return status.Error(codes.InvalidArgument, "volume ID missing in request")
		}
		// Check snapshot name
		if len(req.GetName()) == 0 {
			return status.Error(codes.InvalidArgument, "snapshot name missing in request")
		}

	case *csi.DeleteSnapshotRequest:
		if isValid := s.option.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT); isValid != true {
			klog.Errorf("Invalid create snapshot request: %v", req)
			return status.Error(codes.Unimplemented, "")
		}
		// Check snapshot id
		klog.Info("Check required parameters")
		if len(req.GetSnapshotId()) == 0 {
			return status.Error(codes.InvalidArgument, "snapshot ID missing in request")
		}

	case *csi.NodePublishVolumeRequest:
		// check volume id
		if len(req.GetVolumeId()) == 0 {
			return status.Error(codes.InvalidArgument, "Volume id missing in request")
		}
		// check target path
		if len(req.GetStagingTargetPath()) == 0 {
			return status.Error(codes.InvalidArgument, "Target path missing in request")
		}
		// Check volume capability
		if req.GetVolumeCapability() == nil {
			return status.Error(codes.InvalidArgument, "Volume capabilities missing in request")
		} else if !s.option.ValidateVolumeCapability(req.GetVolumeCapability()) {
			return status.Error(codes.FailedPrecondition, "Exceed capabilities")
		}
		// check stage path
		if len(req.GetStagingTargetPath()) == 0 {
			return status.Error(codes.FailedPrecondition, "Staging target path not set")
		}

	case *csi.NodeUnpublishVolumeRequest:
		// check arguments
		if len(req.GetTargetPath()) == 0 {
			return status.Error(codes.InvalidArgument, "Target path missing in request")
		}
		if len(req.GetVolumeId()) == 0 {
			return status.Error(codes.InvalidArgument, "Volume id missing in request")
		}
	case *csi.NodeStageVolumeRequest:
		if flag := s.option.ValidateNodeServiceRequest(csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME); flag == false {
			return status.Error(codes.Unimplemented, "Node has not stage capability")
		}
		if len(req.GetVolumeId()) == 0 {
			return status.Error(codes.InvalidArgument, "Volume ID missing in request")
		}
		if len(req.GetStagingTargetPath()) == 0 {
			return status.Error(codes.InvalidArgument, "Target path missing in request")
		}
		if req.GetVolumeCapability() == nil {
			return status.Error(codes.InvalidArgument, "Volume capability missing in request")
		}

	case *csi.NodeUnstageVolumeRequest:
		if flag := s.option.ValidateNodeServiceRequest(csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME); flag == false {
			return status.Error(codes.Unimplemented, "Node has not stage capability")
		}
		if len(req.GetVolumeId()) == 0 {
			return status.Error(codes.InvalidArgument, "Volume ID missing in request")
		}
		if len(req.GetStagingTargetPath()) == 0 {
			return status.Error(codes.InvalidArgument, "Target path missing in request")
		}
	case *csi.NodeExpandVolumeRequest:
		if flag := s.option.ValidateNodeServiceRequest(csi.NodeServiceCapability_RPC_EXPAND_VOLUME); flag == false {
			return status.Error(codes.Unimplemented, "Node has not stage capability")
		}
		if len(req.GetVolumeId()) == 0 {
			return status.Error(codes.InvalidArgument, "Volume ID missing in request")
		}
		if len(req.GetVolumePath()) == 0 {
			return status.Error(codes.InvalidArgument, "Volume path missing in request")
		}

	case *csi.NodeGetVolumeStatsRequest:
		if flag := s.option.ValidateNodeServiceRequest(csi.NodeServiceCapability_RPC_GET_VOLUME_STATS); flag == false {
			return status.Error(codes.Unimplemented, "Node has not stage capability")
		}
		if len(req.GetVolumeId()) == 0 {
			return status.Error(codes.InvalidArgument, "Volume ID missing in request")
		}
		if len(req.GetVolumePath()) == 0 {
			return status.Error(codes.InvalidArgument, "Volume path missing in request")
		}
	}

	return nil
}
