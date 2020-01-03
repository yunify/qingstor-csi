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

package service

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	"time"
)

var (
	DefaultBackOff = wait.Backoff{
		Duration: time.Second,
		Factor:   1.5,
		Steps:    20,
		Cap:      time.Minute * 2,
	}
)

type Option struct {
	Name          string
	Version       string
	NodeId        string
	MaxVolume     int64
	VolumeCap     []*csi.VolumeCapability_AccessMode
	ControllerCap []*csi.ControllerServiceCapability
	NodeCap       []*csi.NodeServiceCapability
	NodeCapType   []csi.NodeServiceCapability_RPC_Type
	PluginCap     []*csi.PluginCapability

	RetryTime wait.Backoff
}

// NewOption
func NewOption() *Option {
	return &Option{
		RetryTime: DefaultBackOff,
	}
}

func (o *Option) SetName(name string) *Option {
	o.Name = name
	return o
}

func (o *Option) SetVersion(version string) *Option {
	o.Version = version
	return o
}

func (o *Option) SetNodeId(nodeId string) *Option {
	o.NodeId = nodeId
	return o
}

func (o *Option) SetMaxVolume(maxVolume int64) *Option {
	o.MaxVolume = maxVolume
	return o
}

func (o *Option) SetVolumeCapabilityAccessNodes(vc []csi.VolumeCapability_AccessMode_Mode) *Option {
	var vca []*csi.VolumeCapability_AccessMode
	for _, c := range vc {
		vca = append(vca, &csi.VolumeCapability_AccessMode{Mode: c})
	}
	o.VolumeCap = vca
	return o
}

func (o *Option) SetControllerServiceCapabilities(cl []csi.ControllerServiceCapability_RPC_Type) *Option {
	var csc []*csi.ControllerServiceCapability
	for _, c := range cl {
		klog.V(4).Infof("Enabling controller service capability: %v", c.String())
		csc = append(csc, NewControllerServiceCapability(c))
	}
	o.ControllerCap = csc
	return o
}

func (o *Option) SetNodeServiceCapabilities(nl []csi.NodeServiceCapability_RPC_Type) *Option {
	var nsc []*csi.NodeServiceCapability
	for _, n := range nl {
		klog.V(4).Infof("Enabling node service capability: %v", n.String())
		nsc = append(nsc, NewNodeServiceCapability(n))
	}
	o.NodeCap = nsc
	return o
}

func (o *Option) SetPluginCapabilities(cap []*csi.PluginCapability) *Option {
	o.PluginCap = cap
	return o
}

func (o *Option) SetRetryTime(retryTime wait.Backoff) *Option {
	o.RetryTime = retryTime
	return o
}

func (o *Option) ValidateVolumeCapability(cap *csi.VolumeCapability) bool {
	if !o.ValidateVolumeAccessMode(cap.GetAccessMode().GetMode()) {
		return false
	}
	return true
}

func (o *Option) ValidateVolumeCapabilities(caps []*csi.VolumeCapability) bool {
	for _, capability := range caps {
		if !o.ValidateVolumeAccessMode(capability.GetAccessMode().GetMode()) {
			return false
		}
	}
	return true
}

func (o *Option) ValidateVolumeAccessMode(c csi.VolumeCapability_AccessMode_Mode) bool {
	for _, mode := range o.VolumeCap {
		if c == mode.GetMode() {
			return true
		}
	}
	return false
}
