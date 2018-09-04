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
	"fmt"
	"testing"
	"flag"
	"os"
)

const (
	SnapTestSnapshotName     = "test"
	SnapTestFakeSnapshotName = "fake"
	SnapTestPoolName         = "csi"
	SnapTestVolumeName       = "foo"
	SnapTestFakeVolumeName   = "fake"
)

func TestMain(m *testing.M) {
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "/tmp")
	flag.Set("v", "3")
	flag.Parse()
	ret := m.Run()
	os.Exit(ret)
}

func TestPreparation(t *testing.T){
	CreateVolume(SnapTestVolumeName, SnapTestPoolName, gib,1)
}

func TestCreateSnapshot(t *testing.T) {
	tests := []struct {
		name     string
		input *snapshotInfo
		output *snapshotInfo
		err      error
	}{
		{
			name: "Succeed to create snapshot",
			input: &snapshotInfo{
				snapName:         SnapTestSnapshotName,
				pool:             SnapTestPoolName,
				sourceVolumeName: SnapTestVolumeName,
			},
			output:&snapshotInfo{
				snapName:         SnapTestSnapshotName,
				pool:             SnapTestPoolName,
				sourceVolumeName: SnapTestVolumeName,
			},
			err: nil,
		},
		{
			name: "Recreate snapshot",
			input: &snapshotInfo{
				snapName:         SnapTestSnapshotName,
				pool:             SnapTestPoolName,
				sourceVolumeName: SnapTestVolumeName,
			},
			output: nil,
			err: fmt.Errorf("Raise error"),
		},
		{
			name: "Failed to create snapshot",
			input: &snapshotInfo{
				snapName:         SnapTestSnapshotName,
				pool:             SnapTestPoolName,
				sourceVolumeName: SnapTestFakeVolumeName,
			},
			output: nil,
			err: fmt.Errorf("Raise error"),
		},
	}
	for _, v := range tests {
		snapInfo, err := CreateSnapshot(v.input.snapName, v.input.sourceVolumeName, v.input.pool)
		if (v.err != nil && err == nil) || (v.err == nil && err != nil) {
			t.Errorf("name %s: error expect %v, but actually %v", v.name, v.err, err)
		}else if v.err == nil && err == nil{
			if v.output == nil && snapInfo == nil{
				t.Errorf("name %s: error expect %v, but actually %v", v.name, v.output, snapInfo)
			}else if v.output != nil && snapInfo != nil{
				if v.output.snapName != snapInfo.snapName{
					t.Errorf("name %s: error expect %v, but actually %v", v.name, v.output, snapInfo)
				}
			}else{
				t.Errorf("name %s: error expect %v, but actually %v", v.name, v.output, snapInfo)
			}
		}
	}
}

func TestFindSnapshot(t *testing.T) {
	tests := []struct {
		name     string
		input *snapshotInfo
		output *snapshotInfo
		err      error
	}{
		{
			name: "Succeed to find snapshot",
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
			name: "Not found snapshot",
			input: &snapshotInfo{
				snapName:         SnapTestSnapshotName,
				pool:             SnapTestPoolName,
				sourceVolumeName: SnapTestFakeVolumeName,
			},
			output: nil,
			err: fmt.Errorf("Not found"),
		},
	}
	for _, v:=range tests{
		snapInfo, err := FindSnapshot(v.input.snapName, v.input.sourceVolumeName, v.input.pool)
		if (v.err != nil && err == nil)||(v.err == nil && err != nil){
			t.Errorf("name %s: error expect %v, but actually %v", v.name, v.err, err)
		}
		if v.err == nil && err == nil{
			if v.output == nil && snapInfo == nil{
				t.Errorf("name %s: error expect %v, but actually %v", v.name, v.output, snapInfo)
			}else if v.output != nil && snapInfo != nil{
				if v.output.snapName != snapInfo.snapName{
					t.Errorf("name %s: error expect %v, but actually %v", v.name, v.output, snapInfo)
				}
			}else{
				t.Errorf("name %s: error expect %v, but actually %v", v.name, v.output, snapInfo)
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
			name: "Succeed to create snapshot",
			snapInfo: &snapshotInfo{
				snapName:         SnapTestSnapshotName,
				pool:             SnapTestPoolName,
				sourceVolumeName: SnapTestVolumeName,
			},
			err: nil,
		},
		{
			name: "Recreate snapshot",
			snapInfo: &snapshotInfo{
				snapName:         SnapTestSnapshotName,
				pool:             SnapTestPoolName,
				sourceVolumeName: SnapTestFakeVolumeName,
			},
			err: fmt.Errorf("Raise error"),
		},
		{
			name: "Failed to create snapshot",
			snapInfo: &snapshotInfo{
				snapName:         SnapTestSnapshotName,
				pool:             SnapTestPoolName,
				sourceVolumeName: SnapTestFakeVolumeName,
			},
			err: fmt.Errorf("Raise error"),
		},
	}
	for _, v := range tests {
		err := DeleteSnapshot(v.snapInfo.snapName, v.snapInfo.sourceVolumeName, v.snapInfo.pool)
		if (v.err != nil && err == nil) || (v.err == nil && err != nil) {
			t.Errorf("name %s: error expect %v, but actually %v", v.name, v.err, err)
		}
	}
}
