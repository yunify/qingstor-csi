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

package qbd

import (
	"fmt"
	"github.com/yunify/qingstor-csi/pkg/common"
	"k8s.io/klog"
	"regexp"
	"strconv"
	"strings"
)

const (
	CmdQbd string = "qbd"
)

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

// AttachVolume attach volume to current node
// Input:
//   volume name: string
//   pool name: string
// Return:
//   not nil: failed to attach volume
//   nil: succeed to attach volume
func AttachVolume(confFile, poolName, volName string) (err error) {
	if volName == "" || poolName == "" {
		return fmt.Errorf("invalid input arguments")
	}
	args := []string{"-m", fmt.Sprintf("%s/%s", poolName, volName), "-c", confFile}
	_, err = common.ExecCommand(CmdQbd, args)
	return err
}

// DetachVolume detach volume from current node
// Input:
//   volume name: string
//   pool name: string
// Return:
//   not nil: failed to detach volume
//   nil: succeed to detach volume
func DetachVolume(confFile, poolName, volName string) (err error) {
	if volName == "" || poolName == "" {
		return fmt.Errorf("invalid input arguments")
	}
	args := []string{"-u", fmt.Sprintf("%s/%s", poolName, volName), "-c", confFile}
	_, err = common.ExecCommand(CmdQbd, args)
	return err
}

// ListVolume get attachment volume info
// Input:
//   volume name: string
// Return cases:
//   info, nil: found attached volume
//   nil, nil: not found attached volume
//   nil, err: return error
func ListVolume(confFile, poolName, volName string) (info *AttachInfo, err error) {
	args := []string{"-l", "-c", confFile}
	output, err := common.ExecCommand(CmdQbd, args)
	if err != nil {
		klog.Infof("list attached volume failed")
		return nil, err
	}
	infoArr := parseAttachVolumeList(string(output))
	var infoArrWithName []*AttachInfo
	for i := range infoArr {
		if infoArr[i].Name == volName  && infoArr[i].Pool == poolName{
			infoArrWithName = append(infoArrWithName, infoArr[i])
		}
	}
	switch len(infoArrWithName) {
	case 0:
		return nil, nil
	case 1:
		return infoArrWithName[0], nil
	default:
		return nil, fmt.Errorf("find duplicate volume [%v]", infoArrWithName)
	}
}

// ParseAttachedVolume parse attached volume list text
func parseAttachVolumeList(input string) (infoArr []*AttachInfo) {
	in := strings.Trim(input, "\n")
	lines := strings.Split(in, "\n")
	for i, v := range lines {
		if i > 0 {
			info := readAttachVolumeInfo(v)
			if info != nil {
				infoArr = append(infoArr, info)
			}
		}
	}
	return infoArr
}

func readAttachVolumeInfo(line string) (ret *AttachInfo) {
	fields := regexp.MustCompile("\\s+").Split(line, -1)
	ret = &AttachInfo{}
	for i, v := range fields {
		switch i {
		case 1:
			ret.Id = common.ParseIntToDec(v)
		case 2:
			ret.Device = "/dev/" + v
		case 3:
			args := strings.Split(v, "/")
			if len(args) != 2 {
				klog.Errorf("expect pool/volume, but actually [%s]", v)
				return nil
			}
			ret.Pool = args[0]
			ret.Name = args[1]
		case 5:
			num, err := strconv.ParseInt(v, 0, 64)
			if err != nil {
				klog.Error(err.Error())
				return nil
			}
			ret.ReadBps = num
		case 6:
			num, err := strconv.ParseInt(v, 0, 64)
			if err != nil {
				klog.Error(err.Error())
				return nil
			}
			ret.WriteBps = num
		case 7:
			num, err := strconv.ParseInt(v, 0, 64)
			if err != nil {
				klog.Error(err.Error())
				return nil
			}
			ret.ReadIops = num
		case 8:
			num, err := strconv.ParseInt(v, 0, 64)
			if err != nil {
				klog.Error(err.Error())
				return nil
			}
			ret.WriteIops = num
		}
	}
	return ret
}
