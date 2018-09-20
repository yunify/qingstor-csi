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
	"reflect"
	"testing"
)

func TestParseVolumeList(t *testing.T) {
	tests := []struct {
		name   string
		output string
		pool   string
		list   []*volumeInfo
		err    error
	}{
		{
			name: "one volume list",
			output: `Volume Count:  1
+--------------+------+-------------+-----------+---------------+--------+---------------------+---------------------+
|      ID      | NAME |    SIZE     | REP COUNT | MIN REP COUNT | STATUS |     STATUS TIME     |    CREATED TIME     |
+--------------+------+-------------+-----------+---------------+--------+---------------------+---------------------+
| 251188477952 | foo  | 10737418240 |         1 |             1 | OK     | 2018-07-09 12:18:34 | 2018-07-09 12:18:34 |
+--------------+------+-------------+-----------+---------------+--------+---------------------+---------------------+`,
			pool: "kube",
			list: []*volumeInfo{
				{
					id:       "251188477952",
					name:     "foo",
					size:     10737418240,
					status:   VolumeStatusOk,
					replicas: 1,
					pool:     "kube",
				},
			},
			err: nil,
		},
		{
			name: "two volumes list",
			output: `Volume Count:  2
+--------------+-------------------------+------------+-----------+---------------+--------+---------------------+---------------------+
|      ID      |          NAME           |    SIZE    | REP COUNT | MIN REP COUNT | STATUS |     STATUS TIME     |    CREATED TIME     |
+--------------+-------------------------+------------+-----------+---------------+--------+---------------------+---------------------+
| 395069882368 | foo                     | 2147483648 |         1 |             1 | OK     | 2018-09-03 20:49:46 | 2018-09-03 20:49:46 |
| 395589976064 | pre-provisioning-volume | 5368709120 |         1 |             1 | OK     | 2018-09-03 22:50:03 | 2018-09-03 22:50:03 |
+--------------+-------------------------+------------+-----------+---------------+--------+---------------------+---------------------+
`,
			pool: "kube",
			list: []*volumeInfo{
				{
					id:       "395069882368",
					name:     "foo",
					size:     2147483648,
					status:   VolumeStatusOk,
					replicas: 1,
					pool:     "kube",
				},
				{
					id:       "395589976064",
					name:     "pre-provisioning-volume",
					size:     5368709120,
					status:   VolumeStatusOk,
					replicas: 1,
					pool:     "kube",
				},
			},
			err: nil,
		},
		{
			name: "no volume list",
			output: `Volume Count:0
`,
			pool: "kube",
			list: nil,
			err:  nil,
		},
	}
	for _, v := range tests {
		volList, err := ParseVolumeList(v.output, v.pool)
		if (v.err == nil && err != nil) || (v.err != nil && err == nil) {
			t.Errorf("name [%s]: error expect [%v], but actually [%v]", v.name, v.err, err)
		}
		if !reflect.DeepEqual(v.list, volList) {
			t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.list, volList)
		}
	}
}

func TestParseSnapshotList(t *testing.T) {
	tests := []struct {
		name   string
		output string
		list   []*snapshotInfo
		err    error
	}{
		{
			name: "two snapshot list",
			output: `Snapshot Count:  2
+--------------+-------------+---------------+---------------+---------------------------+--------+
|  VOLUME ID   | SNAPSHOT ID | SNAPSHOT NAME | SNAPSHOT SIZE |        CREATE TIME        | STATUS |
+--------------+-------------+---------------+---------------+---------------------------+--------+
| 274726912000 |       25463 | snapshot      |    2147483648 | 2018-08-23T11:38:19+08:00 | OK     |
| 274726912000 |       25464 | snapshot2     |    2147483648 | 2018-08-23T11:39:39+08:00 | OK     |
+--------------+-------------+---------------+---------------+---------------------------+--------+
`,
			list: []*snapshotInfo{
				{
					snapName:         "snapshot",
					snapID:           "25463",
					sizeByte:         2147483648,
					status:           SnapshotStatusOk,
					createdTime:      1535024299,
					sourceVolumeName: "274726912000",
				},
				{
					snapName:         "snapshot2",
					snapID:           "25464",
					sizeByte:         2147483648,
					status:           SnapshotStatusOk,
					createdTime:      1535024379,
					sourceVolumeName: "274726912000",
				},
			},
			err: nil,
		},
		{
			name: "no volume list",
			output: `Volume Count:0
`,
			list: nil,
			err:  nil,
		},
	}
	for _, v := range tests {
		snapList, err := ParseSnapshotList(v.output)
		if (v.err == nil && err != nil) || (v.err != nil && err == nil) {
			t.Errorf("name [%s]: error expect [%v], but actually [%v]", v.name, v.err, err)
		}
		if !reflect.DeepEqual(v.list, snapList) {
			t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.list, snapList)
		}
	}

}

func TestParsePoolInfo(t *testing.T) {
	tests := []struct {
		name   string
		output string
		pools  *poolInfo
		err    error
	}{
		{
			name: "find csi pool",
			output: `+----------+-----------+-------+------+------+
| POOL ID  | POOL NAME | TOTAL | FREE | USED |
+----------+-----------+-------+------+------+
| 67108864 | csi       |  2982 | 1222 | 1759 |
+----------+-----------+-------+------+------+

`,
			pools: &poolInfo{
				id:    "67108864",
				name:  "csi",
				total: 2982 * gib,
				free:  1222 * gib,
				used:  1759 * gib,
			},
			err: nil,
		},
		{
			name:   "pool not found",
			output: `Pool Count:  0`,
			pools:  nil,
			err:    nil,
		},
	}
	for _, v := range tests {
		poolNames, err := ParsePoolInfo(v.output)
		if (v.err == nil && err != nil) || (v.err != nil && err == nil) {
			t.Errorf("name [%s]: error expect [%v], but actually [%v]", v.name, v.err, err)
		}
		if !reflect.DeepEqual(v.pools, poolNames) {
			t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.pools, poolNames)
		}
	}
}

func TestParsePoolNameList(t *testing.T) {
	tests := []struct {
		name   string
		output string
		pools  []string
		err    error
	}{
		{
			name: "find csi pool",
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
			err: nil,
		},
		{
			name:   "pool not found",
			output: `Pool Count:  0`,
			pools:  nil,
			err:    nil,
		},
		{
			name:   "wrong output",
			output: `wrong output`,
			pools:  nil,
			err:    fmt.Errorf("wrong output"),
		},
	}
	for _, v := range tests {
		poolNames, err := ParsePoolNameList(v.output)
		if (v.err == nil && err != nil) || (v.err != nil && err == nil) {
			t.Errorf("name [%s]: error expect [%v], but actually [%v]", v.name, v.err, err)
		}
		if !reflect.DeepEqual(v.pools, poolNames) {
			t.Errorf("name [%s]: expect [%v], but actually [%v]", v.name, v.pools, poolNames)
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
			name: "two attached volume",
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
			name:   "have 0 volume",
			output: "Volume Count:  0",
			cnt:    0,
			errStr: "",
		},
		{
			name:   "have 1 volume",
			output: "Volume Count:  1",
			cnt:    1,
			errStr: "",
		},
		{
			name:   "have 2 volumes",
			output: "Volume Count:  2",
			cnt:    2,
			errStr: "",
		},
		{
			name:   "not found count number",
			output: "Volume Count:",
			cnt:    0,
			errStr: "strconv.Atoi: parsing \"\": invalid syntax",
		},
		{
			name:   "not found volume count",
			output: "fake",
			cnt:    0,
			errStr: "cannot found volume count",
		},
	}
	for _, v := range tests {
		exCnt, err := readCountNumber(v.output)
		if err != nil {
			if err.Error() != v.errStr {
				t.Errorf("name [%s]: expect error [%s], but actually [%s]", v.name, v.errStr, err.Error())
			}
		}
		if exCnt != v.cnt {
			t.Errorf("name [%s]: expect [%d], but actually [%d]", v.name, v.cnt, exCnt)
		}
	}
}
