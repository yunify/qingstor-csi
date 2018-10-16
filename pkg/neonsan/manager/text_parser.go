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
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParseVolumeList parse a volume info
// 	Return Case:
//   volumes, nil: found volumes info
//   nil, nil: volumes not found
//   nil, err: error
func ParseVolumeList(input string, poolName string) (vols []*VolumeInfo, err error) {
	in := strings.Trim(input, "\n")
	lines := strings.Split(in, "\n")
	for i, v := range lines {
		if i == 0 {
			cnt, err := readCountNumber(v)
			if err != nil {
				glog.Warningf(err.Error())
				return nil, err
			}
			if cnt == 0 {
				glog.Warningf("has 0 volume")
				return nil, nil
			}
		} else if i >= 4 && v[0] != '+' {
			volInfo := readVolumeInfoContent(v)
			volInfo.Pool = poolName
			vols = append(vols, volInfo)
		}
	}
	return vols, nil
}

// ParseSnapshotList
// WARNING: Due to Neonsan CLI tool only returning volume ID, we replace
// volume name field of snapshotInfo by volume id.
func ParseSnapshotList(input string) (snaps []*SnapshotInfo, err error) {
	in := strings.Trim(input, "\n")
	lines := strings.Split(in, "\n")
	for i, v := range lines {
		if i == 0 {
			cnt, err := readCountNumber(v)
			if err != nil {
				glog.Warningf(err.Error())
				return nil, err
			}
			if cnt == 0 {
				glog.Warningf("has 0 snapshot")
				return nil, nil
			}
		} else if i >= 4 && v[0] != '+' {
			snaps = append(snaps, readSnapshotInfoContent(v))
		}
	}
	return snaps, nil
}

// ParsePoolInfo
// Return case:
//   pool info, nil:
func ParsePoolInfo(input string) (poolInfo *PoolInfo, err error) {
	in := strings.Trim(input, "\n")
	lines := strings.Split(in, "\n")
	for i, v := range lines {
		if i >= 3 && v[0] != '+' {
			poolInfo = readPoolInfoContent(v)
		}
	}
	return poolInfo, nil
}

// ParsePoolNameList
func ParsePoolNameList(input string) (poolNames []string, err error) {
	in := strings.Trim(input, "\n")
	lines := strings.Split(in, "\n")
	for i, v := range lines {
		if i == 0 {
			cnt, err := readCountNumber(v)
			if err != nil {
				glog.Warningf(err.Error())
				return nil, err
			}
			if cnt == 0 {
				glog.Warningf("has 0 pool")
				return nil, nil
			}
		} else if i >= 4 && v[0] != '+' {
			poolNames = append(poolNames, readPoolName(v))
		}
	}
	return poolNames, nil
}

//	ParseAttachedVolume
func ParseAttachVolumeList(input string) (infoArr []*AttachInfo) {
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

func readCountNumber(line string) (cnt int, err error) {
	// Because print "count" when list snapshot in lower case and
	// output "Count" when list volume and list pool in upper case.
	if !strings.Contains(line, "ount:") {
		return cnt, fmt.Errorf("cannot found volume count")
	}
	line = strings.Replace(line, " ", "", -1)
	lines := strings.Split(line, ":")
	for i := range lines {
		if i == 1 {
			return strconv.Atoi(lines[i])
		}
	}
	return cnt, fmt.Errorf("cannot found count")
}

// WARNING: not set volume pool
func readVolumeInfoContent(line string) (ret *VolumeInfo) {
	curLine := strings.Replace(line, " ", "", -1)
	curLine = strings.Trim(curLine, "|")
	fields := strings.Split(curLine, "|")
	ret = &VolumeInfo{}
	for i, v := range fields {
		switch i {
		case 0:
			ret.Id = v
		case 1:
			ret.Name = v
		case 2:
			size64, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				glog.Errorf("parse int64 [%s] error in string [%s]", v, line)
				return nil
			}
			ret.SizeByte = size64
		case 3:
			rep, err := strconv.Atoi(v)
			if err != nil {
				glog.Errorf("parse int [%s] error in string [%s]", v, line)
				return nil
			}
			ret.Replicas = rep
		case 5:
			ret.Status = v
		}
	}
	return ret
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

func readPoolInfoContent(line string) (ret *PoolInfo) {
	curLine := strings.Replace(line, " ", "", -1)
	curLine = strings.Trim(curLine, "|")
	fields := strings.Split(curLine, "|")
	ret = &PoolInfo{}
	for i, v := range fields {
		switch i {
		case 0:
			ret.Id = v
		case 1:
			ret.Name = v
		case 2:
			size, err := strconv.Atoi(v)
			if err != nil {
				glog.Error(err.Error())
				return nil
			}
			ret.TotalByte = util.Gib * int64(size)
		case 3:
			size, err := strconv.Atoi(v)
			if err != nil {
				glog.Error(err.Error())
				return nil
			}
			ret.FreeByte = util.Gib * int64(size)
		case 4:
			size, err := strconv.Atoi(v)
			if err != nil {
				glog.Error(err.Error())
				return nil
			}
			ret.UsedByte = util.Gib * int64(size)
		}
	}
	return ret
}

func readSnapshotInfoContent(line string) (ret *SnapshotInfo) {
	curLine := strings.Replace(line, " ", "", -1)
	curLine = strings.Trim(curLine, "|")
	fields := strings.Split(curLine, "|")
	ret = &SnapshotInfo{}
	for i, v := range fields {
		switch i {
		case 0:
			ret.SrcVolName = v
		case 1:
			ret.Id = v
		case 2:
			ret.Name = v
		case 3:
			size64, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				glog.Error(err.Error())
				return nil
			}
			ret.SizeByte = size64
		case 4:
			timeObj, err := time.Parse(util.TimeLayout, v)
			if err != nil {
				glog.Error(err.Error())
				return nil
			}
			ret.CreatedTime = timeObj.Unix()
		case 5:
			if status, ok := SnapshotStatusType[v]; !ok {
				glog.Errorf("Invalid snapshot status [%s]", v)
				return nil
			} else {
				ret.Status = status
			}
		}
	}
	return ret
}

func readAttachVolumeInfo(line string) (ret *AttachInfo) {
	fields := regexp.MustCompile("\\s+").Split(line, -1)
	ret = &AttachInfo{}
	for i, v := range fields {
		switch i {
		case 1:
			ret.Id = util.ParseIntToDec(v)
		case 2:
			ret.Device = "/dev/" + v
		case 3:
			args := strings.Split(v, "/")
			if len(args) != 2 {
				glog.Errorf("expect pool/volume, but actually [%s]", v)
				return nil
			}
			ret.Pool = args[0]
			ret.Name = args[1]
		case 5:
			num, err := strconv.ParseInt(v, 0, 64)
			if err != nil {
				glog.Error(err.Error())
				return nil
			}
			ret.ReadBps = num
		case 6:
			num, err := strconv.ParseInt(v, 0, 64)
			if err != nil {
				glog.Error(err.Error())
				return nil
			}
			ret.WriteBps = num
		case 7:
			num, err := strconv.ParseInt(v, 0, 64)
			if err != nil {
				glog.Error(err.Error())
				return nil
			}
			ret.ReadIops = num
		case 8:
			num, err := strconv.ParseInt(v, 0, 64)
			if err != nil {
				glog.Error(err.Error())
				return nil
			}
			ret.WriteIops = num
		}
	}
	return ret
}
