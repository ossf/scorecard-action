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

	"github.com/caarlos0/env/v6"

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
	// Scorecard options.
	ScorecardOpts *scopts.Options

	// Scorecard command-line options.
	ScorecardBin  string `env:"SCORECARD_BIN"`
	EnabledChecks string `env:"ENABLED_CHECKS"`
	PolicyFile    string `env:"SCORECARD_POLICY_FILE"`
	Format        string `env:"SCORECARD_RESULTS_FORMAT"`
	ResultsFile   string `env:"SCORECARD_RESULTS_FILE"`
	// TODO(options): This may be better as a bool
	PublishResultsStr string `env:"SCORECARD_PUBLISH_RESULTS"`

	// Input options.
	// TODO(options): These input options shadow the some of the SCORECARD_
	//                env vars:
	//                export SCORECARD_RESULTS_FILE="$INPUT_RESULTS_FILE"
	//                export SCORECARD_RESULTS_FORMAT="$INPUT_RESULTS_FORMAT"
	//                export SCORECARD_PUBLISH_RESULTS="$INPUT_PUBLISH_RESULTS"
	//
	//                Let's target them for deletion, but only after confirming
	//                that there isn't anything that surprisingly needs them.
	InputResultsFile    string `env:"INPUT_RESULTS_FILE"`
	InputResultsFormat  string `env:"INPUT_RESULTS_FORMAT"`
	InputPublishResults string `env:"INPUT_PUBLISH_RESULTS"`

	// Scorecard checks.
	EnableSarif             string `env:"ENABLE_SARIF"`
	EnableLicense           string `env:"ENABLE_LICENSE"`
	EnableDangerousWorkflow string `env:"ENABLE_DANGEROUS_WORKFLOW"`

	// GitHub options.
	// TODO(github): Consider making this a separate options set so we can
	//               encapsulate handling
	GithubEventName      string `env:"GITHUB_EVENT_NAME"`
	CheckGithubEventPath string `env:"GITHUB_EVENT_PATH"`
	GithubRef            string `env:"GITHUB_REF"`
	GithubRepository     string `env:"GITHUB_REPOSITORY"`
	GithubWorkspace      string `env:"GITHUB_WORKSPACE"`

	DefaultBranch string `env:"SCORECARD_DEFAULT_BRANCH"`
	// TODO(options): This may be better as a bool
	IsForkStr string `env:"SCORECARD_IS_FORK"`
	// TODO(options): This may be better as a bool
	PrivateRepoStr string `env:"SCORECARD_PRIVATE_REPOSITORY"`
}

const (
	defaultScorecardBin        = "/scorecard"
	defaultScorecardPolicyFile = "./policy.yml"
)

// New TODO(lint): should have comment or be unexported (revive).
func New() (*Options, error) {
	opts := &Options{}
	if err := env.Parse(opts); err != nil {
		return opts, fmt.Errorf("parsing entrypoint env vars: %w", err)
	}

	// TODO(options): Push options into scorecard.Options once/if it supports
	//                validation.
	scOpts := scopts.New()

	// TODO(options): Move this set-or-default logic to its own function.
	scOpts.PolicyFile = opts.PolicyFile
	if scOpts.PolicyFile == "" {
		scOpts.PolicyFile = defaultScorecardPolicyFile
	}

	if opts.ScorecardBin == "" {
		opts.ScorecardBin = defaultScorecardBin
	}

	opts.ScorecardOpts = scOpts

	if opts.ResultsFile == "" {
		opts.ResultsFile = opts.InputResultsFile
		// TODO(options): We should check if this is empty.
	}

	if opts.Format == "" {
		opts.Format = opts.InputResultsFormat
	}
	opts.ScorecardOpts.Format = opts.Format

	if opts.PublishResultsStr == "" {
		opts.PublishResultsStr = opts.InputPublishResults
		if opts.PublishResultsStr == "" {
			opts.PublishResultsStr = "false"
		}
	}

	// TODO(options): Consider running Initialize() before returning.
	// TODO(options): Consider running Validate() before returning.
	return opts, nil
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

	o.EnableSarif = "1"
	o.EnableLicense = "1"
	o.EnableDangerousWorkflow = "1"

	return CheckGithubEventPath()
}

// Validate validates the scorecard configuration.
func (o *Options) Validate(writer io.Writer) error {
	if os.Getenv(EnvGithubAuthToken) == "" {
		fmt.Fprintf(writer, "The 'repo_token' variable is empty.\n")
		if o.IsForkStr == trueStr {
			fmt.Fprintf(writer, "We have detected you are running on a fork.\n")
		}

		fmt.Fprintf(
			writer,
			"Please follow the instructions at https://github.com/ossf/scorecard-action#authentication to create the read-only PAT token.\n", //nolint:lll
		)

		return ErrEmptyGitHubAuthToken
	}

	if strings.Contains(os.Getenv(o.GithubEventName), "pull_request") &&
		os.Getenv(o.GithubRef) == o.DefaultBranch {
		fmt.Fprintf(writer, "%s not supported with %s event.\n", os.Getenv(o.GithubRef), os.Getenv(o.GithubEventName))
		fmt.Fprintf(writer, "Only the default branch %s is supported.\n", o.DefaultBranch)

		return errOnlyDefaultBranchSupported
	}

	return nil
}

// CheckRequired TODO(lint): should have comment or be unexported (revive).
func (o *Options) CheckRequired() error {
	err := CheckRequiredEnv()
	if err != nil {
		return fmt.Errorf("checking if required env vars are set: %w", err)
	}

	return nil
}

// Print is a function to print options.
func (o *Options) Print(writer io.Writer) {
	EnvPrint(writer)
}

// SetRepository TODO(lint): should have comment or be unexported (revive).
func (o *Options) SetRepository() {
	o.ScorecardOpts.Repo = os.Getenv(o.GithubRepository)
}

// Repo TODO(lint): should have comment or be unexported (revive).
func (o *Options) Repo() string {
	return o.ScorecardOpts.Repo
}

// RepoIsSet TODO(lint): should have comment or be unexported (revive).
func (o *Options) RepoIsSet() bool {
	return o.Repo() != ""
}

// SetRepoVisibility sets the repository's visibility.
func (o *Options) SetRepoVisibility(privateRepo bool) {
	o.PrivateRepoStr = strconv.FormatBool(privateRepo)
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
	isPrivateRepo := o.PrivateRepoStr
	if isPrivateRepo == trueStr || isPrivateRepo == "" {
		o.PublishResultsStr = "false"
	} else {
		o.PublishResultsStr = trueStr
	}
}

// GetGithubToken retrieves the GitHub auth token from the environment.
func GetGithubToken() string {
	return os.Getenv(EnvGithubAuthToken)
}

// GetGithubWorkspace retrieves the GitHub auth token from the environment.
func GetGithubWorkspace() string {
	return os.Getenv(EnvGithubWorkspace)
}

// CheckGithubEventPath gets the path to the GitHub event and sets the
// SCORECARD_IS_FORK environment variable.
// TODO(options): Check if this actually needs to be exported.
func CheckGithubEventPath() error {
	var result string
	var exists bool

	if result, exists = os.LookupEnv(EnvGithubEventPath); !exists {
		return ErrGitHubEventPathNotSet
	}

	if result == "" {
		return ErrGitHubEventPathEmpty
	}

	data, err := ioutil.ReadFile(result)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", EnvGithubEventPath, err)
	}

	isFork, err := RepoIsFork(string(data))
	if err != nil {
		return fmt.Errorf("error checking if scorecard is a fork: %w", err)
	}

	isForkStr := strconv.FormatBool(isFork)
	if err := os.Setenv(EnvScorecardFork, isForkStr); err != nil {
		return fmt.Errorf("error setting %s: %w", EnvScorecardFork, err)
	}

	return err
}

// RepoIsFork checks if the current repo is a fork.
func RepoIsFork(ghEventPath string) (bool, error) {
	if ghEventPath == "" {
		return false, ErrGitHubEventPath
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
