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
