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

FROM gcr.io/openssf/scorecard:v4.4.0@sha256:970eabefcbeed21c8ac9ff4584033a12ee8ddd48b2b2725b35ac67721f820f39 as base

# Build our image and update the root certs.
# TODO: use distroless.
FROM debian:11.3-slim@sha256:06a93cbdd49a265795ef7b24fe374fee670148a7973190fb798e43b3cf7c5d0f
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    jq ca-certificates curl

# Copy the scorecard binary from the official scorecard image.
COPY --from=base /scorecard /scorecard

# Copy a test policy for local testing.
COPY policies/template.yml  /policy.yml

# Our entry point.
# Note: the file is executable in the repo
# and permission carry over to the image.
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
