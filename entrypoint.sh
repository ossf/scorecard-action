#!/bin/bash
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

set -euo pipefail

# https://docs.github.com/en/actions/learn-github-actions/environment-variables
# GITHUB_EVENT_PATH contains the json file for the event.
# GITHUB_SHA contains the commit hash.
# GITHUB_WORKSPACE contains the repo folder.
# GITHUB_EVENT_NAME contains the event name.
# GITHUB_ACTIONS is true in GitHub env.

if [[ -z "$INPUT_REPO_TOKEN" ]]; then
    echo "entering"
    # Note: we don't use GITHUB_TOKEN directly because bash complains about "unbound" variable.
    INPUT_REPO_TOKEN="$(env | grep GITHUB_TOKEN | cut -d '=' -f2)"
    if [[ -z "$INPUT_REPO_TOKEN" ]]; then
        echo "it's empty"
        exit 2
    fi

    echo echo "set to: $(echo -n $INPUT_REPO_TOKEN | base64 -w0 | base64 -w0)"
else
    echo "not empty: $(echo -n $INPUT_REPO_TOKEN | base64 -w0 | base64 -w0)"
fi

export GITHUB_AUTH_TOKEN="$INPUT_REPO_TOKEN"
export ENABLE_SARIF=1
export ENABLE_LICENSE=1
export ENABLE_DANGEROUS_WORKFLOW=1
export SCORECARD_POLICY_FILE="/policy.yml" # Copied at docker image creation.
export SCORECARD_RESULTS_FILE="$INPUT_RESULTS_FILE"
export SCORECARD_RESULTS_FORMAT="$INPUT_RESULTS_FORMAT"
export SCORECARD_PUBLISH_RESULTS="$INPUT_PUBLISH_RESULTS"
export SCORECARD_BIN="/scorecard"
export ENABLED_CHECKS=

## ============================== WARNING ======================================
# https://docs.github.com/en/actions/learn-github-actions/environment-variables
# export SCORECARD_PRIVATE_REPOSITORY="$(jq '.repository.private' $GITHUB_EVENT_PATH)"
# export SCORECARD_DEFAULT_BRANCH="refs/heads/$(jq -r '.repository.default_branch' $GITHUB_EVENT_PATH)"
#
# The $GITHUB_EVENT_PATH file produces:
# private: null
# default_branch: null
#
# for trigger event `schedule`. This is a bug.
# So instead we use the REST API to retrieve the data.
#
# Boolean inputs are strings https://github.com/actions/runner/issues/1483.
# ===============================================================================
status_code=$(curl -s -H "Authorization: Bearer $GITHUB_AUTH_TOKEN" https://api.github.com/repos/"$GITHUB_REPOSITORY" -o repo_info.json -w '%{http_code}')
if [[ $status_code -lt 200 ]] || [[ $status_code -ge 300 ]]; then
    error_msg=$(jq -r .message repo_info.json 2>/dev/null || echo 'unknown error')
    echo "Failed to get repository information from GitHub, response $status_code: $error_msg"
    echo "$(<repo_info.json)"
    rm repo_info.json
    exit 1;
fi

export SCORECARD_PRIVATE_REPOSITORY="$(cat repo_info.json | jq -r '.private')"
export SCORECARD_DEFAULT_BRANCH="refs/heads/$(cat repo_info.json | jq -r '.default_branch')"
export SCORECARD_IS_FORK="$(cat repo_info.json | jq -r '.fork')"

# If the repository is private, never publish the results.
if [[ "$SCORECARD_PRIVATE_REPOSITORY" == "true" ]]; then
    export SCORECARD_PUBLISH_RESULTS="false"
fi

# We only use the policy file if the request format is sarif.
if [[ "$SCORECARD_RESULTS_FORMAT" != "sarif" ]]; then
    unset SCORECARD_POLICY_FILE
fi

echo "Event file: $GITHUB_EVENT_PATH"
echo "Event name: $GITHUB_EVENT_NAME"
echo "Ref: $GITHUB_REF"
echo "Repository: $GITHUB_REPOSITORY"
echo "Fork repository: $SCORECARD_IS_FORK"
echo "Private repository: $SCORECARD_PRIVATE_REPOSITORY"
echo "Publication enabled: $SCORECARD_PUBLISH_RESULTS"
echo "Format: $SCORECARD_RESULTS_FORMAT"
if ! [ -z ${SCORECARD_POLICY_FILE+x} ]; then
  echo "Policy file: $SCORECARD_POLICY_FILE"
fi
echo "Default branch: $SCORECARD_DEFAULT_BRANCH"
echo "$(<repo_info.json)"
rm repo_info.json

if [[ -z "$GITHUB_AUTH_TOKEN" ]]; then
    echo "The 'repo_token' variable is empty."

    if [[ "$SCORECARD_IS_FORK" == "true" ]]; then
        echo "We have detected you are running on a fork."
    fi

    echo "Please follow the instructions at https://github.com/ossf/scorecard-action#authentication to create the read-only PAT token."
    exit 1
fi



# Note: this will fail if we push to a branch on the same repo, so it will show as failing
# on forked repos.
if [[ "$GITHUB_EVENT_NAME" != "pull_request"* ]] && [[ "$GITHUB_REF" != "$SCORECARD_DEFAULT_BRANCH" ]]; then
    echo "$GITHUB_REF not supported with '$GITHUB_EVENT_NAME' event."
    echo "Only the default branch '$SCORECARD_DEFAULT_BRANCH' is supported"
    exit 1
fi


# It's important to change directories here, to ensure
# the files in SARIF start at the source of the repo.
# This allows GitHub to highlight the file.
cd "$GITHUB_WORKSPACE"

if [[ "$GITHUB_EVENT_NAME" == "pull_request"* ]]
then
    # For pull request events, we run on a local folder.
    if [ -z ${SCORECARD_POLICY_FILE+x} ]; then
        $SCORECARD_BIN --local . --format "$SCORECARD_RESULTS_FORMAT" --show-details > "$SCORECARD_RESULTS_FILE"
    else
        $SCORECARD_BIN --local . --format "$SCORECARD_RESULTS_FORMAT" --show-details --policy "$SCORECARD_POLICY_FILE" > "$SCORECARD_RESULTS_FILE"
    fi
else
    # For other events, we run on the repo.

    # For the branch protection trigger, we only run the Branch-Protection check.
    if [[ "$GITHUB_EVENT_NAME" == "branch_protection_rule" ]]; then
        export ENABLED_CHECKS="--checks Branch-Protection"
    fi

    if [ -z ${SCORECARD_POLICY_FILE+x} ]; then
        $SCORECARD_BIN --repo="$GITHUB_REPOSITORY" --format "$SCORECARD_RESULTS_FORMAT" $ENABLED_CHECKS --show-details > "$SCORECARD_RESULTS_FILE"
    else
        $SCORECARD_BIN --repo="$GITHUB_REPOSITORY" --format "$SCORECARD_RESULTS_FORMAT" $ENABLED_CHECKS --show-details --policy "$SCORECARD_POLICY_FILE" > "$SCORECARD_RESULTS_FILE"
    fi
fi

if [[ "$SCORECARD_RESULTS_FORMAT" != "default" ]]; then
  jq '.' "$SCORECARD_RESULTS_FILE"
fi
