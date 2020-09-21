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

package sanity

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/yunify/qingstor-csi/pkg/service"
	"github.com/yunify/qingstor-csi/pkg/storage/neonsan"
	"google.golang.org/grpc"
	"k8s.io/klog"
	"os"
	"path/filepath"
	"testing"

	"github.com/kubernetes-csi/csi-test/v3/pkg/sanity"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	uds         = "/tmp/csi.sock"
	udsEndpoint = "unix://" + uds

	defaultConfigPath = "/etc/neonsan/qbd.conf"
	defaultProtocol   = "tcp"
)

var (
	qbdServer                    service.NonBlockingGRPCServer
	nodeCap = []csi.NodeServiceCapability_RPC_Type{
		csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
		csi.NodeServiceCapability_RPC_EXPAND_VOLUME,
		//csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
	}
	serviceOpt = service.NewOption().SetName("mock.neonsan.csi.com").SetVersion("1.1.0").
		SetNodeId("HelloNeonsan").SetMaxVolume(100).
		SetVolumeCapabilityAccessNodes(service.DefaultVolumeAccessModeType).
		SetControllerServiceCapabilities(service.DefaultControllerServiceCapability).
		SetNodeServiceCapabilities(nodeCap).
		SetPluginCapabilities(service.DefaultPluginCapability).
		SetRetryTime(service.DefaultBackOff)
)

var _ = BeforeSuite(func() {
	klog.InitFlags(nil)

	qbdServer = service.NewNonBlockingGRPCServer()
	qbdServer.Start(udsEndpoint, service.New(serviceOpt, neonsan.New(defaultConfigPath, defaultProtocol)))

})

var _ = AfterSuite(func() {

	if qbdServer != nil {
		qbdServer.Stop()
	}
})

func TestCSISanity(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CSI Sanity Test Suite")
}

var _ = Describe("QBD Neonsan CSI Driver -- mount", func() {
	config := &sanity.TestConfig{
		TargetPath:                filepath.Join(os.TempDir(), "/csi-target"),
		StagingPath:               filepath.Join(os.TempDir(), "/csi-staging"),
		TestVolumeSize:            1 << 30, // 10 GiB
		TestVolumeAccessType:      "mount",
		Address:                   udsEndpoint,
		TestNodeVolumeAttachLimit: true,
		TestVolumeParameters:      map[string]string{"pool_name": "testPool", "rep_count": "2", "fsType": "ext3"},
		IDGen:                     &sanity.DefaultIDGenerator{},
		DialOptions:               []grpc.DialOption{grpc.WithInsecure()},
		ControllerDialOptions:     []grpc.DialOption{grpc.WithInsecure()},
	}
	sanity.GinkgoTest(config)
})

var _ = Describe("QBD Neonsan CSI Driver -- block", func() {
	config := &sanity.TestConfig{
		TargetPath:                filepath.Join(os.TempDir(), "/csi-target"),
		StagingPath:               filepath.Join(os.TempDir(), "/csi-staging"),
		TestVolumeSize:            1 << 30, // 10 GiB
		TestVolumeAccessType:      "block",
		Address:                   udsEndpoint,
		TestNodeVolumeAttachLimit: true,
		TestVolumeParameters:      map[string]string{"pool_name": "testPool", "rep_count": "2", "fsType": "ext3"},
		IDGen:                     &sanity.DefaultIDGenerator{},
		DialOptions:               []grpc.DialOption{grpc.WithInsecure()},
		ControllerDialOptions:     []grpc.DialOption{grpc.WithInsecure()},
	}
	sanity.GinkgoTest(config)
})
