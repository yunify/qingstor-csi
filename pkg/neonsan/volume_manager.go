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

const (
	CmdQbd               string = "qbd"
	CmdNeonsan           string = "neonsan"
	VolumeStatusOk              = "OK"
	VolumeStatusError           = "ERROR"
	VolumeStatusDegraded        = "DEGRADED"
)

type VolumeManager interface {
	FindVolume(volName string, volPool string) (outVol *volumeInfo, err error)
	FindVolumeWithoutPool(volName string) (outVol *volumeInfo, err error)
	GetPoolNameList() (pools []string, err error)
	CreateVolume(volName string, volPool string, volSize64 int64, replicas int) (outVol *volumeInfo, err error)
	DeleteVolume(volName string, volPool string) (err error)
}

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
//	Description:	find volume without pool name, volume name must be unique in all pools
//	Return cases:	pool,	nil:	found volume's pool
//					"",		nil:	not found
//					"",		error:	error
func FindVolumeWithoutPool(volName string) (outVol *volumeInfo, err error) {
	pools, err := GetPoolNameList()
	if err != nil {
		return nil, err
	}
	var volInfos []*volumeInfo
	for _, p := range pools {
		vol, err := FindVolume(volName, p)
		if err != nil {
			glog.Errorf("error find volume [%s] in pool [%s]", vol.name, vol.pool)
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
func CreateVolume(volName string, volPool string, volSize64 int64, replicas int) (outVol *volumeInfo, err error) {
	// do create
	args := []string{"create_volume", "--volume", volName, "--pool", volPool, "--size", strconv.FormatInt(volSize64, 10), "--repcount", strconv.Itoa(replicas), "-c", ConfigFilePath}
	_, err = ExecCommand(CmdNeonsan, args)
	if err != nil {
		return nil, err
	}
	// get volume information
	return FindVolume(volName, volPool)
}

// 	DeleteVolume
//	Description:	delete volume through Neonsan client command line tool
//	Input:	volume name:	string
//			volume pool:	string
//	Return:	error:	1. not nil: delete volume failed	2. nil: delete volume success
func DeleteVolume(volName string, volPool string) (err error) {
	args := []string{"delete_volume", "--volume", volName, "--pool", volPool, "-c", ConfigFilePath}
	_, err = ExecCommand(CmdNeonsan, args)
	return err
}

//	AttachVolume
//func AttachVolume(volName string, )
