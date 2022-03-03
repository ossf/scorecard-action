// Copyright OpenSSF Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package options

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// Environment variables.
// TODO(env): Remove once environment variables are not used for config.
//nolint:revive,nolintlint
const (
	EnvEnableSarif             = "ENABLE_SARIF"
	EnvEnableLicense           = "ENABLE_LICENSE"
	EnvEnableDangerousWorkflow = "ENABLE_DANGEROUS_WORKFLOW"
	EnvGithubEventPath         = "GITHUB_EVENT_PATH"
	EnvGithubEventName         = "GITHUB_EVENT_NAME"
	EnvGithubRepository        = "GITHUB_REPOSITORY"
	EnvGithubRef               = "GITHUB_REF"
	EnvGithubWorkspace         = "GITHUB_WORKSPACE"
	EnvGithubAuthToken         = "GITHUB_AUTH_TOKEN" //nolint:gosec
	EnvInputResultsFile        = "INPUT_RESULTS_FILE"
	EnvInputResultsFormat      = "INPUT_RESULTS_FORMAT"
	EnvInputPublishResults     = "INPUT_PUBLISH_RESULTS"
	EnvScorecardFork           = "SCORECARD_IS_FORK"
	EnvScorecardPrivateRepo    = "SCORECARD_PRIVATE_REPOSITORY"
)

// CheckRequiredEnv is a function to check if the required environment variables are set.
func CheckRequiredEnv() error {
	envVariables := make(map[string]bool)
	envVariables[EnvGithubRepository] = true
	envVariables[EnvGithubAuthToken] = true

	for key := range envVariables {
		// TODO(env): Refactor to use helpers
		if _, exists := os.LookupEnv(key); !exists {
			return errRequiredEnvNotSet
		}
	}

	return nil
}

// envPrint is a function to print the ENV variables.
func envPrint(writer io.Writer) {
	fmt.Fprintf(writer, "GITHUB_EVENT_PATH=%s\n", os.Getenv(EnvGithubEventPath))
	fmt.Fprintf(writer, "GITHUB_EVENT_NAME=%s\n", os.Getenv(EnvGithubEventName))
	fmt.Fprintf(writer, "GITHUB_REPOSITORY=%s\n", os.Getenv(EnvGithubRepository))
	fmt.Fprintf(writer, "SCORECARD_IS_FORK=%s\n", os.Getenv(EnvScorecardFork))
	fmt.Fprintf(writer, "Ref=%s\n", os.Getenv(EnvGithubRef))
}

// Adapted from sigs.k8s.io/release-utils/env

// TODO(env): Consider making these methods on an env var type.

// Lookup returns either the provided environment variable for the given key
// or the default value def if not set.
func Lookup(envVar, def string, mustExist, mustNotBeEmpty bool) (string, error) {
	value, ok := os.LookupEnv(envVar)
	if !ok {
		if mustExist {
			return value, errEnvVarNotSetWithKey(envVar)
		}
	}

	if value == "" {
		if mustNotBeEmpty {
			return value, errEnvVarIsEmptyWithKey(envVar)
		}

		return def, nil
	}

	return value, nil
}

// Errors

var (
	// Errors.
	// TODO(env): Determine if these errors actually need to be named.

	// ErrGitHubEventPath TODO(lint): should have comment or be unexported (revive).
	ErrGitHubEventPath = errors.New("error getting GITHUB_EVENT_PATH")
	// ErrGitHubEventPathEmpty TODO(lint): should have comment or be unexported (revive).
	ErrGitHubEventPathEmpty = errEnvVarIsEmptyWithKey(EnvGithubEventPath)
	// ErrGitHubEventPathNotSet TODO(lint): should have comment or be unexported (revive).
	ErrGitHubEventPathNotSet = errEnvVarNotSetWithKey(EnvGithubEventPath)
	// ErrEmptyGitHubAuthToken TODO(lint): should have comment or be unexported (revive).
	ErrEmptyGitHubAuthToken = errEnvVarIsEmptyWithKey(EnvGithubAuthToken)

	errEnvVarNotSet  = errors.New("env var is not set")
	errEnvVarIsEmpty = errors.New("env var is empty")

	errRequiredEnvNotSet = errors.New("required environment variables are not set")
)

func errEnvVarNotSetWithKey(envVar string) error {
	return fmt.Errorf("%w: %s", errEnvVarNotSet, envVar)
}

func errEnvVarIsEmptyWithKey(envVar string) error {
	return fmt.Errorf("%w: %s", errEnvVarIsEmpty, envVar)
}
