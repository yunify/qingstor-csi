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
	"strconv"
)

// NewDefaulNeonsanStorageClass create default Neonsan StorageClass object
func NewDefaultNeonsanStorageClass() *NeonsanStorageClass {
	return &NeonsanStorageClass{
		Replicas:     1,
		StepSize:     1,
		Pool:         PoolNameDefault,
		VolumeFsType: util.FileSystemDefault,
		Protocol:     util.ProtocolDefault,
	}
}

//	NewNeonsanStorageClassFromMap create a Neonsan StorageClass object from map
func NewNeonsanStorageClassFromMap(opt map[string]string) (*NeonsanStorageClass, error) {
	sc := NewDefaultNeonsanStorageClass()

	//	Get volume replicas
	if sReplica, ok := opt["replicas"]; ok {
		iReplica, err := strconv.Atoi(sReplica)
		if err != nil {
			return nil, err
		}
		sc.Replicas = iReplica
	}

	// Get minimal volume increase size
	if sStepSize, ok := opt["stepSize"]; ok {
		iStepSize, err := strconv.Atoi(sStepSize)
		if err != nil {
			return nil, err
		}
		if iStepSize <= 0 {
			return nil, fmt.Errorf("step size must greate than zero")
		}
		sc.StepSize = iStepSize
	}

	//	Get volume pool
	if sPool, ok := opt["pool"]; ok {
		sc.Pool = sPool
	}

	// Get volume FsType
	// Default is ext4
	if sFsType, ok := opt["fsType"]; ok {
		if !util.IsValidFileSystemType(sFsType) {
			return nil, fmt.Errorf("does not support fsType \"%s\"", sFsType)
		}
		sc.VolumeFsType = sFsType
	}
	return sc, nil
}
