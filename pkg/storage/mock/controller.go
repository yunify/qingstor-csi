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
	"github.com/yunify/qingstor-csi/pkg/common"
	"time"
)

var (
	errorNotImplement = errors.New("method not implement")
	errorNotToCalled  = errors.New("method should not to be called")
)

//requestSize G
func (p *mockStorageProvider) CreateVolume(volName string, requestSize int64, replicas int) (volId string, err error) {
	vol, err := p.FindVolumeByName(volName)
	if err != nil{
		return "",err
	}
	if vol != nil{
		return "", errors.New("volume exist")
	}

	volId = common.GenerateHashInEightBytes(time.Now().UTC().String())
	vol = &csi.Volume{
		CapacityBytes: requestSize,
		VolumeId:volId,
	}
	p.volumes[volName] = vol
	return volId,nil
}

func (p *mockStorageProvider) DeleteVolume(volId string) (err error) {
	vol,err := p.FindVolume(volId)
	if vol == nil{
		return errors.New("delete not exist volume")
	}
	delete(p.volumes, volId)
	return nil
}

func (p *mockStorageProvider) FindVolume(volId string) (*csi.Volume, error) {
	for _,vol := range p.volumes {
		if vol.VolumeId == volId{
			return vol,nil
		}
	}
	return nil,nil
}

func (p *mockStorageProvider) FindVolumeByName(volName string) (*csi.Volume, error) {
	return p.volumes[volName],nil
}

func (p *mockStorageProvider) AttachVolume(volId string, instanceId string) (err error) {
	return errorNotToCalled
}

func (p *mockStorageProvider) DetachVolume(volId string, instanceId string) (err error) {
	return errorNotToCalled
}

func (p *mockStorageProvider) ResizeVolume(volId string, requestSize int64) (err error) {
	v, err := p.FindVolume(volId)
	if err != nil {
		return err
	}
	if v == nil{
		return errors.New("not found")
	}
	v.CapacityBytes = requestSize
	return nil
}

func (p *mockStorageProvider) CloneVolume(volName string, volType int, srcVolId string, zone string) (volId string, err error) {
	return "", errorNotImplement
}
