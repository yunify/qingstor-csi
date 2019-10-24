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

type Topology struct {
	zone         string
	instanceType InstanceType
}

func NewTopology(zone string, instanceType InstanceType) *Topology {
	return &Topology{zone, instanceType}
}

func (t *Topology) GetZone() string {
	return t.zone
}

func (t *Topology) GetInstanceType() InstanceType {
	return t.instanceType
}

func (t *Topology) SetZone(zone string) {
	t.zone = zone
}

func (t *Topology) SetInstanceType(instanceType InstanceType) {
	t.instanceType = instanceType
}
