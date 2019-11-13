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

NEONSAN_IMAGE_NAME=csiplugin/csi-neonsan
NEONSAN_VERSION=canary
ROOT_PATH=$(pwd)
PACKAGE_LIST=./cmd/... ./pkg/...

neonsan-plugin:
	go build -ldflags "-w -s" -mod=vendor  -o deploy/neonsan/kubernetes/release/neonsan-plugin ./cmd/neonsan

neonsan-plugin-debug:
	go build  -gcflags "all=-N -l" -mod=vendor  -o deploy/neonsan/kubernetes/release/neonsan-plugin-debug ./cmd/neonsan

neonsan-container:
	docker build -t ${NEONSAN_IMAGE_NAME}:${NEONSAN_VERSION} -f deploy/neonsan/docker/Dockerfile  .

yaml:
	kustomize build deploy/neonsan/kubernetes/base > deploy/neonsan/kubernetes/release/csi-neonsan-${NEONSAN_VERSION}.yaml

release:
	cd deploy/neonsan/kubernetes/ && tar -zcvf csi-neonsan-${NEONSAN_VERSION}.tar.gz release/*

mod:
	go build ./...
	go mod download
	go mod tidy
	go mod vendor

test:
	go test -cover -mod=vendor -gcflags=-l ./pkg/...

clean:
	go clean -r -x ./...
	rm -rf ./_output
