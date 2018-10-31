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
	// add csi pool
	manager.Pools = append(manager.Pools, TestPool)
}

func TestManager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Manager Suite")
}
