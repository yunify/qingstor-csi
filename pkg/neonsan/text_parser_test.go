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
	"reflect"
	"testing"
)

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
		exVol := ParseVolumeInfo(v.output)
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
		exPools := ParsePoolList(v.output)
		if len(exPools) != len(v.pools) {
			t.Errorf("name %s: expect pools len %d, but actually len %d", v.name, len(v.pools), len(exPools))
		} else {
			for i := range v.pools {
				if v.pools[i] != exPools[i] {
					t.Errorf("name %s: expect pools %v, but actually %v", v.name, v.pools, exPools)
				}
			}
		}
	}
}

func TestParsePoolInfo(t *testing.T) {
	tests := []struct {
		name  string
		input string
		info  *poolInfo
	}{
		{
			name: "CSI pool info",
			input: `+----------+-----------+-------+------+------+
| POOL ID  | POOL NAME | TOTAL | FREE | USED |
+----------+-----------+-------+------+------+
| 67108864 | csi       |  2982 | 2767 |  214 |
+----------+-----------+-------+------+------+
`,
			info: &poolInfo{
				id:    "67108864",
				name:  "csi",
				total: gib * 2982,
				free:  gib * 2767,
				used:  gib * 214,
			},
		},
		{
			name:  "nil pool info",
			input: ``,
			info:  nil,
		},
	}
	for _, v := range tests {
		ret := ParsePoolInfo(v.input)
		if v.info == nil && ret == nil {
			continue
		} else if v.info != nil && ret != nil {
			if !reflect.DeepEqual(*v.info, *ret) {
				t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.info, ret)
			}
		} else {
			t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.info, ret)
		}
	}
}

func TestParseAttachedVolumeList(t *testing.T) {
	tests := []struct {
		name  string
		input string
		infos []attachInfo
	}{
		{
			name: "Two attached volume",
			input: `dev_id  vol_id  device  volume  config  read_bps    write_bps   read_iops   write_iops
0   0x3ff7000000    qbd0    csi/foo1    /etc/neonsan/qbd.conf   0   0   0   0
1   0x3a7c000000    qbd1    csi/foo /etc/neonsan/qbd.conf   0   0   0   0

`,
			infos: []attachInfo{
				{
					id:        "274726912000",
					name:      "foo1",
					device:    "/dev/qbd0",
					pool:      "csi",
					readBps:   0,
					writeBps:  0,
					readIops:  0,
					writeIops: 0,
				},
				{
					id:        "251188477952",
					name:      "foo",
					device:    "/dev/qbd1",
					pool:      "csi",
					readBps:   0,
					writeBps:  0,
					readIops:  0,
					writeIops: 0,
				},
			},
		},
	}

	for _, v := range tests {
		ret := ParseAttachVolumeList(v.input)
		if len(v.infos) != len(ret) {
			t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.infos, ret)
		} else {
			if !reflect.DeepEqual(v.infos, ret) {
				t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.infos, ret)
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
