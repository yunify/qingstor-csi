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
	"strings"
)

const (
	snapSep   = "@"
	volumeSep = "/"

	scReplicaNameOld = "replica"
	scPoolNameOld    = "pool"
	scPoolNameNew    = "pool_name"
)

func SplitSnapshotName(fullSnapshotName string) (poolName, volumeName, snapshotName string) {
	s := strings.Split(fullSnapshotName, snapSep)
	if len(s) == 2 {
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

func GetPoolName(parameters map[string]string) string {
	if poolNameNew, okNew := parameters[scPoolNameNew]; okNew {
		return poolNameNew
	} else if poolNameOld, okOld := parameters[scPoolNameOld]; okOld {
		return poolNameOld
	}
	return ""
}

func TuneUpParameter(parameter map[string]string) {
	//Backward: "pool" -> "pool_name", "replica" -> "rep_count"
	backward(parameter, scPoolNameOld, scPoolNameNew)
	backward(parameter, scReplicaNameOld, "rep_count")
	// set default if the parameter is empty
	setDefaultIfEmpty(parameter, "rep_count", "1")
	// delete unnecessary
	delUnnecessary(parameter, "fsType")
}

func backward(parameters map[string]string, oldKey, newKey string) {
	if poolName, okOld := parameters[oldKey]; okOld {
		if _, okNew := parameters[newKey]; !okNew {
			parameters[newKey] = poolName
		}
		delete(parameters, oldKey)
	}
}

func setDefaultIfEmpty(parameters map[string]string, key, defaultValue string) {
	if parameters == nil {
		return
	}
	if _, ok := parameters[key]; !ok {
		parameters[key] = defaultValue
	}
}

func delUnnecessary(parameters map[string]string, key string) {
	if parameters != nil {
		delete(parameters, key)
	}
}
