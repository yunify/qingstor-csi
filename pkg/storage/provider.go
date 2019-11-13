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

package storage

import "github.com/container-storage-interface/spec/lib/go/csi"

type ControllerOperator interface {
	// Volume Management
	// FindVolume finds and gets volume information by volume ID.
	// Return:
	//   nil,  nil:  volume does not exist
	//   volume, nil: found volume and return volume info
	//   nil,  error: storage system internal error
	FindVolume(volId string) (*csi.Volume, error)
	// FindVolumeByName finds and gets volume information by its name.
	// It will filter volume in deleted and ceased status and return first discovered item.
	// Return:
	//   nil, nil: volume does not exist
	//   volume, nil: found volume and return first discovered volume info
	//   nil, error: storage system internal error
	FindVolumeByName(volName string) (*csi.Volume, error)
	// CreateVolume creates volume with specified name, size, replicas, type and zone and returns volume id.
	// Return:
	//   volume id, nil: succeed to create volume and return volume id
	//   nil, error: failed to create volume
	CreateVolume(volName string, requestSize int64, replicas int) (volId string, err error)
	// DeleteVolume deletes volume by id.
	// Return:
	//   nil: succeed to delete volume
	//   error: failed to delete volume
	DeleteVolume(volId string) error
	// AttachVolume attaches volume on specified node.
	// Return:
	//   nil: succeed to attach volume
	//   error: failed to attach volume
	AttachVolume(volId string, instanceId string) error
	// DetachVolume detaches volume from node.
	// Return:
	//   nil: succeed to detach volume
	//   error: failed to detach volume
	DetachVolume(volId string, instanceId string) error
	// ResizeVolume expands volume to specified capacity.
	// Return:
	//   nil: succeed to expand volume
	//   error: failed to expand volume
	ResizeVolume(volId string, requestSize int64) error
	// CloneVolume clones a volume
	// Return:
	//   volume id, nil: succeed to clone volume and return volume id
	//   nil, error: failed to clone volume
	CloneVolume(volName string, volType int, srcVolId string, zone string) (volId string, err error)
}

type NodeOperator interface {
	NodeAttachVolume(volId string) error
	NodeDetachVolume(volId string) error
	NodeGetDevice(volId string) (device string, err error)
}

type Provider interface {
	ControllerOperator
	NodeOperator
}
