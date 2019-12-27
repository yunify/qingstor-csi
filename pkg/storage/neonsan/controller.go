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
	"github.com/yunify/qingstor-csi/pkg/common"
	"github.com/yunify/qingstor-csi/pkg/storage/neonsan/api"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	"time"
)

var (
	retryBackOff = wait.Backoff{
		Duration: time.Second,
		Factor:   1.5,
		Steps:    20,
		Cap:      time.Minute * 10,
	}
)

func (v *neonsan) CreateVolume(volumeName string, requestSize int64, replicas int) error {
	return api.CreateVolume(v.confFile, v.poolName, volumeName, requestSize, replicas)
}

func (v *neonsan) DeleteVolume(volumeName string) (err error) {
	_, err = api.DeleteVolume(v.confFile, v.poolName, volumeName)
	return err
}

func (v *neonsan) ListVolume(volumeName string) (*csi.Volume, error) {
	vol, err := api.ListVolume(v.confFile, v.poolName, volumeName)
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

func (v *neonsan) ResizeVolume(volumeName string, requestSize int64) (err error) {
	return api.ResizeVolume(v.confFile, v.poolName, volumeName, requestSize)
}

func (v *neonsan) CloneVolume(sourceVolumeName, snapshotName, targetVolumeName string) error {
	sourceVolume, err := api.ListVolume(v.confFile, v.poolName, sourceVolumeName)
	if err != nil {
		return err
	}
	if sourceVolume == nil {
		return errors.New("source volume not exist")
	}
	err = api.CreateVolume(v.confFile, v.poolName, targetVolumeName, int64(sourceVolume.Size), sourceVolume.ReplicationCount)
	if err != nil {
		return err
	}
	err = api.CloneVolume(v.confFile, sourceVolumeName, snapshotName, v.poolName, targetVolumeName, v.poolName)
	if err != nil {
		return err
	}
	//Wait until clone SYNCED
	err = retry.OnError(retryBackOff, common.DefaultRetryErrorFunc, func() error {
		cloneInfo, listCloneErr := api.ListClone(v.confFile, sourceVolumeName, v.poolName, targetVolumeName, v.poolName)
		if listCloneErr != nil {
			return err
		}
		if cloneInfo.Status != "SYNCED" {
			return errors.New("clone not synced")
		}
		return nil
	})
	if err != nil {
		return  err
	}
	return api.DetachCloneRelationship(v.confFile, sourceVolumeName, v.poolName, targetVolumeName, v.poolName)
}

func (v *neonsan) CreateSnapshot(volumeName, snapshotName string) error {
	return api.CreateSnapshot(v.confFile, volumeName, snapshotName, v.poolName)
}

func (v *neonsan) ListSnapshot(volumeName, snapshotName string) (*csi.Snapshot, error) {
	return api.ListSnapshot(v.confFile, volumeName, snapshotName, v.poolName)
}

func (v *neonsan) DeleteSnapshot(volumeName, snapshotName string) error {
	return api.DeleteSnapshot(v.confFile, volumeName, snapshotName, v.poolName)
}
