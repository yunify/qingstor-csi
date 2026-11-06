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
	"fmt"
	"strings"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

// VolumeInUseError reports that a volume is opened by a client of another node
// and must not be attached here. The CSI service maps it to
// codes.FailedPrecondition, every other NodeCheckVolumeExclusive error is an
// infrastructure failure and is mapped to codes.Internal.
type VolumeInUseError struct {
	VolumeID string
	Clients  []string
}

func (e *VolumeInUseError) Error() string {
	return fmt.Sprintf("volume %s is already opened by %s, refuse to attach it again",
		e.VolumeID, strings.Join(e.Clients, ","))
}

// ClonePendingError reports that an existing volume is still being copied by a
// center side clone job. The CSI service maps it to codes.Aborted so that the
// provisioner retries later instead of binding a partially copied volume.
type ClonePendingError struct {
	VolumeID string
	Status   string
	Percent  int
	Age      time.Duration
}

func (e *ClonePendingError) Error() string {
	return fmt.Sprintf("volume %s is still being cloned (status %s, percent %d%%, started %s ago), retry later",
		e.VolumeID, e.Status, e.Percent, e.Age.Round(time.Second))
}

// CloneFailedError reports that the clone of an existing volume failed. Cleaned
// is true when the failed volume was removed from the center: the next
// CreateVolume call then starts from a clean state and can heal by itself.
type CloneFailedError struct {
	VolumeID string
	Reason   string
	Cleaned  bool
}

func (e *CloneFailedError) Error() string {
	if e.Cleaned {
		return fmt.Sprintf("clone of volume %s failed (%s), the failed volume was removed: the next retry starts a new clone",
			e.VolumeID, e.Reason)
	}
	return fmt.Sprintf("clone of volume %s failed (%s), automatic cleanup failed: manual action required, detach the clone relation of %s and delete it in the center",
		e.VolumeID, e.Reason, e.VolumeID)
}

type ControllerOperator interface {
	CreateVolume(volumeName string, requestSize int64, parameters map[string]string) (volumeID string, err error)
	CreateVolumeFromSnapshot(volumeName, snapshotID string, requestSize int64, parameters map[string]string) (volumeID string, err error)
	CreateVolumeByClone(volumeName, sourceVolumeID string, requestSize int64, parameters map[string]string) (volumeID string, err error)
	FindVolumeByName(volumeName string, parameters map[string]string) (*csi.Volume, error)

	FindVolume(volumeID string) (*csi.Volume, error)
	DeleteVolume(volumeID string) error
	ResizeVolume(volumeID string, requestSize int64) error

	CreateSnapshot(volumeID, snapshotName string) error
	DeleteSnapshot(snapshotID string) error
	FindSnapshot(snapshotID string) (*csi.Snapshot, error)
	FindSnapshotByName(volumeID, snapshotName string) (*csi.Snapshot, error)

	// CloneReady checks whether an existing volume created from contentSource
	// can be handed back to the caller. It returns nil when the volume may be
	// reported as success, a *ClonePendingError while the clone is still
	// running, and a *CloneFailedError when the clone failed.
	CloneReady(volumeID string, contentSource *csi.VolumeContentSource) error
}

type NodeOperator interface {
	// NodeAttachVolume attaches the volume to this node and returns the device
	// path. knownDevice is the device already observed by the caller, it is
	// reused when non empty so a retried NodeStageVolume does not list all
	// attached volumes of the node again.
	NodeAttachVolume(volumeID, knownDevice string) (string, error)
	NodeDetachVolume(volumeID string) error
	NodeGetDevice(volumeID string) (device string, err error)
	// NodeCheckVolumeExclusive returns an error when the volume is opened by a
	// client on another node, so that a ReadWriteOnce volume can never be
	// attached by two nodes at the same time.
	NodeCheckVolumeExclusive(volumeID string) error
}

type Provider interface {
	ControllerOperator
	NodeOperator
	// HealthCheck reports whether the driver backend (qbd and at least one
	// center) is usable. It is called by Probe.
	HealthCheck() error
}
