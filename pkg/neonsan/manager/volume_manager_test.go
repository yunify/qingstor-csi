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
	"github.com/yunify/qingstor-csi/pkg/neonsan/util"
)

var _ = Describe("VolumeManager", func() {
	BeforeEach(func() {
		if hasCli == false {
			Skip(UnsupportCli)
		}
		By("clean volume")
		info, err := manager.FindVolume(TestVolume1, TestPool)
		Expect(err).To(BeNil())
		if info != nil {
			manager.DeleteVolume(info.Name, info.Pool)
		}
		info, err = manager.FindVolume(TestVolume2, TestPool)
		Expect(err).To(BeNil())
		if info != nil {
			manager.DeleteVolume(info.Name, info.Pool)
		}
		info, err = manager.FindVolume(TestVolumeFake, TestPool)
		Expect(err).To(BeNil())
		if info != nil {
			manager.DeleteVolume(info.Name, info.Pool)
		}
	})

	Describe("CreateVolume", func() {
		It("can create volume", func() {
			By("creating volume")
			volInfo1, err := manager.CreateVolume(TestVolume1, TestPool, util.Gib*2, 1)
			Expect(err).To(BeNil())
			Expect(volInfo1.Name).To(Equal(TestVolume1))

			By("finding volume")
			exVolInfo, err := manager.FindVolume(volInfo1.Name, volInfo1.Pool)
			Expect(err).To(BeNil())
			Expect(exVolInfo).To(Equal(volInfo1))

			By("re-creating volume")
			volInfo2, err := manager.CreateVolume(TestVolume1, TestPool, util.Gib*2, 1)
			Expect(err).NotTo(BeNil())
			Expect(volInfo2).To(BeNil())
		})
	})

	Describe("DeleteVolume", func() {
		BeforeEach(func() {
			By("creating volume")
			volInfo1, err := manager.CreateVolume(TestVolume1, TestPool, util.Gib*2, 1)
			Expect(err).To(BeNil())
			Expect(volInfo1.Name).To(Equal(TestVolume1))
		})

		It("can delete volume", func() {
			By("deleting volume")
			err := manager.DeleteVolume(TestVolume1, TestPool)
			Expect(err).To(BeNil())
		})

		It("cannot re-delete volume", func() {
			By("deleting volume")
			err := manager.DeleteVolume(TestVolume1, TestPool)
			Expect(err).To(BeNil())

			By("re-deleting volume")
			err = manager.DeleteVolume(TestVolume1, TestPool)
			Expect(err).NotTo(BeNil())
		})

		It("cannot delete fake volume", func() {
			By("deleting fake volume")
			err := manager.DeleteVolume(TestVolumeFake, TestPool)
			Expect(err).NotTo(BeNil())
		})
	})

	Describe("Attach and Detach Volume", func() {
		BeforeEach(func() {
			By("creating volume")
			volInfo1, err := manager.CreateVolume(TestVolume1, TestPool, util.Gib*2, 1)
			Expect(err).To(BeNil())
			Expect(volInfo1.Name).To(Equal(TestVolume1))
		})

		It("can attach volume", func() {
			By("succeed to attach volume")
			err := manager.AttachVolume(TestVolume1, TestPool)
			Expect(err).To(BeNil())

			By("cleaner")
			err = manager.DetachVolume(TestVolume1, TestPool)
			Expect(err).To(BeNil())
		})

		It("cannot re-attach volume", func() {
			By("succeed to attach volume")
			err := manager.AttachVolume(TestVolume1, TestPool)
			Expect(err).To(BeNil())

			By("failed to re-attach volume")
			err = manager.AttachVolume(TestVolume1, TestPool)
			Expect(err).NotTo(BeNil())

			By("cleaner")
			err = manager.DetachVolume(TestVolume1, TestPool)
			Expect(err).To(BeNil())
		})

		It("cannot attach and detach not-exist volume", func() {
			err := manager.AttachVolume(TestVolumeFake, TestPool)
			Expect(err).NotTo(BeNil())
			err = manager.DetachVolume(TestVolumeFake, TestPool)
			Expect(err).NotTo(BeNil())
		})
	})

	Describe("Consult Volume", func() {
		BeforeEach(func() {
			By("creating volume")
			volInfo1, err := manager.CreateVolume(TestVolume1, TestPool, util.Gib*2, 1)
			Expect(err).To(BeNil())
			Expect(volInfo1.Name).To(Equal(TestVolume1))

			volInfo2, err := manager.CreateVolume(TestVolume2, TestPool, util.Gib*2, 1)
			Expect(err).To(BeNil())
			Expect(volInfo2.Name).To(Equal(TestVolume2))
		})

		It("find volume", func() {
			volInfo, err := manager.FindVolume(TestVolume1, TestPool)
			Expect(err).To(BeNil())
			Expect(volInfo).NotTo(BeNil())
		})

		It("find non-exist volume", func() {
			volInfo, err := manager.FindVolume(TestVolumeFake, TestPool)
			Expect(err).To(BeNil())
			Expect(volInfo).To(BeNil())
		})

		It("find volume in non-exist pool", func() {
			_, err := manager.FindVolume(TestVolume1, TestPoolFake)
			Expect(err).NotTo(BeNil())
		})

		It("find volume without pool", func() {
			volInfo, err := manager.FindVolumeWithoutPool(TestVolume1)
			Expect(err).To(BeNil())
			Expect(volInfo).NotTo(BeNil())
		})

		It("find non-exist volume without pool", func() {
			volInfo, err := manager.FindVolumeWithoutPool(TestVolumeFake)
			Expect(err).To(BeNil())
			Expect(volInfo).To(BeNil())
		})

		It("list volume by pool", func() {
			volInfos, err := manager.ListVolumeByPool(TestPool)
			Expect(err).To(BeNil())
			Expect(len(volInfos) > 0).To(Equal(true))
		})

		It("list volume by non-exist pool", func() {
			_, err := manager.ListVolumeByPool(TestPoolFake)
			Expect(err).NotTo(BeNil())
		})

		It("find attached volume without pool", func() {
			err := manager.AttachVolume(TestVolume1, TestPool)
			Expect(err).To(BeNil())
			volInfo, err := manager.FindAttachedVolumeWithoutPool(TestVolume1)
			Expect(err).To(BeNil())
			Expect(volInfo.Name).To(Equal(TestVolume1))

			By("cleaner")
			err = manager.DetachVolume(TestVolume1,TestPool)
			Expect(err).To(BeNil())
		})

		It("find non-attached volume without pool", func() {
			volInfo, err := manager.FindAttachedVolumeWithoutPool(TestVolumeFake)
			Expect(err).To(BeNil())
			Expect(volInfo).To(BeNil())
		})
	})
})
