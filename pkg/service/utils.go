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

package service

import (
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/yunify/qingstor-csi/pkg/common"
	"io/ioutil"
	"k8s.io/klog"
	"k8s.io/utils/mount"
	"os"
	"strings"
)

func GetInstanceIdFromFile(filepath string) (instanceId string, err error) {
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	instanceId = string(bytes[:])
	instanceId = strings.Replace(instanceId, "\n", "", -1)
	klog.Infof("Getting instance-id: \"%s\"", instanceId)
	return instanceId, nil
}

func NewControllerServiceCapability(cap csi.ControllerServiceCapability_RPC_Type) *csi.ControllerServiceCapability {
	return &csi.ControllerServiceCapability{
		Type: &csi.ControllerServiceCapability_Rpc{
			Rpc: &csi.ControllerServiceCapability_RPC{
				Type: cap,
			},
		},
	}
}

func NewNodeServiceCapability(cap csi.NodeServiceCapability_RPC_Type) *csi.NodeServiceCapability {
	return &csi.NodeServiceCapability{
		Type: &csi.NodeServiceCapability_Rpc{
			Rpc: &csi.NodeServiceCapability_RPC{
				Type: cap,
			},
		},
	}
}

// Required Volume Size
func GetRequiredVolumeSizeByte(capRange *csi.CapacityRange) (int64, error) {
	if capRange == nil {
		return common.Gib, nil
	}
	res := int64(0)
	if capRange.GetRequiredBytes() > 0 {
		res = capRange.GetRequiredBytes()
	}
	if capRange.GetLimitBytes() > 0 && res > capRange.GetLimitBytes() {
		return -1, fmt.Errorf("volume required bytes %d greater than limit bytes %d", res, capRange.GetLimitBytes())
	}
	return res, nil
}

func createTargetMountPath(mounter mount.Interface, mountPath string, isBlock bool) (bool, error) {
	// Check if that mount path exists properly
	notMnt, err := mount.IsNotMountPoint(mounter, mountPath)
	if err != nil {
		if os.IsNotExist(err) {
			if isBlock {
				//#nosec
				pathFile, e := os.OpenFile(mountPath, os.O_CREATE|os.O_RDWR, 0750)
				if e != nil {
					klog.Errorf("Failed to create mountPath:%s with error: %v", mountPath, e)
					return notMnt, e
				}
				if err = pathFile.Close(); err != nil {
					klog.Errorf("Failed to close mountPath:%s with error: %v", mountPath, err)
					return notMnt, err
				}
			} else {
				// Create a directory
				if err = os.MkdirAll(mountPath, 0750); err != nil {
					return notMnt, err
				}
			}
			notMnt = true
		} else {
			return false, err
		}
	}
	return notMnt, err
}

