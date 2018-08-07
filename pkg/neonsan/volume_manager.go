package neonsan

import (
	"fmt"
	"strings"
	"strconv"
	"github.com/golang/glog"
)

type volumeInfo struct {
	id string
	name string
	size int64
	status string
	replicas int
	pool string
}

const (
	CmdQbd string = "qbd"
	CmdNeonsan string = "neonsan"
	VolumeStatusOk = "OK"
	VolumeStatusError = "ERROR"
	VolumeStatusDegraded = "DEGRADED"
)

// FindVolume
// Return Value: 	vol, 	nil:	found volume
//					nil,	nil:	volume not found
//					nil,	err:	error
func FindVolume(volName string, volPool string) (outVol *volumeInfo, err error){
	args := []string{"list_volume", "--volume", volName, "--pool", volPool, "--detail", "-c", ConfigFilePath}
	output, err := execCommand(CmdNeonsan, args)
	if err != nil{
		return nil, err
	}
	outVol = parseVolumeInfoFromString(string(output))
	if outVol == nil{
		return nil, nil
	}
	if outVol.name != volName{
		return nil, fmt.Errorf("Mismatch volume name: expect %s, but actually %s", volName, outVol.name)
	}
	outVol.pool = volPool
	return outVol, nil
}

// CreateVolume
func CreateVolume(volName string, volPool string, volSize64 int64, replicas int) (outVol *volumeInfo, err error){
	// do create
	args := []string{"create_volume", "--volume", volName, "--pool", volPool, "--size", strconv.FormatInt(volSize64, 10), "--repcount", strconv.Itoa(replicas), "-c", ConfigFilePath}
	output, err := execCommand(CmdNeonsan, args)
	glog.Infof("output = %s", output)
	if err != nil{
		return nil, err
	}
	// get volume information
	return FindVolume(volName, volPool)
}

// DeleteVolume
func DeleteVolume(volName string, volPool string) (err error){
	args := []string{"delete_volume", "--volume", volName, "--pool", volPool, "-c", ConfigFilePath}
	output, err:=execCommand(CmdNeonsan, args)
	glog.Infof("output = %s", output)
	return err
}

// ParseVolumeInfoFromByte parse a volume info
// Return Value:	vol: 	found one volume info
//					nil:	not found volume
func parseVolumeInfoFromString(output string) (vol *volumeInfo){
	lines:=strings.Split(output, "\n")
	for i, v:= range lines{
		switch i {
		case 0:
			cnt, err := readVolumeInfoCount(v)
			if err != nil || cnt != 1{
				return nil
			}
		case 4:
			vol = readVolumeInfoContent(v)
		}
	}
	return vol
}

func readVolumeInfoCount(line string) (cnt int, err error){
	if !strings.Contains(line, "Volume Count:"){
		return cnt, fmt.Errorf("Cannot found volume count")
	}
	line = strings.Replace(line, " ", "", -1)
	lines := strings.Split(line, ":")
	for i, _:=range lines{
		if i == 1{
			return strconv.Atoi(lines[i])
		}
	}
	return cnt, fmt.Errorf("Cannot found volume count")
}

func readVolumeInfoContent(line string) (ret *volumeInfo) {
	curLine := strings.Replace(line, " ", "", -1)
	curLine = strings.Trim(curLine, "|")
	fields := strings.Split(curLine, "|")
	volInfo := volumeInfo{}
	for i, v:= range fields{
		switch i {
		case 0:
			volInfo.id = v
		case 1:
			volInfo.name = v
		case 2:
			size64, err := strconv.ParseInt(v, 10, 64)
			if err != nil{
				glog.Errorf("parse int64 [%d] error in string [%s]", v, line)
				return nil
			}
			volInfo.size = size64
		case 3:
			rep, err := strconv.Atoi(v)
			if err != nil{
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