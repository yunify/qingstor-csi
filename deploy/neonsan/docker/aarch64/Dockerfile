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
ARG qbd_version
ENV QBD_VERSION=${qbd_version}
WORKDIR /qingstor-csi
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -mod=vendor  -ldflags "-s -w" -o  _output/neonsan-csi-driver ./cmd/neonsan
COPY deploy/neonsan/qbd/aarch64/${QBD_VERSION}/pitrix-dep-qbd-util-*.deb  _output/


FROM ubuntu:18.04
LABEL maintainers="Yunify"
LABEL description="NeonSAN CSI plugin"
# libcurl4 and libicu60 for qbd
RUN apt-get update -y && \
    apt-get install -y libcurl4 libicu60 && \
    apt-get install -y e2fsprogs xfsprogs mount ca-certificates udev
COPY --from=builder /qingstor-csi/_output/neonsan-csi-driver /neonsan-csi-driver
COPY --from=builder /qingstor-csi/_output/pitrix-dep-qbd-util-*.deb /root/
RUN dpkg -i /root/pitrix-dep-qbd-util-*.deb && rm -rf  /root/pitrix-dep-qbd-util-*.deb
RUN chmod +x /neonsan-csi-driver && \
    mkdir -p /var/log/neonsan-csi-driver
ENTRYPOINT ["/neonsan-csi-driver"]
