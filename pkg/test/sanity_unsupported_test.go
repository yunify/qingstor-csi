//+build !linux

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

package test

import (
	"github.com/yunify/qingstor-csi/pkg/common"
	"github.com/yunify/qingstor-csi/pkg/service"
	"github.com/yunify/qingstor-csi/pkg/storage/mock"
	"k8s.io/klog"
	"os"
	"path/filepath"
	"testing"

	"github.com/kubernetes-csi/csi-test/pkg/sanity"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	socket      = "127.0.0.1:10086"
	tcpEndpoint = "tcp://" + socket
	uds         = "/tmp/csi.sock"
	udsEndpoint = "unix://" + uds

	defaultConfigPath = "/etc/neonsan/qbd.conf"
	defaultPoolName   = "kube"
)

var (
	mockServer service.NonBlockingGRPCServer
	serviceOpt = service.NewOption().SetName("mock.neonsan.csi.com").SetVersion("1.1.0").
		SetNodeId("HelloNeonsan").SetMaxVolume(10).
		SetVolumeCapabilityAccessNodes(service.DefaultVolumeAccessModeType).
		SetControllerServiceCapabilities(service.DefaultControllerServiceCapability).
		SetNodeServiceCapabilities(service.DefaultNodeServiceCapability).
		SetPluginCapabilities(service.DefaultPluginCapability).
		SetRetryTime(service.DefaultBackOff).
		SetRetryCnt(service.DefaultRetryCnt)
)

var _ = BeforeSuite(func() {
	klog.InitFlags(nil)
	mockServer = service.NewNonBlockingGRPCServer()
	mockServer.Start(tcpEndpoint, service.New(serviceOpt, mock.New(), common.NewFakeFormatAndMount()))

})

var _ = AfterSuite(func() {
	if mockServer != nil {
		mockServer.Stop()
	}

})

func TestCSISanity(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CSI Sanity Test Suite")
}

var _ = Describe("Mock Neonsan CSI Driver", func() {
	config := &sanity.Config{
		TargetPath:                filepath.Join(os.TempDir(), "/csi-target"),
		StagingPath:               filepath.Join(os.TempDir(), "/csi-staging"),
		Address:                   socket,
		TestNodeVolumeAttachLimit: true,
		IDGen:                     &sanity.DefaultIDGenerator{},
	}

	sanity.GinkgoTest(config)
})

