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

import "github.com/yunify/qingstor-csi/pkg/storage/neonsan/qbd"

func (v *neonsan) NodeAttachVolume(volId string) error {
	return qbd.AttachVolume(v.confFile, v.poolName, volId)
}

func (v *neonsan) NodeDetachVolume(volId string) error {
	return qbd.DetachVolume(v.confFile, v.poolName, volId)
}

func (v *neonsan) NodeGetDevice(volId string) (string, error) {
	attachInfo, err := qbd.ListVolume(v.confFile, v.poolName, volId)
	if err != nil {
		return "", err
	}
	if attachInfo != nil {
		return attachInfo.Device, nil
	}
	return "", nil
}
