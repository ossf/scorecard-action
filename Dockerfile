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

#v1.19 go
FROM golang:1.19.2@sha256:c2a98a509c3d901aed78332fa0bf6144b4f9ac2bceff2bc77ddc6bc3b70276a5 AS builder
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.* ./
RUN go mod download
COPY . ./

FROM base AS build
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 make build

# TODO: use distroless:
# FROM gcr.io/distroless/base:nonroot@sha256:02f667185ccf78dbaaf79376b6904aea6d832638e1314387c2c2932f217ac5cb
FROM debian:11.5-slim@sha256:b46fc4e6813f6cbd9f3f6322c72ab974cc0e75a72ca02730a8861e98999875c7

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    # For debugging.
    jq ca-certificates curl
COPY --from=build /src/scorecard-action /

# Copy a test policy for local testing.
COPY policies/template.yml  /policy.yml

ENTRYPOINT [ "/scorecard-action" ]
