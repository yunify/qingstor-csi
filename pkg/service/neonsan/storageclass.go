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
	"fmt"
	"github.com/yunify/qingstor-csi/pkg/common"
	"strconv"
)

const (
	StorageClassFsTypeName  = "fsType"
	StorageClassReplicaName = "replica"
)

type StorageClass struct {
	FsType  string
	Replica int
}

// NewStorageClassFromMap create qingStorageClass object from map
func NewStorageClassFromMap(opt map[string]string) (*StorageClass, error) {
	sFsType, fsTypeOk := opt[StorageClassFsTypeName]
	sReplica, replicaOk := opt[StorageClassReplicaName]

	sc := &StorageClass{
		FsType:  common.DefaultFileSystem,
		Replica: common.DefaultDiskReplica,
	}
	if fsTypeOk {
		if !common.IsValidFileSystemType(sFsType) {
			return nil, fmt.Errorf("unsupported filesystem type %s", sFsType)
		}
		sc.FsType = sFsType
	}

	// Get volume replicas
	if replicaOk {
		iReplica, err := strconv.Atoi(sReplica)
		if err != nil {
			return nil, err
		}
		sc.Replica = iReplica
	}
	return sc, nil
}
