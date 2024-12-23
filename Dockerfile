# Copyright 2021 Security Scorecard Authors
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

# Testing: docker run -e GITHUB_REF=refs/heads/main \
#           -e GITHUB_EVENT_NAME=branch_protection_rule \
#           -e INPUT_RESULTS_FORMAT=sarif \
#           -e INPUT_RESULTS_FILE=results.sarif \
#           -e GITHUB_WORKSPACE=/ \
#           -e INPUT_POLICY_FILE="/policy.yml" \
#           -e INPUT_REPO_TOKEN=$GITHUB_AUTH_TOKEN \
#           -e GITHUB_REPOSITORY="ossf/scorecard" \
#           laurentsimon/scorecard-action:latest

FROM golang:1.23.4@sha256:70031844b8c225351d0bb63e2c383f80db85d92ba894e3da7e13bcf80efa9a37 AS builder
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.* ./
RUN go mod download
COPY . ./

FROM builder AS build
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 make build

# Need root for GitHub Actions support
FROM gcr.io/distroless/base@sha256:e9d0321de8927f69ce20e39bfc061343cce395996dfc1f0db6540e5145bc63a5
COPY --from=build /src/scorecard-action /
COPY policies/template.yml /policy.yml
ENTRYPOINT [ "/scorecard-action" ]
