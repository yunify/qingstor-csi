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

package manager

const (
	SnapshotStatusOk string = "OK"
	DefaultPoolName  string = "kube"
)

const (
	CmdQbd               string = "qbd"
	CmdNeonsan           string = "neonsan"
	VolumeStatusOk       string = "OK"
	VolumeStatusError    string = "ERROR"
	VolumeStatusDegraded string = "DEGRADED"
)

var SnapshotStatusType = map[string]string{
	SnapshotStatusOk: SnapshotStatusOk,
}

var Pools []string = []string{"kube"}

type VolumeInfo struct {
	Id       string
	Name     string
	SizeByte int64
	Status   string
	Replicas int
	Pool     string
}

// poolInfo: stats pool
// total, free, used: pool size in bytes
type PoolInfo struct {
	Id        string
	Name      string
	TotalByte int64
	FreeByte  int64
	UsedByte  int64
}

type SnapshotInfo struct {
	// Neonsan's snapshot id and name in Kubernetes will correspond to
	// snapshot name in NeonSAN.
	Name string

	// NeonSAN will auto generate this ID, snapshot ID will not be used in
	// Kubernetes.
	Id       string
	SizeByte int64
	Status   string
	Pool     string

	// Timestamp when the point-in-time snapshot is taken on the storage
	// system. The format of this field should be a Unix nanoseconds time
	// encoded as an int64. On Unix, the command `date +%s%N` returns the
	// current time in nanoseconds since 1970-01-01 00:00:00 UTC. This
	// field is REQUIRED.
	CreatedTime int64
	SrcVolName  string
}

type AttachInfo struct {
	Id        string
	Name      string
	Device    string
	Pool      string
	ReadBps   int64
	WriteBps  int64
	ReadIops  int64
	WriteIops int64
}

type NeonsanStorageClass struct {
	Replicas     int    `json:"replicas"`
	VolumeFsType string `json:"fsType"`
	Pool         string `json:"pool"`
	StepSize     int    `json:"stepSize"`
	Protocol     string `json:"protocol"`
}

type SnapshotCacheType struct {
	Snaps map[string]*SnapshotInfo
}

type SnapshotCache interface {
	// Add snapshot into map
	// 1. snapshot name does not exist, add snapshot information normally.
	// 2. snapshot name exists but snapshot info is not equal to input
	// snapshot info, add snapshot failed.
	// 3. snapshot name exists and snapshot info is equal to input snapshot
	// info, add snapshot succeed.
	Add(info *SnapshotInfo) bool
	// Find snapshot information by snapshot name
	// If founded snapshot, return snapshot info
	// If not founded snapshot, return nil
	Find(snapName string) *SnapshotInfo
	// Delete snapshot information form map
	Delete(snapName string)
	// Add all snapshot information into map
	Sync() error
	// List all snapshot info
	List() []*SnapshotInfo
}

type TextParser interface {
	ParseVolumeList(input string) (volList []*VolumeInfo)

	ParsePoolList(input string) (pools []*PoolInfo)

	ParseSnapshotList(input string) (snaps []*SnapshotInfo)

	ParsePoolNameList(input string) (pools []string)
}
