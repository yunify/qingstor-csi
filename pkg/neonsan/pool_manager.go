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

// poolInfo: stats pool
// total, free, used: pool size in bytes
type poolInfo struct {
	id    string
	name  string
	total int64
	free  int64
	used  int64
}

// FindPool
// Description: get pool detail information
// Input: pool name: string
// Return cases:
//   pool, nil: found pool
//   nil, nil: pool not found
//   nil, err: error
func FindPool(poolName string) (outPool *poolInfo, err error) {
	args := []string{"stats_pool", "-pool", poolName, "-c", ConfigFilePath}
	output, err := ExecCommand(CmdNeonsan, args)
	if err != nil {
		return nil, err
	}
	poolInfo, err := ParsePoolInfo(string(output))
	return poolInfo, err
}

func ListPoolName() (pools []string, err error) {
	args := []string{"list_pool",  "-c", ConfigFilePath}
	output, err := ExecCommand(CmdNeonsan, args)
	if err != nil {
		return nil, err
	}
	pools, err = ParsePoolNameList(string(output))
	return pools, err
}
