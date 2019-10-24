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

package driver

import (
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/yunify/qingcloud-csi/pkg/common"
	"strconv"
	"strings"
)

const (
	StorageClassTypeName     = "type"
	StorageClassMaxSizeName  = "maxSize"
	StorageClassMinSizeName  = "minSize"
	StorageClassStepSizeName = "stepSize"
	StorageClassFsTypeName   = "fsType"
	StorageClassReplicaName  = "replica"
	StorageClassTagsName     = "tags"
)

type QingStorageClass struct {
	DiskType VolumeType
	MaxSize  int
	MinSize  int
	StepSize int
	FsType   string
	Replica  int
	Tags     []string
}

// NewDefaultQingStorageClass create default QingStorageClass object
func NewDefaultQingStorageClass() *QingStorageClass {
	return NewDefaultQingStorageClassFromType(DefaultVolumeType)
}

// NewDefaultQingStorageClassFromType create default qingStorageClass by specified volume type
func NewDefaultQingStorageClassFromType(diskType VolumeType) *QingStorageClass {
	if diskType.IsValid() != true {
		return nil
	}
	return &QingStorageClass{
		DiskType: diskType,
		MaxSize:  VolumeTypeToMaxSize[diskType],
		MinSize:  VolumeTypeToMinSize[diskType],
		StepSize: VolumeTypeToStepSize[diskType],
		FsType:   common.DefaultFileSystem,
		Replica:  DefaultDiskReplicaType,
	}
}

// NewQingStorageClassFromMap create qingStorageClass object from map
func NewQingStorageClassFromMap(opt map[string]string) (*QingStorageClass, error) {
	sVolType, volTypeOk := opt[StorageClassTypeName]
	sMaxSize, maxSizeOk := opt[StorageClassMaxSizeName]
	sMinSize, minSizeOk := opt[StorageClassMinSizeName]
	sStepSize, stepSizeOk := opt[StorageClassStepSizeName]
	sFsType, fsTypeOk := opt[StorageClassFsTypeName]
	sReplica, replicaOk := opt[StorageClassReplicaName]
	sTags, tagsOk := opt[StorageClassTagsName]

	sc := NewDefaultQingStorageClass()

	if volTypeOk {
		// Convert volume type to integer
		iVolType, err := strconv.Atoi(sVolType)
		if err != nil {
			return nil, err
		}
		if !VolumeType(iVolType).IsValid() {
			return nil, fmt.Errorf("invalid volume type %d", iVolType)
		}
		sc.DiskType = VolumeType(iVolType)
	}

	if maxSizeOk && minSizeOk && stepSizeOk {
		// Get volume max size
		iMaxSize, err := strconv.Atoi(sMaxSize)
		if err != nil {
			return nil, err
		}
		if iMaxSize <= 0 {
			return nil, fmt.Errorf("max size must greater than zero")
		}
		sc.MaxSize = iMaxSize
		// Get volume min size
		iMinSize, err := strconv.Atoi(sMinSize)
		if err != nil {
			return nil, err
		}
		if iMinSize <= 0 {
			return nil, fmt.Errorf("min size must greater than zero")
		}
		sc.MinSize = iMinSize
		// Ensure volume minSize less than volume maxSize
		if sc.MaxSize < sc.MinSize {
			return nil, fmt.Errorf("max size must greater than or equal to min size")
		}
		// Get volume step size
		iStepSize, err := strconv.Atoi(sStepSize)
		if err != nil {
			return nil, err
		}
		if iStepSize <= 0 {
			return nil, fmt.Errorf("step size must greater than zero")
		}
		sc.StepSize = iStepSize
	}

	if fsTypeOk {
		if !IsValidFileSystemType(sFsType) {
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
		if !IsValidReplica(iReplica) {
			return nil, fmt.Errorf("unsupported replica %s", sReplica)
		}
		sc.Replica = iReplica
	}

	if tagsOk && len(sTags) > 0 {
		sc.Tags = strings.Split(strings.ReplaceAll(sTags, " ", ""), ",")
	}
	return sc, nil
}

func (sc QingStorageClass) GetMinSizeByte() int64 {
	return int64(sc.MinSize) * common.Gib
}

func (sc QingStorageClass) GetMaxSizeByte() int64 {
	return int64(sc.MaxSize) * common.Gib
}
func (sc QingStorageClass) GetStepSizeByte() int64 {
	return int64(sc.StepSize) * common.Gib
}

func (sc QingStorageClass) GetTags() []string {
	return sc.Tags
}

// FormatVolumeSize transfer to proper volume size
func (sc QingStorageClass) FormatVolumeSizeByte(sizeByte int64) int64 {
	if sizeByte <= sc.GetMinSizeByte() {
		return sc.GetMinSizeByte()
	} else if sizeByte > sc.GetMaxSizeByte() {
		return sc.GetMaxSizeByte()
	}
	if sizeByte%sc.GetStepSizeByte() != 0 {
		sizeByte = (sizeByte/sc.GetStepSizeByte() + 1) * sc.GetStepSizeByte()
	}
	if sizeByte > sc.GetMaxSizeByte() {
		return sc.GetMaxSizeByte()
	}
	return sizeByte
}

// Required Volume Size
func (sc QingStorageClass) GetRequiredVolumeSizeByte(capRange *csi.CapacityRange) (int64, error) {
	if capRange == nil {
		return int64(sc.MinSize) * common.Gib, nil
	}
	res := int64(0)
	if capRange.GetRequiredBytes() > 0 {
		res = capRange.GetRequiredBytes()
	}
	res = sc.FormatVolumeSizeByte(res)
	if capRange.GetLimitBytes() > 0 && res > capRange.GetLimitBytes() {
		return -1, fmt.Errorf("volume required bytes %d greater than limit bytes %d", res, capRange.GetLimitBytes())
	}
	return res, nil
}
