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
	"github.com/yunify/qingstor-csi/pkg/common"
	"github.com/yunify/qingstor-csi/pkg/storage"
	"google.golang.org/grpc"
	"k8s.io/kubernetes/pkg/util/mount"
)


type Service interface {
	csi.ControllerServer
	csi.IdentityServer
	csi.NodeServer
	Interceptor() grpc.UnaryServerInterceptor
}

type service struct {
	option          *Option
	storageProvider storage.Provider
	mounter         *mount.SafeFormatAndMount
	locks           *common.ResourceLocks
}

func New(option *Option, cloud storage.Provider, mounter *mount.SafeFormatAndMount) Service {
	return &service{
		option:          option,
		storageProvider: cloud,
		mounter:         mounter,
		locks:           common.NewResourceLocks(),
	}
}
