package neonsan

import (
	"flag"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "/tmp")
	flag.Set("v", "3")
	flag.Parse()
	ret := m.Run()
	os.Exit(ret)
}

func TestCreateVolume(t *testing.T) {
	tests := []struct {
		name      string
		volName   string
		volPool   string
		volSize64 int64
		replicas  int
		infoExist bool
		errStr    string
	}{
		{
			name:      "Create succeed",
			volName:   "foo",
			volPool:   "csi",
			volSize64: 2 * gib,
			replicas:  1,
			infoExist: true,
			errStr:    "",
		},
		{
			name:      "Create failed",
			volName:   "foo",
			volPool:   "csi",
			volSize64: 2 * gib,
			replicas:  1,
			infoExist: false,
			errStr:    "Volume already existed",
		},
	}
	for _, v := range tests {
		volInfo, err := CreateVolume(v.volName, v.volPool, v.volSize64, v.replicas)

		// check volume info
		if (v.infoExist == false && volInfo != nil) || (v.infoExist == true && volInfo == nil) {
			t.Errorf("name %s:  volume info expect [%t], but actually [%t]", v.name, v.infoExist, volInfo == nil)
		}

		// check error
		if v.errStr != "" && err != nil {
			if !strings.Contains(err.Error(), v.errStr) {
				t.Errorf("name %s: error expect [%s], but actually [%s]", v.name, v.errStr, err.Error())
			}
		} else if v.errStr == "" && err == nil {
			continue
		} else {
			t.Errorf("name %s: error expect [%s], but actually [%v]", v.name, v.errStr, err)
		}
	}
}

func TestFindVolume(t *testing.T) {
	tests := []struct {
		name    string
		volName string
		volPool string
		info    *volumeInfo
	}{
		{
			name:    "Found volume",
			volName: "foo",
			volPool: "csi",
			info: &volumeInfo{
				name: "foo",
				pool: "csi",
				size: 2 * gib,
			},
		},
		{
			name:    "Not found volume",
			volName: "nofound",
			volPool: "csi",
			info:    nil,
		},
	}
	for _, v := range tests {
		volInfo, err := FindVolume(v.volName, v.volPool)
		if err != nil {
			t.Errorf("name %s: volume error [%s]", v.name, err.Error())
		}

		// check volume info
		if v.info != nil && volInfo != nil {
			if v.info.name != volInfo.name || v.info.pool != volInfo.pool {
				t.Errorf("name %s: volume info expect [%v], but actually [%v]", v.name, v.info, volInfo)
			}
		}
	}
}

func TestFindVolumeWithoutPool(t *testing.T) {
	tests := []struct {
		name    string
		volName string
		volPool string
	}{
		{
			name:    "Find volume foo in pool csi",
			volName: "foo",
			volPool: "csi",
		},
		{
			name:    "Find volume nofound in pool csi",
			volName: "nofound",
			volPool: "",
		},
	}
	for _, v := range tests {
		ret, err := FindVolumeWithoutPool(v.volName)
		if err != nil {
			t.Errorf("name %s: volume error [%s]", v.name, err.Error())
		}
		if v.volPool != "" && ret != nil {
			if v.volPool != ret.pool {
				t.Errorf("name %s: volume pool expect [%s], but actually [%s]", v.name, v.volPool, ret.pool)
			}
		} else if v.volPool == "" && ret == nil {
			continue
		} else {
			t.Errorf("name %s: volume pool expect [%s], but actually [%v]", v.name, v.volPool, ret)
		}
	}
}

func TestDeleteVolume(t *testing.T) {
	tests := []struct {
		name    string
		volName string
		volPool string
		errStr  string
	}{
		{
			name:    "Delete success",
			volName: "foo",
			volPool: "csi",
			errStr:  "",
		},
		{
			name:    "Delete failed",
			volName: "nofound",
			volPool: "csi",
			errStr:  "Volume not exists",
		},
	}
	for _, v := range tests {
		err := DeleteVolume(v.volName, v.volPool)
		if v.errStr == "" && err == nil {
			continue
		} else if v.errStr != "" && err != nil {
			if !strings.Contains(err.Error(), v.errStr) {
				t.Errorf("name %s: error expect [%s], but actually [%s]", v.name, v.errStr, err.Error())
			}
		} else {
			t.Errorf("name %s: error expect [%s], but actually [%v]", v.name, v.errStr, err)
		}
	}
}

func TestParseVolumeInfo(t *testing.T) {
	tests := []struct {
		name   string
		output string
		vol    *volumeInfo
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
				id:       "251188477952",
				name:     "foo",
				size:     10737418240,
				status:   VolumeStatusOk,
				replicas: 1,
			},
		},
	}
	for _, v := range tests {
		exVol := parseVolumeInfo(v.output)
		if (v.vol == nil && exVol != nil) || (v.vol != nil && exVol == nil) {
			t.Errorf("name %s: parse error, expect %v, but actually %v", v.name, v.vol, exVol)
		} else if !reflect.DeepEqual(*v.vol, *exVol) {
			t.Errorf("name %s: parse error, expect %v, but actually %v", v.name, v.vol, exVol)
		}
	}
}

func TestParsePoolList(t *testing.T) {
	tests := []struct {
		name   string
		output string
		pools  []string
	}{
		{
			name: "Find csi pool",
			output: `Pool Count:  4
+----------+
|   NAME   |
+----------+
| pool     |
| vol      |
| neonpool |
| csi      |
+----------+
`,
			pools: []string{
				"pool",
				"vol",
				"neonpool",
				"csi",
			},
		},
		{
			name:   "Pool not found",
			output: `Pool Count:  0`,
			pools:  []string{},
		},
		{
			name:   "Wrong output",
			output: `wrong output`,
			pools:  []string{},
		},
	}
	for _, v := range tests {
		exPools := parsePoolList(v.output)
		if len(exPools) != len(v.pools) {
			t.Errorf("name %s: expect pools len %d, but actually len %d", v.name, len(v.pools), len(exPools))
		} else {
			for i, _ := range v.pools {
				if v.pools[i] != exPools[i] {
					t.Errorf("name %s: expect pools %v, but actually %v", v.name, v.pools, exPools)
				}
			}
		}
	}
}

func TestReadCountNumber(t *testing.T) {
	tests := []struct {
		name   string
		output string
		cnt    int
		errStr string
	}{
		{
			name:   "Have 0 volume",
			output: "Volume Count:  0",
			cnt:    0,
			errStr: "",
		},
		{
			name:   "Have 1 volume",
			output: "Volume Count:  1",
			cnt:    1,
			errStr: "",
		},
		{
			name:   "Have 2 volumes",
			output: "Volume Count:  2",
			cnt:    2,
			errStr: "",
		},
		{
			name:   "Not found count number",
			output: "Volume Count:",
			cnt:    0,
			errStr: "strconv.Atoi: parsing \"\": invalid syntax",
		},
		{
			name:   "Not found volume count",
			output: "fake",
			cnt:    0,
			errStr: "cannot found volume count",
		},
	}
	for _, v := range tests {
		exCnt, err := readCountNumber(v.output)
		if err != nil {
			if err.Error() != v.errStr {
				t.Errorf("name %s: expect error %s, but actually %s", v.name, v.errStr, err.Error())
			}
		}
		if exCnt != v.cnt {
			t.Errorf("name %s: expect %d, but actually %d", v.name, v.cnt, exCnt)
		}
	}
}
