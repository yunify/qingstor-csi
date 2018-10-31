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

package cache_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/yunify/qingstor-csi/pkg/neonsan/cache"
	"github.com/yunify/qingstor-csi/pkg/neonsan/manager"
	"sort"
	"strconv"
)

var _ = Describe("Snapshot Cache", func() {
	var cache SnapshotCacheType
	BeforeEach(func() {
		By("new snapshot cache")
		cache = SnapshotCacheType{}
		cache.New()

		By("add snapshots")
		m, n := 5, 4
		for i := 1; i <= m; i++ {
			for j := 1; j <= n; j++ {
				num := j + 100*i
				cache.Add(
					&manager.SnapshotInfo{
						Name:        "snapshot" + strconv.Itoa(j) + "-volume" + strconv.Itoa(i),
						Id:          "25463",
						SizeByte:    2121210000 + int64(num),
						Status:      manager.SnapshotStatusOk,
						CreatedTime: 1535020000 + int64(num),
						SrcVolName:  "volume" + strconv.Itoa(i),
					})
			}
		}
		Expect(len(cache.List())).To(Equal(m * n))
	})

	It("can add snapshot info", func() {
		By("generate snapshot")
		length := len(cache.List())
		exSnap := cache.List()[0]
		addedSnap := &manager.SnapshotInfo{
			Name:        exSnap.Name + "-add",
			Id:          exSnap.Id,
			SizeByte:    exSnap.SizeByte,
			Status:      exSnap.Status,
			CreatedTime: exSnap.CreatedTime,
			SrcVolName:  exSnap.SrcVolName,
		}
		Expect(addedSnap).NotTo(Equal(exSnap))

		By("add snapshot")
		flag := cache.Add(addedSnap)
		findSnap := cache.Find(addedSnap.Name)

		Expect(flag).To(Equal(true))
		Expect(findSnap).To(Equal(addedSnap))
		Expect(len(cache.List())).To(Equal(length + 1))
	})

	It("can re-add snapshot info", func() {
		By("generate snapshot")
		length := len(cache.List())
		exSnap := cache.List()[0]

		By("re-add snapshot")
		flag := cache.Add(exSnap)
		findSnap := cache.Find(exSnap.Name)

		Expect(flag).To(Equal(true))
		Expect(findSnap).To(Equal(exSnap))
		Expect(len(cache.List())).To(Equal(length))

	})

	It("cannot re-add snapshot but incompatible", func() {
		By("generate snapshot")
		length := len(cache.List())
		exSnap := cache.List()[0]
		addedSnap := &manager.SnapshotInfo{
			Name:        exSnap.Name,
			Id:          exSnap.Id,
			SizeByte:    exSnap.SizeByte + 1,
			Status:      exSnap.Status,
			CreatedTime: exSnap.CreatedTime,
			SrcVolName:  exSnap.SrcVolName,
		}
		Expect(addedSnap).NotTo(Equal(exSnap))

		By("add snapshot")
		flag := cache.Add(addedSnap)
		findSnap := cache.Find(addedSnap.Name)

		Expect(flag).To(Equal(false))
		Expect(findSnap).To(Equal(exSnap))
		Expect(len(cache.List())).To(Equal(length))
	})

	It("can delete snapshot info", func() {
		for i := len(cache.List()); i > 0; i-- {
			exSnap := cache.List()[0]

			By("delete snapshot")
			cache.Delete(exSnap.Name)
			findSnap := cache.Find(exSnap.Name)

			Expect(findSnap).To(BeNil())
			Expect(len(cache.List())).To(Equal(i - 1))

			By("re-delete snapshot")
			cache.Delete(exSnap.Name)
			findSnap = cache.Find(exSnap.Name)

			Expect(findSnap).To(BeNil())
			Expect(len(cache.List())).To(Equal(i - 1))
		}
	})

	It("can list sorted info", func() {
		snapList := cache.List()
		nameList := make([]string, 0)
		for i := range snapList {
			nameList = append(nameList, snapList[i].Name)
		}
		By(fmt.Sprintf("%v", nameList))
		Expect(sort.StringsAreSorted(nameList)).To(Equal(true))
	})

	It("can sync cache", func() {
		if hasCli == false {
			Skip(UnsupportCli)
		}
		manager.Pools = []string{"csi", "kube"}
		cache := SnapshotCacheType{}
		cache.New()
		err := cache.Sync()
		fmt.Fprintf(GinkgoWriter, "len %d", len(cache.List()))
		for _, v := range cache.List() {
			fmt.Fprintf(GinkgoWriter, "snapshot %v\n", v)
		}
		Expect(err).To(BeNil())

	})
})
