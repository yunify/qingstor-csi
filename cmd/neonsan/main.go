// +-------------------------------------------------------------------------
// | Copyright (C) 2018 Yunify, Inc.
// +-------------------------------------------------------------------------
// | Licensed under the Apache License, Version 2.0 (the "License");
// | you may not use this work except in compliance with the License.
// | You may obtain a copy of the License in the LICENSE file, or at:
// |
// | http://www.apache.org/licenses/LICENSE-2.0
// |
// | Unless required by applicable law or agreed to in writing, software
// | distributed under the License is distributed on an "AS IS" BASIS,
// | WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// | See the License for the specific language governing permissions and
// | limitations under the License.
// +-------------------------------------------------------------------------

package main

import (
	"flag"
	"math/rand"
	"os"
	"time"

	"github.com/yunify/qingstor-csi/pkg/service"
	"github.com/yunify/qingstor-csi/pkg/storage/neonsan"
	"k8s.io/klog"
)

//noinspection ALL
const (
	version                   = "v1.2.0"
	defaultProvisionName      = "neonsan.csi.qingstor.com"
	defaultConfigPath         = "/etc/neonsan/qbd.conf"
	defaultProtocol           = "tcp"
	defaultPoolName           = "kube"
	defaultInstanceIdFilePath = "/etc/qingcloud/instance-id"
	defaultVolumeArchive      = false
)

//noinspection ALL
var (
	endpoint         = flag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	nodeId           = flag.String("nodeid", "", "If driver cannot get instance ID from /etc/qingcloud/instance-id, we would use this flag.")
	configPath       = flag.String("config", defaultConfigPath, "Neonsan server config file path")
	driverName       = flag.String("drivername", defaultProvisionName, "name of the driver")
	protocol         = flag.String("protocol", defaultProtocol, "protocol of qbd")
	maxVolume        = flag.Int64("maxvolume", 100, "Maximum number of volumes that controller can publish to the node.")
	retryIntervalMax = flag.Duration("retry-interval-max", 2*time.Minute, "Maximum retry interval(s) of failed deletion.")
	volumeArchive    = flag.Bool("volumearchive", defaultVolumeArchive, "Archive volume when delete volume.")
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()
	rand.Seed(time.Now().UTC().UnixNano())
	handle()
	os.Exit(0)
}

func handle() {
	// Get Instance Id
	instanceId, err := service.GetInstanceIdFromFile(defaultInstanceIdFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			klog.Warningf("Failed to get instance id from file, use --nodeId flag. error: %s", err)
			instanceId = *nodeId
		} else {
			klog.Fatalf("Failed to get instance id from file, error: %s", err)
		}
	}
	// Get neonsan config
	storageProvider := neonsan.New(*configPath, *protocol, *volumeArchive)

	klog.Infof("Version: %s", version)

	// Set BackOff
	rt := service.DefaultBackOff
	rt.Cap = *retryIntervalMax

	// Option
	serviceOpt := service.NewOption().SetName(*driverName).SetVersion(version).
		SetNodeId(instanceId).SetMaxVolume(*maxVolume).
		SetVolumeCapabilityAccessNodes(service.DefaultVolumeAccessModeType).
		SetControllerServiceCapabilities(service.DefaultControllerServiceCapability).
		SetNodeServiceCapabilities(service.DefaultNodeServiceCapability).
		SetPluginCapabilities(service.DefaultPluginCapability).
		SetRetryTime(rt)

	// service
	service.Run(serviceOpt, storageProvider, *endpoint)
}
