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

FROM golang:1.22.1@sha256:af65374fc66d5752364535f761408af0b7852d1223fe4af200033b12c958715b AS builder
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
FROM gcr.io/distroless/base@sha256:280852156756ea3f39f9e774a30346f2e756244e1f432aea3061c4ac85d90a66
COPY --from=build /src/scorecard-action /
COPY policies/template.yml /policy.yml
ENTRYPOINT [ "/scorecard-action" ]
