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

IMAGE=csiplugin/csi-neonsan
TAG=v1.2.0-rc2
IMAGE_UBUNTU=csiplugin/csi-neonsan-ubuntu
TAG_UBUNTU=v1.2.0
RELEASE_VERSION=v1.2.0
ROOT_PATH=$(pwd)
PACKAGE_LIST=./cmd/... ./pkg/...

neonsan-plugin:
	go build -ldflags "-w -s" -mod=vendor  -o deploy/neonsan/plugin/neonsan-plugin ./cmd/neonsan

neonsan-plugin-debug:
	go build  -gcflags "all=-N -l" -mod=vendor  -o deploy/neonsan/plugin/neonsan-plugin-debug ./cmd/neonsan

container:
	docker build -t ${IMAGE}:${TAG} -f deploy/neonsan/docker/Dockerfile  .

container-ubuntu:
	docker build -t ${IMAGE_UBUNTU}:${TAG_UBUNTU} -f deploy/neonsan/docker/ubuntu/Dockerfile  .

yaml:
	kustomize build deploy/neonsan/kubernetes/base > deploy/neonsan/kubernetes/release/csi-neonsan-${RELEASE_VERSION}.yaml

yaml-ubuntu:
	kustomize build deploy/neonsan/kubernetes/base-ubuntu > deploy/neonsan/kubernetes/release/csi-neonsan-${RELEASE_VERSION}-ubuntu.yaml

release:
	cp deploy/neonsan/plugin/* deploy/neonsan/kubernetes/release && cd deploy/neonsan/kubernetes/ && tar -zcvf csi-neonsan-${RELEASE_VERSION}.tar.gz release/*

mod:
	go build ./...
	go mod download
	go mod tidy
	go mod vendor

test:
	go test -cover -mod=vendor -gcflags=-l -count=1 ./pkg/...

clean:
	go clean -r -x ./...
	rm -rf ./_output
