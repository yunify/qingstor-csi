package manager_test

import (
	"flag"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/yunify/qingstor-csi/pkg/neonsan/manager"
	"testing"
)

const (
	TestPool     = "csi"
	TestPoolFake = "fake"

	TestVolume1    = "vol1"
	TestVolume2    = "vol2"
	TestVolumeFake = "fake"

	TestSnap1    = "snap1"
	TestSnap2    = "snap2"
	TestSnapFake = "fake"
)

var hasCli bool = false

const (
	UnsupportCli = "Unsupport NeonSAN CLI"
)

func init() {
	flag.BoolVar(&hasCli, "hasCli", false, "current environment support NeonSAN CLI")
	manager.Pools = append(manager.Pools, TestPool)
}

func TestManager(t *testing.T) {

	RegisterFailHandler(Fail)
	RunSpecs(t, "Manager Suite")
}
