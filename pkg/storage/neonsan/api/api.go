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
	"io/ioutil"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pelletier/go-toml"
	"github.com/samuel/go-zookeeper/zk"
	"k8s.io/klog"
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
	Role                string    `json:"role"`
	Policy              string    `json:"policy"`
	CreateTime          time.Time `json:"create_time" format:"ISO 8601"`
	StatusTime          time.Time `json:"status_time" format:"ISO 8601"`
	MetroReplica        string    `json:"metro_replica"`
	ProvisionType       string    `json:"provision_type"` // thin or thick
	MaxBs               int       `json:"max_bs"`
	VolumeAllocated     int       `json:"volume_allocated"`
	RgName              string    `json:"rg_name"`
	Encrypted           string    `json:"encrypted"`
}

type VolumeForClone struct {
	ID                  int    `json:"id"`
	Size                int    `json:"size"`
	ReplicationCount    int    `json:"replication_count"`
	MinReplicationCount int    `json:"min_replication_count"`
	Role                string `json:"role"`
	MaxBs               int    `json:"max_bs"`
	Encrypte            string `json:"encrypte"`
	KeyName             string `json:"key_name"`
	Rg                  string `json:"rg"`
}

func ListVolume(confFile, poolName, volumeName string) (*Volume, error) {
	if len(poolName) == 0 || len(volumeName) == 0 {
		return nil, nil
	}
	request := ListVolumeRequest{
		Op:       "list_volume",
		PoolName: poolName,
		Name:     volumeName,
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

func CreateVolume(confFile, volumeName string, size int64, parameters map[string]string) error {
	request := CreateVolumeRequest{
		Op:         "create_volume",
		Name:       volumeName,
		Size:       size,
		Parameters: parameters,
	}
	response := &CreateVolumeResponse{}
	return httpGet(confFile, request, response)
}

func DeleteVolume(confFile, poolName, volumeName string) error {
	request := DeleteVolumeRequest{
		Op:       "delete_volume",
		PoolName: poolName,
		Name:     volumeName,
	}
	response := &DeleteVolumeResponse{}
	return httpGet(confFile, request, response)
}

func RenameVolume(confFile, poolName, volumeName, newName string) error {
	request := RenameVolumeRequest{
		Op:       "rename_volume",
		PoolName: poolName,
		Name:     volumeName,
		NewName:  newName,
		IsForce:  false,
	}
	response := &RenameVolumeResponse{}
	return httpGet(confFile, request, response)
}

func ResizeVolume(confFile, poolName, volumeName string, size int64) error {
	request := ResizeVolumeRequest{
		Op:       "resize_volume",
		PoolName: poolName,
		Name:     volumeName,
		Size:     size,
		IsForce:  true,
	}
	response := &ResizeVolumeResponse{}
	return httpGet(confFile, request, response)
}

func GetVolumeForClone(confFile, poolName, volumeName string) (*VolumeForClone, error) {
	request := GetVolumeForCloneRequest{
		Op:       "get_volume_info",
		PoolName: poolName,
		Name:     volumeName,
	}
	response := &GetVolumeForCloneResponse{}
	err := httpGet(confFile, request, response)
	if err != nil {
		return nil, err
	}
	return &response.VolumeInfo, nil
}

func CloneVolume(confFile, sourcePoolName, sourceVolumeName, snapshotName, targetVolumeName, targetPoolName string) error {
	request := CloneVolumeRequest{
		Op:           "clone_volume",
		SourcePool:   sourcePoolName,
		SourceVol:    sourceVolumeName,
		SnapshotName: snapshotName,
		TargetPool:   targetPoolName,
		TargetVol:    targetVolumeName,
	}
	response := &CloneVolumeResponse{}
	return httpGet(confFile, request, response)
}

func ListClone(confFile, sourcePoolName, sourceVolumeName, targetPoolName, targeVolumeName string) (*CloneInfo, error) {
	request := ListCloneRequest{
		Op:           "list_clone",
		SvolFullname: sourcePoolName + "/" + sourceVolumeName,
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

func ListClone220(confFile, sourcePoolName, sourceVolumeName, targetPoolName, targetVolumeName string) (*CloneInfo, error) {
	request := ListCloneRequest220{
		Op:        "list_clone",
		SourceVol: sourcePoolName + "/" + sourceVolumeName,
		TargetVol: targetPoolName + "/" + targetVolumeName,
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

func DetachCloneRelationship(confFile, sourcePoolName, sourceVolumeName, targetPoolName, targetVolumeName string) error {
	request := DetachCloneRelationshipRequest{
		Op:        "detach_clone_relationship",
		SourceVol: sourcePoolName + "/" + sourceVolumeName,
		TargetVol: targetPoolName + "/" + targetVolumeName,
	}
	response := &DetachCloneRelationshipResponse{}
	return httpGet(confFile, request, response)
}

func CreateSnapshot(confFile, poolName, volumeName, snapshotName string) error {
	request := CreateSnapshotRequest{
		Op:           "create_snapshot",
		PoolName:     poolName,
		VolumeName:   volumeName,
		SnapshotName: snapshotName,
	}
	response := &CreateSnapshotResponse{}
	return httpGet(confFile, request, response)
}

func DeleteSnapshot(confFile, poolName, volumeName, snapshotName string) error {
	request := DeleteSnapshotRequest{
		Op:           "delete_snapshot",
		PoolName:     poolName,
		VolumeName:   volumeName,
		SnapshotName: snapshotName,
	}
	response := &DeleteSnapshotResponse{}
	return httpGet(confFile, request, response)
}

func ListSnapshot(confFile, poolName, volumeName, snapshotName string) (*SnapshotInfo, error) {
	if len(poolName) == 0 || len(volumeName) == 0 || len(snapshotName) == 0 {
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
	return &response.Snapshots[0], nil
}

func buildParameters(request interface{}) string {
	t, v := reflect.TypeOf(request), reflect.ValueOf(request)
	sb := strings.Builder{}
	parameter := make(map[string]string)
	for k := 0; k < t.NumField(); k++ {
		switch val := v.Field(k).Interface().(type) {
		case int:
			parameter[t.Field(k).Tag.Get("json")] = strconv.Itoa(val)
		case int64:
			parameter[t.Field(k).Tag.Get("json")] = strconv.Itoa(int(val))
		case string:
			parameter[t.Field(k).Tag.Get("json")] = val
		case bool:
			parameter[t.Field(k).Tag.Get("json")] = strconv.FormatBool(val)
		case map[string]string:
			for k1, v1 := range val {
				parameter[k1] = v1
			}
		default:
			klog.Warning("invalidType: ", reflect.TypeOf(val))
		}
	}
	for k2, v2 := range parameter {
		sb.WriteString(fmt.Sprintf("%s=%s&", k2, v2))
	}
	return sb.String()
}

func httpGet(confFile string, request interface{}, response Response) error {
	apiHost, err := getApiHost(confFile)
	if err != nil {
		return err
	}
	url := "http://" + apiHost + ":2600/qfa?" + buildParameters(request)
	klog.Infof("NeonsanApi request:%s", url)
	http.DefaultClient.Timeout = time.Second * 30
	ret, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() {
		if ret != nil && ret.Body != nil {
			_ = ret.Body.Close()
		}
	}()
	// http 400 is param error, let it show ret.Body in error
	if ret.StatusCode != 200 && ret.StatusCode != 400 {
		return fmt.Errorf("NeonsanAPI http code:%d", ret.StatusCode)
	}
	body, err := ioutil.ReadAll(ret.Body)
	if err != nil {
		return err
	}
	klog.Infof("NeonsanApi response:%s, request:%s", body, url)
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
