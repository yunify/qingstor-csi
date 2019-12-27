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

package mock

import (
	"errors"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/protobuf/ptypes"
	"github.com/yunify/qingstor-csi/pkg/common"
)

func (p *mockStorageProvider) CreateVolume(volumeName string, requestSize int64, replicas int) error {
	vol, _ := p.ListVolume(volumeName)
	if vol != nil {
		return errors.New("volume exist")
	}
	vol = &csi.Volume{
		CapacityBytes: requestSize,
		VolumeId:      volumeName,
	}
	p.volumes[volumeName] = vol
	return nil
}

func (p *mockStorageProvider) DeleteVolume(volumeName string) error {
	vol, _ := p.ListVolume(volumeName)
	if vol == nil {
		return errors.New("delete not exist volume")
	}
	delete(p.volumes, volumeName)
	return nil
}

func (p *mockStorageProvider) ListVolume(volumeName string) (*csi.Volume, error) {
	return p.volumes[volumeName], nil
}

func (p *mockStorageProvider) ResizeVolume(volId string, requestSize int64) (err error) {
	v, _ := p.ListVolume(volId)
	if v == nil {
		return errors.New("not found")
	}
	v.CapacityBytes = requestSize
	return nil
}

func (p *mockStorageProvider) CloneVolume(sourceVolName, snapshotName, targetVolName string) error {
	srcVol, err := p.ListVolume(sourceVolName)
	if err != nil {
		return err
	}
	if srcVol == nil {
		return errors.New("src vol not exist")
	}
	return p.CreateVolume(targetVolName, srcVol.CapacityBytes, 0)
}

func (p *mockStorageProvider) CreateSnapshot(volumeName, snapshotName string) error {
	ptime := ptypes.TimestampNow()
	snap := &csi.Snapshot{
		SizeBytes:      0,
		SnapshotId:     common.JoinSnapshotName(volumeName,snapshotName),
		SourceVolumeId: volumeName,
		CreationTime:   ptime,
		ReadyToUse:     true,
	}
	p.snapshots[common.JoinSnapshotName(volumeName, snapshotName)] = snap
	return  nil
}

func (p *mockStorageProvider) ListSnapshot(volumeName, snapshotName string) (*csi.Snapshot, error) {
	if snap, ok := p.snapshots[common.JoinSnapshotName(volumeName,snapshotName)]; ok {
		return snap, nil
	}
	return nil, nil
}

func (p *mockStorageProvider) DeleteSnapshot(volumeName, snapshotName string) error {
	delete(p.snapshots, common.JoinSnapshotName(volumeName, snapshotName))
	return nil
}
