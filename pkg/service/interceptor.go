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
		// Check sanity of request Name, Volume Capabilities
		if len(req.GetName()) == 0 {
			return status.Error(codes.InvalidArgument, "volume name missing in request")
		}
		// Required volume capability
		if req.GetVolumeCapabilities() == nil {
			return status.Error(codes.InvalidArgument, "volume capabilities missing in request")
		} else if !s.option.ValidateVolumeCapabilities(req.GetVolumeCapabilities()) {
			return status.Error(codes.InvalidArgument, "volume capabilities not match")
		}
	case *csi.DeleteVolumeRequest:
		// Check sanity of request Name, Volume Capabilities
		if len(req.GetVolumeId()) == 0 {
			return status.Error(codes.InvalidArgument, "Volume id missing in request")
		}
	case *csi.CreateSnapshotRequest:
		// Check source volume id
		if len(req.GetSourceVolumeId()) == 0 {
			return status.Error(codes.InvalidArgument, "volume ID missing in request")
		}
		// Check snapshot name
		if len(req.GetName()) == 0 {
			return status.Error(codes.InvalidArgument, "snapshot name missing in request")
		}
	case *csi.DeleteSnapshotRequest:
		if len(req.GetSnapshotId()) == 0 {
			return status.Error(codes.InvalidArgument, "snapshot ID missing in request")
		}
	case *csi.ControllerExpandVolumeRequest:
		if len(req.GetVolumeId()) == 0 {
			return status.Error(codes.InvalidArgument, "No volume id is provided")
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
		if len(req.GetVolumeId()) == 0 {
			return status.Error(codes.InvalidArgument, "Volume ID missing in request")
		}
		if len(req.GetStagingTargetPath()) == 0 {
			return status.Error(codes.InvalidArgument, "Target path missing in request")
		}
		if req.GetVolumeCapability() == nil {
			return status.Error(codes.InvalidArgument, "Volume capability missing in request")
		}
	case *csi.NodeExpandVolumeRequest:
		if len(req.GetVolumeId()) == 0 {
			return status.Error(codes.InvalidArgument, "Volume ID missing in request")
		}
		if len(req.GetVolumePath()) == 0 {
			return status.Error(codes.InvalidArgument, "Volume path missing in request")
		}
	case *csi.NodeUnstageVolumeRequest:
		if len(req.GetVolumeId()) == 0 {
			return status.Error(codes.InvalidArgument, "Volume ID missing in request")
		}
		if len(req.GetStagingTargetPath()) == 0 {
			return status.Error(codes.InvalidArgument, "Target path missing in request")
		}
	case *csi.NodeGetVolumeStatsRequest:
		if len(req.GetVolumeId()) == 0 {
			return status.Error(codes.InvalidArgument, "Volume ID missing in request")
		}
		if len(req.GetVolumePath()) == 0 {
			return status.Error(codes.InvalidArgument, "Volume path missing in request")
		}
	}

	return nil
}
