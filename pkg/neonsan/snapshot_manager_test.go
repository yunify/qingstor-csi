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

package neonsan

import (
	"errors"
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"reflect"
	"testing"
)

const (
	SnapTestSnapshotName     = "test"
	SnapTestFakeSnapshotName = "fake"
	SnapTestPoolName         = "csi"
	SnapTestFakePoolName     = "fake"
	SnapTestVolumeName       = "foo"
	SnapTestFakeVolumeName   = "fake"
	SnapTestVolumeNameNoSnap = "nosnap"
)

var cache []*snapshotInfo = []*snapshotInfo{
	&snapshotInfo{
		snapName:         "vol1-snap1",
		snapID:           "399616507911",
		sizeByte:         10737418240,
		status:           SnapshotStatusOk,
		pool:             "kube",
		createdTime:      1535024379,
		sourceVolumeName: "vol1",
	},
	&snapshotInfo{
		snapName:         "vol1-snap2",
		snapID:           "399616507912",
		sizeByte:         10737418240,
		status:           SnapshotStatusOk,
		pool:             "kube",
		createdTime:      1535024379,
		sourceVolumeName: "vol1",
	},
	&snapshotInfo{
		snapName:         "vol2-snap1",
		snapID:           "399616507921",
		sizeByte:         10737418240,
		status:           SnapshotStatusOk,
		pool:             "kube",
		createdTime:      1535024379,
		sourceVolumeName: "vol1",
	},
	&snapshotInfo{
		snapName:         "vol2-snap2",
		snapID:           "399616507922",
		sizeByte:         10737418240,
		status:           SnapshotStatusOk,
		pool:             "kube",
		createdTime:      1535024379,
		sourceVolumeName: "vol1",
	},
}

func TestSnapshotPrepare(t *testing.T) {
	if _, err := CreateVolume(SnapTestVolumeName, SnapTestPoolName, gib, 1); err != nil {
		t.Errorf("Failed to create volume [%s], error [%v]", SnapTestVolumeName, err)
	}
	if _, err := CreateVolume(SnapTestVolumeNameNoSnap, SnapTestPoolName, gib, 1); err != nil {
		t.Errorf("Failed to create volume [%s], error [%v]", SnapTestVolumeNameNoSnap, err)
	}
}

func TestSnapshotCheck(t *testing.T) {
	if vol, err := FindVolume(SnapTestVolumeName, SnapTestPoolName); err != nil || vol == nil {
		t.Errorf("Not found volume [%s], error [%v]", SnapTestVolumeName, err)
	}
	if vol, err := FindVolume(SnapTestVolumeNameNoSnap, SnapTestPoolName); err != nil || vol == nil {
		t.Errorf("Not found volume [%s], error [%v]", SnapTestVolumeNameNoSnap, err)
	}
}

func TestSnapshotCleaner(t *testing.T) {
	if err := DeleteVolume(SnapTestVolumeName, SnapTestPoolName); err != nil {
		t.Errorf("Failed to delete volume [%s], error [%v]", SnapTestVolumeName, err)
	}
	if err := DeleteVolume(SnapTestVolumeNameNoSnap, SnapTestPoolName); err != nil {
		t.Errorf("Failed to delete volume [%s], error [%v]", SnapTestVolumeNameNoSnap, err)
	}
}

func TestCreateSnapshot(t *testing.T) {
	tests := []struct {
		name   string
		input  *snapshotInfo
		output *snapshotInfo
		err    error
	}{
		{
			name: "succeed to create snapshot",
			input: &snapshotInfo{
				snapName:         SnapTestSnapshotName,
				pool:             SnapTestPoolName,
				sourceVolumeName: SnapTestVolumeName,
			},
			output: &snapshotInfo{
				snapName:         SnapTestSnapshotName,
				pool:             SnapTestPoolName,
				sourceVolumeName: SnapTestVolumeName,
			},
			err: nil,
		},
		{
			name: "recreate snapshot",
			input: &snapshotInfo{
				snapName:         SnapTestSnapshotName,
				pool:             SnapTestPoolName,
				sourceVolumeName: SnapTestVolumeName,
			},
			output: nil,
			err:    errors.New("raise error"),
		},
		{
			name: "fake volume name",
			input: &snapshotInfo{
				snapName:         SnapTestSnapshotName,
				pool:             SnapTestPoolName,
				sourceVolumeName: SnapTestFakeVolumeName,
			},
			output: nil,
			err:    errors.New("raise error"),
		},
		{
			name: "fake pool name",
			input: &snapshotInfo{
				snapName:         SnapTestSnapshotName,
				pool:             SnapTestFakePoolName,
				sourceVolumeName: SnapTestVolumeName,
			},
			output: nil,
			err:    errors.New("raise error"),
		},
	}
	for _, v := range tests {
		snapInfo, err := CreateSnapshot(v.input.snapName, v.input.sourceVolumeName, v.input.pool)
		if (v.err != nil && err == nil) || (v.err == nil && err != nil) {
			t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.err, err)
		}

		if v.err == nil && err == nil {
			if v.output != nil && snapInfo != nil {
				if v.output.snapName != snapInfo.snapName {
					t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.output, snapInfo)
				}
			} else {
				t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.output, snapInfo)
			}
		}
	}
}

func TestFindSnapshot(t *testing.T) {
	tests := []struct {
		name   string
		input  *snapshotInfo
		output *snapshotInfo
		err    error
	}{
		{
			name: "succeed to find snapshot",
			input: &snapshotInfo{
				snapName:         SnapTestSnapshotName,
				pool:             SnapTestPoolName,
				sourceVolumeName: SnapTestVolumeName,
			},
			output: &snapshotInfo{
				snapName:         SnapTestSnapshotName,
				pool:             SnapTestPoolName,
				sourceVolumeName: SnapTestVolumeName,
			},
			err: nil,
		},
		{
			name: "volume doesn't contains any snapshot",
			input: &snapshotInfo{
				snapName:         SnapTestSnapshotName,
				pool:             SnapTestPoolName,
				sourceVolumeName: SnapTestVolumeNameNoSnap,
			},
			output: nil,
			err:    nil,
		},
		{
			name: "fake snapshot name",
			input: &snapshotInfo{
				snapName:         SnapTestFakeSnapshotName,
				pool:             SnapTestPoolName,
				sourceVolumeName: SnapTestVolumeName,
			},
			output: nil,
			err:    nil,
		},
		{
			name: "fake volume name",
			input: &snapshotInfo{
				snapName:         SnapTestSnapshotName,
				pool:             SnapTestPoolName,
				sourceVolumeName: SnapTestFakeVolumeName,
			},
			output: nil,
			err:    errors.New("raise error"),
		},
		{
			name: "fake pool name",
			input: &snapshotInfo{
				snapName:         SnapTestSnapshotName,
				pool:             SnapTestFakePoolName,
				sourceVolumeName: SnapTestVolumeName,
			},
			output: nil,
			err:    errors.New("raise error"),
		},
	}
	for _, v := range tests {
		snapInfo, err := FindSnapshot(v.input.snapName, v.input.sourceVolumeName, v.input.pool)
		if (v.err != nil && err == nil) || (v.err == nil && err != nil) {
			t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.err, err)
		}
		if v.err == nil && err == nil {
			if v.output != nil && snapInfo != nil {
				if v.output.snapName != snapInfo.snapName {
					t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.output, snapInfo)
				}
			} else if (v.output != nil && snapInfo == nil) || (v.output == nil && snapInfo != nil) {
				t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.output, snapInfo)
			}
		}
	}
}

func TestFindSnapshotWithoutPool(t *testing.T) {
	tests := []struct {
		name   string
		input  *snapshotInfo
		output *snapshotInfo
		err    error
	}{
		{
			name: "succeed to find snapshot",
			input: &snapshotInfo{
				snapName: SnapTestSnapshotName,
			},
			output: &snapshotInfo{
				snapName:         SnapTestSnapshotName,
				pool:             SnapTestPoolName,
				sourceVolumeName: SnapTestVolumeName,
			},
			err: nil,
		},
		{
			name: "not found snapshot",
			input: &snapshotInfo{
				snapName: SnapTestFakeSnapshotName,
			},
			output: nil,
			err:    nil,
		},
	}
	for _, v := range tests {
		snapInfo, err := FindSnapshotWithoutPool(v.input.snapName)
		if (v.err != nil && err == nil) || (v.err == nil && err != nil) {
			t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.err, err)
		}
		if v.err == nil && err == nil {
			if v.output != nil && snapInfo != nil {
				if v.output.snapName != snapInfo.snapName || v.output.pool != snapInfo.pool {
					t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.output, snapInfo)
				}
			} else if (v.output != nil && snapInfo == nil) || (v.output == nil && snapInfo != nil) {
				t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.output, snapInfo)
			}
		}
	}
}

func TestListSnapshotByVolume(t *testing.T) {
	tests := []struct {
		name   string
		input  *snapshotInfo
		output []*snapshotInfo
		err    error
	}{
		{
			name: "succeed to find snapshot",
			input: &snapshotInfo{
				pool:             SnapTestPoolName,
				sourceVolumeName: SnapTestVolumeName,
			},
			output: []*snapshotInfo{
				{
					snapName:         SnapTestSnapshotName,
					pool:             SnapTestPoolName,
					sourceVolumeName: SnapTestVolumeName,
				},
			},
			err: nil,
		},
		{
			name: "find no snapshot",
			input: &snapshotInfo{
				pool:             SnapTestPoolName,
				sourceVolumeName: SnapTestVolumeNameNoSnap,
			},
			output: nil,
			err:    nil,
		},
		{
			name: "find fake pool",
			input: &snapshotInfo{
				pool:             SnapTestFakePoolName,
				sourceVolumeName: SnapTestVolumeName,
			},
			output: nil,
			err:    errors.New("raise error"),
		},
		{
			name: "find fake volume",
			input: &snapshotInfo{
				pool:             SnapTestPoolName,
				sourceVolumeName: SnapTestFakeVolumeName,
			},
			output: nil,
			err:    errors.New("raise error"),
		},
	}
	for _, v := range tests {
		snapList, err := ListSnapshotByVolume(v.input.sourceVolumeName, v.input.pool)
		if (v.err != nil && err == nil) || (v.err == nil && err != nil) {
			t.Errorf("name [%s]: error expect [%v], but actually [%v]", v.name, v.err, err)
		}
		if v.err == nil && err == nil {
			if len(v.output) != len(snapList) {
				t.Errorf("name [%s]: error expect [%d], but actually [%d]", v.name, len(v.output), len(snapList))
			} else {
				for i := range v.output {
					if v.output[i].snapName != snapList[i].snapName || v.output[i].pool != snapList[i].pool {
						t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.output, snapList)
						break
					}
				}
			}
		}
	}
}

func TestDeleteSnapshot(t *testing.T) {
	tests := []struct {
		name     string
		snapInfo *snapshotInfo
		err      error
	}{
		{
			name: "succeed to delete snapshot",
			snapInfo: &snapshotInfo{
				snapName:         SnapTestSnapshotName,
				pool:             SnapTestPoolName,
				sourceVolumeName: SnapTestVolumeName,
			},
			err: nil,
		},
		{
			name: "re-delete snapshot",
			snapInfo: &snapshotInfo{
				snapName:         SnapTestSnapshotName,
				pool:             SnapTestPoolName,
				sourceVolumeName: SnapTestFakeVolumeName,
			},
			err: errors.New("raise error"),
		},
		{
			name: "failed to delete snapshot",
			snapInfo: &snapshotInfo{
				snapName:         SnapTestSnapshotName,
				pool:             SnapTestPoolName,
				sourceVolumeName: SnapTestFakeVolumeName,
			},
			err: errors.New("raise error"),
		},
	}
	for _, v := range tests {
		err := DeleteSnapshot(v.snapInfo.snapName, v.snapInfo.sourceVolumeName, v.snapInfo.pool)
		if (v.err != nil && err == nil) || (v.err == nil && err != nil) {
			t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.err, err)
		}
	}
}

func TestConvertNeonToCsiSnap(t *testing.T) {
	tests := []struct {
		caseName string
		neonSnap *snapshotInfo
		csiSnap  *csi.Snapshot
	}{
		{
			caseName: "valid NeonSAN snapshot",
			neonSnap: &snapshotInfo{
				snapName:         SnapTestSnapshotName,
				snapID:           "25463",
				sizeByte:         2147483648,
				status:           SnapshotStatusOk,
				pool:             SnapTestPoolName,
				createdTime:      1535024379,
				sourceVolumeName: SnapTestVolumeName,
			},
			csiSnap: &csi.Snapshot{
				SizeBytes:      2147483648,
				Id:             SnapTestSnapshotName,
				SourceVolumeId: SnapTestVolumeName,
				CreatedAt:      1535024379,
				Status: &csi.SnapshotStatus{
					Type: csi.SnapshotStatus_READY,
				},
			},
		},
		{
			caseName: "without snap name",
			neonSnap: &snapshotInfo{
				snapID:           "25463",
				sizeByte:         2147483648,
				status:           SnapshotStatusOk,
				pool:             SnapTestPoolName,
				createdTime:      1535024379,
				sourceVolumeName: SnapTestVolumeName,
			},
			csiSnap: &csi.Snapshot{
				SizeBytes:      2147483648,
				SourceVolumeId: SnapTestVolumeName,
				CreatedAt:      1535024379,
				Status: &csi.SnapshotStatus{
					Type: csi.SnapshotStatus_READY,
				},
			},
		},
		{
			caseName: "zero value snap info",
			neonSnap: &snapshotInfo{},
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
	tests := []struct {
		caseName  string
		neonSnaps []*snapshotInfo
		respList  []*csi.ListSnapshotsResponse_Entry
	}{
		{
			caseName: "normal snapshot info array",
			neonSnaps: []*snapshotInfo{
				{
					snapName:         "snapshot1",
					snapID:           "25463",
					sizeByte:         2147483648,
					status:           SnapshotStatusOk,
					createdTime:      1535024299,
					sourceVolumeName: "volume1",
				},
				{
					snapName:         "snapshot2",
					snapID:           "25464",
					sizeByte:         2147483648,
					status:           SnapshotStatusOk,
					createdTime:      1535024379,
					sourceVolumeName: "volume2",
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
	exampleFullList := []*snapshotInfo{
		{
			snapName:         "snapshot1",
			snapID:           "25463",
			sizeByte:         2147483648,
			status:           SnapshotStatusOk,
			createdTime:      1535024299,
			sourceVolumeName: "volume1",
		},
		{
			snapName:         "snapshot2",
			snapID:           "25466",
			sizeByte:         2147483648,
			status:           SnapshotStatusOk,
			createdTime:      1535024266,
			sourceVolumeName: "volume1",
		},
		{
			snapName:         "snapshot3",
			snapID:           "25472",
			sizeByte:         2147483648,
			status:           SnapshotStatusOk,
			createdTime:      1535024272,
			sourceVolumeName: "volume1",
		},
		{
			snapName:         "snapshot2",
			snapID:           "25564",
			sizeByte:         2147485648,
			status:           SnapshotStatusOk,
			createdTime:      1535025379,
			sourceVolumeName: "volume2",
		},
		{
			snapName:         "snapshot3",
			snapID:           "25564",
			sizeByte:         2143285648,
			status:           SnapshotStatusOk,
			createdTime:      1535024379,
			sourceVolumeName: "volume2",
		},
		{
			snapName:         "snapshot1",
			snapID:           "25664",
			sizeByte:         2147485648,
			status:           SnapshotStatusOk,
			createdTime:      1535026379,
			sourceVolumeName: "volume3",
		},
		{
			snapName:         "snapshot4",
			snapID:           "25564",
			sizeByte:         2141285648,
			status:           SnapshotStatusOk,
			createdTime:      1535064379,
			sourceVolumeName: "volume1",
		},
	}
	tests := []struct {
		caseName    string
		fullList    []*snapshotInfo
		page        int
		itemPerPage int
		pageList    []*snapshotInfo
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
	var snapArr []snapshotInfo
	for vol := 0; vol < 10; vol++ {
		for snap := 0; snap < 20; snap++ {
			tmp := snapshotInfo{
				snapName:         fmt.Sprintf("vol-%d-snap-%d", vol, snap),
				snapID:           fmt.Sprintf("3996165079%d%d", vol, snap),
				sizeByte:         10737418240,
				status:           SnapshotStatusOk,
				pool:             "kube",
				createdTime:      1535024379,
				sourceVolumeName: fmt.Sprintf("vol-%d", vol),
			}
			snapArr = append(snapArr, tmp)
		}
	}
	cache := snapshotCache{}
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
		v.sizeByte = v.sizeByte - 1
		flag = cache.Add(&v)
		if flag != false {
			t.Errorf("re-add snapshot cache expect [%t], but actually [%t]", false, flag)
		}
	}

	// Find
	for _, v := range snapArr {
		snapInfo := cache.Find(v.snapName)
		if !reflect.DeepEqual(v, *snapInfo) {
			t.Errorf("find snapshot cache expect [%v], but actually [%v]", v, *snapInfo)
		}
	}
	// Delete
	for _, v := range snapArr {
		// normal delete
		cache.Delete(v.snapName)
		// re-delete
		cache.Delete(v.snapName)
		// find
		snapInfo := cache.Find(v.snapName)
		if snapInfo != nil {
			t.Errorf("find snapshot cache expect [%v], but actually [%v]", nil, snapInfo)
		}
	}
}

func TestSnapshotCache_Sync(t *testing.T) {
	pools := []string{"kube", "csi"}
	cache := snapshotCache{}
	cache.New()
	err := cache.Sync(pools)
	if err != nil {
		t.Errorf("sync snapshot cache expect [%v], but actually [%v]", nil, err)
	}
	snapList := cache.List()
	t.Logf("list snapshot cache count [%d]", len(snapList))
}
