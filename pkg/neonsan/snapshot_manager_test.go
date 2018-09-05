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
	"flag"
	"os"
	"testing"
	"errors"
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

func TestMain(m *testing.M) {
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "/tmp")
	flag.Set("v", "3")
	flag.Parse()
	ret := m.Run()
	os.Exit(ret)
}

func TestSnapshotPrepare(t *testing.T) {
	if _, err := CreateVolume(SnapTestVolumeName, SnapTestPoolName, gib, 1); err != nil{
		t.Errorf("Failed to create volume [%s], error [%v]", SnapTestVolumeName, err)
	}
	if _,err:= CreateVolume(SnapTestVolumeNameNoSnap, SnapTestPoolName, gib, 1); err != nil{
		t.Errorf("Failed to create volume [%s], error [%v]", SnapTestVolumeNameNoSnap, err)
	}
}

func TestSnapshotCheck(t *testing.T){
	if vol, err := FindVolume(SnapTestVolumeName, SnapTestPoolName); err != nil || vol == nil{
		t.Errorf("Not found volume [%s], error [%v]", SnapTestVolumeName, err)
	}
	if vol, err := FindVolume(SnapTestVolumeNameNoSnap, SnapTestPoolName); err != nil || vol == nil{
		t.Errorf("Not found volume [%s], error [%v]", SnapTestVolumeNameNoSnap, err)
	}
}

func TestSnapshotCleaner(t *testing.T){
	if  err := DeleteVolume(SnapTestVolumeName, SnapTestPoolName); err != nil{
		t.Errorf("Failed to delete volume [%s], error [%v]", SnapTestVolumeName, err)
	}
	if err := DeleteVolume(SnapTestVolumeNameNoSnap, SnapTestPoolName); err != nil{
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
