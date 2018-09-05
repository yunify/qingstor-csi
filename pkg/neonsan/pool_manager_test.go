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
	"errors"
	"testing"
)

const (
	PoolTestPoolName     = "csi"
	PoolTestFakePoolName = "fake"
)

func TestFindPool(t *testing.T) {
	tests := []struct {
		name     string
		poolName string
		output   *poolInfo
		err      error
	}{
		{
			name:     "find pool",
			poolName: PoolTestPoolName,
			output: &poolInfo{
				name: PoolTestPoolName,
			},
			err: nil,
		},
		{
			name:     "no found pool",
			poolName: PoolTestFakePoolName,
			output:   nil,
			err:      errors.New("not found"),
		},
	}
	for _, v := range tests {
		poolInfo, err := FindPool(v.poolName)
		if (v.err != nil && err == nil) && (v.err == nil && err != nil) {
			t.Errorf("name %s: error expect %v, but actually %v", v.name, v.err, err)
		} else if v.err == nil && err == nil {
			if v.output.name != poolInfo.name {
				t.Errorf("name %s: error expect %v, but actually %v", v.name, v.output, poolInfo)
			}
		}
	}
}
