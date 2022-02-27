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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/ossf/scorecard-action/env"
	scopts "github.com/ossf/scorecard/v4/options"
)

var (
	// Errors.
	errDefaultBranchEmpty         = errors.New("default branch is empty")
	errOnlyDefaultBranchSupported = errors.New("only default branch is supported")

	trueStr = "true"
)

// Options TODO(lint): should have comment or be unexported (revive).
type Options struct {
	ScorecardOpts   *scopts.Options
	GithubEventName string
	ScorecardBin    string
	DefaultBranch   string

	// TODO(options): This may be better as a bool
	PrivateRepo string
	// TODO(options): This may be better as a bool
	PublishResults string

	ResultsFile string
}

const (
	defaultScorecardBin        = "/scorecard"
	defaultScorecardPolicyFile = "./policy.yml"
)

// New TODO(lint): should have comment or be unexported (revive).
func New() *Options {
	scOpts := scopts.New()
	scOpts.PolicyFile = defaultScorecardPolicyFile

	// TODO: Populate options constructor
	opts := &Options{
		ScorecardOpts: scOpts,
		ScorecardBin:  defaultScorecardBin,
	}

	return opts
}

// Initialize initializes the environment variables required for the action.
func (o *Options) Initialize() error {
	/*
	 https://docs.github.com/en/actions/learn-github-actions/environment-variables
	   GITHUB_EVENT_PATH contains the json file for the event.
	   GITHUB_SHA contains the commit hash.
	   GITHUB_WORKSPACE contains the repo folder.
	   GITHUB_EVENT_NAME contains the event name.
	   GITHUB_ACTIONS is true in GitHub env.
	*/

	envvars := make(map[string]string)
	envvars[env.EnableSarif] = "1"
	envvars[env.EnableLicense] = "1"
	envvars[env.EnableDangerousWorkflow] = "1"

	for key, val := range envvars {
		if err := os.Setenv(key, val); err != nil {
			return fmt.Errorf("error setting %s: %w", key, err)
		}
	}

	err := setFromEnvVarStrict(&o.ResultsFile, env.InputResultsFile)
	if err != nil {
		return fmt.Errorf("setting %s: %w", o.ResultsFile, err)
	}

	err = setFromEnvVarStrict(&o.ScorecardOpts.Format, env.InputResultsFormat)
	if err != nil {
		return fmt.Errorf("setting %s: %w", o.ScorecardOpts.Format, err)
	}

	err = setFromEnvVarStrict(&o.PrivateRepo, env.ScorecardPrivateRepo)
	if err != nil {
		return fmt.Errorf("setting %s: %w", o.PrivateRepo, err)
	}

	err = setFromEnvVarStrict(&o.PublishResults, env.InputPublishResults)
	if err != nil {
		return fmt.Errorf("setting %s: %w", o.PublishResults, err)
	}

	return GithubEventPath()
}

// Validate validates the scorecard configuration.
func (o *Options) Validate(writer io.Writer) error {
	if os.Getenv(env.GithubAuthToken) == "" {
		fmt.Fprintf(writer, "The 'repo_token' variable is empty.\n")
		if os.Getenv(env.ScorecardFork) == trueStr {
			fmt.Fprintf(writer, "We have detected you are running on a fork.\n")
		}

		fmt.Fprintf(
			writer,
			"Please follow the instructions at https://github.com/ossf/scorecard-action#authentication to create the read-only PAT token.\n", //nolint:lll
		)

		return env.ErrEmptyGitHubAuthToken
	}

	if strings.Contains(os.Getenv(env.GithubEventName), "pull_request") &&
		os.Getenv(env.GithubRef) == o.DefaultBranch {
		fmt.Fprintf(writer, "%s not supported with %s event.\n", os.Getenv(env.GithubRef), os.Getenv(env.GithubEventName))
		fmt.Fprintf(writer, "Only the default branch %s is supported.\n", o.DefaultBranch)

		return errOnlyDefaultBranchSupported
	}

	return nil
}

// CheckRequired TODO(lint): should have comment or be unexported (revive).
func (o *Options) CheckRequired() error {
	err := env.CheckRequired()
	if err != nil {
		return fmt.Errorf("checking if required env vars are set: %w", err)
	}

	return nil
}

// Print is a function to print options.
func (o *Options) Print(writer io.Writer) {
	env.Print(writer)
}

// SetRepository TODO(lint): should have comment or be unexported (revive).
func (o *Options) SetRepository() {
	o.ScorecardOpts.Repo = os.Getenv(env.GithubRepository)
}

// Repo TODO(lint): should have comment or be unexported (revive).
func (o *Options) Repo() string {
	return o.ScorecardOpts.Repo
}

// SetRepoVisibility sets the repository's visibility.
func (o *Options) SetRepoVisibility(privateRepo bool) {
	o.PrivateRepo = strconv.FormatBool(privateRepo)
}

// SetDefaultBranch sets the default branch.
func (o *Options) SetDefaultBranch(defaultBranch string) error {
	if defaultBranch == "" {
		return errDefaultBranchEmpty
	}

	o.DefaultBranch = fmt.Sprintf("refs/heads/%s", defaultBranch)
	return nil
}

// SetPublishResults sets whether results should be published based on a
// repository's visibility.
func (o *Options) SetPublishResults() {
	isPrivateRepo := o.PrivateRepo
	if isPrivateRepo == trueStr || isPrivateRepo == "" {
		o.PublishResults = "false"
	} else {
		o.PublishResults = trueStr
	}
}

// GetGithubToken retrieves the GitHub auth token from the environment.
func GetGithubToken() string {
	return os.Getenv(env.GithubAuthToken)
}

// GetGithubWorkspace retrieves the GitHub auth token from the environment.
func GetGithubWorkspace() string {
	return os.Getenv(env.GithubWorkspace)
}

// GithubEventPath gets the path to the GitHub event and sets the
// SCORECARD_IS_FORK environment variable.
// TODO(options): Check if this actually needs to be exported.
func GithubEventPath() error {
	var result string
	var exists bool

	if result, exists = os.LookupEnv(env.GithubEventPath); !exists {
		return env.ErrGitHubEventPathNotSet
	}

	if result == "" {
		return env.ErrGitHubEventPathEmpty
	}

	data, err := ioutil.ReadFile(result)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", env.GithubEventPath, err)
	}

	isFork, err := RepoIsFork(string(data))
	if err != nil {
		return fmt.Errorf("error checking if scorecard is a fork: %w", err)
	}

	isForkStr := strconv.FormatBool(isFork)
	if err := os.Setenv(env.ScorecardFork, isForkStr); err != nil {
		return fmt.Errorf("error setting %s: %w", env.ScorecardFork, err)
	}

	return err
}

// RepoIsFork checks if the current repo is a fork.
func RepoIsFork(ghEventPath string) (bool, error) {
	if ghEventPath == "" {
		return false, env.ErrGitHubEventPath
	}
	/*
	 https://docs.github.com/en/actions/reference/workflow-commands-for-github-actions#github_repository_is_fork
	   GITHUB_REPOSITORY_IS_FORK is true if the repository is a fork.
	*/
	type repo struct {
		Repository struct {
			Fork bool `json:"fork"`
		} `json:"repository"`
	}
	var r repo
	if err := json.Unmarshal([]byte(ghEventPath), &r); err != nil {
		return false, fmt.Errorf("error unmarshalling ghEventPath: %w", err)
	}

	return r.Repository.Fork, nil
}

// setFromEnvVarStrict TODO(lint): should have comment or be unexported (revive).
func setFromEnvVarStrict(option *string, envVar string) error {
	return setFromEnvVar(option, envVar, "", true, true)
}

// TODO(env): Refactor
//            - Convert to method
//            - Only fail if both the config value and env var is empty.
func setFromEnvVar(option *string, envVar, def string, mustExist, mustNotBeEmpty bool) error {
	value, err := env.Lookup(envVar, def, mustExist, mustNotBeEmpty)
	if err != nil {
		return fmt.Errorf("setting value for option %s: %w", *option, err)
	}

	*option = value
	return nil
}
