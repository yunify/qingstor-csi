package neonsan

import (
	"testing"
	"reflect"
	"flag"
	"os"
)

func TestMain(m *testing.M) {
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "/tmp")
	flag.Set("v", "3")
	flag.Parse()
	ret := m.Run()
	os.Exit(ret)
}

func TestFindVolume(t *testing.T){
	tests := []struct{
		name string
		volName string
		volPool string
		info *volumeInfo
		errStr string
	}{
		{
			name: "Found volume",
			volName: "foo",
			volPool: "csi",
			info: &volumeInfo{
				name: "foo",
				pool: "csi",
				size: 2*gib,
			},
			errStr: "",
		},
	}
	for _, v:=range tests{
		volInfo, err := FindVolume(v.volName, v.volPool)
		if (err != nil && v.errStr == "" ) || (err == nil && v.errStr != ""){
			t.Errorf("name %s: expect error string is \"%s\", but actually \"%s\"", v.name, v.errStr, err.Error())
		}else if (volInfo == nil && v.info != nil) || (volInfo != nil && v.info == nil){
			t.Errorf("name %s: expect volume %v, but actually %v", v.name, v.info, volInfo)
		}
	}
}

func TestCreateVolume(t *testing.T) {
	tests := []struct{
		name string
		volName string
		volPool string
		volSize64 int64
		replicas int
		nilInfo bool
		errStr string
	}{
		{
			name: "Create success",
			volName: "foo",
			volPool: "csi",
			volSize64: 2*gib,
			replicas: 1,
			nilInfo: false,
			errStr: "",
		},
	}
	for _, v:=range tests{
		volInfo, err := CreateVolume(v.volName, v.volPool, v.volSize64, v.replicas)
		if (err != nil && v.errStr == "" ) || (err == nil && v.errStr != ""){
			t.Errorf("name %s: expect error string is \"%s\", but actually \"%s\"", v.name, v.errStr, err.Error())
		}else if (volInfo == nil && !v.nilInfo) || (volInfo != nil && v.nilInfo){
			t.Errorf("name %s: expect volume pointer %t, but actually %v", v.name, v.nilInfo, volInfo)
		}
	}
}

func TestDeleteVolume(t *testing.T){
	tests := []struct{
		name string
		volName string
		volPool string
		errStr string
	}{
		{
			name: "Delete success",
			volName: "foo",
			volPool: "csi",
			errStr: "",
		},
	}
	for _, v:= range tests{
		err := DeleteVolume(v.volName, v.volPool)
		if (err != nil && v.errStr == "" ) || (err == nil && v.errStr != ""){
			t.Errorf("name %s: expect error string is \"%s\", but actually \"%s\"", v.name, v.errStr, err.Error())
		}
	}
}

func TestParseVolumeInfoFromString(t *testing.T){
	tests := []struct{
		name string
		output string
		vol *volumeInfo
	}{
		{
			name: "Found volume",
			output: `Volume Count:  1 
+--------------+------+-------------+-----------+---------------+--------+---------------------+---------------------+ 
|      ID      | NAME |    SIZE     | REP COUNT | MIN REP COUNT | STATUS |     STATUS TIME     |    CREATED TIME     | 
+--------------+------+-------------+-----------+---------------+--------+---------------------+---------------------+ 
| 251188477952 | foo  | 10737418240 |         1 |             1 | OK     | 2018-07-09 12:18:34 | 2018-07-09 12:18:34 | 
+--------------+------+-------------+-----------+---------------+--------+---------------------+---------------------+`,
			vol: &volumeInfo{
				id: "251188477952",
				name: "foo",
				size: 10737418240,
				status: VolumeStatusOk,
				replicas: 1,
			},
		},
	}
	for _, v := range tests{
		exVol := parseVolumeInfoFromString(v.output)
		if (v.vol == nil && exVol != nil) || (v.vol != nil && exVol == nil){
			t.Errorf("name %s: parse error, expect %v, but actually %v", v.name, v.vol, exVol)
		}else if !reflect.DeepEqual(*v.vol, *exVol){
			t.Errorf("name %s: parse error, expect %v, but actually %v", v.name, v.vol, exVol)
		}
	}
}

func TestReadVolumeInfoCount(t *testing.T){
	tests := []struct{
		name string
		output string
		cnt int
		errStr string
	}{
		{
			name: "Have 0 volume",
			output: "Volume Count:  0",
			cnt: 0,
			errStr: "",
		},
		{
			name: "Have 1 volume",
			output: "Volume Count:  1",
			cnt: 1,
			errStr: "",
		},
		{
			name: "Have 2 volumes",
			output: "Volume Count:  2",
			cnt: 2,
			errStr: "",
		},
		{
			name: "Not found count number",
			output: "Volume Count:",
			cnt: 0,
			errStr: "strconv.Atoi: parsing \"\": invalid syntax",
		},
		{
			name: "Not found volume count",
			output: "fake",
			cnt: 0,
			errStr: "Cannot found volume count",
		},
	}
	for _, v:= range tests{
		exCnt, err := readVolumeInfoCount(v.output)
		if err != nil{
			if err.Error() != v.errStr{
				t.Errorf("name %s: expect error %s, but actually %s", v.name, v.errStr, err.Error())
			}
		}
		if exCnt != v.cnt{
			t.Errorf("name %s: expect %d, but actually %d", v.name, v.cnt, exCnt)
		}
	}
}