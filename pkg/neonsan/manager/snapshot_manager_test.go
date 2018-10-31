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
	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/yunify/qingstor-csi/pkg/neonsan/manager"
	"github.com/yunify/qingstor-csi/pkg/neonsan/util"
	"os"
	"path"
)

var _ = Describe("Snapshot", func() {
	BeforeEach(func() {
		if hasCli == false {
			Skip(UnsupportCli)
		}
		By("checking pools")
		poolInfo, err := manager.FindPool(TestPool)
		Expect(err).To(BeNil())
		Expect(poolInfo).NotTo(BeNil())

		poolInfo, err = manager.FindPool(TestPoolFake)
		Expect(err).To(BeNil())
		Expect(poolInfo).To(BeNil())

		By("creating volumes")
		info, err := manager.FindVolume(TestVolume1, TestPool)
		Expect(err).To(BeNil())
		if info == nil {
			manager.CreateVolume(TestVolume1, TestPool, 10*util.Gib, 1)
		}
		info, err = manager.FindVolume(TestVolume2, TestPool)
		Expect(err).To(BeNil())
		if info == nil {
			manager.CreateVolume(TestVolume2, TestPool, 10*util.Gib, 1)
		}
		info, err = manager.FindVolume(TestVolumeFake, TestPool)
		Expect(err).To(BeNil())
		if info != nil {
			manager.DeleteVolume(TestVolumeFake, TestPool)
		}
	})

	AfterEach(func() {
		By("removing volumes")
		err := manager.DeleteVolume(TestVolume1, TestPool)
		Expect(err).To(BeNil())

		err = manager.DeleteVolume(TestVolume2, TestPool)
		Expect(err).To(BeNil())
	})

	It("can create snapshot", func() {
		By("creating snapshot normally")
		info, err := manager.CreateSnapshot(TestSnap1, TestVolume1, TestPool)
		Expect(err).To(BeNil())
		exInfo, err := manager.FindSnapshot(TestSnap1, TestVolume1, TestPool)
		Expect(err).To(BeNil())
		Expect(exInfo).To(Equal(info))

		By("recreate snapshot")
		_, err = manager.CreateSnapshot(TestSnap1, TestVolume1, TestPool)
		Expect(err).NotTo(BeNil())

		By("create snapshot in fake volume")
		_, err = manager.CreateSnapshot(TestSnap2, TestVolumeFake, TestPool)
		Expect(err).NotTo(BeNil())

		By("create snapshot in fake pool")
		_, err = manager.CreateSnapshot(TestSnap2, TestVolume1, TestPoolFake)
		Expect(err).NotTo(BeNil())
	})

	It("can delete snapshot", func() {
		By("creating snapshot for deleting")
		info, err := manager.CreateSnapshot(TestSnap1, TestVolume1, TestPool)
		Expect(err).To(BeNil())
		Expect(info).NotTo(BeNil())
		info, err = manager.CreateSnapshot(TestSnap2, TestVolume1, TestPool)
		Expect(err).To(BeNil())
		Expect(info).NotTo(BeNil())

		By("deleting snapshot")
		err = manager.DeleteSnapshot(TestSnap1, TestVolume1, TestPool)
		Expect(err).To(BeNil())
		info, err = manager.FindSnapshot(TestSnap1, TestVolume1, TestPool)
		Expect(err).To(BeNil())
		Expect(info).To(BeNil())

		By("deleting snapshot again")
		err = manager.DeleteSnapshot(TestSnap1, TestVolume1, TestPool)
		Expect(err).NotTo(BeNil())
		info, err = manager.FindSnapshot(TestSnap1, TestVolume1, TestPool)
		Expect(err).To(BeNil())
		Expect(info).To(BeNil())

		By("deleting non-existed snapshot")
		err = manager.DeleteSnapshot(TestSnap2, TestVolumeFake, TestPool)
		Expect(err).NotTo(BeNil())
	})

	Describe("finding snapshot", func() {
		var info1, info2 *manager.SnapshotInfo
		BeforeEach(func() {
			By("creating snapshot")
			var err error
			info1, err = manager.CreateSnapshot(TestSnap1, TestVolume1, TestPool)
			Expect(err).To(BeNil())
			Expect(info1).NotTo(BeNil())
			info2, err = manager.CreateSnapshot(TestSnap2, TestVolume1, TestPool)
			Expect(err).To(BeNil())
			Expect(info2).NotTo(BeNil())
		})

		It("can find specified snapshot", func() {
			By("existed snapshot")
			exInfo, err := manager.FindSnapshot(TestSnap1, TestVolume1, TestPool)
			Expect(err).To(BeNil())
			Expect(exInfo).To(Equal(info1))
			exInfo, err = manager.FindSnapshot(TestSnap2, TestVolume1, TestPool)
			Expect(err).To(BeNil())
			Expect(exInfo).To(Equal(info2))

			By("volume doesn't contains any snapshot")
			exInfo, err = manager.FindSnapshot(TestSnap1, TestVolume2, TestPool)
			Expect(err).To(BeNil())
			Expect(exInfo).To(BeNil())

			By("fake snapshot name")
			exInfo, err = manager.FindSnapshot(TestSnapFake, TestVolume1, TestPool)
			Expect(err).To(BeNil())
			Expect(exInfo).To(BeNil())

			By("fake volume name")
			exInfo, err = manager.FindSnapshot(TestSnap1, TestVolumeFake, TestPool)
			Expect(err).NotTo(BeNil())
		})

		It("can list snapshot by volume", func() {
			By("finding two snapshot")
			infoList, err := manager.ListSnapshotByVolume(TestVolume1, TestPool)
			Expect(err).To(BeNil())
			Expect(len(infoList)).To(Equal(2))

			By("finding no snapshot")
			infoList, err = manager.ListSnapshotByVolume(TestVolume2, TestPool)
			Expect(err).To(BeNil())
			Expect(len(infoList)).To(Equal(0))

			By("find fake pool")
			_, err = manager.ListSnapshotByVolume(TestVolume1, TestPoolFake)
			Expect(err).NotTo(BeNil())

			By("find fake volume")
			_, err = manager.ListSnapshotByVolume(TestVolumeFake, TestPool)
			Expect(err).NotTo(BeNil())
		})
	})

	It("create volume from snapshot", func() {
		By("creating snapshot")
		info1, err := manager.CreateSnapshot(TestSnap1, TestVolume1, TestPool)
		Expect(err).To(BeNil())
		Expect(info1).NotTo(BeNil())

		By("exporting snapshot")
		err = manager.ExportSnapshot(manager.ExportSnapshotRequest{
			SnapName:   info1.Name,
			SrcVolName: info1.SrcVolName,
			PoolName:   info1.Pool,
			FilePath:   path.Join(".", info1.Name),
			Protocol:   util.ProtocolTCP,
		})
		Expect(err).To(BeNil())

		By("checking exporting file")
		_, err = os.Stat(path.Join(".", info1.Name))
		Expect(err).To(BeNil())

		By("importing snapshot")
		err = manager.ImportSnapshot(manager.ImportSnapshotRequest{
			VolName:  TestVolume2,
			PoolName: TestPool,
			FilePath: path.Join(".", info1.Name),
			Protocol: util.ProtocolTCP,
		})
		Expect(err).To(BeNil())

		By("checking snapshot")
		exInfo, err := manager.FindSnapshot(info1.Name, TestVolume2, TestPool)
		Expect(err).To(BeNil())
		Expect(exInfo).NotTo(BeNil())

		By("rollback snapshot")
		err = manager.RollbackSnapshot(manager.RollbackSnapshotRequest{
			VolumeName: TestVolume2,
			Pool:       TestPool,
			SnapName:   TestSnap1,
		})
		Expect(err).To(BeNil())
	})
})

var _ = Describe("Snapshot Info", func() {
	DescribeTable("convert NeonSAN snapshot to CSI",
		func(neonInfo *manager.SnapshotInfo, csiInfo *csi.Snapshot) {
			resInfo := manager.ConvertNeonToCsiSnap(neonInfo)
			Expect(resInfo).To(Equal(csiInfo))
		},
		Entry("valid NeonSAN snapshot",
			&manager.SnapshotInfo{
				Name:        TestSnap1,
				Id:          "25463",
				SizeByte:    2147483648,
				Status:      manager.SnapshotStatusOk,
				Pool:        TestPool,
				CreatedTime: 1535024379,
				SrcVolName:  TestVolume1,
			},
			&csi.Snapshot{
				SizeBytes:      2147483648,
				Id:             TestSnap1,
				SourceVolumeId: TestVolume1,
				CreatedAt:      1535024379,
				Status: &csi.SnapshotStatus{
					Type: csi.SnapshotStatus_READY,
				},
			}),
		Entry("without snapshot id",
			&manager.SnapshotInfo{
				Name:        TestSnap1,
				SizeByte:    2147483648,
				Status:      manager.SnapshotStatusOk,
				Pool:        TestPool,
				CreatedTime: 1535024379,
				SrcVolName:  TestVolume1,
			},
			&csi.Snapshot{
				SizeBytes:      2147483648,
				Id:             TestSnap1,
				SourceVolumeId: TestVolume1,
				CreatedAt:      1535024379,
				Status: &csi.SnapshotStatus{
					Type: csi.SnapshotStatus_READY,
				},
			}),
		Entry("zero value snapshot info",
			&manager.SnapshotInfo{},
			&csi.Snapshot{}),
		Entry("nil snapshot info", nil, nil),
	)

	DescribeTable("convert NeonSAN snap list to CSI",
		func(neonInfo []*manager.SnapshotInfo, csiInfo []*csi.ListSnapshotsResponse_Entry) {
			resInfo := manager.ConvertNeonSnapToListSnapResp(neonInfo)
			Expect(resInfo).To(Equal(csiInfo))
		},
		Entry("normal snapshot info array",
			[]*manager.SnapshotInfo{
				{
					Name:        TestSnap1,
					Id:          "25463",
					SizeByte:    2147483648,
					Status:      manager.SnapshotStatusOk,
					CreatedTime: 1535024299,
					SrcVolName:  TestVolume1,
				},
				{
					Name:        TestSnap2,
					Id:          "25464",
					SizeByte:    2147483648,
					Status:      manager.SnapshotStatusOk,
					CreatedTime: 1535024379,
					SrcVolName:  TestVolume2,
				},
			},
			[]*csi.ListSnapshotsResponse_Entry{
				{
					Snapshot: &csi.Snapshot{
						Id:             TestSnap1,
						SizeBytes:      2147483648,
						SourceVolumeId: TestVolume1,
						CreatedAt:      1535024299,
						Status: &csi.SnapshotStatus{
							Type: csi.SnapshotStatus_READY,
						},
					},
				},
				{
					Snapshot: &csi.Snapshot{
						Id:             TestSnap2,
						SizeBytes:      2147483648,
						SourceVolumeId: TestVolume2,
						CreatedAt:      1535024379,
						Status: &csi.SnapshotStatus{
							Type: csi.SnapshotStatus_READY,
						},
					},
				},
			},
		),
		Entry("nil array", nil, nil),
	)
})
