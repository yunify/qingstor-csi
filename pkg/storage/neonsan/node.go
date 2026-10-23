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

package neonsan

import (
	"fmt"
	"strings"

	"github.com/yunify/qingstor-csi/pkg/storage/neonsan/api"
	"github.com/yunify/qingstor-csi/pkg/storage/neonsan/qbd"
)

func (v *neonsan) NodeAttachVolume(volumeID string) error {
	poolName, volumeName := SplitVolumeName(volumeID)
	return qbd.AttachVolume(v.confFile, v.protocol, poolName, volumeName)
}

func (v *neonsan) NodeDetachVolume(volumeID string) error {
	poolName, volumeName := SplitVolumeName(volumeID)
	return qbd.DetachVolume(v.confFile, v.protocol, poolName, volumeName)
}

func (v *neonsan) NodeIsVolumeAttached(volumeID string) (bool, error) {
	return true, nil
}

func (v *neonsan) NodeGetDevice(volumeID string) (string, error) {
	poolName, volumeName := SplitVolumeName(volumeID)
	return qbd.GetDevice(v.confFile, poolName, volumeName)
}

// NodeCheckVolumeExclusive ensures the volume is not opened by any client.
// NeonSAN itself allows one volume to be opened by several clients, so without
// this check a ReadWriteOnce volume could be attached by two nodes at the same
// time and both sides would write to the same data. An already attached volume
// (including the one attached by this node) is treated as busy: qbd refuses to
// open it twice anyway, so the caller returns the error to the CO.
func (v *neonsan) NodeCheckVolumeExclusive(volumeID string) error {
	poolName, volumeName := SplitVolumeName(volumeID)
	if poolName == "" || volumeName == "" {
		return fmt.Errorf("invalid volume id [%s]", volumeID)
	}
	connections, err := api.ListVolumeConns(v.confFile, poolName, volumeName)
	if err != nil {
		return fmt.Errorf("list connections of volume %s failed: %v", volumeID, err)
	}
	if len(connections) == 0 {
		return nil
	}
	clients := make([]string, 0, len(connections))
	seen := make(map[string]bool, len(connections))
	for _, connection := range connections {
		if connection.RemoteAddr == "" || seen[connection.RemoteAddr] {
			continue
		}
		seen[connection.RemoteAddr] = true
		clients = append(clients, connection.RemoteAddr)
	}
	return fmt.Errorf("volume %s is already opened by %s, refuse to attach it again",
		volumeID, strings.Join(clients, ","))
}
