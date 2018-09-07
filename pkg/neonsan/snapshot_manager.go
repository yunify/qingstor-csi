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
	"errors"
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/golang/glog"
)

type snapshotInfo struct {
	// Neonsan's snapshot id and name in Kubernetes will correspond to
	// snapshot name in NeonSAN.
	snapName string

	// NeonSAN will auto generate this ID, snapshot ID will not be used in
	// Kubernetes.
	snapID   string
	sizeByte int64
	status   string
	pool     string

	// Timestamp when the point-in-time snapshot is taken on the storage
	// system. The format of this field should be a Unix nanoseconds time
	// encoded as an int64. On Unix, the command `date +%s%N` returns the
	// current time in nanoseconds since 1970-01-01 00:00:00 UTC. This
	// field is REQUIRED.
	createdTime      int64
	sourceVolumeName string
}

const (
	SnapshotStatusOk string = "OK"
)

var SnapshotStatusType = map[string]string{
	SnapshotStatusOk: SnapshotStatusOk,
}

// FindSnapshot gets snapshot information in specified pool
// srcVolName must be a valid volume name
// Return case:
//   snap, nil: succeed to find a snapshot
//   nil, nil: cannot find snapshot
//   nil, err: find snapshot error
func FindSnapshot(snapName, srcVolName, pool string) (outSnap *snapshotInfo, err error) {
	snapList, err := ListSnapshotByVolume(srcVolName, pool)
	if err != nil {
		glog.Errorf("List snapshot error: [%v]", err.Error())
		return nil, err
	}
	for i := range snapList {
		glog.Infof("snapList[%d]=[%v], %s,%s", i, snapList[i], snapList[i].snapName, snapName)
		if snapList[i].snapName == snapName {
			return snapList[i], nil
		}
	}
	return nil, nil
}

// FindSnapshotWithoutPool gets snapshot information in all pools
// CAUTION: the execution time is extremely long.
// Return case:
//   snap, nil: find a snapshot
//   nil, nil: cannot find snapshot
//   nil, err: find snapshot error or find duplicate snapshots
func FindSnapshotWithoutPool(snapName string) (outSnap *snapshotInfo, err error) {
	poolNames, err := ListPoolName()
	if err != nil {
		return nil, err
	}
	glog.Infof("pools [%v]", poolNames)
	for _, poolName := range poolNames {
		glog.Infof("pool [%s]", poolName)
		vols, err := ListVolumeByPool(poolName)
		if err != nil {
			return nil, err
		}
		glog.Infof("vols [%v]", vols)
		for _, volInfo := range vols {
			glog.Infof("vol [%s]", volInfo.name)
			snapInfo, err := FindSnapshot(snapName, volInfo.name, poolName)
			if err != nil || snapInfo != nil {
				return snapInfo, err
			}
		}
	}
	return nil, nil
}

// ListSnapshotByVolume lists snapshots by volume name and pool
// Return case:
//   snapshot list, nil: find snapshots in specific volume
//   nil, nil: find no snapshots in specific volume
//   nil, err: find snapshot error
func ListSnapshotByVolume(srcVolName, pool string) (snaps []*snapshotInfo, err error) {
	args := []string{"list_snapshot", "--volume", srcVolName, "--pool", pool, "-c", ConfigFilePath}
	output, err := ExecCommand(CmdNeonsan, args)
	if err != nil {
		glog.Errorf("Failed to find snapshot, args [%v].", args)
		return nil, err
	}
	snaps, err = ParseSnapshotList(string(output))
	if err != nil {
		return nil, err
	}
	for i := range snaps {
		snaps[i].pool = pool
		snaps[i].sourceVolumeName = srcVolName
	}
	return snaps, nil
}

// CreateSnapshot create snapshot
// Return case:
//   snap, nil: succeed to create the snapshot
//   nil, err: failed to create the snapshot
func CreateSnapshot(snapName, srcVolName, pool string) (outSnap *snapshotInfo, err error) {
	args := []string{"create_snapshot", "--snapshot", fmt.Sprintf("%s@%s", srcVolName, snapName), "--pool", pool,
		"-c", ConfigFilePath}
	_, err = ExecCommand(CmdNeonsan, args)
	if err != nil {
		return nil, err
	}
	snapInfo, err := FindSnapshot(snapName, srcVolName, pool)
	if err != nil {
		return nil, fmt.Errorf("CreateSnapshot: [%v]", err)
	}
	if snapInfo == nil {
		return nil, fmt.Errorf("CreateSnapshot error: cannot find snapshot [%s] after creating", snapName)
	}
	return snapInfo, nil
}

// DeleteSnapshot delete snapshot
// Return case:
//   nil: succeed to delete snapshot
//   err: failed to delete snapshot
func DeleteSnapshot(snapName, srcVolName, pool string) (err error) {
	args := []string{"delete_snapshot", "--snapshot", fmt.Sprintf("%s@%s", srcVolName, snapName), "--pool", pool,
		"-c", ConfigFilePath}
	_, err = ExecCommand(CmdNeonsan, args)
	if err != nil {
		glog.Errorf("Failed to delete snapshot, args [%v], error [%v].", args, err)
		return err
	}
	glog.Infof("Succeed to delete snapshot, args [%v].", args)
	return nil
}

// ConvertNeonsanToCsiSnap convert snapshot info to csi.Snapshot
func ConvertNeonToCsiSnap(neonSnap *snapshotInfo) (csiSnap *csi.Snapshot) {
	if neonSnap == nil {
		return nil
	}
	csiSnap = &csi.Snapshot{}
	csiSnap.SizeBytes = neonSnap.sizeByte
	csiSnap.Id = neonSnap.snapName
	csiSnap.SourceVolumeId = neonSnap.sourceVolumeName
	csiSnap.CreatedAt = neonSnap.createdTime
	if neonSnap.status == SnapshotStatusOk {
		csiSnap.Status = &csi.SnapshotStatus{
			Type: csi.SnapshotStatus_READY,
		}
	}
	return csiSnap
}

// ConvertNeonSnapToListSnapResp covert snapshot info array to csi.ListSnapshotsResponse_Entry array
func ConvertNeonSnapToListSnapResp(neonSnaps []*snapshotInfo) (respList []*csi.ListSnapshotsResponse_Entry) {
	for i := range neonSnaps {
		resp := &csi.ListSnapshotsResponse_Entry{
			Snapshot: ConvertNeonToCsiSnap(neonSnaps[i]),
		}
		respList = append(respList, resp)
	}
	return respList
}

// ReadListPage
// Parameters:
//   page: page number begins with 1.
func ReadListPage(fullList []*snapshotInfo, page int, itemPerPage int) (pageList []*snapshotInfo, err error) {
	if fullList == nil {
		return nil, nil
	}
	if page < 0 || itemPerPage <= 0 {
		return nil, errors.New("ReadListPage: input argument error")
	}
	// [headIndex, tailIndex)
	headIndex := itemPerPage * (page - 1)
	tailIndex := headIndex + itemPerPage
	totalLength := len(fullList)
	if totalLength < tailIndex {
		tailIndex = totalLength
	}
	if totalLength < headIndex {
		return nil, errors.New("ReadListPage: head index must not exceed list length")
	}
	return fullList[headIndex:tailIndex], nil
}
