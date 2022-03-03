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

	"github.com/ossf/scorecard/v4/options"
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
	EnabledChecks string `env:"ENABLED_CHECKS"`

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
	EnableLicense           string `env:"ENABLE_LICENSE"`
	EnableDangerousWorkflow string `env:"ENABLE_DANGEROUS_WORKFLOW"`

	// GitHub options.
	// TODO(github): Consider making this a separate options set so we can
	//               encapsulate handling
	GithubEventName  string `env:"GITHUB_EVENT_NAME"`
	GithubEventPath  string `env:"GITHUB_EVENT_PATH"`
	GithubRef        string `env:"GITHUB_REF"`
	GithubRepository string `env:"GITHUB_REPOSITORY"`
	GithubWorkspace  string `env:"GITHUB_WORKSPACE"`

	DefaultBranch string `env:"SCORECARD_DEFAULT_BRANCH"`
	// TODO(options): This may be better as a bool
	IsForkStr string `env:"SCORECARD_IS_FORK"`
	// TODO(options): This may be better as a bool
	PrivateRepoStr string `env:"SCORECARD_PRIVATE_REPOSITORY"`
}

const (
	defaultScorecardPolicyFile = "policy.yml"
	formatSarif                = options.FormatSarif
)

// New TODO(lint): should have comment or be unexported (revive).
func New() *Options {
	opts := &Options{}
	if err := env.Parse(opts); err != nil {
		// TODO(options): Consider making this an error.
		fmt.Printf("parsing entrypoint env vars: %+v", err)
	}

	// TODO(options): Push options into scorecard.Options once/if it supports
	//                validation.
	scOpts := scopts.New()

	if err := opts.Initialize(); err != nil {
		// TODO(options): Consider making this an error.
		fmt.Printf("initializing scorecard-action options: %+v\n", err)
	}

	// TODO(options): Move this set-or-default logic to its own function.
	if opts.InputResultsFormat != "" {
		scOpts.Format = opts.InputResultsFormat
	} else {
		scOpts.EnableSarif = true
		scOpts.Format = formatSarif
		os.Setenv(options.EnvVarEnableSarif, trueStr)
		if scOpts.Format == "" {
			// Default the scorecard command to using SARIF format.
			if scOpts.PolicyFile == "" {
				// TODO(policy): Should we default or error here?
				scOpts.PolicyFile = defaultScorecardPolicyFile
			}
		}
	}

	if err := scOpts.Validate(); err != nil {
		// TODO(options): Consider making this an error.
		fmt.Printf("validating scorecard options: %+v\n", err)
	}

	opts.ScorecardOpts = scOpts
	opts.SetPublishResults()

	if opts.ScorecardOpts.PublishResults {
		if scOpts.ResultsFile == "" {
			scOpts.ResultsFile = opts.InputResultsFile
			// TODO(options): We should check if this is empty.
		}
	}

	// TODO(options): Consider running Validate() before returning.
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

	o.EnableLicense = "1"
	o.EnableDangerousWorkflow = "1"

	return o.SetRepoInfo()
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
func (o *Options) Print() {
	fmt.Printf("Event file: %s\n", o.GithubEventPath)
	fmt.Printf("Event name: %s\n", o.GithubEventName)
	fmt.Printf("Ref: %s\n", o.ScorecardOpts.Commit)
	fmt.Printf("Repository: %s\n", o.ScorecardOpts.Repo)
	fmt.Printf("Fork repository: %s\n", o.IsForkStr)
	fmt.Printf("Private repository: %s\n", o.PrivateRepoStr)
	fmt.Printf("Publication enabled: %+v\n", o.ScorecardOpts.PublishResults)
	fmt.Printf("Format: %s\n", o.ScorecardOpts.Format)
	fmt.Printf("Policy file: %s\n", o.ScorecardOpts.PolicyFile)
	fmt.Printf("Default branch: %s\n", o.DefaultBranch)
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
	// Check INPUT_PUBLISH_RESULTS
	var inputBool bool
	var inputErr error
	input := os.Getenv(EnvInputPublishResults)
	if input != "" {
		inputBool, inputErr = strconv.ParseBool(o.InputPublishResults)
		if inputErr != nil {
			// TODO(options): Consider making this an error.
			fmt.Printf(
				"could not parse a valid bool from %s (%s): %+v\n",
				input,
				EnvInputPublishResults,
				inputErr,
			)
		}
	}

	privateRepo, err := strconv.ParseBool(o.PrivateRepoStr)
	if err != nil {
		// TODO(options): Consider making this an error.
		fmt.Printf(
			"parsing bool from %s: %+v\n",
			o.PrivateRepoStr,
			err,
		)
	}

	if privateRepo {
		o.ScorecardOpts.PublishResults = false
	} else {
		o.ScorecardOpts.PublishResults = o.ScorecardOpts.PublishResults || inputBool
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
// TODO(options): Choose a more accurate name for what this does.
func (o *Options) SetRepoInfo() error {
	eventPath := o.GithubEventPath
	if eventPath == "" {
		return ErrGitHubEventPathEmpty
	}

	repoInfo, err := ioutil.ReadFile(eventPath)
	if err != nil {
		return fmt.Errorf("reading GitHub event path: %w", err)
	}

	/*
	 https://docs.github.com/en/actions/reference/workflow-commands-for-github-actions#github_repository_is_fork
	   GITHUB_REPOSITORY_IS_FORK is true if the repository is a fork.
	*/
	type repo struct {
		Repository struct {
			DefaultBranch string `json:"default_branch"`
			Fork          bool   `json:"fork"`
			Private       bool   `json:"private"`
		} `json:"repository"`
	}
	var r repo
	if err := json.Unmarshal([]byte(repoInfo), &r); err != nil {
		return fmt.Errorf("unmarshalling repo info: %w", err)
	}

	o.PrivateRepoStr = strconv.FormatBool(r.Repository.Private)
	o.IsForkStr = strconv.FormatBool(r.Repository.Fork)
	o.DefaultBranch = r.Repository.DefaultBranch

	return nil
}
