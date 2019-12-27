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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/protobuf/ptypes"
	"github.com/pelletier/go-toml"
	"github.com/samuel/go-zookeeper/zk"
	"github.com/yunify/qingstor-csi/pkg/common"
	"io/ioutil"
	"k8s.io/klog"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	RetCodeOK = 0
)

type Volume struct {
	Id                  int       `json:"id"`
	Name                string    `json:"name"`
	PoolName            string    `json:"pool_name"`
	Size                int       `json:"size"`
	ReplicationCount    int       `json:"replication_count"`
	Status              string    `json:"status"`
	MinReplicationCount int       `json:"min_replication_count"`
	CreateTime          time.Time `json:"create_time" format:"ISO 8601"`
	StatusTime          time.Time `json:"status_time" format:"ISO 8601"`
	MetroReplica        string    `json:"metro_replica"`
	VolumeAllocated     int       `json:"volume_allocated"`
	ProvisionType       string    `json:"provision_type"`
	Role                string    `json:"role"`
}

func ListVolume(confFile, poolName, volName string) (*Volume, error) {
	request := ListVolumeRequest{
		Op:       "list_volume",
		PoolName: poolName,
		Name:     volName,
	}
	response := &ListVolumeResponse{}
	err := httpGet(confFile, request, response)
	if err != nil {
		return nil, err
	}
	if len(response.Volumes) == 0 {
		return nil, nil
	}
	return &response.Volumes[0], nil
}

func CreateVolume(confFile, poolName, volName string, size int64, repCount int) error {
	request := CreateVolumeRequest{
		Op:       "create_volume",
		PoolName: poolName,
		Name:     volName,
		Size:     size,
		RepCount: repCount,
	}
	response := &CreateVolumeResponse{}
	return httpGet(confFile, request, response)
}

func DeleteVolume(confFile, poolName, volName string) (int, error) {
	request := DeleteVolumeRequest{
		Op:       "delete_volume",
		PoolName: poolName,
		Name:     volName,
	}
	response := &DeleteVolumeResponse{}
	err := httpGet(confFile, request, response)
	if err != nil {
		return -1, err
	}
	return response.Id, nil
}

func ResizeVolume(confFile, poolName, volName string, size int64) error {
	request := ResizeVolumeRequest{
		Op:       "resize_volume",
		PoolName: poolName,
		Name:     volName,
		Size:     size,
	}
	response := &ResizeVolumeResponse{}
	return httpGet(confFile, request, response)
}

func CloneVolume(confFile, sourceVol, snapshotName, sourcePool, targetVol, targetPool string) error {
	request := CloneVolumeRequest{
		Op:           "clone_volume",
		SourcePool:   sourcePool,
		SourceVol:    sourceVol,
		SnapshotName: snapshotName,
		TargetPool:   targetPool,
		TargetVol:    targetVol,
	}
	response := &CloneVolumeResponse{}
	return httpGet(confFile, request, response)
}

func ListClone(confFile, sourceVol, sourcePool, targetVol, targetPool string) (*CloneInfo, error) {
	request := ListCloneRequest{
		Op:           "list_clone",
		SvolFullname: sourcePool + "/" + sourceVol,
	}
	response := &ListCloneResponse{}
	err := httpGet(confFile, request, response)
	if err != nil {
		return nil, err
	}
	if len(response.CloneVolumes) == 0 {
		return nil, errors.New("no clone ")
	}
	return &response.CloneVolumes[0], nil

}

func DetachCloneRelationship(confFile, sourceVol, sourcePool, targetVol, targetPool string) error {
	request := DetachCloneRelationshipRequest{
		Op:        "detach_clone_relationship",
		SourceVol: sourcePool + "/" + sourceVol,
		TargetVol: targetPool + "/" + targetVol,
	}
	response := &DetachCloneRelationshipResponse{}
	return httpGet(confFile, request, response)
}

func CreateSnapshot(confFile, volumeName, snapshotName, poolName string) error {
	request := CreateSnapshotRequest{
		Op:           "create_snapshot",
		PoolName:     poolName,
		VolumeName:   volumeName,
		SnapshotName: snapshotName,
	}
	response := &CreateSnapshotResponse{}
	return httpGet(confFile, request, response)
}

func DeleteSnapshot(confFile, volumeName, snapshotName, poolName string) error {
	request := DeleteSnapshotRequest{
		Op:           "delete_snapshot",
		PoolName:     poolName,
		VolumeName:   volumeName,
		SnapshotName: snapshotName,
	}
	response := &DeleteSnapshotResponse{}
	return httpGet(confFile, request, response)
}

func ListSnapshot(confFile, volumeName, snapshotName, poolName string) (*csi.Snapshot, error) {
	if len(volumeName) == 0 || len(snapshotName) == 0{
		return nil, nil
	}
	request := ListSnapshotRequest{
		Op:           "list_snapshot",
		PoolName:     poolName,
		VolumeName:   volumeName,
		SnapshotName: snapshotName,
	}
	response := &ListSnapshotResponse{}
	err := httpGet(confFile, request, response)
	if err != nil {
		return nil, err
	}
	if len(response.Snapshots) == 0 {
		return nil, nil
	}
	snapshot := response.Snapshots[0]
	creationTime, _ := ptypes.TimestampProto(snapshot.CreateTime)
	csiSnapshot := &csi.Snapshot{
		SizeBytes:      snapshot.SnapshotSize,
		SnapshotId:     common.JoinSnapshotName(volumeName, snapshot.SnapshotName),
		SourceVolumeId: volumeName,
		CreationTime:   creationTime,
		ReadyToUse:     snapshot.Status == "OK",
	}
	return csiSnapshot, nil

}

func buildParameters(request interface{}) string {
	t, v := reflect.TypeOf(request), reflect.ValueOf(request)
	sb := strings.Builder{}
	for k := 0; k < t.NumField(); k++ {
		sb.WriteString(t.Field(k).Tag.Get(`json`))
		sb.WriteString("=")
		switch val := v.Field(k).Interface().(type) {
		case int:
			sb.WriteString(strconv.Itoa(val))
		case int64:
			sb.WriteString(strconv.Itoa(int(val)))
		case string:
			sb.WriteString(val)
		default:
			klog.Warning("invalidType: ", reflect.TypeOf(val))
		}
		sb.WriteString("&")
	}
	return sb.String()
}

func httpGet(confFile string, request interface{}, response Response) error {
	apiHost, err := getApiHost(confFile)
	if err != nil {
		return err
	}
	url := "http://" + apiHost + ":2600/qfa?" + buildParameters(request)
	klog.Infof("NeonsanApi [Begin] request:%s", url)
	ret, err := http.Get(url)
	if err != nil {
		return err
	}
	if ret.StatusCode != 200 {
		return fmt.Errorf("NeonsanAPI http code:%d", ret.StatusCode)
	}
	defer func() {
		_ = ret.Body.Close()
	}()

	body, err := ioutil.ReadAll(ret.Body)
	if err != nil {
		return err
	}
	klog.Infof("NeonsanApi [End] request:%s, response:%s", url, body)
	err = json.Unmarshal(body, response)
	if err != nil {
		return err
	}
	rspHeader := response.Header()
	if rspHeader != nil && rspHeader.RetCode != RetCodeOK {
		return errors.New(rspHeader.Reason)
	}
	return nil
}

func getApiHost(confFile string) (string, error) {
	config, err := toml.LoadFile(confFile)
	if err != nil {
		return "", err
	}
	zkIp_ := config.Get("zookeeper.ip")
	if zkIp_ == nil {
		return "", errors.New("no zookeeper.ip in config file")
	}
	zkIp := zkIp_.(string)
	c, _, err := zk.Connect(strings.Split(zkIp, ","), time.Second, zk.WithLogInfo(false))
	if err != nil {
		return "", err
	}
	defer c.Close()

	clusterName_ := config.Get("zookeeper.cluster_name")
	if clusterName_ == nil {
		return "", errors.New("no zookeeper.cluster_name in config file")
	}
	clusterName := clusterName_.(string)
	if clusterName == "" {
		return "", errors.New("no zookeeper.cluster_name in config")
	}

	zkCenterPath := "/neonsan" + "/" + clusterName + "/centers"
	children, _, err := c.Children(zkCenterPath)
	if err != nil {
		return "", err
	}

	if len(children) == 0 {
		return "", errors.New("no center found in zk")
	}

	sort.Strings(children)
	ip, _, err := c.Get(zkCenterPath + "/" + children[0])
	if err != nil {
		return "", err
	}
	return string(ip), nil
}
