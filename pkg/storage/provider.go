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

package storage

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
)

type ControllerOperator interface {
	CreateVolume(volumeName string, requestSize int64, parameters map[string]string) (volumeID string, err error)
	CreateVolumeFromSnapshot(volumeName, snapshotID string, parameters map[string]string) (volumeID string, err error)
	CreateVolumeByClone(volumeName, sourceVolumeID string, parameters map[string]string) (volumeID string, err error)
	FindVolumeByName(volumeName string, parameters map[string]string)(*csi.Volume, error)

	FindVolume(volumeID string) (*csi.Volume, error)
	DeleteVolume(volumeID string) error
	ResizeVolume(volumeID string, requestSize int64) error

	CreateSnapshot(volumeID, snapshotName string) error
	DeleteSnapshot(snapshotID string) error
	FindSnapshot(snapshotID string) (*csi.Snapshot, error)
	FindSnapshotByName(volumeID, snapshotName string)(*csi.Snapshot, error)
}

type NodeOperator interface {
	NodeAttachVolume(volumeID string) error
	NodeDetachVolume(volumeID string) error
	NodeGetDevice(volumeID string) (device string, err error)
}

type Provider interface {
	ControllerOperator
	NodeOperator
}
