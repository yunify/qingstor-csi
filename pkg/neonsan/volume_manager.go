package neonsan

import (
	"fmt"
	"github.com/golang/glog"
	"strconv"
	"strings"
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

// 	FindVolume
// 	Description:	get volume detail information
//	Input:	volume name:	string
//			volume pool:	string
// 	Return cases: 	vol, 	nil:	found volume
//					nil,	nil:	volume not found
//					nil,	err:	error
func FindVolume(volName string, volPool string) (outVol *volumeInfo, err error) {
	args := []string{"list_volume", "--volume", volName, "--pool", volPool, "--detail", "-c", ConfigFilePath}
	output, err := execCommand(CmdNeonsan, args)
	if err != nil {
		return nil, err
	}
	outVol = parseVolumeInfo(string(output))
	if outVol == nil {
		return nil, nil
	}
	if outVol.name != volName {
		return nil, fmt.Errorf("Mismatch volume name: expect %s, but actually %s", volName, outVol.name)
	}
	outVol.pool = volPool
	return outVol, nil
}

//	FindVolumePool
//	Description:	find volume pool name, volume name must be unique in all pools
//	Return cases:	pool,	nil:	found volume's pool
//					"",		nil:	not found
//					"",		error:	error
func FindVolumePool(volName string) (volPool string, err error) {
	pools, err := GetPoolNameList()
	if err != nil {
		return "", err
	}
	for _, p := range pools {
		volumes, err := GetVolumeNameList(p)
		if err != nil {
			return "", err
		}
		for _, v := range volumes {
			if v == volName {
				return p, nil
			}
		}
	}
	return "", nil
}

//	GetVolumeNameList
//	Description:	get volume list
//	Return cases:	volumes,	nil:	found volume name list
//					nil,		nil:	volume list not found
//					nil,		nil:	error
func GetVolumeNameList(volPool string) (volumes []string, err error) {
	args := []string{"list_volume", "--pool", volPool, "-c", ConfigFilePath}
	output, err := execCommand(CmdNeonsan, args)
	if err != nil {
		return nil, err
	}
	return parseVolumeList(string(output)), nil
}

//	GetPoolList
//	Description:	get pool list
//	Return cases:	pool,	nil:	found pool
//					nil,	nil:	pool not found
//					nil,	err:	error
func GetPoolNameList() (pools []string, err error) {
	args := []string{"list_pool", "-c", ConfigFilePath}
	output, err := execCommand(CmdNeonsan, args)
	if err != nil {
		return nil, err
	}
	return parsePoolList(string(output)), nil
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
	output, err := execCommand(CmdNeonsan, args)
	glog.Infof("output = %s", output)
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
	output, err := execCommand(CmdNeonsan, args)
	glog.Infof("output = %s", output)
	return err
}

// 	ParseVolumeInfo parse a volume info
//	Input arguments:	string to be parsed:	string
// 	Return Values:	vol: 	1. not nil: found one volume info	2. nil:	volume not found
func parseVolumeInfo(output string) (vol *volumeInfo) {
	out := strings.Trim(output, "\n")
	lines := strings.Split(out, "\n")
	for i, v := range lines {
		switch i {
		case 0:
			cnt, err := readCountNumber(v)
			if err != nil || cnt != 1 {
				return nil
			}
		case 4:
			vol = readVolumeInfoContent(v)
		}
	}
	return vol
}

func readCountNumber(line string) (cnt int, err error) {
	if !strings.Contains(line, "Count:") {
		return cnt, fmt.Errorf("cannot found volume count")
	}
	line = strings.Replace(line, " ", "", -1)
	lines := strings.Split(line, ":")
	for i, _ := range lines {
		if i == 1 {
			return strconv.Atoi(lines[i])
		}
	}
	return cnt, fmt.Errorf("cannot found volume count")
}

func readVolumeInfoContent(line string) (ret *volumeInfo) {
	curLine := strings.Replace(line, " ", "", -1)
	curLine = strings.Trim(curLine, "|")
	fields := strings.Split(curLine, "|")
	volInfo := volumeInfo{}
	for i, v := range fields {
		switch i {
		case 0:
			volInfo.id = v
		case 1:
			volInfo.name = v
		case 2:
			size64, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				glog.Errorf("parse int64 [%d] error in string [%s]", v, line)
				return nil
			}
			volInfo.size = size64
		case 3:
			rep, err := strconv.Atoi(v)
			if err != nil {
				glog.Errorf("parse int [%d] error in string [%s]", v, line)
				return nil
			}
			volInfo.replicas = rep
		case 5:
			volInfo.status = v
		}
	}
	ret = &volInfo
	return ret
}

func parsePoolList(output string) (pools []string) {
	out := strings.Trim(output, "\n")
	lines := strings.Split(out, "\n")
	for i, v := range lines {
		if i == 0 {
			cnt, err := readCountNumber(v)
			if err != nil {
				glog.Error(err.Error())
				return nil
			}
			if cnt == 0 {
				glog.Warningf("server has 0 pool")
				return nil
			}
		} else if i >= 4 && v[0] != '+' {
			pools = append(pools, readPoolName(v))
		}
	}
	return pools
}

func readPoolName(line string) (pool string) {
	curLine := strings.Replace(line, " ", "", -1)
	curLine = strings.Trim(curLine, "|")
	fields := strings.Split(curLine, "|")
	if len(fields) == 1 {
		return fields[0]
	}
	return ""

}

func parseVolumeList(output string) (volumes []string) {
	out := strings.Trim(output, "\n")
	lines := strings.Split(out, "\n")
	for i, v := range lines {
		if i == 0 {
			cnt, err := readCountNumber(v)
			if err != nil {
				glog.Errorf(err.Error())
				return nil
			}
			if cnt == 0 {
				glog.Warningf("server has 0 pool")
				return nil
			}
		} else if i >= 4 && v[0] == '|' {
			volumes = append(volumes, readVolumeName(v))
		}
	}
	return volumes
}

func readVolumeName(line string) (volume string) {
	curLine := strings.Replace(line, " ", "", -1)
	curLine = strings.Trim(curLine, "|")
	fields := strings.Split(curLine, "|")
	if len(fields) == 1 {
		return fields[0]
	}
	return ""
}
