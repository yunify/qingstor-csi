# +-------------------------------------------------------------------------
# | Copyright (C) 2018 Yunify, Inc.
# +-------------------------------------------------------------------------
# | Licensed under the Apache License, Version 2.0 (the "License");
# | you may not use this work except in compliance with the License.
# | You may obtain a copy of the License in the LICENSE file, or at:
# |
# | http://www.apache.org/licenses/LICENSE-2.0
# |
# | Unless required by applicable law or agreed to in writing, software
# | distributed under the License is distributed on an "AS IS" BASIS,
# | WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# | See the License for the specific language governing permissions and
# | limitations under the License.
# +-------------------------------------------------------------------------

.PHONY: all disk

DISK_IMAGE_NAME=csiplugin/csi-qingcloud
DISK_VERSION=v1.1.0-rc.1
NEONSAN_IMAGE_NAME=csi-neonsan
NEONSAN_VERSION=v0.4.0.1
ROOT_PATH=$(pwd)
PACKAGE_LIST=./cmd/... ./pkg/...

disk: mod
	docker build -t ${DISK_IMAGE_NAME}-builder:${DISK_VERSION} -f deploy/disk/docker/Dockerfile . --target builder

disk-container:
	docker build -t ${DISK_IMAGE_NAME}:${DISK_VERSION} -f deploy/disk/docker/Dockerfile  .

neonsan-container:
	docker build -t ${NEONSAN_IMAGE_NAME} -f deploy/neonsan/docker/Dockerfile  .

install-dev:
	cp /root/.qingcloud/config.yaml deploy/disk/kubernetes/base/config.yaml
	kustomize build  deploy/disk/kubernetes/overlays/dev|kubectl apply -f -

uninstall-dev:
	kustomize build  deploy/disk/kubernetes/overlays/dev|kubectl delete -f -

gen-dev:
	cp /root/.qingcloud/config.yaml deploy/disk/kubernetes/base/config.yaml
	kustomize build deploy/disk/kubernetes/overlays/dev

gen-prod:
	kustomize build deploy/disk/kubernetes/overlays/prod > deploy/disk/kubernetes/releases/qingcloud-csi-disk-${DISK_VERSION}.yaml

mod:
	go build ./...
	go mod download
	go mod tidy
	go mod vendor

fmt:
	go fmt ${PACKAGE_LIST}

fmt-deep: fmt
	gofmt -s -w -l ./pkg/cloud/ ./pkg/common/ ./pkg/disk/driver ./pkg/disk/rpcserver

sanity-test:
	nohup ${ROOT_PATH}/csi-sanity --csi.endpoint /var/lib/kubelet/plugins/disk.csi.qingcloud.com/csi.sock -csi.testvolumeexpandsize 21474836480  -ginkgo.noColor &

clean:
	go clean -r -x ./...
	rm -rf ./_output
