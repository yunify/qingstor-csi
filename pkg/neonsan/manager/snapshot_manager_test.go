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

package manager

import (
	"errors"
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/yunify/qingstor-csi/pkg/neonsan/util"
	"reflect"
	"testing"
)

const (
	SnapTestSnapTest = "test"
	SnapTestSnapFake = "fake"

	SnapTestPoolCsi  = "csi"
	SnapTestPoolFake = "fake"

	SnapTestVolumeFoo    = "foo"
	SnapTestVolumeFake   = "fake"
	SnapTestVolumeNosnap = "nosnap"
)

func TestCheck(t *testing.T) {
	Pools = []string{SnapTestPoolCsi}
	vol, err := FindVolume(SnapTestVolumeFoo, SnapTestPoolCsi)
	if err != nil {
		t.Error(err)
	}
	if vol == nil {
		t.Errorf("volume [%s] does not exist", SnapTestVolumeFoo)
	}
	vol, err = FindVolume(SnapTestVolumeNosnap, SnapTestPoolCsi)
	if err != nil {
		t.Error(err)
	}
	if vol == nil {
		t.Errorf("volume [%s] does not exist", SnapTestVolumeFoo)
	}
}

func TestCleaner(t *testing.T) {
	Pools = []string{SnapTestPoolCsi}
	DeleteVolume(SnapTestVolumeFoo, SnapTestPoolCsi)
	DeleteVolume(SnapTestVolumeNosnap, SnapTestPoolCsi)
}

func TestInit(t *testing.T) {
	Pools = []string{SnapTestPoolCsi}
	if _, err := CreateVolume(SnapTestVolumeFoo, SnapTestPoolCsi, util.Gib*10, 1); err != nil {
		t.Error(err)
	}
	if _, err := CreateVolume(SnapTestVolumeNosnap, SnapTestPoolCsi, util.Gib*10, 1); err != nil {
		t.Error(err)
	}
}

func TestCreateSnapshot(t *testing.T) {
	Pools = []string{SnapTestPoolCsi}
	tests := []struct {
		name   string
		input  *SnapshotInfo
		output *SnapshotInfo
		err    error
	}{
		{
			name: "succeed to create snapshot",
			input: &SnapshotInfo{
				Name:       SnapTestSnapTest,
				Pool:       SnapTestPoolCsi,
				SrcVolName: SnapTestVolumeFoo,
			},
			output: &SnapshotInfo{
				Name:       SnapTestSnapTest,
				Pool:       SnapTestPoolCsi,
				SrcVolName: SnapTestVolumeFoo,
			},
			err: nil,
		},
		{
			name: "recreate snapshot",
			input: &SnapshotInfo{
				Name:       SnapTestSnapTest,
				Pool:       SnapTestPoolCsi,
				SrcVolName: SnapTestVolumeFoo,
			},
			output: nil,
			err:    errors.New("raise error"),
		},
		{
			name: "fake volume name",
			input: &SnapshotInfo{
				Name:       SnapTestSnapTest,
				Pool:       SnapTestPoolCsi,
				SrcVolName: SnapTestVolumeFake,
			},
			output: nil,
			err:    errors.New("raise error"),
		},
		{
			name: "fake pool name",
			input: &SnapshotInfo{
				Name:       SnapTestSnapTest,
				Pool:       SnapTestPoolFake,
				SrcVolName: SnapTestVolumeFoo,
			},
			output: nil,
			err:    errors.New("raise error"),
		},
	}
	for _, v := range tests {
		snapInfo, err := CreateSnapshot(v.input.Name, v.input.SrcVolName, v.input.Pool)
		if (v.err != nil && err == nil) || (v.err == nil && err != nil) {
			t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.err, err)
		}

		if v.err == nil && err == nil {
			if v.output != nil && snapInfo != nil {
				if v.output.Name != snapInfo.Name {
					t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.output, snapInfo)
				}
			} else {
				t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.output, snapInfo)
			}
		}
	}
}

func TestFindSnapshot(t *testing.T) {
	Pools = []string{SnapTestPoolCsi}
	tests := []struct {
		name   string
		input  *SnapshotInfo
		output *SnapshotInfo
		err    error
	}{
		{
			name: "succeed to find snapshot",
			input: &SnapshotInfo{
				Name:       SnapTestSnapTest,
				Pool:       SnapTestPoolCsi,
				SrcVolName: SnapTestVolumeFoo,
			},
			output: &SnapshotInfo{
				Name:       SnapTestSnapTest,
				Pool:       SnapTestPoolCsi,
				SrcVolName: SnapTestVolumeFoo,
			},
			err: nil,
		},
		{
			name: "volume doesn't contains any snapshot",
			input: &SnapshotInfo{
				Name:       SnapTestSnapTest,
				Pool:       SnapTestPoolCsi,
				SrcVolName: SnapTestVolumeNosnap,
			},
			output: nil,
			err:    nil,
		},
		{
			name: "fake snapshot name",
			input: &SnapshotInfo{
				Name:       SnapTestSnapFake,
				Pool:       SnapTestPoolCsi,
				SrcVolName: SnapTestVolumeFoo,
			},
			output: nil,
			err:    nil,
		},
		{
			name: "fake volume name",
			input: &SnapshotInfo{
				Name:       SnapTestSnapTest,
				Pool:       SnapTestPoolCsi,
				SrcVolName: SnapTestVolumeFake,
			},
			output: nil,
			err:    errors.New("raise error"),
		},
		{
			name: "fake pool name",
			input: &SnapshotInfo{
				Name:       SnapTestSnapTest,
				Pool:       SnapTestPoolFake,
				SrcVolName: SnapTestVolumeFoo,
			},
			output: nil,
			err:    errors.New("raise error"),
		},
	}
	for _, v := range tests {
		snapInfo, err := FindSnapshot(v.input.Name, v.input.SrcVolName, v.input.Pool)
		if (v.err != nil && err == nil) || (v.err == nil && err != nil) {
			t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.err, err)
		}
		if v.err == nil && err == nil {
			if v.output != nil && snapInfo != nil {
				if v.output.Name != snapInfo.Name {
					t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.output, snapInfo)
				}
			} else if (v.output != nil && snapInfo == nil) || (v.output == nil && snapInfo != nil) {
				t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.output, snapInfo)
			}
		}
	}
}

func TestFindSnapshotWithoutPool(t *testing.T) {
	Pools = []string{SnapTestPoolCsi}
	tests := []struct {
		name   string
		input  *SnapshotInfo
		output *SnapshotInfo
		err    error
	}{
		{
			name: "succeed to find snapshot",
			input: &SnapshotInfo{
				Name: SnapTestSnapTest,
			},
			output: &SnapshotInfo{
				Name:       SnapTestSnapTest,
				Pool:       SnapTestPoolCsi,
				SrcVolName: SnapTestVolumeFoo,
			},
			err: nil,
		},
		{
			name: "not found snapshot",
			input: &SnapshotInfo{
				Name: SnapTestSnapFake,
			},
			output: nil,
			err:    nil,
		},
	}
	for _, v := range tests {
		snapInfo, err := FindSnapshotWithoutPool(v.input.Name)
		if (v.err != nil && err == nil) || (v.err == nil && err != nil) {
			t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.err, err)
		}
		if v.err == nil && err == nil {
			if v.output != nil && snapInfo != nil {
				if v.output.Name != snapInfo.Name || v.output.Pool != snapInfo.Pool {
					t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.output, snapInfo)
				}
			} else if (v.output != nil && snapInfo == nil) || (v.output == nil && snapInfo != nil) {
				t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.output, snapInfo)
			}
		}
	}
}

func TestListSnapshotByVolume(t *testing.T) {
	Pools = []string{SnapTestPoolCsi}
	tests := []struct {
		name   string
		input  *SnapshotInfo
		output []*SnapshotInfo
		err    error
	}{
		{
			name: "succeed to find snapshot",
			input: &SnapshotInfo{
				Pool:       SnapTestPoolCsi,
				SrcVolName: SnapTestVolumeFoo,
			},
			output: []*SnapshotInfo{
				{
					Name:       SnapTestSnapTest,
					Pool:       SnapTestPoolCsi,
					SrcVolName: SnapTestVolumeFoo,
				},
			},
			err: nil,
		},
		{
			name: "find no snapshot",
			input: &SnapshotInfo{
				Pool:       SnapTestPoolCsi,
				SrcVolName: SnapTestVolumeNosnap,
			},
			output: nil,
			err:    nil,
		},
		{
			name: "find fake pool",
			input: &SnapshotInfo{
				Pool:       SnapTestPoolFake,
				SrcVolName: SnapTestVolumeFoo,
			},
			output: nil,
			err:    errors.New("raise error"),
		},
		{
			name: "find fake volume",
			input: &SnapshotInfo{
				Pool:       SnapTestPoolCsi,
				SrcVolName: SnapTestVolumeFake,
			},
			output: nil,
			err:    errors.New("raise error"),
		},
	}
	for _, v := range tests {
		snapList, err := ListSnapshotByVolume(v.input.SrcVolName, v.input.Pool)
		if (v.err != nil && err == nil) || (v.err == nil && err != nil) {
			t.Errorf("name [%s]: error expect [%v], but actually [%v]", v.name, v.err, err)
		}
		if v.err == nil && err == nil {
			if len(v.output) != len(snapList) {
				t.Errorf("name [%s]: error expect [%d], but actually [%d]", v.name, len(v.output), len(snapList))
			} else {
				for i := range v.output {
					if v.output[i].Name != snapList[i].Name || v.output[i].Pool != snapList[i].Pool {
						t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.output, snapList)
						break
					}
				}
			}
		}
	}
}

func TestDeleteSnapshot(t *testing.T) {
	Pools = []string{SnapTestPoolCsi}
	tests := []struct {
		name     string
		snapInfo *SnapshotInfo
		err      error
	}{
		{
			name: "succeed to delete snapshot",
			snapInfo: &SnapshotInfo{
				Name:       SnapTestSnapTest,
				Pool:       SnapTestPoolCsi,
				SrcVolName: SnapTestVolumeFoo,
			},
			err: nil,
		},
		{
			name: "re-delete snapshot",
			snapInfo: &SnapshotInfo{
				Name:       SnapTestSnapTest,
				Pool:       SnapTestPoolCsi,
				SrcVolName: SnapTestVolumeFoo,
			},
			err: errors.New("raise error"),
		},
		{
			name: "failed to delete snapshot",
			snapInfo: &SnapshotInfo{
				Name:       SnapTestSnapTest,
				Pool:       SnapTestPoolCsi,
				SrcVolName: SnapTestVolumeFake,
			},
			err: errors.New("raise error"),
		},
	}
	for _, v := range tests {
		err := DeleteSnapshot(v.snapInfo.Name, v.snapInfo.SrcVolName, v.snapInfo.Pool)
		if (v.err != nil && err == nil) || (v.err == nil && err != nil) {
			t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.err, err)
		}
	}
}

func TestConvertNeonToCsiSnap(t *testing.T) {
	Pools = []string{SnapTestPoolCsi}
	tests := []struct {
		caseName string
		neonSnap *SnapshotInfo
		csiSnap  *csi.Snapshot
	}{
		{
			caseName: "valid NeonSAN snapshot",
			neonSnap: &SnapshotInfo{
				Name:        SnapTestSnapTest,
				Id:          "25463",
				SizeByte:    2147483648,
				Status:      SnapshotStatusOk,
				Pool:        SnapTestPoolCsi,
				CreatedTime: 1535024379,
				SrcVolName:  SnapTestVolumeFoo,
			},
			csiSnap: &csi.Snapshot{
				SizeBytes:      2147483648,
				Id:             SnapTestSnapTest,
				SourceVolumeId: SnapTestVolumeFoo,
				CreatedAt:      1535024379,
				Status: &csi.SnapshotStatus{
					Type: csi.SnapshotStatus_READY,
				},
			},
		},
		{
			caseName: "without snap name",
			neonSnap: &SnapshotInfo{
				Name:        SnapTestSnapTest,
				SizeByte:    2147483648,
				Status:      SnapshotStatusOk,
				Pool:        SnapTestPoolCsi,
				CreatedTime: 1535024379,
				SrcVolName:  SnapTestVolumeFoo,
			},
			csiSnap: &csi.Snapshot{
				SizeBytes:      2147483648,
				Id:             SnapTestSnapTest,
				SourceVolumeId: SnapTestVolumeFoo,
				CreatedAt:      1535024379,
				Status: &csi.SnapshotStatus{
					Type: csi.SnapshotStatus_READY,
				},
			},
		},
		{
			caseName: "zero value snap info",
			neonSnap: &SnapshotInfo{},
			csiSnap:  &csi.Snapshot{},
		},
		{
			caseName: "nil snap info",
			neonSnap: nil,
			csiSnap:  nil,
		},
	}
	for _, v := range tests {
		csiSnap := ConvertNeonToCsiSnap(v.neonSnap)
		if !reflect.DeepEqual(v.csiSnap, csiSnap) {
			t.Errorf("name [%s]: expect [%v], but actually [%v]", v.caseName, v.csiSnap, csiSnap)
		}
	}
}

func TestConvertNeonSnapToListSnapResp(t *testing.T) {
	Pools = []string{SnapTestPoolCsi}
	tests := []struct {
		caseName  string
		neonSnaps []*SnapshotInfo
		respList  []*csi.ListSnapshotsResponse_Entry
	}{
		{
			caseName: "normal snapshot info array",
			neonSnaps: []*SnapshotInfo{
				{
					Name:        "snapshot1",
					Id:          "25463",
					SizeByte:    2147483648,
					Status:      SnapshotStatusOk,
					CreatedTime: 1535024299,
					SrcVolName:  "volume1",
				},
				{
					Name:        "snapshot2",
					Id:          "25464",
					SizeByte:    2147483648,
					Status:      SnapshotStatusOk,
					CreatedTime: 1535024379,
					SrcVolName:  "volume2",
				},
			},
			respList: []*csi.ListSnapshotsResponse_Entry{
				{
					Snapshot: &csi.Snapshot{
						Id:             "snapshot1",
						SizeBytes:      2147483648,
						SourceVolumeId: "volume1",
						CreatedAt:      1535024299,
						Status: &csi.SnapshotStatus{
							Type: csi.SnapshotStatus_READY,
						},
					},
				},
				{
					Snapshot: &csi.Snapshot{
						Id:             "snapshot2",
						SizeBytes:      2147483648,
						SourceVolumeId: "volume2",
						CreatedAt:      1535024379,
						Status: &csi.SnapshotStatus{
							Type: csi.SnapshotStatus_READY,
						},
					},
				},
			},
		},
		{
			caseName:  "nil array",
			neonSnaps: nil,
			respList:  nil,
		},
	}
	for _, v := range tests {
		respList := ConvertNeonSnapToListSnapResp(v.neonSnaps)
		if !reflect.DeepEqual(v.respList, respList) {
			t.Errorf("name [%s]: expect [%v], but actually [%v]", v.caseName, v.respList, respList)
		}
	}
}

func TestReadListPage(t *testing.T) {
	Pools = []string{SnapTestPoolCsi}
	exampleFullList := []*SnapshotInfo{
		{
			Name:        "snapshot1",
			Id:          "25463",
			SizeByte:    2147483648,
			Status:      SnapshotStatusOk,
			CreatedTime: 1535024299,
			SrcVolName:  "volume1",
		},
		{
			Name:        "snapshot2",
			Id:          "25466",
			SizeByte:    2147483648,
			Status:      SnapshotStatusOk,
			CreatedTime: 1535024266,
			SrcVolName:  "volume1",
		},
		{
			Name:        "snapshot3",
			Id:          "25472",
			SizeByte:    2147483648,
			Status:      SnapshotStatusOk,
			CreatedTime: 1535024272,
			SrcVolName:  "volume1",
		},
		{
			Name:        "snapshot2",
			Id:          "25564",
			SizeByte:    2147485648,
			Status:      SnapshotStatusOk,
			CreatedTime: 1535025379,
			SrcVolName:  "volume2",
		},
		{
			Name:        "snapshot3",
			Id:          "25564",
			SizeByte:    2143285648,
			Status:      SnapshotStatusOk,
			CreatedTime: 1535024379,
			SrcVolName:  "volume2",
		},
		{
			Name:        "snapshot1",
			Id:          "25664",
			SizeByte:    2147485648,
			Status:      SnapshotStatusOk,
			CreatedTime: 1535026379,
			SrcVolName:  "volume3",
		},
		{
			Name:        "snapshot4",
			Id:          "25564",
			SizeByte:    2141285648,
			Status:      SnapshotStatusOk,
			CreatedTime: 1535064379,
			SrcVolName:  "volume1",
		},
	}
	tests := []struct {
		caseName    string
		fullList    []*SnapshotInfo
		page        int
		itemPerPage int
		pageList    []*SnapshotInfo
	}{
		{
			caseName:    "normal read page 1 and 3 item/page",
			fullList:    exampleFullList,
			page:        1,
			itemPerPage: 3,
			pageList:    exampleFullList[:3],
		},
		{
			caseName:    "normal read page 2 and 3 item/page",
			fullList:    exampleFullList,
			page:        2,
			itemPerPage: 3,
			pageList:    exampleFullList[3:6],
		},
		{
			caseName:    "normal read page 3 and 3 item/page",
			fullList:    exampleFullList,
			page:        3,
			itemPerPage: 3,
			pageList:    exampleFullList[6:],
		},
		{
			caseName:    "normal read page 3 and 2 item/page",
			fullList:    exampleFullList,
			page:        3,
			itemPerPage: 2,
			pageList:    exampleFullList[4:6],
		},
		{
			caseName:    "normal read page 4 and 2 item/page",
			fullList:    exampleFullList,
			page:        4,
			itemPerPage: 2,
			pageList:    exampleFullList[6:],
		},
		{
			caseName:    "nil info",
			fullList:    nil,
			page:        3,
			itemPerPage: 3,
			pageList:    nil,
		},
	}
	for _, v := range tests {
		pageList, err := ReadListPage(v.fullList, v.page, v.itemPerPage)
		if err != nil {
			continue
		} else if !reflect.DeepEqual(v.pageList, pageList) {
			t.Errorf("name [%s]: expect [%v], but actually [%v]", v.caseName, v.pageList, pageList)
		}
	}
}

func TestSnapshotCache(t *testing.T) {
	Pools = []string{SnapTestPoolCsi}
	var snapArr []SnapshotInfo
	for vol := 0; vol < 10; vol++ {
		for snap := 0; snap < 20; snap++ {
			tmp := SnapshotInfo{
				Name:        fmt.Sprintf("vol-%d-snap-%d", vol, snap),
				Id:          fmt.Sprintf("3996165079%d%d", vol, snap),
				SizeByte:    10737418240,
				Status:      SnapshotStatusOk,
				Pool:        "kube",
				CreatedTime: 1535024379,
				SrcVolName:  fmt.Sprintf("vol-%d", vol),
			}
			snapArr = append(snapArr, tmp)
		}
	}
	cache := SnapshotCacheType{}
	cache.New()
	// Add
	for _, v := range snapArr {
		// Normal add
		tmp := v
		flag := cache.Add(&tmp)
		if flag != true {
			t.Errorf("add snapshot cache expect [%t], but actually [%t]", true, flag)
		}
		// Re-add
		flag = cache.Add(&tmp)
		if flag != true {
			t.Errorf("re-add snapshot cache expect [%t], but actually [%t]", true, flag)
		}
		// Re-add but incompatible
		v.SizeByte = v.SizeByte - 1
		flag = cache.Add(&v)
		if flag != false {
			t.Errorf("re-add snapshot cache expect [%t], but actually [%t]", false, flag)
		}
	}

	// Find
	for _, v := range snapArr {
		snapInfo := cache.Find(v.Name)
		if !reflect.DeepEqual(v, *snapInfo) {
			t.Errorf("find snapshot cache expect [%v], but actually [%v]", v, *snapInfo)
		}
	}
	// Delete
	for _, v := range snapArr {
		// normal delete
		cache.Delete(v.Name)
		// re-delete
		cache.Delete(v.Name)
		// find
		snapInfo := cache.Find(v.Name)
		if snapInfo != nil {
			t.Errorf("find snapshot cache expect [%v], but actually [%v]", nil, snapInfo)
		}
	}
}

func TestSnapshotCache_Sync(t *testing.T) {
	Pools = []string{SnapTestPoolCsi}
	cache := SnapshotCacheType{}
	cache.New()
	err := cache.Sync()
	if err != nil {
		t.Errorf("sync snapshot cache expect [%v], but actually [%v]", nil, err)
	}
	snapList := cache.List()
	t.Logf("list snapshot cache count [%d]", len(snapList))
}
