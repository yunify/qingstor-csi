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

	// Ret codes the driver has to interpret, they are defined by the center
	// (qfcenter src/qfc/consts.go).
	RetCodeVolumeNotExists = -102
	RetCodeCloneNotFound   = -111
	RetCodeCloneOngoing    = -112
	RetCodeVolumeHasClone  = -114

	defaultApiHostTimeout = time.Second * 10
)

// APIError keeps the ret_code of a center answer next to its reason. Callers
// need it to tell an idempotent no-op, for example a clone relation which is
// already detached (-111), from a real failure.
type APIError struct {
	RetCode int
	Reason  string
}

func (e *APIError) Error() string { return e.Reason }

// IsRetCode reports whether err is an APIError carrying retCode.
func IsRetCode(err error, retCode int) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.RetCode == retCode
}

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

// ListVolumeConns lists the clients which have the volume opened. The
// returned remote addresses belong to the hosts running the qbd daemons.
func ListVolumeConns(confFile, poolName, volumeName string) ([]VolumeConnInfo, error) {
	if len(poolName) == 0 || len(volumeName) == 0 {
		return nil, fmt.Errorf("invalid pool or volume name: %q/%q", poolName, volumeName)
	}
	request := ListVolumeConnsRequest{
		Op:       "list_volume_conns",
		PoolName: poolName,
		Name:     volumeName,
	}
	response := &ListVolumeConnsResponse{}
	err := httpGet(confFile, request, response)
	if err != nil {
		return nil, err
	}
	return response.Connections, nil
}

// GetActiveCenter returns the qfc ip of the current leader center. It is used
// as the center liveness check of the CSI Probe call.
func GetActiveCenter(confFile string, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	apiHost, err := getApiHostWithTimeout(confFile, timeout)
	if err != nil {
		return "", err
	}
	// Discovery and the request share one budget, otherwise a probe could spend
	// timeout on zookeeper and another timeout on the center.
	remaining := time.Until(deadline)
	if remaining <= 0 {
		return "", fmt.Errorf("get active center from %s timeout before request", apiHost)
	}
	client := &http.Client{Timeout: remaining}
	url := "http://" + apiHost + ":2600/control?op=get_active_center"
	response := &GetActiveCenterResponse{}
	if err = doRequest(client, url, response, false); err != nil {
		return "", fmt.Errorf("get active center from %s failed: %v", apiHost, err)
	}
	if response.IP == "" {
		return "", fmt.Errorf("get active center from %s failed: empty ip", apiHost)
	}
	return response.IP, nil
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

func CloneVolume(confFile, sourcePoolName, sourceVolumeName, snapshotName, targetVolumeName, targetPoolName string, linkClone bool, size int64) error {
	request := CloneVolumeRequest{
		Op:           "clone_volume",
		SourcePool:   sourcePoolName,
		SourceVol:    sourceVolumeName,
		SnapshotName: snapshotName,
		TargetPool:   targetPoolName,
		TargetVol:    targetVolumeName,
		Size:         size,
		LinkClone:    linkClone,
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
		return nil, nil
	}
	return &response.CloneVolumes[0], nil
}

// ListCloneByTarget lists the clone relations in which the volume is the
// target. The source is not known when a volume is deleted, and the center
// filters on target_vol alone when source_vol is empty.
func ListCloneByTarget(confFile, targetPoolName, targetVolumeName string) ([]CloneInfo, error) {
	request := ListCloneRequest220{
		Op:        "list_clone",
		TargetVol: targetPoolName + "/" + targetVolumeName,
	}
	response := &ListCloneResponse{}
	if err := httpGet(confFile, request, response); err != nil {
		return nil, err
	}
	return response.CloneVolumes, nil
}

func DetachCloneRelationship(confFile, sourcePoolName, sourceVolumeName, targetPoolName, targetVolumeName string) error {
	request := DetachCloneRelationshipRequest{
		Op:        "detach_clone_relationship",
		SourceVol: sourcePoolName + "/" + sourceVolumeName,
		TargetVol: targetPoolName + "/" + targetVolumeName,
	}
	response := &DetachCloneRelationshipResponse{}
	return httpGetWithClient(confFile, request, response, cloneDetachClient)
}

// DetachCloneRelationshipByTarget detaches every clone relation of the target
// volume, without knowing the source. It is used when the target volume is
// deleted: the center refuses to delete a volume whose clone relation is still
// attached (-114). Detach is metadata only, so it uses the short timeout.
func DetachCloneRelationshipByTarget(confFile, targetPoolName, targetVolumeName string) error {
	request := DetachCloneRelationshipRequest{
		Op:        "detach_clone_relationship",
		TargetVol: targetPoolName + "/" + targetVolumeName,
	}
	response := &DetachCloneRelationshipResponse{}
	return httpGetWithClient(confFile, request, response, cloneDetachClient)
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

// httpClient is shared by the volume/snapshot APIs. It must not be
// http.DefaultClient: writing its Timeout from several goroutines is a data
// race, and the health check needs a different (shorter) budget than the
// volume operations.
var httpClient = &http.Client{Timeout: time.Second * 30}

// cloneDetachClient bounds one detach call. Detach is a metadata operation and
// it runs inside CreateVolume, which has to finish before the provisioner
// timeout (90s in the deployed csi-provisioner).
var cloneDetachClient = &http.Client{Timeout: time.Second * 15}

// getURL is only a thin wrapper around client.Get, it exists so the unit
// tests can replace the HTTP layer.
func getURL(client *http.Client, url string) (*http.Response, error) {
	return client.Get(url)
}

// doRequest performs one center API request: send, read the body, decode it
// into response and check ret_code. A 400 answer is allowed to pass through so
// that the body reason is returned to the caller, the center reports parameter
// errors that way.
func doRequest(client *http.Client, url string, response Response, logInfo bool) error {
	logRequest := klog.V(5).Infof
	if logInfo {
		logRequest = klog.Infof
	}
	logRequest("NeonsanApi request:%s", url)
	ret, err := getURL(client, url)
	if err != nil {
		return err
	}
	defer func() {
		if ret != nil && ret.Body != nil {
			_ = ret.Body.Close()
		}
	}()
	// http 400 is param error, let it show ret.Body in error
	if ret.StatusCode != http.StatusOK && ret.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("NeonsanAPI http code:%d", ret.StatusCode)
	}
	body, err := ioutil.ReadAll(ret.Body)
	if err != nil {
		return err
	}
	logRequest("NeonsanApi response:%s, request:%s", body, url)
	if err = json.Unmarshal(body, response); err != nil {
		return err
	}
	rspHeader := response.Header()
	if rspHeader != nil && rspHeader.RetCode != RetCodeOK {
		return &APIError{RetCode: rspHeader.RetCode, Reason: rspHeader.Reason}
	}
	return nil
}

func httpGet(confFile string, request interface{}, response Response) error {
	return httpGetWithClient(confFile, request, response, httpClient)
}

func httpGetWithClient(confFile string, request interface{}, response Response, client *http.Client) error {
	apiHost, err := getApiHost(confFile)
	if err != nil {
		return err
	}
	url := "http://" + apiHost + ":2600/qfa?" + buildParameters(request)
	return doRequest(client, url, response, true)
}

func getApiHost(confFile string) (string, error) {
	return getApiHostWithTimeout(confFile, defaultApiHostTimeout)
}

// getApiHostWithTimeout resolves the center address from zookeeper, bounding
// the whole discovery (connect + lookup) by timeout.
func getApiHostWithTimeout(confFile string, timeout time.Duration) (string, error) {
	config, err := toml.LoadFile(confFile)
	if err != nil {
		return "", err
	}
	zkIp_ := config.Get("zookeeper.ip")
	if zkIp_ == nil {
		return "", errors.New("no zookeeper.ip in config file")
	}
	zkIp := zkIp_.(string)
	sessionTimeout := timeout
	if sessionTimeout < time.Second {
		sessionTimeout = time.Second
	}
	c, events, err := zk.Connect(strings.Split(zkIp, ","), sessionTimeout, zk.WithLogInfo(false))
	if err != nil {
		return "", err
	}
	defer c.Close()
	if err = waitForZKConnection(events, timeout); err != nil {
		return "", err
	}

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

// waitForZKConnection waits until the zookeeper client has established a
// session.
func waitForZKConnection(events <-chan zk.Event, timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case event, ok := <-events:
			if !ok {
				return errors.New("zookeeper connection closed")
			}
			switch event.State {
			case zk.StateHasSession:
				return nil
			case zk.StateAuthFailed:
				return errors.New("zookeeper auth failed")
			}
		case <-timer.C:
			return errors.New("connect zookeeper timeout")
		}
	}
}
