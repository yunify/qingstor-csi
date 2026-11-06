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
	"net"

	"github.com/yunify/qingstor-csi/pkg/storage"
	"github.com/yunify/qingstor-csi/pkg/storage/neonsan/api"
	"github.com/yunify/qingstor-csi/pkg/storage/neonsan/qbd"
)

func (v *neonsan) NodeAttachVolume(volumeID, knownDevice string) (string, error) {
	if len(knownDevice) != 0 {
		return knownDevice, nil
	}
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

// NodeCheckVolumeExclusive ensures the volume is not opened by a client of
// another node. NeonSAN itself allows one volume to be opened by several
// clients, so without this check a ReadWriteOnce volume could be attached by
// two nodes at the same time and both sides would write to the same data.
// Connections opened by this node are ignored: a retried NodeStageVolume
// (kubelet retry, driver restart) must keep working on the volume this node has
// already attached, and a leftover connection of this node (node reboot, qbd
// crash) must not block the volume forever - qbd decides whether it can be
// attached again.
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
	local, err := localIPs()
	if err != nil {
		return fmt.Errorf("list local addresses failed: %v", err)
	}
	clients := remoteClients(connections, local)
	if len(clients) == 0 {
		return nil
	}
	return &storage.VolumeInUseError{VolumeID: volumeID, Clients: clients}
}

// remoteClients returns the distinct remote addresses of the connections which
// were not opened by a client of this node.
func remoteClients(connections []api.VolumeConnInfo, local map[string]bool) []string {
	clients := make([]string, 0, len(connections))
	seen := make(map[string]bool, len(connections))
	for _, connection := range connections {
		addr := connection.RemoteAddr
		if addr == "" || seen[addr] {
			continue
		}
		seen[addr] = true
		if local[connectionIP(addr)] {
			continue
		}
		clients = append(clients, addr)
	}
	return clients
}

// connectionIP returns the IP part of an "ip:port" remote address, the address
// is returned unchanged when it carries no port.
func connectionIP(addr string) string {
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return host
	}
	return addr
}

// localIPs returns the IP addresses of this node, a connection of this node is
// identified by the address it comes from.
func localIPs() (map[string]bool, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	local := make(map[string]bool, len(addrs))
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		local[ipNet.IP.String()] = true
	}
	return local, nil
}
