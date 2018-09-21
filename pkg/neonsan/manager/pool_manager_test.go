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
	"testing"
	"github.com/yunify/qingstor-csi/pkg/neonsan/util"
)

const (
	PoolTestPoolName     = "csi"
	PoolTestFakePoolName = "fake"
)

func TestFindPool(t *testing.T) {
	Pools = append(Pools, PoolTestPoolName)
	tests := []struct {
		name     string
		poolName string
		output   *PoolInfo
	}{
		{
			name:     "find pool",
			poolName: PoolTestPoolName,
			output: &PoolInfo{
				Name: PoolTestPoolName,
			},
		},
		{
			name:     "not found pool",
			poolName: PoolTestFakePoolName,
			output:   nil,
		},
	}
	for _, v := range tests {
		poolInfo, err := FindPool(v.poolName)
		if err != nil {
			t.Errorf("name [%s]: find pool error [%v]", v.name, err)
		}

		if v.output == nil && poolInfo == nil {
			// no found volume
		} else if v.output != nil && poolInfo != nil {
			// found volume, check volume info
			if v.output.Name != poolInfo.Name {
				t.Errorf("name [%s]: error expect [%v], but actually [%v]", v.name, v.output, poolInfo)
			}
		} else {
			// return value mismatch
			t.Errorf("name [%s]: error expect [%v], but actually [%v]", v.name, v.output, poolInfo)
		}
	}
}

func TestListPoolName(t *testing.T) {
	Pools = append(Pools, PoolTestPoolName)
	tests := []struct {
		name   string
		output string
	}{
		{
			name:   "find pool",
			output: PoolTestPoolName,
		},
	}
	for _, v := range tests {
		pools := ListPoolName()

		// check return pool list
		if !util.ContainsString(pools, v.output) {
			t.Errorf("name [%s]: expect pool [%s] must in return pool list [%v], but actually not", v.name, v.output, pools)
		}
	}
}
