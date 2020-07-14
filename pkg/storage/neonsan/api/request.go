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

package api

import (
	"time"
)

type ResponseHeader struct {
	Op      string `json:"op"`
	RetCode int    `json:"ret_code"`
	Reason  string `json:"reason"`
}

func (r *ResponseHeader) Header() *ResponseHeader {
	return r
}

type Response interface {
	Header() *ResponseHeader
}

type CreateVolumeRequest struct {
	Op         string            `json:"op"`
	Name       string            `json:"name"`
	Size       int64             `json:"size"`
	Parameters map[string]string `json:"parameters"`
	/*
		RepCount     int    `json:"rep_count"`
		PoolName     string `json:"pool_name"`
		Mrip         string `json:"mrip"`
		Mrport       int    `json:"mrport"`
		Rpo          int    `json:"rpo"`
		Encrypte     string `json:"encrypte"`
		KeyName      string `json:"key_name"`
		Rg           string `json:"rg"`
		Label        string `json:"label"`
		Policy       string `json:"policy"`
		Dc           string `json:"dc"`
		SampleVolume string `json:"sample_volume"`
		MutexGroup   string `json:"mutex_group"`
		Role         string `json:"role"`
		MinRepCount  int    `json:"min_rep_count"`
		MaxBs        int    `json:"max_bs"`
		ThickProv    bool   `json:"thick_prov"`
	*/
}

type CreateVolumeResponse struct {
	ResponseHeader
	Id   int `json:"id"`
	Size int `json:"size"`
}

type DeleteVolumeRequest struct {
	Op       string `json:"op"`
	PoolName string `json:"pool_name"`
	Name     string `json:"name"`
}

type DeleteVolumeResponse struct {
	ResponseHeader
	Id int `json:"id"`
}

type ResizeVolumeRequest struct {
	Op       string `json:"op"`
	PoolName string `json:"pool_name"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
}

type ResizeVolumeResponse struct {
	ResponseHeader
	Id   int `json:"id"`
	Size int `json:"size"`
}

type ListVolumeRequest struct {
	Op       string `json:"op"`
	PoolName string `json:"pool_name"`
	Name     string `json:"name"`
}

type ListVolumeResponse struct {
	ResponseHeader
	Count   int      `json:"count"`
	Volumes []Volume `json:"volumes"`
}

type GetVolumeForCloneRequest struct {
	Op       string `json:"op"`
	Name     string `json:"name"`
	PoolName string `json:"pool_name"`
}

type GetVolumeForCloneResponse struct {
	ResponseHeader
	VolumeInfo VolumeForClone `json:"VolumeInfo"`
}

type CloneVolumeRequest struct {
	Op           string `json:"op"`
	SourcePool   string `json:"source_pool"`
	SourceVol    string `json:"source_vol"`
	SnapshotName string `json:"snapshot_name"`
	TargetPool   string `json:"target_pool"`
	TargetVol    string `json:"target_vol"`
	Size         int64  `json:"size"`
}

type CloneVolumeResponse struct {
	ResponseHeader
	Id        string `json:"id"`
	QueueName string `json:"queue_name"`
}

type ListCloneRequest struct {
	Op           string `json:"op"`
	SvolFullname string `json:"svol_fullname"`
}

type ListCloneRequest220 struct {
	Op        string `json:"op"`
	SourceVol string `json:"source_vol"`
	TargetVol string `json:"target_vol"`
}

type CloneInfo struct {
	Id         int       `json:"id"`
	SourceVol  string    `json:"source_vol"`
	TargetVol  string    `json:"target_vol"`
	CreateTime time.Time `json:"create_time" format:"ISO 8601"`
	Status     string    `json:"status"`
	StatusTime time.Time `json:"status_time" format:"ISO 8601"`
}

type ListCloneResponse struct {
	ResponseHeader
	CloneVolumes []CloneInfo `json:"CloneVolumes"`
}

type DetachCloneRelationshipRequest struct {
	Op        string `json:"op"`
	SourceVol string `json:"source_vol"`
	TargetVol string `json:"target_vol"`
}

type CloneRelation struct {
	SourceVol int `json:"SourceVol"`
	TargetVol int `json:"TargetVol"`
}

type DetachCloneRelationshipResponse struct {
	ResponseHeader
	CloneRelations []CloneRelation `json:"CloneRelations"`
}

type CreateSnapshotRequest struct {
	Op           string `json:"op"`
	PoolName     string `json:"pool_name"`
	VolumeName   string `json:"volume_name"`
	SnapshotName string `json:"snapshot_name"`
}

type CreateSnapshotResponse struct {
	ResponseHeader
}

type DeleteSnapshotRequest struct {
	Op           string `json:"op"`
	PoolName     string `json:"pool_name"`
	VolumeName   string `json:"volume_name"`
	SnapshotName string `json:"snapshot_name"`
}

type DeleteSnapshotResponse struct {
	ResponseHeader
	Id int `json:"id"`
}

type ListSnapshotRequest struct {
	Op           string `json:"op"`
	PoolName     string `json:"pool_name"`
	VolumeName   string `json:"volume_name"`
	SnapshotName string `json:"snapshot_name"` //Optional
}

type SnapshotInfo struct {
	SnapshotId   int       `json:"snapshot_id"`
	SnapshotName string    `json:"snapshot_name"`
	SnapshotSize int64     `json:"snapshot_size"`
	Status       string    `json:"status"`
	CreateTime   time.Time `json:"create_time" format:"ISO 8601"`
}

type ListSnapshotResponse struct {
	ResponseHeader
	VolumeId      int            `json:"volume_id"`
	SnapshotCount int            `json:"snapshot_count"`
	Snapshots     []SnapshotInfo `json:"snapshots"`
}
