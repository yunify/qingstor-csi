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

FROM golang:1.14.4-alpine as builder
WORKDIR /qingstor-csi
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -mod=vendor  -ldflags "-s -w" -o  _output/neonsan-csi-driver ./cmd/neonsan

FROM centos:7
LABEL maintainers="Yunify"
LABEL description="NeonSAN CSI plugin"
# libcurl3 and libicu55 for qbd
RUN yum -y install libcurl libicu
RUN yum -y install e2fsprogs xfsprogs mount ca-certificates udev
COPY --from=builder /qingstor-csi/_output/neonsan-csi-driver /neonsan-csi-driver
RUN chmod +x /neonsan-csi-driver && \
    mkdir -p /var/log/neonsan-csi-driver
ENTRYPOINT ["/neonsan-csi-driver"]