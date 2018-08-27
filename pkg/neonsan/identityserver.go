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
	"github.com/kubernetes-csi/drivers/pkg/csi-common"
)

type identityServer struct {
	*csicommon.DefaultIdentityServer
}

// Disable Probe implementation due to the timeout setting of external attacher is too short.
//func (ids *identityServer) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {
//	glog.Info("*************** Start Probe ***************")
//	defer glog.Info("=============== End Probe ===============")
//
//	// check dependencies
//	glog.Info("Verify dependencies")
//	if err := ProbeNeonsanCommand(); err != nil {
//		glog.Error("Missing required dependency [neonsan]")
//		return nil, status.Error(codes.FailedPrecondition, err.Error())
//	}
//	glog.Info("Succeed to execute Neonsan command.")
//
//	// verify configuration
//	glog.Info("Verify configuration")
//	if _, err := os.Stat(ConfigFilePath); err != nil {
//		glog.Error("Stat config file path [%s] error [%v].", ConfigFilePath, err.Error())
//		if os.IsNotExist(err) {
//			return nil, status.Error(codes.FailedPrecondition, err.Error())
//		} else {
//			return nil, status.Error(codes.Internal, err.Error())
//		}
//	}
//	glog.Infof("Config file [%s] exists.", ConfigFilePath)
//
//	if _, err := GetPoolNameList(); err != nil {
//		glog.Error("Failed to list pool name.")
//		return nil, status.Error(codes.FailedPrecondition, err.Error())
//	}
//	glog.Info("Succeed to list pool name.")
//
//	return &csi.ProbeResponse{Ready: &wrappers.BoolValue{Value: true}}, nil
//}
