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
	"github.com/yunify/qingstor-csi/pkg/neonsan/util"
)

// FindPool
// Description: get pool detail information.
// Input: pool name: string
// Return cases:
//   pool, nil: found pool
//   nil, nil: not found pool
//   nil, err: error
func FindPool(poolName string) (outPool *PoolInfo, err error) {
	// check whether the pool exists
	if !util.ContainsString(ListPoolName(), poolName) {
		// the pool doesn't exist
		return nil, nil
	}
	// get pool info
	args := []string{"stats_pool", "-pool", poolName, "-c", util.ConfigFilePath}
	output, err := util.ExecCommand(CmdNeonsan, args)
	if err != nil {
		return nil, fmt.Errorf("call FindPool [%s]: %v", poolName, err.Error())
	}
	poolInfo, err := ParsePoolInfo(string(output))
	return poolInfo, err
}

// ListPoolName list pool name array
// Return cases:
//   pool list, nil: found pools
//   nil, nil: no found pools
//   nil, err: error
func ListPoolName() (pools []string) {
	return Pools
}
