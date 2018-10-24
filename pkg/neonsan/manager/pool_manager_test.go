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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/yunify/qingstor-csi/pkg/neonsan/manager"
)

var _ = Describe("Pool", func() {
	It("can find pool", func() {
		By("csi pool")
		poolInfo, err := manager.FindPool(TestPool)
		Expect(err).To(BeNil())
		Expect(poolInfo.Name).To(Equal(TestPool))

		By("fake pool")
		poolInfo, err = manager.FindPool(TestPoolFake)
		Expect(err).NotTo(BeNil())
	})

	It("can list pool", func() {
		pools := manager.ListPoolName()
		Expect(len(pools)).NotTo(Equal(0))
	})

})
