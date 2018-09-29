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

.PHONY: all neonsan

NEONSAN_IMAGE_NAME=dockerhub.qingcloud.com/csiplugin/csi-neonsan
NEONSAN_IMAGE_VERSION=v0.3.0
NEONSAN_PLUGIN_NAME=neonsan-plugin

neonsan:
	if [ ! -d ./vendor ]; then dep ensure; fi
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o  _output/${NEONSAN_PLUGIN_NAME} ./cmd/neonsan

neonsan-container: neonsan
	cp _output/${NEONSAN_PLUGIN_NAME} deploy/neonsan/docker
	docker build -t $(NEONSAN_IMAGE_NAME):latest deploy/neonsan/docker
	docker tag  $(NEONSAN_IMAGE_NAME):latest $(NEONSAN_IMAGE_NAME):$(NEONSAN_IMAGE_VERSION)

fmt:
	go fmt ./cmd/neonsan
	go fmt ./pkg/neonsan/ ./pkg/neonsan/manager/ ./pkg/neonsan/util/

fmt-deep: fmt
	gofmt -s -w -l ./cmd/neonsan/ ./pkg/neonsan/ ./pkg/neonsan/manager/ ./pkg/neonsan/util/

clean:
	go clean -r -x
	rm -rf ./_output
	rm -rf deploy/block/docker/${NEONSAN_PLUGIN_NAME}
