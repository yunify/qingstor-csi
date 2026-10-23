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
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/protobuf/ptypes/wrappers"
	"golang.org/x/net/context"
	"k8s.io/klog"
)

const (
	// probeCacheTTL keeps Probe cheap: kubelet and the sidecars call it
	// periodically, each call would otherwise hit zookeeper and the center.
	probeCacheTTL = 10 * time.Second
	// probeFailureThreshold avoids flipping the driver to not ready on a single
	// transient failure of the center or of the local qbd daemon.
	probeFailureThreshold = 2
)

// Plugin MUST implement this RPC call
func (s *service) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	ready := s.isReady()
	klog.V(5).Infof("Probe ready:%t", ready)
	return &csi.ProbeResponse{Ready: &wrappers.BoolValue{Value: ready}}, nil
}

// isReady checks qbd and at least one center, with a short result cache.
func (s *service) isReady() bool {
	s.probeMu.Lock()
	defer s.probeMu.Unlock()
	if !s.probeChecked.IsZero() && time.Since(s.probeChecked) < probeCacheTTL {
		return s.probeReady
	}
	s.probeChecked = time.Now()
	err := s.storageProvider.HealthCheck()
	if err != nil {
		s.probeFailures++
		klog.V(4).Infof("Probe health check failed (%d/%d): %v", s.probeFailures, probeFailureThreshold, err)
		if s.probeFailures >= probeFailureThreshold {
			if s.probeReady {
				klog.Errorf("Driver is not ready: %v", err)
			}
			s.probeReady = false
		}
		return s.probeReady
	}
	s.probeFailures = 0
	if !s.probeReady {
		klog.Infof("Driver is ready again")
	}
	s.probeReady = true
	return true
}

// Get plugin capabilities: CONTROLLER, ACCESSIBILITY, EXPANSION
func (s *service) GetPluginCapabilities(ctx context.Context, req *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	klog.V(5).Infof("Using default capabilities")
	return &csi.GetPluginCapabilitiesResponse{Capabilities: s.option.PluginCap}, nil
}

func (s *service) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	klog.V(5).Infof("Using GetPluginInfo")
	return &csi.GetPluginInfoResponse{
		Name:          s.option.Name,
		VendorVersion: s.option.Version,
	}, nil
}
