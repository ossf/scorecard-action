# Copyright 2023 Security Scorecard Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Testing: docker run -e GITHUB_REPOSITORY_OWNER=naveensrinivasan \
# -e GITHUB_REPOSITORY=scorecard-action \
# -e GITHUB_SHA=3fd6b13799a3e63276d0913fefa90c0e9ca32e31 \
# -e GITHUB_TOKEN=GH_TOKEN \
# -e GITHUB_PR_NUMBER=9 \

#v1.19 go
FROM golang:1.19.5@sha256:bb9811fad43a7d6fd2173248d8331b2dcf5ac9af20976b1937ecd214c5b8c383 AS builder
WORKDIR /
ENV CGO_ENABLED=0
COPY go.mod go.sum ./
COPY dependency-analysis/*.go /

FROM builder AS build
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /dependency-analysis /

FROM gcr.io/distroless/base@sha256:122585ba4c098993df9f8dc7285433e8a19974de32528ee3a4b07308808c84ce
COPY --from=build /dependency-analysis /dependency-analysis
ENTRYPOINT ["/dependency-analysis"]
