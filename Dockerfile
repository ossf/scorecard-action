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

# See docs/development.md for details on how to test this image.

# TODO: Prefer SHA for builder image.
# TODO: Upgrade to go1.18 once this repo is compatible.
FROM golang:1.17-bullseye as builder

WORKDIR /workspace

# TODO: Revisit directory structure to make this a more lightweight copy.
COPY ./ ./

# Copied from make build target
RUN CGO_ENABLED=0 go build -o scorecard -trimpath -a -tags netgo -ldflags '-w -extldflags'

# Build our image and update the root certs.
# TODO: use distroless.
FROM debian:11.3-slim@sha256:fbaacd55d14bd0ae0c0441c2347217da77ad83c517054623357d1f9d07f79f5e
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    jq ca-certificates curl

# Copy the scorecard binary from the intermediate builder image.
COPY --from=builder /workspace/scorecard /scorecard

# Copy a test policy for local testing.
COPY --from=builder /workspace/policies/template.yml /policy.yml

# Our entry point.
# Note: the file is executable in the repo
# and permission carry over to the image.
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
