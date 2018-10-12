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
	"github.com/yunify/qingstor-csi/pkg/neonsan/cache"
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
	var snapCache cache.SnapshotCacheType
	snapCache.New()
	snapCache.Sync()
	return &controllerServer{
		DefaultControllerServer: csicommon.NewDefaultControllerServer(d),
		cache:                   &snapCache,
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
		csi.ControllerServiceCapability_RPC_GET_CAPACITY,
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
		csi.ControllerServiceCapability_RPC_LIST_SNAPSHOTS,
	})
	neons.driver.AddVolumeCapabilityAccessModes([]csi.VolumeCapability_AccessMode_Mode{
		csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER})
	// Create GRPC servers
	neons.ids = NewIdentityServer(neons.driver)
	neons.ns = NewNodeServer(neons.driver)
	neons.cs = NewControllerServer(neons.driver)

	s := csicommon.NewNonBlockingGRPCServer()

	// Initialize snapshot cache

	s.Start(endpoint, neons.ids, neons.cs, neons.ns)
	s.Wait()
}
