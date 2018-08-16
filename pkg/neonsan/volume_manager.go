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

// 	FindVolume
// 	Description:	get volume detail information
//	Input:	volume name:	string
//			volume pool:	string
// 	Return cases: 	vol, 	nil:	found volume
//					nil,	nil:	volume not found
//					nil,	err:	error
func FindVolume(volName string, volPool string) (outVol *volumeInfo, err error) {
	args := []string{"list_volume", "--volume", volName, "--pool", volPool, "--detail", "-c", ConfigFilePath}
	output, err := ExecCommand(CmdNeonsan, args)
	if err != nil {
		return nil, err
	}
	outVol = ParseVolumeInfo(string(output))
	if outVol == nil {
		return nil, nil
	}
	if outVol.name != volName {
		return nil, fmt.Errorf("mismatch volume name: expect %s, but actually %s", volName, outVol.name)
	}
	outVol.pool = volPool
	return outVol, nil
}

//	FindVolumeWithoutPool
//	Description:	find volume info in all pools
//	Return cases:	volumes,	nil:	found volumes
//					nil,		nil:	not found
//					nil,		error:	error
func FindVolumeWithoutPool(volName string) (volInfo *volumeInfo, err error) {
	pools, err := GetPoolNameList()
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

//	GetPoolList
//	Description:	get pool list
//	Return cases:	pool,	nil:	found pool
//					nil,	nil:	pool not found
//					nil,	err:	error
func GetPoolNameList() (pools []string, err error) {
	args := []string{"list_pool", "-c", ConfigFilePath}
	output, err := ExecCommand(CmdNeonsan, args)
	if err != nil {
		return nil, err
	}
	return ParsePoolList(string(output)), nil
}

// 	CreateVolume
//	Description:	create volume through Neonsan client commandline tool and return volume information
// 	Input:	volume name:	string
//			volume pool:	string
//			volume bytes:	int64
//			replica count:	int
//	Return:	volume information pointer:	*volumeInfo
//			error:	error
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

// 	DeleteVolume
//	Description:	delete volume through Neonsan client command line tool
//	Input:	volume name:	string
//			volume pool:	string
//	Return:	error:	1. not nil: delete volume failed	2. nil: delete volume success
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

// FindAttachedVolumeWithoutPool
// Description:	get attachment volume info
// Input:	volume name:	string
//	Return cases:	info,	nil:	found attached volume
//					nil,	nil:	attached volume not found
//					nil,	err:	return error
func FindAttachedVolumeWithoutPool(volName string) (info *attachInfo, err error) {
	args := []string{"-l"}
	output, err := ExecCommand(CmdQbd, args)
	if err != nil {
		glog.Infof("list attached volume failed")
		return nil, err
	}
	infoArr := ParseAttachVolumeList(string(output))
	var infoArrWithName []*attachInfo
	for i, _ := range infoArr {
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
