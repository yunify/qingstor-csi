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
)

func (p *mockStorageProvider) CreateVolume(volumeName string, requestSize int64, parameters map[string]string) (volumeID string, err error) {
	volume, _ := p.FindVolumeByName(volumeName, parameters)
	if volume != nil {
		return "", errors.New("volume exist")
	}
	volume = &csi.Volume{
		CapacityBytes: requestSize,
		VolumeId:      volumeName,
	}
	p.volumes[volumeName] = volume
	return volumeName, nil
}

func (p *mockStorageProvider) CreateVolumeFromSnapshot(volumeName, snapshotID string, parameters map[string]string) (volumeID string, err error) {
	snapshot, _ := p.FindSnapshot(snapshotID)
	if snapshot == nil {
		return "", errors.New("snapshot not exist")
	}
	sourceVolume, _ := p.FindVolume(snapshot.SourceVolumeId)
	if sourceVolume == nil {
		return "", errors.New("source volume not exist")
	}
	return p.CreateVolume(volumeName, sourceVolume.CapacityBytes, nil)
}

func (p *mockStorageProvider) CreateVolumeByClone(volumeName, sourceVolumeID string, parameters map[string]string) (volumeID string, err error) {
	srcVol, _ := p.FindVolume(sourceVolumeID)
	if srcVol == nil {
		return "", errors.New("source volume not exist")
	}
	return p.CreateVolume(volumeName, srcVol.CapacityBytes, nil)
}

func (p *mockStorageProvider) FindVolumeByName(volumeName string, parameters map[string]string) (*csi.Volume, error) {
	return p.FindVolume(volumeName)
}

func (p *mockStorageProvider) FindVolume(volumeID string) (*csi.Volume, error) {
	return p.volumes[volumeID], nil
}

func (p *mockStorageProvider) DeleteVolume(volumeID string) error {
	vol, _ := p.FindVolume(volumeID)
	if vol == nil {
		return errors.New("delete not exist volume")
	}
	delete(p.volumes, volumeID)
	return nil
}

func (p *mockStorageProvider) ResizeVolume(volumeID string, requestSize int64) error {
	v, _ := p.FindVolume(volumeID)
	if v == nil {
		return errors.New("not found")
	}
	v.CapacityBytes = requestSize
	return nil
}

func (p *mockStorageProvider) CreateSnapshot(volumeID, snapshotName string) error {
	ptime := ptypes.TimestampNow()
	snap := &csi.Snapshot{
		SizeBytes:      0,
		SnapshotId:     snapshotName,
		SourceVolumeId: volumeID,
		CreationTime:   ptime,
		ReadyToUse:     true,
	}
	p.snapshots[snapshotName] = snap
	return nil
}

func (p *mockStorageProvider) DeleteSnapshot(snapshotID string) error {
	delete(p.snapshots, snapshotID)
	return nil
}

func (p *mockStorageProvider) FindSnapshot(snapshotID string) (*csi.Snapshot, error) {
	if snap, ok := p.snapshots[snapshotID]; ok {
		return snap, nil
	}
	return nil, nil
}

func (p *mockStorageProvider) FindSnapshotByName(volumeID, snapshotName string) (*csi.Snapshot, error) {
	return p.FindSnapshot(snapshotName)
}
