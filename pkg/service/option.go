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
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"k8s.io/klog"
)

type Option struct {
	name          string
	version       string
	nodeId        string
	maxVolume     int64
	volumeCap     []*csi.VolumeCapability_AccessMode
	controllerCap []*csi.ControllerServiceCapability
	nodeCap       []*csi.NodeServiceCapability
	pluginCap     []*csi.PluginCapability
}

type OptionInput struct {
	Name          string
	Version       string
	NodeId        string
	MaxVolume     int64
	VolumeCap     []csi.VolumeCapability_AccessMode_Mode
	ControllerCap []csi.ControllerServiceCapability_RPC_Type
	NodeCap       []csi.NodeServiceCapability_RPC_Type
	PluginCap     []*csi.PluginCapability
}

// GetOption
// Create disk driver
func GetOption() *Option {
	return &Option{}
}

func (d *Option) InitOption(input *OptionInput) {
	d.name = input.Name
	d.version = input.Version
	// Setup Node Id
	d.nodeId = input.NodeId
	// Setup max volume
	d.maxVolume = input.MaxVolume
	// Setup cap
	d.addVolumeCapabilityAccessModes(input.VolumeCap)
	d.addControllerServiceCapabilities(input.ControllerCap)
	d.addNodeServiceCapabilities(input.NodeCap)
	d.addPluginCapabilities(input.PluginCap)
}

func (d *Option) addVolumeCapabilityAccessModes(vc []csi.VolumeCapability_AccessMode_Mode) {
	var vca []*csi.VolumeCapability_AccessMode
	for _, c := range vc {
		klog.V(4).Infof("Enabling volume access mode: %v", c.String())
		vca = append(vca, NewVolumeCapabilityAccessMode(c))
	}
	d.volumeCap = vca
}

func (d *Option) addControllerServiceCapabilities(cl []csi.ControllerServiceCapability_RPC_Type) {
	var csc []*csi.ControllerServiceCapability
	for _, c := range cl {
		klog.V(4).Infof("Enabling controller service capability: %v", c.String())
		csc = append(csc, NewControllerServiceCapability(c))
	}
	d.controllerCap = csc
}

func (d *Option) addNodeServiceCapabilities(nl []csi.NodeServiceCapability_RPC_Type) {
	var nsc []*csi.NodeServiceCapability
	for _, n := range nl {
		klog.V(4).Infof("Enabling node service capability: %v", n.String())
		nsc = append(nsc, NewNodeServiceCapability(n))
	}
	d.nodeCap = nsc
}

func (d *Option) addPluginCapabilities(cap []*csi.PluginCapability) {
	d.pluginCap = cap
}

func (d *Option) ValidateControllerServiceRequest(c csi.ControllerServiceCapability_RPC_Type) bool {
	if c == csi.ControllerServiceCapability_RPC_UNKNOWN {
		return true
	}

	for _, cap := range d.controllerCap {
		if c == cap.GetRpc().Type {
			return true
		}
	}
	return false
}

func (d *Option) ValidateNodeServiceRequest(c csi.NodeServiceCapability_RPC_Type) bool {
	if c == csi.NodeServiceCapability_RPC_UNKNOWN {
		return true
	}
	for _, cap := range d.nodeCap {
		if c == cap.GetRpc().Type {
			return true
		}
	}
	return false

}

func (d *Option) ValidateVolumeCapability(cap *csi.VolumeCapability) bool {
	if !d.ValidateVolumeAccessMode(cap.GetAccessMode().GetMode()) {
		return false
	}
	return true
}

func (d *Option) ValidateVolumeCapabilities(caps []*csi.VolumeCapability) bool {
	for _, cap := range caps {
		if !d.ValidateVolumeAccessMode(cap.GetAccessMode().GetMode()) {
			return false
		}
	}
	return true
}

func (d *Option) ValidateVolumeAccessMode(c csi.VolumeCapability_AccessMode_Mode) bool {
	for _, mode := range d.volumeCap {
		if c == mode.GetMode() {
			return true
		}
	}
	return false
}

func (d *Option) ValidatePluginCapabilityService(cap csi.PluginCapability_Service_Type) bool {
	for _, v := range d.GetPluginCapability() {
		if v.GetService() != nil && v.GetService().GetType() == cap {
			return true
		}
	}
	return false
}

func (d *Option) GetName() string {
	return d.name
}

func (d *Option) GetVersion() string {
	return d.version
}

func (d *Option) GetInstanceId() string {
	return d.nodeId
}

func (d *Option) GetMaxVolumePerNode() int64 {
	return d.maxVolume
}

func (d *Option) GetControllerCapability() []*csi.ControllerServiceCapability {
	return d.controllerCap
}

func (d *Option) GetNodeCapability() []*csi.NodeServiceCapability {
	return d.nodeCap
}

func (d *Option) GetPluginCapability() []*csi.PluginCapability {
	return d.pluginCap
}

func (d *Option) GetVolumeCapability() []*csi.VolumeCapability_AccessMode {
	return d.volumeCap
}

func (d *Option) GetTopologyZoneKey() string {
	return fmt.Sprintf("topology.%s/zone", d.GetName())
}

func (d *Option) GetTopologyInstanceTypeKey() string {
	return fmt.Sprintf("topology.%s/instance-type", d.GetName())
}
