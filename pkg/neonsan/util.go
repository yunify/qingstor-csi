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
	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/golang/glog"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	Int64Max        int64  = int64(^uint64(0) >> 1)
	PluginFolder    string = "/var/lib/kubelet/plugins/"
	DefaultPoolName string = "kube"
	TimeLayout      string = "2006-01-02T15:04:05+08:00"
)

const (
	kib    int64 = 1024
	mib    int64 = kib * 1024
	gib    int64 = mib * 1024
	gib100 int64 = gib * 100
	tib    int64 = gib * 1024
	tib100 int64 = tib * 100
)

const (
	FileSystemExt3    string = "ext3"
	FileSystemExt4    string = "ext4"
	FileSystemXfs     string = "xfs"
	FileSystemDefault string = FileSystemExt4
)

const (
	ProtocolRDMA    string = "RDMA"
	ProtocolTCP     string = "TCP"
	ProtocolDefault string = ProtocolRDMA
)

var (
	ConfigFilePath  string = "/etc/neonsan/qbd.conf"
	TempSnapshotDir string = "/tmp"
)

// ExecCommand
// Return cases:	normal output,	nil:	normal output
//					error logs,		error:	command execute error
func ExecCommand(command string, args []string) ([]byte, error) {
	glog.Infof("execCommand: command = \"%s\", args = \"%v\"", command, args)
	time.Sleep(time.Second)
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("code [%s]: message [%s]", err.Error(), output)
	}
	return output, nil
}

// ContainsVolumeCapability
// Does Array of VolumeCapability_AccessMode contain the volume capability of subCaps
func ContainsVolumeCapability(accessModes []*csi.VolumeCapability_AccessMode, subCaps *csi.VolumeCapability) bool {
	for _, cap := range accessModes {
		if cap.GetMode() == subCaps.GetAccessMode().GetMode() {
			return true
		}
	}
	return false
}

// ContainsVolumeCapabilities
// Does array of VolumeCapability_AccessMode contain volume capabilities of subCaps
func ContainsVolumeCapabilities(accessModes []*csi.VolumeCapability_AccessMode, subCaps []*csi.VolumeCapability) bool {
	for _, v := range subCaps {
		if !ContainsVolumeCapability(accessModes, v) {
			return false
		}
	}
	return true
}

// ContainsNodeServiceCapability
// Does array of NodeServiceCapability contain node service capability of subCap
func ContainsNodeServiceCapability(nodeCaps []*csi.NodeServiceCapability, subCap csi.NodeServiceCapability_RPC_Type) bool {
	for _, v := range nodeCaps {
		if strings.Contains(v.String(), subCap.String()) {
			return true
		}
	}
	return false
}

// FormatVolumeSize convert volume size properly
func FormatVolumeSize(inputSize int64, step int64) int64 {
	if inputSize <= gib || step < 0 {
		return gib
	}
	remainder := inputSize % step
	if remainder != 0 {
		return inputSize - remainder + step
	}
	return inputSize
}

// Check file system type
// Support: ext3, ext4 and xfs
func IsValidFileSystemType(fs string) bool {
	switch fs {
	case FileSystemExt3:
		return true
	case FileSystemExt4:
		return true
	case FileSystemXfs:
		return true
	default:
		return false
	}
}

//	ParseIntToDec convert number string to decimal number string
func ParseIntToDec(hex string) (dec string) {
	i64, err := strconv.ParseInt(hex, 0, 64)
	if err != nil {
		return ""
	}
	return strconv.FormatInt(i64, 10)
}
