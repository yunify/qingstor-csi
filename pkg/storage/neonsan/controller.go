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
	"strconv"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/protobuf/ptypes"
	"github.com/yunify/qingstor-csi/pkg/common"
	"github.com/yunify/qingstor-csi/pkg/storage/neonsan/api"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
)

var (
	retryBackOff = wait.Backoff{
		Duration: time.Second,
		Factor:   1.5,
		Steps:    20,
		Cap:      time.Minute * 10,
	}
)

func (v *neonsan) CreateVolume(volumeName string, requestSize int64, parameters map[string]string) (string, error) {
	TuneUpParameter(parameters)
	err := api.CreateVolume(v.confFile, volumeName, requestSize, parameters)
	if err != nil {
		return "", err
	}
	poolName := GetPoolName(parameters)
	return JoinVolumeName(poolName, volumeName), nil
}

func (v *neonsan) CreateVolumeFromSnapshot(volumeName, snapshotID string, parameters map[string]string) (string, error) {
	targetPoolName := GetPoolName(parameters)
	sourcePoolName, sourceVolumeName, snapshotName := SplitSnapshotName(snapshotID)
	err := v.cloneVolume(sourcePoolName, sourceVolumeName, snapshotName, targetPoolName, volumeName)
	if err != nil {
		return "", err
	}
	return JoinVolumeName(targetPoolName, volumeName), nil
}

func (v *neonsan) CreateVolumeByClone(volumeName, sourceVolumeID string, parameters map[string]string) (string, error) {
	targetPoolName := GetPoolName(parameters)
	sourcePoolName, sourceVolumeName := SplitVolumeName(sourceVolumeID)
	err := v.cloneVolume(sourcePoolName, sourceVolumeName, "", targetPoolName, volumeName)
	if err != nil {
		return "", err
	}
	return JoinVolumeName(targetPoolName, volumeName), nil
}

func (v *neonsan) FindVolumeByName(volumeName string, parameters map[string]string) (*csi.Volume, error) {
	volumeID := JoinVolumeName(GetPoolName(parameters), volumeName)
	return v.FindVolume(volumeID)
}

func (v *neonsan) FindVolume(volumeID string) (*csi.Volume, error) {
	poolName, volumeName := SplitVolumeName(volumeID)
	vol, err := api.ListVolume(v.confFile, poolName, volumeName)
	if err != nil {
		return nil, err
	}
	if vol == nil {
		return nil, nil
	}
	return &csi.Volume{
		VolumeId:      volumeID,
		CapacityBytes: int64(vol.Size),
	}, nil
}

func (v *neonsan) DeleteVolume(volumeID string) (err error) {
	poolName, volumeName := SplitVolumeName(volumeID)
	if v.volumeArchive {
		return api.RenameVolume(v.confFile, poolName, volumeName, volumeName+"_archive_"+time.Now().Format("20060102150405")+"_.img")
	}
	return api.DeleteVolume(v.confFile, poolName, volumeName)
}

func (v *neonsan) ResizeVolume(volumeID string, requestSize int64) error {
	poolName, volumeName := SplitVolumeName(volumeID)
	return api.ResizeVolume(v.confFile, poolName, volumeName, requestSize)
}

func (v *neonsan) CreateSnapshot(volumeID, snapshotName string) error {
	poolName, volumeName := SplitVolumeName(volumeID)
	return api.CreateSnapshot(v.confFile, poolName, volumeName, snapshotName)
}

func (v *neonsan) FindSnapshot(snapshotID string) (*csi.Snapshot, error) {
	poolName, volumeName, snapshotName := SplitSnapshotName(snapshotID)
	snapshot, err := api.ListSnapshot(v.confFile, poolName, volumeName, snapshotName)
	if err != nil {
		return nil, err
	}
	if snapshot == nil {
		return nil, nil
	}
	creationTime, err := ptypes.TimestampProto(snapshot.CreateTime)
	if err != nil {
		return nil, err
	}
	csiSnapshot := &csi.Snapshot{
		SizeBytes:      snapshot.SnapshotSize,
		SnapshotId:     JoinSnapshotName(poolName, volumeName, snapshot.SnapshotName),
		SourceVolumeId: volumeName,
		CreationTime:   creationTime,
		ReadyToUse:     snapshot.Status == "OK",
	}
	return csiSnapshot, nil
}

func (v *neonsan) FindSnapshotByName(volumeID, snapshotName string) (*csi.Snapshot, error) {
	poolName, volumeName := SplitVolumeName(volumeID)
	snapshotID := JoinSnapshotName(poolName, volumeName, snapshotName)
	return v.FindSnapshot(snapshotID)
}

func (v *neonsan) DeleteSnapshot(snapshotID string) error {
	poolName, volumeName, snapshotName := SplitSnapshotName(snapshotID)
	return api.DeleteSnapshot(v.confFile, poolName, volumeName, snapshotName)
}

func (v *neonsan) cloneVolume(sourcePoolName, sourceVolumeName, snapshotName, targetPoolName, targetVolumeName string) error {
	volumeForClone, err := api.GetVolumeForClone(v.confFile, sourcePoolName, sourceVolumeName)
	if err != nil {
		return err
	}
	if volumeForClone == nil {
		return errors.New("source volume not exist")
	}
	parameters := map[string]string{
		"rep_count":  strconv.Itoa(volumeForClone.ReplicationCount),
		"thick_prov": "false",
		"pool_name":  targetPoolName,
		"rg":         volumeForClone.Rg,
		"encrypte":   volumeForClone.Encrypte,
		"key_name":   volumeForClone.KeyName,
	}
	err = api.CreateVolume(v.confFile, targetVolumeName, int64(volumeForClone.Size), parameters)
	if err != nil {
		return err
	}
	err = api.CloneVolume(v.confFile, sourcePoolName, sourceVolumeName, snapshotName, targetVolumeName, targetPoolName)
	if err != nil {
		return err
	}
	//Wait until clone SYNCED
	err = retry.OnError(retryBackOff, common.DefaultRetryErrorFunc, func() error {
		cloneInfo, listCloneErr := api.ListClone220(v.confFile, sourcePoolName, sourceVolumeName, targetPoolName, targetVolumeName)
		if listCloneErr != nil {
			return listCloneErr
		}
		if cloneInfo.Status != "SYNCED" {
			return errors.New("clone not synced")
		}
		return nil
	})
	if err != nil {
		return err
	}
	return api.DetachCloneRelationship(v.confFile, sourcePoolName, sourceVolumeName, targetPoolName, targetVolumeName)
}
