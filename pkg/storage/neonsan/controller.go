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
	"fmt"
	"strconv"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/protobuf/ptypes"
	"github.com/yunify/qingstor-csi/pkg/storage"
	"github.com/yunify/qingstor-csi/pkg/storage/neonsan/api"
	"k8s.io/klog"
)

var (
	// cloneWaitPerCall bounds how long one CreateVolume call watches a clone.
	// The copy runs as a center side job and is not interrupted, the CSI call
	// only polls its status. The budget plus one detach (15s) stays below the
	// 90s timeout of the deployed csi-provisioner: a longer copy is resumed by
	// the next retry, the center job keeps running in between.
	cloneWaitPerCall  = time.Second * 60
	clonePollInterval = time.Second * 2
	// cloneTempGrace is how long a temporary volume without a clone record is
	// trusted to belong to a concurrent CreateVolume which has not issued
	// clone_volume yet. Older volumes are leftovers of a failed attempt.
	cloneTempGrace = time.Second * 10
	// rollbackDeleteRetries is the number of attempts used to detach the clone
	// relation and delete the target volume of a failed clone.
	rollbackDeleteRetries = 3
)

const (
	cloneStatusInit       = "INIT"
	cloneStatusSyncing    = "SYNCING"
	cloneStatusSynced     = "SYNCED"
	cloneStatusSyncFailed = "SYNC_FAILED"
	// cloneTempSuffix keeps a volume under copy away from its final name:
	// FindVolume only succeeds after the rename, so a crash between create and
	// SYNCED can never bind a half copied volume to a PVC. The name is stable
	// so a retried CreateVolume resumes the running center side copy instead of
	// starting a new one. The suffix carries a dot so that a volume a user
	// created by hand is extremely unlikely to collide with it: a collision
	// would make the resume path delete a foreign volume.
	cloneTempSuffix = ".csi-cloning"
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

func (v *neonsan) CreateVolumeFromSnapshot(volumeName, snapshotID string, requestSize int64, parameters map[string]string) (string, error) {
	targetPoolName := GetPoolName(parameters)
	sourcePoolName, sourceVolumeName, snapshotName := SplitSnapshotName(snapshotID)
	// Link clone keeps a data dependency on the source snapshot/volume, only do
	// it when the storage class asks for it explicitly. Default is a full clone.
	err := v.cloneVolume(sourcePoolName, sourceVolumeName, snapshotName, targetPoolName, volumeName, requestSize, parameters["link_clone"] == "true")
	if err != nil {
		return "", err
	}
	return JoinVolumeName(targetPoolName, volumeName), nil
}

func (v *neonsan) CreateVolumeByClone(volumeName, sourceVolumeID string, requestSize int64, parameters map[string]string) (string, error) {
	targetPoolName := GetPoolName(parameters)
	sourcePoolName, sourceVolumeName := SplitVolumeName(sourceVolumeID)

	linkClone := parameters["link_clone"] == "true"
	var snapshotName string
	if linkClone {
		snapshotName = fmt.Sprintf("csi-link-clone-%d", time.Now().UnixNano())
		err := api.CreateSnapshot(v.confFile, sourcePoolName, sourceVolumeName, snapshotName)
		if err != nil {
			return "", err
		}
	}

	err := v.cloneVolume(sourcePoolName, sourceVolumeName, snapshotName, targetPoolName, volumeName, requestSize, linkClone)
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
	if poolName == "" || volumeName == "" {
		return fmt.Errorf("invalid volume id [%s]", volumeID)
	}
	if v.volumeArchive {
		// The archived volume keeps its data, so detach without deleting the target.
		// Renaming a volume which is still the target of a clone relation would
		// leave the relation pointing at a volume the center no longer knows
		// under its original name.
		if detachErr := v.detachCloneTarget(poolName, volumeName); detachErr != nil {
			return fmt.Errorf("cannot detach the clone relation of volume %s before archiving it: %w", volumeID, detachErr)
		}
		return api.RenameVolume(v.confFile, poolName, volumeName, volumeName+"_archive_"+time.Now().Format("20060102150405")+"_.img")
	}
	return v.deleteVolume(poolName, volumeName)
}

// deleteVolume asks the center to delete the volume. It intentionally does not
// detach clone relations: a volume which is still the source or target of a
// clone is left for an operator to inspect and detach manually, so deleting a
// user volume never changes a relation of another volume.
func (v *neonsan) deleteVolume(poolName, volumeName string) error {
	volumeID := JoinVolumeName(poolName, volumeName)
	err := api.DeleteVolume(v.confFile, poolName, volumeName)
	if err == nil || api.IsRetCode(err, api.RetCodeVolumeNotExists) {
		return nil
	}
	if api.IsRetCode(err, api.RetCodeVolumeHasClone) {
		return fmt.Errorf("volume %s cannot be deleted: a clone relationship is still attached; detach it in the center and retry the delete: %w", volumeID, err)
	}
	return err
}

// RemoveCloneLeftover deletes the temporary volume an earlier clone attempt
// left behind. The clone source disappeared, so the copy cannot be completed
// and the temporary volume is unreachable: the final volume name is only
// created by the rename at the end of a successful copy.
func (v *neonsan) RemoveCloneLeftover(volumeName string, parameters map[string]string) error {
	poolName := GetPoolName(parameters)
	if poolName == "" {
		return fmt.Errorf("the pool of volume %s is unknown, cannot clean its clone leftovers", volumeName)
	}
	tempName := volumeName + cloneTempSuffix
	tempVolume, err := api.ListVolume(v.confFile, poolName, tempName)
	if err != nil {
		return err
	}
	if tempVolume == nil {
		return nil
	}
	// A copy which is still running owns the volume: the center refuses to
	// detach it (-112) and to delete it (-114). Leave the volume to the next
	// CreateVolume retry, which runs this cleanup again after the job reached
	// SYNCED or SYNC_FAILED.
	relations, err := api.ListCloneByTarget(v.confFile, poolName, tempName)
	if api.IsRetCode(err, api.RetCodeVolumeNotExists) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, relation := range relations {
		if relation.Status != cloneStatusSynced && relation.Status != cloneStatusSyncFailed {
			klog.V(4).Infof("Temporary volume %s/%s is still being copied (status %s), keeping it for the next retry", poolName, tempName, relation.Status)
			return nil
		}
	}
	klog.Infof("Removing the temporary volume %s/%s left behind by an interrupted clone of %s", poolName, tempName, volumeName)
	if !v.rollbackCloneVolume(poolName, tempName) {
		return fmt.Errorf("cannot remove the temporary volume %s/%s, manual action required", poolName, tempName)
	}
	return nil
}

// detachCloneTarget detaches every clone relation in which the volume is the
// target. A volume without relation is a no-op (-111). The archive path uses it
// because the archived volume keeps its data and must not remain the target of
// a clone relation. The delete path deliberately does not call it.
func (v *neonsan) detachCloneTarget(poolName, volumeName string) error {
	relations, err := api.ListCloneByTarget(v.confFile, poolName, volumeName)
	if api.IsRetCode(err, api.RetCodeVolumeNotExists) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(relations) == 0 {
		return nil
	}
	err = api.DetachCloneRelationshipByTarget(v.confFile, poolName, volumeName)
	if err != nil && !api.IsRetCode(err, api.RetCodeCloneNotFound) {
		return err
	}
	return nil
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
		// The center answers -102 when the volume of the snapshot is gone
		// instead of returning an empty list. The snapshot is gone with it,
		// so report "not found" and keep DeleteSnapshot idempotent.
		if api.IsRetCode(err, api.RetCodeVolumeNotExists) {
			return nil, nil
		}
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
		SourceVolumeId: JoinVolumeName(poolName, volumeName),
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
	err := api.DeleteSnapshot(v.confFile, poolName, volumeName, snapshotName)
	if api.IsRetCode(err, api.RetCodeVolumeNotExists) {
		// The snapshot disappeared together with its volume between the check
		// and the delete. DeleteSnapshot must be idempotent, report success so
		// external-snapshotter can drop the VolumeSnapshotContent.
		return nil
	}
	return err
}

func (v *neonsan) cloneVolume(sourcePoolName, sourceVolumeName, snapshotName, targetPoolName, targetVolumeName string, requestSize int64, linkClone bool) error {
	if linkClone {
		return v.linkCloneVolume(sourcePoolName, sourceVolumeName, snapshotName, targetPoolName, targetVolumeName, requestSize)
	}
	return v.fullCloneVolume(sourcePoolName, sourceVolumeName, snapshotName, targetPoolName, targetVolumeName, requestSize)
}

func (v *neonsan) linkCloneVolume(sourcePoolName, sourceVolumeName, snapshotName, targetPoolName, targetVolumeName string, requestSize int64) error {
	if snapshotName == "" {
		return errors.New("link clone require a snapshot")
	}
	err := api.CloneVolume(v.confFile, sourcePoolName, sourceVolumeName, snapshotName, targetVolumeName, targetPoolName, true, requestSize)
	if err != nil {
		return err
	}
	return nil
}

func (v *neonsan) fullCloneVolume(sourcePoolName, sourceVolumeName, snapshotName, targetPoolName, targetVolumeName string, requestSize int64) error {
	volumeForClone, err := api.GetVolumeForClone(v.confFile, sourcePoolName, sourceVolumeName)
	if err != nil {
		return err
	}
	if volumeForClone == nil {
		return errors.New("source volume not exist")
	}
	effectiveSize := requestSize
	if int64(volumeForClone.Size) > effectiveSize {
		effectiveSize = int64(volumeForClone.Size)
	}
	parameters := map[string]string{
		"rep_count":  strconv.Itoa(volumeForClone.ReplicationCount),
		"thick_prov": "false",
		"pool_name":  targetPoolName,
		"rg":         volumeForClone.Rg,
		"encrypte":   volumeForClone.Encrypte,
		"key_name":   volumeForClone.KeyName,
	}
	// The volume is created under a stable temporary name and renamed once the
	// copy is SYNCED: the rename is the commit point. FindVolume never sees a
	// half copied volume, so a crash or a timeout can never bind empty data to
	// a PVC, and a retry resumes the copy which is already running.
	tempName := targetVolumeName + cloneTempSuffix
	tempVolume, err := api.ListVolume(v.confFile, targetPoolName, tempName)
	if err != nil {
		return err
	}
	needCreate := true
	if tempVolume != nil {
		resumable, resumeErr := v.resumeFullClone(sourcePoolName, sourceVolumeName, snapshotName, targetPoolName, tempName, tempVolume)
		if resumeErr != nil {
			return resumeErr
		}
		needCreate = !resumable
	}
	if needCreate {
		if err = api.CreateVolume(v.confFile, tempName, effectiveSize, parameters); err != nil {
			return err
		}
		if err = v.cloneAndWait(sourcePoolName, sourceVolumeName, snapshotName, targetPoolName, tempName, effectiveSize); err != nil {
			var pending *storage.ClonePendingError
			var resume *storage.CloneResumeError
			if errors.As(err, &pending) || errors.As(err, &resume) {
				// The copy is still running or already complete, the next call
				// resumes it instead of starting over.
				return err
			}
			return v.cloneFailed(targetPoolName, targetVolumeName, tempName, err)
		}
	}
	if err = v.verifyCloneSource(sourcePoolName, sourceVolumeName, snapshotName); err != nil {
		var resume *storage.CloneResumeError
		if errors.As(err, &resume) {
			return err
		}
		return v.cloneFailed(targetPoolName, targetVolumeName, tempName, err)
	}
	// The copy is complete, so the rename is the commit point: the volume becomes
	// visible under its final name and the clone relation follows the rename.
	// Detaching first would lose the copied data when the rename fails or the
	// driver dies in between: the temporary volume has no clone record then, so
	// the next CreateVolume would delete it and copy again. Both steps keep the
	// copied volume on failure and are simply repeated by the next call.
	if err = api.RenameVolume(v.confFile, targetPoolName, tempName, targetVolumeName); err != nil {
		return &storage.CloneResumeError{Err: fmt.Errorf("rename_volume: %w", err)}
	}
	if err = api.DetachCloneRelationship(v.confFile, sourcePoolName, sourceVolumeName, targetPoolName, targetVolumeName); err != nil && !api.IsRetCode(err, api.RetCodeCloneNotFound) {
		return &storage.CloneResumeError{Err: fmt.Errorf("detach_clone_relationship: %w", err)}
	}
	return nil
}

// resumeFullClone looks after the temporary volume of an earlier attempt. It
// returns true when the volume is ready to be renamed to its final name, false
// when the volume was incomplete (it is deleted then).
func (v *neonsan) resumeFullClone(sourcePoolName, sourceVolumeName, snapshotName, targetPoolName, tempName string, tempVolume *api.Volume) (bool, error) {
	cloneInfo, err := api.ListClone220(v.confFile, sourcePoolName, sourceVolumeName, targetPoolName, tempName)
	if err != nil {
		// The state of the copy cannot be read (center failover). The volume
		// may still be written by the center job, so it must be neither
		// deleted nor reported as ready. Keep the caller retrying, the same
		// way CloneReady and cloneAndWait do.
		return false, &storage.ClonePendingError{
			VolumeID: JoinVolumeName(targetPoolName, tempName),
			Status:   "UNKNOWN",
			Reason:   err.Error(),
		}
	}
	if cloneInfo == nil {
		// A concurrent CreateVolume may have created this volume and not issued
		// clone_volume yet, never delete a fresh volume.
		if !tempVolume.CreateTime.IsZero() && time.Since(tempVolume.CreateTime) < cloneTempGrace {
			return false, &storage.ClonePendingError{
				VolumeID: JoinVolumeName(targetPoolName, tempName),
				Status:   "CREATING",
				Age:      time.Since(tempVolume.CreateTime),
			}
		}
		if !v.rollbackCloneVolume(targetPoolName, tempName) {
			return false, fmt.Errorf("cannot delete the incomplete volume %s/%s, manual action required", targetPoolName, tempName)
		}
		return false, nil
	}
	switch cloneInfo.Status {
	case cloneStatusSynced:
		if err := v.verifyCloneSource(sourcePoolName, sourceVolumeName, snapshotName); err != nil {
			var resume *storage.CloneResumeError
			if errors.As(err, &resume) {
				return false, err
			}
			if !v.rollbackCloneVolume(targetPoolName, tempName) {
				return false, fmt.Errorf("the copy of a deleted source %s/%s cannot be removed from %s/%s, manual action required", sourcePoolName, sourceVolumeName, targetPoolName, tempName)
			}
			return false, err
		}
		// The copy is complete: the volume is renamed first, the relation is
		// detached afterwards by fullCloneVolume.
		return true, nil
	case cloneStatusSyncFailed:
		if !v.rollbackCloneVolume(targetPoolName, tempName) {
			return false, fmt.Errorf("cannot clean the failed clone volume %s/%s, manual action required", targetPoolName, tempName)
		}
		return false, nil
	default:
		age := time.Duration(0)
		if !cloneInfo.StatusTime.IsZero() {
			age = time.Since(cloneInfo.StatusTime)
		}
		return false, &storage.ClonePendingError{
			VolumeID: JoinVolumeName(targetPoolName, tempName),
			Status:   cloneInfo.Status,
			Percent:  cloneInfo.Percent,
			Age:      age,
		}
	}
}

// cloneFailed rolls the temporary volume of a failed clone back and explains
// in the error whether the cleanup succeeded, so the operator knows whether the
// next retry is safe or a manual action is required.
// verifyCloneSource makes sure the data a finished copy is based on still
// exists. Deleting the source snapshot or volume while the copy runs lets the
// center report SYNCED for a target which was never filled, and binding such a
// volume would silently lose the data. A read error is reported as retryable:
// it says nothing about the source.
func (v *neonsan) verifyCloneSource(sourcePoolName, sourceVolumeName, snapshotName string) error {
	if snapshotName != "" {
		snapshot, err := api.ListSnapshot(v.confFile, sourcePoolName, sourceVolumeName, snapshotName)
		if err != nil {
			return &storage.CloneResumeError{Err: fmt.Errorf("list_snapshot: %w", err)}
		}
		if snapshot == nil {
			return fmt.Errorf("the source snapshot %s/%s@%s was deleted while it was copied", sourcePoolName, sourceVolumeName, snapshotName)
		}
		return nil
	}
	sourceVolume, err := api.ListVolume(v.confFile, sourcePoolName, sourceVolumeName)
	if err != nil {
		return &storage.CloneResumeError{Err: fmt.Errorf("list_volume: %w", err)}
	}
	if sourceVolume == nil {
		return fmt.Errorf("the source volume %s/%s was deleted while it was copied", sourcePoolName, sourceVolumeName)
	}
	return nil
}
func (v *neonsan) cloneFailed(poolName, targetVolumeName, tempName string, cause error) error {
	if !v.rollbackCloneVolume(poolName, tempName) {
		return fmt.Errorf("full clone of %s/%s failed: %w; cleanup of the partial volume %s failed, manual action required: detach its clone relation and delete %s/%s in the center",
			poolName, targetVolumeName, cause, tempName, poolName, tempName)
	}
	return fmt.Errorf("full clone of %s/%s failed: %w; the partial volume %s was removed, the next retry starts from a clean state",
		poolName, targetVolumeName, cause, tempName)
}

// cloneAndWait starts a center side full copy and waits for it at most
// cloneWaitPerCall. The copy itself keeps running when the wait expires, the
// caller (csi-provisioner) retries and the next call resumes from the clone
// status. A SYNC_FAILED clone is reported at once, it never becomes SYNCED by
// itself.
func (v *neonsan) cloneAndWait(sourcePoolName, sourceVolumeName, snapshotName, targetPoolName, targetVolumeName string, size int64) error {
	if err := api.CloneVolume(v.confFile, sourcePoolName, sourceVolumeName, snapshotName, targetVolumeName, targetPoolName, false, size); err != nil {
		return fmt.Errorf("clone_volume: %w", err)
	}
	start := time.Now()
	deadline := start.Add(cloneWaitPerCall)
	listFailures := 0
	var lastListErr error
	for {
		cloneInfo, err := api.ListClone220(v.confFile, sourcePoolName, sourceVolumeName, targetPoolName, targetVolumeName)
		if err != nil {
			// A center failover must not abort a running copy, and it must not
			// roll the volume back either: the copy still writes it, so the
			// delete would fail and the volume would be reported as needing a
			// manual cleanup. Report it as pending and keep the volume.
			listFailures++
			lastListErr = err
			if time.Now().After(deadline) {
				klog.Warningf("Clone %s/%s: list_clone failed %d times within %s, keeping the volume and resuming on the next CreateVolume, last error: %v",
					targetPoolName, targetVolumeName, listFailures, cloneWaitPerCall, lastListErr)
				return &storage.ClonePendingError{
					VolumeID: JoinVolumeName(targetPoolName, targetVolumeName),
					Status:   "UNKNOWN",
					Age:      time.Since(start),
					Reason:   lastListErr.Error(),
				}
			}
			time.Sleep(clonePollInterval)
			continue
		}
		if listFailures > 0 {
			klog.Infof("Clone %s/%s: list_clone recovered after %d failures, last error: %v",
				targetPoolName, targetVolumeName, listFailures, lastListErr)
			listFailures = 0
		}
		if cloneInfo == nil {
			if time.Now().After(deadline) {
				return errors.New("list_clone: clone record not found")
			}
			time.Sleep(clonePollInterval)
			continue
		}
		switch cloneInfo.Status {
		case cloneStatusSynced:
			// The copy is complete. Renaming and detaching are done by the
			// caller, a stale relation never makes the copy unusable.
			return nil
		case cloneStatusSyncFailed:
			return fmt.Errorf("clone finished with status %s", cloneStatusSyncFailed)
		}
		if time.Now().After(deadline) {
			age := time.Duration(0)
			if !cloneInfo.StatusTime.IsZero() {
				age = time.Since(cloneInfo.StatusTime)
			}
			return &storage.ClonePendingError{
				VolumeID: JoinVolumeName(targetPoolName, targetVolumeName),
				Status:   cloneInfo.Status,
				Percent:  cloneInfo.Percent,
				Age:      age,
			}
		}
		time.Sleep(clonePollInterval)
	}
}

// rollbackCloneVolume deletes the partial volume of a failed clone. The center
// refuses to delete a volume whose clone relation is attached (-114), so the
// relation is detached between the attempts. It returns true when the volume is
// gone, either deleted or reported as not existing.
func (v *neonsan) rollbackCloneVolume(poolName, volumeName string) bool {
	var lastErr error
	for attempt := 0; attempt < rollbackDeleteRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
		err := api.DeleteVolume(v.confFile, poolName, volumeName)
		if err == nil || api.IsRetCode(err, api.RetCodeVolumeNotExists) {
			return true
		}
		lastErr = err
		if !api.IsRetCode(err, api.RetCodeVolumeHasClone) {
			continue
		}
		if err = api.DetachCloneRelationshipByTarget(v.confFile, poolName, volumeName); err != nil && !api.IsRetCode(err, api.RetCodeCloneNotFound) {
			lastErr = err
		}
	}
	klog.Errorf("rollback clone: delete volume %s/%s failed after %d attempts: %v", poolName, volumeName, rollbackDeleteRetries, lastErr)
	return false
}

// CloneReady tells whether an already existing target volume can be returned to
// the caller. A running clone is never ready: its data is incomplete, so the
// caller has to retry later instead of binding a half copied volume to a PVC.
func (v *neonsan) CloneReady(volumeID string, contentSource *csi.VolumeContentSource) error {
	targetPoolName, targetVolumeName := SplitVolumeName(volumeID)
	if targetPoolName == "" || targetVolumeName == "" {
		return fmt.Errorf("invalid volume id [%s]", volumeID)
	}
	sourcePoolName, sourceVolumeName, err := cloneSource(contentSource)
	if err != nil {
		return err
	}
	cloneInfo, err := api.ListClone220(v.confFile, sourcePoolName, sourceVolumeName, targetPoolName, targetVolumeName)
	if err != nil {
		// The state of the clone cannot be read (center failover): the volume
		// must not be reported as ready, because it may still be half copied.
		// Keep the caller retrying instead of raising a driver failure.
		return &storage.ClonePendingError{
			VolumeID: volumeID,
			Status:   "UNKNOWN",
			Reason:   err.Error(),
		}
	}
	if cloneInfo == nil {
		return nil
	}
	switch cloneInfo.Status {
	case cloneStatusSynced:
		if err = api.DetachCloneRelationship(v.confFile, sourcePoolName, sourceVolumeName, targetPoolName, targetVolumeName); err != nil && !api.IsRetCode(err, api.RetCodeCloneNotFound) {
			// The copy is complete, only the metadata is left.
			return &storage.CloneResumeError{Err: fmt.Errorf("detach_clone_relationship: %w", err)}
		}
		return nil
	case cloneStatusSyncFailed:
		if !v.rollbackCloneVolume(targetPoolName, targetVolumeName) {
			return &storage.CloneFailedError{VolumeID: volumeID, Reason: cloneInfo.Status, Cleaned: false}
		}
		return &storage.CloneFailedError{VolumeID: volumeID, Reason: cloneInfo.Status, Cleaned: true}
	default:
		age := time.Duration(0)
		if !cloneInfo.StatusTime.IsZero() {
			age = time.Since(cloneInfo.StatusTime)
		}
		return &storage.ClonePendingError{
			VolumeID: volumeID,
			Status:   cloneInfo.Status,
			Percent:  cloneInfo.Percent,
			Age:      age,
		}
	}
}

// cloneSource returns the source volume of a CreateVolume request which carries
// a content source, the snapshot parent volume is returned for snapshot
// sources because the clone relationship is tracked per volume.
func cloneSource(contentSource *csi.VolumeContentSource) (poolName, volumeName string, err error) {
	switch contentSource.GetType().(type) {
	case *csi.VolumeContentSource_Snapshot:
		poolName, volumeName, _ = SplitSnapshotName(contentSource.GetSnapshot().GetSnapshotId())
	case *csi.VolumeContentSource_Volume:
		poolName, volumeName = SplitVolumeName(contentSource.GetVolume().GetVolumeId())
	}
	if poolName == "" || volumeName == "" {
		return "", "", fmt.Errorf("invalid content source [%v]", contentSource)
	}
	return poolName, volumeName, nil
}
