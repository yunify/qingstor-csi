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
	"time"

	"github.com/yunify/qingstor-csi/pkg/storage"
	"github.com/yunify/qingstor-csi/pkg/storage/neonsan/api"
)

const healthCheckTimeout = time.Second * 3

type neonsan struct {
	confFile      string
	protocol      string
	volumeArchive bool
}

func New(confFile, protocol string, volumeArchive bool) storage.Provider {
	return &neonsan{
		confFile:      confFile,
		protocol:      protocol,
		volumeArchive: volumeArchive,
	}
}

// HealthCheck reports the driver as ready when at least one center is
// reachable. qbd is not part of the check: the controller plugin does not use
// it and a node side qbd failure surfaces on the volume operations itself.
func (v *neonsan) HealthCheck() error {
	if _, err := api.GetActiveCenter(v.confFile, healthCheckTimeout); err != nil {
		return fmt.Errorf("no reachable neonsan center: %v", err)
	}
	return nil
}
