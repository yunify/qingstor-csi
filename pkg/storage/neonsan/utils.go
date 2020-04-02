/*
Copyright (C) 2018 Yunify, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this work except in compliance with the License.
You may obtain a copy of the License in the LICENSE file, or at:

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
	"strconv"
	"strings"
)

const (
	snapSep   = "@"
	volumeSep = "/"

	scReplicaName  = "replica"
	defaultReplica = 1
	scPoolName       = "pool"
	defaultPool    = "kube"
)

var (
	errorInvalidArgument = errors.New("invalid argument")
)

func SplitSnapshotName(fullSnapshotName string) (poolName, volumeName, snapshotName string) {
	s := strings.Split(fullSnapshotName, snapSep)
	if len(s) == 2{
		poolName, volumeName = SplitVolumeName(s[0])
		snapshotName = s[1]
	}
	return
}

func JoinSnapshotName(poolName, volumeName, snapshotName string) (fullSnapshotName string) {
	return poolName + volumeSep + volumeName + snapSep + snapshotName
}

func SplitVolumeName(fullVolumeName string) (poolName, volumeName string) {
	s := strings.Split(fullVolumeName, volumeSep)
	if len(s) == 2 {
		poolName, volumeName = s[0], s[1]
	}
	return
}

func JoinVolumeName(poolName, volumeName string) (fullVolumeName string) {
	return poolName + volumeSep + volumeName
}

func GetReplica(parameters map[string]string) (int, error) {
	sReplica, ok := parameters[scReplicaName]
	if ok {
		iReplica, err := strconv.Atoi(sReplica)
		if err != nil || iReplica <= 0 {
			return -1, err
		}
		return iReplica, nil
	}
	return defaultReplica, nil
}

func GetPoolName(parameters map[string]string) string {
	poolName, ok := parameters[scPoolName]
	if ok {
		return poolName
	}
	return ""
}
