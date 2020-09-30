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
package e2e

import (
	. "github.com/onsi/ginkgo"
	_ "github.com/onsi/gomega"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/kubernetes/test/e2e/framework/testfiles"
	"k8s.io/kubernetes/test/e2e/storage/testsuites"
	"k8s.io/kubernetes/test/e2e/storage/utils"
	"path"
)

var CSITestSuites = []func() testsuites.TestSuite{
	testsuites.InitDisruptiveTestSuite,
	testsuites.InitEphemeralTestSuite,
	testsuites.InitMultiVolumeTestSuite,
	testsuites.InitProvisioningTestSuite,
	testsuites.InitSnapshottableTestSuite,
	testsuites.InitSubPathTestSuite,
	testsuites.InitTopologyTestSuite,
	testsuites.InitVolumeExpandTestSuite,
	testsuites.InitVolumeIOTestSuite,
	testsuites.InitVolumeLimitsTestSuite,
	testsuites.InitVolumesTestSuite,
	testsuites.InitVolumeModeTestSuite,
}

// This executes testSuites for csi volumes.
var _ = utils.SIGDescribe("CSI Volumes", func() {
	testfiles.AddFileSource(testfiles.RootFileSource{Root: path.Join(framework.TestContext.RepoRoot, "../../deploy/kubernetes/")})

	curDriver := initDriver("")
	Context(testsuites.GetDriverNameWithFeatureTags(curDriver), func() {
		testsuites.DefineTestSuite(curDriver, CSITestSuites)
	})
})
