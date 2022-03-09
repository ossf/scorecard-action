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

#v1.17 go
FROM golang@sha256:bd9823cdad5700fb4abe983854488749421d5b4fc84154c30dae474100468b85 AS base
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.* ./
RUN go mod download
COPY . ./

FROM base AS build
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 make build-scorecard-action

# FROM gcr.io/distroless/base:nonroot@sha256:02f667185ccf78dbaaf79376b6904aea6d832638e1314387c2c2932f217ac5cb
FROM debian:11.2-slim@sha256:d5cd7e54530a8523168473a2dcc30215f2c863bfa71e09f77f58a085c419155b
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    jq ca-certificates curl
COPY --from=build /src/scorecard-action /

# Copy a test policy for local testing.
COPY policies/template.yml  /policy.yml

ENTRYPOINT [ "/scorecard-action" ]
