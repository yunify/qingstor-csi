//+build !linux

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

