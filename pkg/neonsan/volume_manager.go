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
	"fmt"
	"github.com/golang/glog"
	"strconv"
)

type volumeInfo struct {
	id       string
	name     string
	size     int64
	status   string
	replicas int
	pool     string
}

type attachInfo struct {
	id        string
	name      string
	device    string
	pool      string
	readBps   int64
	writeBps  int64
	readIops  int64
	writeIops int64
}

const (
	CmdQbd               string = "qbd"
	CmdNeonsan           string = "neonsan"
	VolumeStatusOk       string = "OK"
	VolumeStatusError    string = "ERROR"
	VolumeStatusDegraded string = "DEGRADED"
)

// FindVolume get volume detail information
// Input:
//   volume name: string
//   volume pool: string
// Return cases:
//   vol, nil: found volume
//   nil, nil: volume not found
//   nil, err: error
func FindVolume(volName string, volPool string) (outVol *volumeInfo, err error) {
	args := []string{"list_volume", "--volume", volName, "--pool", volPool, "--detail", "-c", ConfigFilePath}
	output, err := ExecCommand(CmdNeonsan, args)
	if err != nil {
		return nil, err
	}
	volList, err := ParseVolumeList(string(output))
	if err != nil {
		return outVol, err
	}
	glog.Infof("Found [%d] volume in [%v].", len(volList), args)
	switch len(volList) {
	case 0:
		return nil, nil
	case 1:
		return volList[0], nil
	default:
		return nil, fmt.Errorf("found duplicated volumes [%s] in pool [%s]", volName, volPool)
	}
}

// FindVolumeWithoutPool find volume info in all pools
// Return cases:
//  volumes, nil: found volumes
//  nil, nil: not found
//  nil, error: error
func FindVolumeWithoutPool(volName string) (volInfo *volumeInfo, err error) {
	pools, err := ListPoolName()
	if err != nil {
		return nil, err
	}
	var volInfos []*volumeInfo
	for _, pool := range pools {
		vol, err := FindVolume(volName, pool)
		if err != nil {
			glog.Errorf("error find volume [%s] in pool [%s]", volName, pool)
			return nil, err
		}
		if vol != nil {
			glog.Infof("found volume [%s] in pool [%s]", vol.name, vol.pool)
			volInfos = append(volInfos, vol)
		}
	}
	switch len(volInfos) {
	case 0:
		return nil, nil
	case 1:
		return volInfos[0], nil
	default:
		return nil, fmt.Errorf("find duplicate volume [%s] in [%d] pools", volName, len(volInfos))
	}
}

// ListVolumeByPool list volume in specific pool
// Return case:
//   volList, nil: found volumes
//   nil, nil: not found volume
//   nil, err: error
func ListVolumeByPool(pool string) (volList []*volumeInfo, err error) {
	args := []string{"list_volume", "--pool", pool, "--detail", "-c", ConfigFilePath}
	output, err := ExecCommand(CmdNeonsan, args)
	if err != nil {
		return nil, err
	}
	volList, err = ParseVolumeList(string(output))
	if err != nil {
		return nil, err
	}
	glog.Infof("Found [%d] volume in [%v].", len(volList), args)
	return volList, nil
}

// CreateVolume create volume through NeonSAN client commandline tool and return volume information
// Input:
//  volume name: string
//  volume pool: string
//  volume bytes: int64
//  replica count: int
// Return case:
//   volume info, nil: succeed to create volume
//   nil, err: failed to create volume
func CreateVolume(volName string, poolName string, volSize64 int64, replicas int) (outVol *volumeInfo, err error) {
	if volName == "" || poolName == "" || volSize64 == 0 {
		return nil, fmt.Errorf("invalid input arguments")
	}
	// do create
	args := []string{"create_volume", "--volume", volName, "--pool", poolName, "--size", strconv.FormatInt(volSize64, 10), "--repcount", strconv.Itoa(replicas), "-c", ConfigFilePath}
	_, err = ExecCommand(CmdNeonsan, args)
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
	args := []string{"delete_volume", "--volume", volName, "--pool", poolName, "-c", ConfigFilePath}
	_, err = ExecCommand(CmdNeonsan, args)
	return err
}

//	AttachVolume
func AttachVolume(volName string, poolName string) (err error) {
	if volName == "" || poolName == "" {
		return fmt.Errorf("invalid input arguments")
	}
	args := []string{"-m", fmt.Sprintf("%s/%s", poolName, volName), "-c", ConfigFilePath}
	_, err = ExecCommand(CmdQbd, args)
	return err
}

//	DetachVolume
func DetachVolume(volName string, poolName string) (err error) {
	if volName == "" || poolName == "" {
		return fmt.Errorf("invalid input arguments")
	}
	args := []string{"-u", fmt.Sprintf("%s/%s", poolName, volName), "-c", ConfigFilePath}
	_, err = ExecCommand(CmdQbd, args)
	return err
}

// FindAttachedVolumeWithoutPool get attachment volume info
// Input:
//   volume name: string
// Return cases:
//   info, nil: found attached volume
//   nil, nil: attached volume not found
//   nil, err: return error
func FindAttachedVolumeWithoutPool(volName string) (info *attachInfo, err error) {
	args := []string{"-l"}
	output, err := ExecCommand(CmdQbd, args)
	if err != nil {
		glog.Infof("list attached volume failed")
		return nil, err
	}
	infoArr := ParseAttachVolumeList(string(output))
	var infoArrWithName []*attachInfo
	for i := range infoArr {
		if infoArr[i].name == volName {
			infoArrWithName = append(infoArrWithName, &infoArr[i])
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

// Probe Qbd command
func ProbeQbdCommand() (err error) {
	args := []string{"-h"}
	_, err = ExecCommand(CmdQbd, args)
	if err != nil {
		glog.Error("Probe Qbd command failed.")
		return err
	}
	return nil
}

// Probe Neonsan command
func ProbeNeonsanCommand() (err error) {
	args := []string{"-h"}
	_, err = ExecCommand(CmdNeonsan, args)
	if err != nil {
		glog.Error("Probe Neonsan command failed.")
		return err
	}
	return nil
}
