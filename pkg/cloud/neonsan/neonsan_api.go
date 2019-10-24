package neonsan

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
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

type ResponseHeader struct {
	Op      string `json:"op"`
	RetCode int    `json:"ret_code"`
	Reason  string `json:"reason"`
}

type CreateVolumeRequest struct {
	Op       string `json:"op"`
	Name     string `json:"name"`
	Size     int    `json:"size"`
	RepCount int    `json:"rep_count"`
	PoolName string `json:"pool_name"`
}

type CreateVolumeResponse struct {
	ResponseHeader
	Id   int `json:"id"`
	Size int `json:"size"`
}

type DeleteVolumeRequest struct {
	Op       string `json:"op"`
	Name     string `json:"name"`
	PoolName string `json:"pool_name"`
}

type DeleteVolumeResponse struct {
	ResponseHeader
	Id int `json:"id"`
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

func buildParameters(request interface{}) string {
	t, v := reflect.TypeOf(request), reflect.ValueOf(request)
	sb := strings.Builder{}
	for k := 0; k < t.NumField(); k++ {
		sb.WriteString(t.Field(k).Tag.Get(`json`))
		sb.WriteString("=")
		switch v.Field(k).Interface().(type) {
		case int:
			sb.WriteString(strconv.Itoa(int(v.Field(k).Int())))
		case string:
			sb.WriteString(v.Field(k).String())
		default:
			sb.WriteString("invalidType")
		}
		sb.WriteString("&")
	}
	return sb.String()
}

func httpGet(request, response interface{}) error {
	url := "http://139.198.126.193:32600/qfa?"
	params := buildParameters(request)
	fmt.Println(params)
	ret, err := http.Get(url + params)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(ret.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, response)
	if err != nil {
		return err
	}
	fmt.Println(string(body))
	return nil
}

func ListVolume(volName string) (*Volume, error) {
	request := ListVolumeRequest{
		Op:       "list_volume",
		PoolName: "kube",
		Name:     volName,
	}
	response := &ListVolumeResponse{}
	err := httpGet(request, response)
	if err != nil {
		return nil, err
	}
	return &response.Volumes[0], nil
}

func CreateVolume(volName, poolName string, size, repCount int) (int, error) {
	request := CreateVolumeRequest{
		Op:       "create_volume",
		Name:     volName,
		Size:     size,
		RepCount: repCount,
		PoolName: poolName,
	}
	response := &CreateVolumeResponse{}
	err := httpGet(request, response)
	if err != nil {
		return -1, err
	}
	return response.Id, nil
}

func DeleteVolume(volName, poolName string) (int, error) {
	request := DeleteVolumeRequest{
		Op:       "delete_volume",
		Name:     volName,
		PoolName: poolName,
	}
	response := &DeleteVolumeResponse{}
	err := httpGet(request, response)
	if err != nil {
		return -1, err
	}
	return response.Id, nil
}
