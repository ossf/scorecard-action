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

package env

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
	EnableSarif             = "ENABLE_SARIF"
	EnableLicense           = "ENABLE_LICENSE"
	EnableDangerousWorkflow = "ENABLE_DANGEROUS_WORKFLOW"
	GithubEventPath         = "GITHUB_EVENT_PATH"
	GithubEventName         = "GITHUB_EVENT_NAME"
	GithubRepository        = "GITHUB_REPOSITORY"
	GithubRef               = "GITHUB_REF"
	GithubWorkspace         = "GITHUB_WORKSPACE"
	GithubAuthToken         = "GITHUB_AUTH_TOKEN" //nolint:gosec
	InputResultsFile        = "INPUT_RESULTS_FILE"
	InputResultsFormat      = "INPUT_RESULTS_FORMAT"
	InputPublishResults     = "INPUT_PUBLISH_RESULTS"
	ScorecardFork           = "SCORECARD_IS_FORK"
	ScorecardPrivateRepo    = "SCORECARD_PRIVATE_REPOSITORY"
)

// CheckRequired is a function to check if the required environment variables are set.
func CheckRequired() error {
	envVariables := make(map[string]bool)
	envVariables[GithubRepository] = true
	envVariables[GithubAuthToken] = true

	for key := range envVariables {
		// TODO(env): Refactor to use helpers
		if _, exists := os.LookupEnv(key); !exists {
			return errRequiredEnvNotSet
		}
	}

	return nil
}

// Print is a function to print the ENV variables.
func Print(writer io.Writer) {
	fmt.Fprintf(writer, "GITHUB_EVENT_PATH=%s\n", os.Getenv(GithubEventPath))
	fmt.Fprintf(writer, "GITHUB_EVENT_NAME=%s\n", os.Getenv(GithubEventName))
	fmt.Fprintf(writer, "GITHUB_REPOSITORY=%s\n", os.Getenv(GithubRepository))
	fmt.Fprintf(writer, "SCORECARD_IS_FORK=%s\n", os.Getenv(ScorecardFork))
	fmt.Fprintf(writer, "Ref=%s\n", os.Getenv(GithubRef))
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
	ErrGitHubEventPathEmpty = errEnvVarIsEmptyWithKey(GithubEventPath)
	// ErrGitHubEventPathNotSet TODO(lint): should have comment or be unexported (revive).
	ErrGitHubEventPathNotSet = errEnvVarNotSetWithKey(InputPublishResults)
	// ErrEmptyGitHubAuthToken TODO(lint): should have comment or be unexported (revive).
	ErrEmptyGitHubAuthToken = errEnvVarIsEmptyWithKey(GithubAuthToken)

	errEnvVarNotSet  = errors.New("env var is not set")
	errEnvVarIsEmpty = errors.New("env var is empty")

	errRequiredEnvNotSet = errors.New("required environment variables are not set")
	// TODO(env): Remove if not needed.
	/*
		errInputResultFileNotSet     = errEnvVarNotSet(InputResultsFile)
		errInputResultFileEmpty      = errEnvVarIsEmpty(InputResultsFile)
		errInputResultFormatNotSet   = errEnvVarNotSet(InputResultsFormat)
		errInputResultFormatEmpty    = errEnvVarIsEmpty(InputResultsFormat)
		errInputPublishResultsNotSet = errEnvVarNotSet(InputPublishResults)
		errInputPublishResultsEmpty  = errEnvVarIsEmpty(InputPublishResults)
	*/
)

func errEnvVarNotSetWithKey(envVar string) error {
	return fmt.Errorf("%w: %s", errEnvVarNotSet, envVar)
}

func errEnvVarIsEmptyWithKey(envVar string) error {
	return fmt.Errorf("%w: %s", errEnvVarIsEmpty, envVar)
}
