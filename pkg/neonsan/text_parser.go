package neonsan

import (
	"fmt"
	"github.com/golang/glog"
	"strconv"
	"strings"
)

type TextParser interface {
	ParseVolumeInfo(input string) (vol *volumeInfo)
	ParsePoolList(input string) (pools []string)
	ParsePoolInfo(input string) (poll *poolInfo)
}

// 	ParseVolumeInfo parse a volume info
//	Input arguments:	string to be parsed:	string
// 	Return Values:	vol: 	1. not nil: found one volume info	2. nil:	volume not found
func ParseVolumeInfo(input string) (vol *volumeInfo) {
	in := strings.Trim(input, "\n")
	lines := strings.Split(in, "\n")
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

//	ParseVolumeList
func ParsePoolList(input string) (pools []string) {
	in := strings.Trim(input, "\n")
	lines := strings.Split(in, "\n")
	for i, v := range lines {
		if i == 0 {
			cnt, err := readCountNumber(v)
			if err != nil {
				glog.Warningf(err.Error())
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

//	ParsePoolInfo
func ParsePoolInfo(input string) (pool *poolInfo) {
	in := strings.Trim(input, "\n")
	lines := strings.Split(in, "\n")
	for i, v := range lines {
		switch i {
		case 3:
			pool = readPoolInfoContent(v)
		}
	}
	return pool
}

func readCountNumber(line string) (cnt int, err error) {
	if !strings.Contains(line, "Count:") {
		return cnt, fmt.Errorf("cannot found volume count")
	}
	line = strings.Replace(line, " ", "", -1)
	lines := strings.Split(line, ":")
	for i := range lines {
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

func readPoolName(line string) (pool string) {
	curLine := strings.Replace(line, " ", "", -1)
	curLine = strings.Trim(curLine, "|")
	fields := strings.Split(curLine, "|")
	if len(fields) == 1 {
		return fields[0]
	}
	return ""
}

func readPoolInfoContent(line string) (ret *poolInfo) {
	curLine := strings.Replace(line, " ", "", -1)
	curLine = strings.Trim(curLine, "|")
	fields := strings.Split(curLine, "|")
	poolInfo := poolInfo{}
	for i, v := range fields {
		switch i {
		case 0:
			poolInfo.id = v
		case 1:
			poolInfo.name = v
		case 2:
			size, err := strconv.Atoi(v)
			if err != nil {
				glog.Error(err.Error())
				return nil
			}
			poolInfo.total = gib * int64(size)
		case 3:
			size, err := strconv.Atoi(v)
			if err != nil {
				glog.Error(err.Error())
				return nil
			}
			poolInfo.free = gib * int64(size)
		case 4:
			size, err := strconv.Atoi(v)
			if err != nil {
				glog.Error(err.Error())
				return nil
			}
			poolInfo.used = gib * int64(size)
		}
	}
	ret = &poolInfo
	return ret
}
