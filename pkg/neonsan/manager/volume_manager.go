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

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/yunify/qingstor-csi/pkg/neonsan/util"
	"strconv"
)

// FindVolume get volume detail information
// Input:
//   volume name: string
//   volume pool: string
// Return cases:
//   vol, nil: found volume
//   nil, nil: volume not found
//   nil, err: error
func FindVolume(volName string, poolName string) (volInfo *VolumeInfo, err error) {
	// check input args
	if !util.ContainsString(ListPoolName(), poolName) {
		return nil, fmt.Errorf("invalid pool name [%s]", poolName)
	}
	args := []string{"list_volume", "--volume", volName, "--pool", poolName, "--detail", "-c", util.ConfigFilePath}
	output, err := util.ExecCommand(CmdNeonsan, args)
	if err != nil {
		return nil, err
	}
	volList, err := ParseVolumeList(string(output))
	if err != nil {
		return nil, err
	}
	glog.Infof("Found [%d] volume in [%v].", len(volList), args)
	switch len(volList) {
	case 0:
		return nil, nil
	case 1:
		return volList[0], nil
	default:
		return nil, fmt.Errorf("found duplicated volumes [%s] in pool [%s]", volName, poolName)
	}
}

// FindVolumeWithoutPool find volume info in pools
// Return cases:
//  volumes, nil: found volumes
//  nil, nil: not found
//  nil, error: error
func FindVolumeWithoutPool(volName string) (volInfo *VolumeInfo, err error) {
	var volList []*VolumeInfo
	for _, pool := range ListPoolName() {
		if vol, err := FindVolume(volName, pool); err != nil {
			glog.Errorf("error find volume [%s] in pool [%s]", volName, pool)
			return nil, err
		} else if vol != nil {
			glog.Infof("found volume [%s] in pool [%s]", vol.Name, vol.Pool)
			volList = append(volList, vol)
		}
	}

	switch len(volList) {
	case 0:
		return nil, nil
	case 1:
		return volList[0], nil
	default:
		return nil, fmt.Errorf("found duplicate volume [%s] in [%d] pools", volName, len(volList))
	}
}

// ListVolumeByPool list volume in specific pool
// Return case:
//   volList, nil: found volumes
//   nil, nil: not found volume
//   nil, err: error
func ListVolumeByPool(poolName string) (volList []*VolumeInfo, err error) {
	// check input args
	if !util.ContainsString(ListPoolName(), poolName) {
		return nil, fmt.Errorf("invalid pool name [%s]", poolName)
	}
	args := []string{"list_volume", "--pool", poolName, "--detail", "-c", util.ConfigFilePath}
	output, err := util.ExecCommand(CmdNeonsan, args)
	if err != nil {
		return nil, err
	}
	volList, err = ParseVolumeList(string(output))
	if err != nil {
		return nil, err
	}
	glog.Infof("Found [%d] volume in [%v].", len(volList), volList)
	return volList, nil
}

// CreateVolume create volume
// Input:
//  volume name: string
//  volume pool: string
//  volume bytes: int64
//  replica count: int
// Return case:
//   volume info, nil: succeed to create volume
//   nil, err: failed to create volume
func CreateVolume(volName string, poolName string, volSize64 int64, replicas int) (outVol *VolumeInfo, err error) {
	if volName == "" || poolName == "" || volSize64 == 0 {
		return nil, fmt.Errorf("invalid input arguments")
	}
	// check input args
	if !util.ContainsString(ListPoolName(), poolName) {
		return nil, fmt.Errorf("invalid pool name [%s]", poolName)
	}
	// do create
	args := []string{"create_volume", "--volume", volName, "--pool", poolName, "--size", strconv.FormatInt(volSize64,
		10), "--repcount", strconv.Itoa(replicas), "-c", util.ConfigFilePath}
	_, err = util.ExecCommand(CmdNeonsan, args)
	if err != nil {
		return nil, err
	}
	// get volume information
	return FindVolume(volName, poolName)
}

// DeleteVolume delete volume through Neonsan client command line tool
// Input:
//   volume name: string
//   volume pool: string
// Return:
//   not nil: failed to delete volume
//   nil: succeed to delete volume
func DeleteVolume(volName string, poolName string) (err error) {
	if volName == "" || poolName == "" {
		return fmt.Errorf("invalid input arguments")
	}
	// check input args
	if !util.ContainsString(ListPoolName(), poolName) {
		return fmt.Errorf("invalid pool name [%s]", poolName)
	}
	args := []string{"delete_volume", "--volume", volName, "--pool", poolName, "-c", util.ConfigFilePath}
	_, err = util.ExecCommand(CmdNeonsan, args)
	return err
}

// AttachVolume attach volume to current node
// Input:
//   volume name: string
//   pool name: string
// Return:
//   not nil: failed to attach volume
//   nil: succeed to attach volume
func AttachVolume(volName string, poolName string) (err error) {
	if volName == "" || poolName == "" {
		return fmt.Errorf("invalid input arguments")
	}
	// check input args
	if !util.ContainsString(ListPoolName(), poolName) {
		return fmt.Errorf("invalid pool name [%s]", poolName)
	}
	args := []string{"-m", fmt.Sprintf("%s/%s", poolName, volName), "-c", util.ConfigFilePath}
	_, err = util.ExecCommand(CmdQbd, args)
	return err
}

// DetachVolume detach volume from current node
// Input:
//   volume name: string
//   pool name: string
// Return:
//   not nil: failed to detach volume
//   nil: succeed to detach volume
func DetachVolume(volName string, poolName string) (err error) {
	if volName == "" || poolName == "" {
		return fmt.Errorf("invalid input arguments")
	}
	// check input args
	if !util.ContainsString(ListPoolName(), poolName) {
		return fmt.Errorf("invalid pool name [%s]", poolName)
	}
	args := []string{"-u", fmt.Sprintf("%s/%s", poolName, volName), "-c", util.ConfigFilePath}
	_, err = util.ExecCommand(CmdQbd, args)
	return err
}

// FindAttachedVolumeWithoutPool get attachment volume info
// Input:
//   volume name: string
// Return cases:
//   info, nil: found attached volume
//   nil, nil: not found attached volume
//   nil, err: return error
func FindAttachedVolumeWithoutPool(volName string) (info *AttachInfo, err error) {
	args := []string{"-l"}
	output, err := util.ExecCommand(CmdQbd, args)
	if err != nil {
		glog.Infof("list attached volume failed")
		return nil, err
	}
	infoArr := ParseAttachVolumeList(string(output))
	var infoArrWithName []*AttachInfo
	for i := range infoArr {
		if infoArr[i].Name == volName {
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

// ProbeQbdCommand
func ProbeQbdCommand() (err error) {
	args := []string{"-h"}
	_, err = util.ExecCommand(CmdQbd, args)
	if err != nil {
		glog.Error("Probe Qbd command failed.")
		return err
	}
	return nil
}

// ProbeNeonsanCommand
func ProbeNeonsanCommand() (err error) {
	args := []string{"-h"}
	_, err = util.ExecCommand(CmdNeonsan, args)
	if err != nil {
		glog.Error("Probe Neonsan command failed.")
		return err
	}
	return nil
}
