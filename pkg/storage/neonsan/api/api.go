package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pelletier/go-toml"
	"github.com/samuel/go-zookeeper/zk"
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

func CreateVolume(confFile, poolName, volName string, size int64, repCount int) (int, error) {
	request := CreateVolumeRequest{
		Op:       "create_volume",
		PoolName: poolName,
		Name:     volName,
		Size:     size,
		RepCount: repCount,
	}
	response := &CreateVolumeResponse{}
	err := httpGet(confFile, request, response)
	if err != nil {
		return -1, err
	}
	return response.Id, nil
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
	apiUrl, err := getApiUrl(confFile)
	if err != nil {
		return err
	}
	url := "http://" + apiUrl + ":2600/qfa?"
	params := buildParameters(request)
	ret, err := http.Get(url + params)
	if err != nil {
		return err
	}
	if ret.StatusCode != 200 {
		return fmt.Errorf("neonsan API, http code:%d", ret.StatusCode)
	}
	defer func() {
		_ = ret.Body.Close()
	}()

	body, err := ioutil.ReadAll(ret.Body)
	if err != nil {
		return err
	}
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

func getApiUrl(confFile string) (string, error) {
	config, err := toml.LoadFile(confFile)
	if err != nil {
		return "", err
	}
	zkIp_ := config.Get("zookeeper.ip")
	if zkIp_ == nil {
		return "", errors.New("no zookeeper.ip in config file")
	}
	zkIp := zkIp_.(string)
	c, _, err := zk.Connect(strings.Split(zkIp, ","), time.Second)
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
