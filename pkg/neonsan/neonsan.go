package neonsan

import (
	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/golang/glog"
	"github.com/kubernetes-csi/drivers/pkg/csi-common"
)

const version = "0.3.0"

type neonsan struct {
	driver *csicommon.CSIDriver

	ids *identityServer
	ns  *nodeServer
	cs  *controllerServer

	cap   []*csi.VolumeCapability_AccessMode
	cscap []*csi.ControllerServiceCapability
}

func GetNeonsanDriver() *neonsan {
	return &neonsan{}
}

// NewIdentityServer
// Create identity server
func NewIdentityServer(d *csicommon.CSIDriver) *identityServer {
	return &identityServer{
		DefaultIdentityServer: csicommon.NewDefaultIdentityServer(d),
	}
}

// NewControllerServer
// Create controller server
func NewControllerServer(d *csicommon.CSIDriver) *controllerServer {
	return &controllerServer{
		DefaultControllerServer: csicommon.NewDefaultControllerServer(d),
	}
}

// NewNodeServer
// Create node server
func NewNodeServer(d *csicommon.CSIDriver) *nodeServer {
	return &nodeServer{
		DefaultNodeServer: csicommon.NewDefaultNodeServer(d),
	}
}

// Run: Initial and start CSI driver
func (neons *neonsan) Run(driverName, nodeId, endpoint string) {
	glog.Infof("Driver: %v version: %v", driverName, version)

	// Initialize default library driver
	neons.driver = csicommon.NewCSIDriver(driverName, version, nodeId)
	if neons.driver == nil {
		glog.Fatalln("Failed to initialize CSI Driver.")
	}
	neons.driver.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_GET_CAPACITY,})
	neons.driver.AddVolumeCapabilityAccessModes([]csi.VolumeCapability_AccessMode_Mode{
		csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER})
	// Create GRPC servers
	neons.ids = NewIdentityServer(neons.driver)
	neons.ns = NewNodeServer(neons.driver)
	neons.cs = NewControllerServer(neons.driver)

	s := csicommon.NewNonBlockingGRPCServer()
	s.Start(endpoint, neons.ids, neons.cs, neons.ns)
	s.Wait()
}
