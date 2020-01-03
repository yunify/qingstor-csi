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
package neonsan

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestUtils(t *testing.T) {
	RegisterFailHandler(Fail)

	Describe("JoinName and SplitName", func() {
		It("VolumeName", func() {
			poolName, volumeName, fullVolumeName := "testPool", "testVolume", "testPool/testVolume"
			Expect(JoinVolumeName(poolName,volumeName)).To(Equal(fullVolumeName))
			poolName2,volumeName2 := SplitVolumeName(fullVolumeName)
			Expect(poolName2).To(Equal(poolName))
			Expect(volumeName2).To(Equal(volumeName))
		})
		It("SnapshotName", func() {
			poolName, volumeName,snapshotName, fullSnapshotName := "testPool", "testVolume", "testSnapshot","testPool/testVolume@testSnapshot"
			Expect(JoinSnapshotName(poolName,volumeName,snapshotName )).To(Equal(fullSnapshotName))
			poolName2,volumeName2,snapshotName2 := SplitSnapshotName(fullSnapshotName)
			Expect(poolName2).To(Equal(poolName))
			Expect(volumeName2).To(Equal(volumeName))
			Expect(snapshotName2).To(Equal(snapshotName))
		})
	})

	Describe("GetPoolName", func() {
		It("nil map", func() {
			Expect(GetPoolName(nil)).To(Equal(""))
		})
		It("map with key ", func() {
			poolName := "beautifulPool"
			Expect(GetPoolName(map[string]string{scPoolName:poolName})).To(Equal(poolName))
		})
		It("map without key", func() {
			Expect(GetPoolName(map[string]string{scPoolName+"x":"y"})).To(Equal(""))
		})
	})

	Describe("GetReplica", func() {
		It("nil map", func() {
			replica, err := GetReplica(nil)
			Expect(err).To(BeNil())
			Expect(replica).To(Equal(defaultReplica))
		})
		It("map with key and right value", func() {
			sReplica, iReplica:= "2",2
			replica, err := GetReplica(map[string]string{scReplicaName:sReplica})
			Expect(err).To(BeNil())
			Expect(replica).To(Equal(iReplica))
		})
		It("map with key and wrong value", func() {
			sReplica := "xx"
			_, err := GetReplica(map[string]string{scReplicaName:sReplica})
			Expect(err).NotTo(BeNil())
		})
		It("map without key", func() {
			replica, err := GetReplica(map[string]string{scReplicaName+"xx":"y"})
			Expect(err).To(BeNil())
			Expect(replica).To(Equal(defaultReplica))
		})
	})
	RunSpecs(t, "service.util")
}
