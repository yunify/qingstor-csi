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
	"errors"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/yunify/qingstor-csi/pkg/storage/neonsan/api"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	"time"
)

var (
	maxRetryCnt      = 10
	retryBackOff     = wait.Backoff{
		Duration: time.Second,
		Factor:   1.5,
		Steps:    20,
		Cap:      time.Minute * 10,
	}
)

func (v *neonsan) CreateVolume(volName string, requestSize int64, replicas int) (string, error) {
	_, err := api.CreateVolume(v.confFile, v.poolName, volName, requestSize, replicas)
	if err != nil {
		return "", err
	}
	return volName, nil
}

func (v *neonsan) DeleteVolume(volId string) (err error) {
	_, err = api.DeleteVolume(v.confFile, v.poolName, volId)
	return err
}

func (v *neonsan) FindVolume(volId string) (*csi.Volume, error) {
	return v.FindVolumeByName(volId)
}

func (v *neonsan) FindVolumeByName(volName string) (*csi.Volume, error) {
	vol, err := api.ListVolume(v.confFile, v.poolName, volName)
	if err != nil {
		return nil, err
	}
	if vol == nil {
		return nil, nil
	}
	return &csi.Volume{
		VolumeId:      vol.Name,
		CapacityBytes: int64(vol.Size),
	}, nil
}

func (v *neonsan) ResizeVolume(volId string, requestSize int64) (err error) {
	return api.ResizeVolume(v.confFile, v.poolName, volId, requestSize)
}

func (v *neonsan) CloneVolume(volName string, srcVolId string) (volId string, err error) {
	srcVol, err := api.ListVolume(v.confFile, v.poolName, srcVolId)
	if err != nil {
		return "", err
	}

	if srcVol == nil {
		return "", errors.New("source volume not exist")
	}

	_, err = api.CreateVolume(v.confFile, v.poolName, volName, int64(srcVol.Size), srcVol.ReplicationCount)
	if err != nil {
		return "", err
	}

	err = api.CloneVolume(v.confFile, srcVolId, v.poolName, volName, v.poolName)
	if err != nil {
		return "", err
	}

	retryCnt := 0
	retryFun := func(e error) bool {
		retryCnt++
		return e != nil && retryCnt <= maxRetryCnt
	}
	//Wait until clone SYNCED
	err = retry.OnError(retryBackOff, retryFun, func() error {
		cloneInfo, listCloneErr := api.ListClone(v.confFile, srcVolId, v.poolName, volName, v.poolName)
		if listCloneErr != nil {
			return err
		}
		if cloneInfo.Status != "SYNCED" {
			return errors.New("clone not synced")
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	err = api.DetachCloneRelationship(v.confFile, srcVolId, v.poolName, volName, v.poolName)
	if err != nil {
		return "", err
	}

	return volName, nil
}
