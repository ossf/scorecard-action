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
	"os"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v6"
	"golang.org/x/net/context"

	"github.com/ossf/scorecard-action/github"
	scopts "github.com/ossf/scorecard/v4/options"
)

const (
	defaultScorecardPolicyFile = "/policy.yml"
	trueStr                    = "true"
	formatSarif                = scopts.FormatSarif

	pullRequestEvent      = "pull_request"
	pushEvent             = "push"
	branchProtectionEvent = "branch_protection_rule"
)

var (
	// Errors.
	errGithubEventPathEmpty       = errors.New("GitHub event path is empty")
	errResultsPathEmpty           = errors.New("results path is empty")
	errGitHubRepoInfoUnavailable  = errors.New("GitHub repo info inaccessible")
	errOnlyDefaultBranchSupported = errors.New("only default branch is supported")
)

// Options are options for running scorecard via GitHub Actions.
type Options struct {
	// Scorecard options.
	ScorecardOpts *scopts.Options

	// Scorecard command-line options.
	EnabledChecks string `env:"ENABLED_CHECKS"`

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
	GithubAPIURL     string `env:"GITHUB_API_URL"`

	DefaultBranch string `env:"SCORECARD_DEFAULT_BRANCH"`
	// TODO(options): This may be better as a bool
	IsForkStr string `env:"SCORECARD_IS_FORK"`
	// TODO(options): This may be better as a bool
	PrivateRepoStr string `env:"SCORECARD_PRIVATE_REPOSITORY"`

	// Input parameters
	InputResultsFile   string `env:"INPUT_RESULTS_FILE"`
	InputResultsFormat string `env:"INPUT_RESULTS_FORMAT"`

	PublishResults bool
}

// New creates a new options set for running scorecard via GitHub Actions.
func New() (*Options, error) {
	opts := &Options{}
	if err := env.Parse(opts); err != nil {
		return opts, fmt.Errorf("parsing entrypoint env vars: %w", err)
	}
	// GITHUB_AUTH_TOKEN
	// Needs to be set *before* setRepoInfo() is invoked.
	// setRepoInfo() uses the GITHUB_AUTH_TOKEN env for querying the REST API.
	if _, tokenSet := os.LookupEnv(EnvGithubAuthToken); !tokenSet {
		inputToken := os.Getenv(EnvInputRepoToken)
		os.Setenv(EnvGithubAuthToken, inputToken)
	}
	if err := opts.setRepoInfo(); err != nil {
		return opts, fmt.Errorf("parsing repo info: %w", err)
	}
	opts.setScorecardOpts()
	opts.setPublishResults()
	return opts, nil
}

// Validate validates the scorecard configuration.
func (o *Options) Validate() error {
	fmt.Println("EnvGithubAuthToken:", EnvGithubAuthToken, os.Getenv(EnvGithubAuthToken))
	if os.Getenv(EnvGithubAuthToken) == "" {
		fmt.Printf("%s variable is empty.\n", EnvGithubAuthToken)
		if o.IsForkStr == trueStr {
			fmt.Printf("We have detected you are running on a fork.\n")
		}

		fmt.Printf(
			"Please follow the instructions at https://github.com/ossf/scorecard-action#authentication to create the read-only PAT token.\n", //nolint:lll
		)

		return errEmptyGitHubAuthToken
	}

	if !o.isPullRequestEvent() &&
		!o.isDefaultBranch() {
		fmt.Printf("%s not supported with %s event.\n", o.GithubRef, o.GithubEventName)
		fmt.Printf("Only the default branch %s is supported.\n", o.DefaultBranch)

		return errOnlyDefaultBranchSupported
	}
	if err := o.ScorecardOpts.Validate(); err != nil {
		return fmt.Errorf("validating scorecard options: %w", err)
	}
	if o.ScorecardOpts.ResultsFile == "" {
		// TODO(test): Reassess test case for this code path
		return errResultsPathEmpty
	}
	return nil
}

// Print is a function to print options.
func (o *Options) Print() {
	// Scorecard options
	fmt.Println("Scorecard options:")
	fmt.Printf("Ref: %s\n", o.ScorecardOpts.Commit)
	fmt.Printf("Repository: %s\n", o.ScorecardOpts.Repo)
	fmt.Printf("Local: %s\n", o.ScorecardOpts.Local)
	fmt.Printf("Format: %s\n", o.ScorecardOpts.Format)
	fmt.Printf("Policy file: %s\n", o.ScorecardOpts.PolicyFile)
	fmt.Println()
	fmt.Println("Event / repo information:")
	fmt.Printf("Event file: %s\n", o.GithubEventPath)
	fmt.Printf("Event name: %s\n", o.GithubEventName)
	fmt.Printf("Fork repository: %s\n", o.IsForkStr)
	fmt.Printf("Private repository: %s\n", o.PrivateRepoStr)
	fmt.Printf("Publication enabled: %+v\n", o.PublishResults)
	fmt.Printf("Default branch: %s\n", o.DefaultBranch)
}

func (o *Options) setScorecardOpts() {
	o.ScorecardOpts = scopts.New()
	// Set GITHUB_AUTH_TOKEN
	inputToken := os.Getenv(EnvInputRepoToken)
	if inputToken == "" {
		fmt.Printf("The 'repo_token' variable is empty.\n")
		fmt.Printf("Using the '%s' variable instead.\n", EnvInputInternalRepoToken)
		inputToken := os.Getenv(EnvInputInternalRepoToken)
		os.Setenv(EnvGithubAuthToken, inputToken)
	}

	// --repo= | --local
	// This section restores functionality that was removed in
	// https://github.com/ossf/scorecard/pull/1898.
	// TODO(options): Consider moving this to its own function.
	if !o.isPullRequestEvent() {
		o.ScorecardOpts.Repo = o.GithubRepository
	} else {
		o.ScorecardOpts.Local = "."
	}

	// --format=
	// Enable scorecard command to use SARIF format (default format).
	os.Setenv(scopts.EnvVarEnableSarif, trueStr)
	o.ScorecardOpts.EnableSarif = true
	o.ScorecardOpts.Format = formatSarif
	if o.InputResultsFormat != "" {
		o.ScorecardOpts.Format = o.InputResultsFormat
	}
	if o.ScorecardOpts.Format == formatSarif && o.ScorecardOpts.PolicyFile == "" {
		// TODO(policy): Should we default or error here?
		o.ScorecardOpts.PolicyFile = defaultScorecardPolicyFile
	}

	// --show-details
	o.ScorecardOpts.ShowDetails = true

	// --commit=
	// TODO(scorecard): Reset commit options. Fix this in scorecard.
	o.ScorecardOpts.Commit = scopts.DefaultCommit

	// --out-file=
	if o.ScorecardOpts.ResultsFile == "" {
		o.ScorecardOpts.ResultsFile = o.InputResultsFile
	}
}

// setPublishResults sets whether results should be published based on a
// repository's visibility.
func (o *Options) setPublishResults() {
	inputVal := o.PublishResults
	o.PublishResults = false
	privateRepo, err := strconv.ParseBool(o.PrivateRepoStr)
	if err != nil {
		// TODO(options): Consider making this an error.
		fmt.Printf(
			"parsing bool from %s: %+v\n",
			o.PrivateRepoStr,
			err,
		)
		return
	}

	o.PublishResults = inputVal && !privateRepo
}

// setRepoInfo gets the path to the GitHub event and sets the
// SCORECARD_IS_FORK environment variable.
// TODO(options): Check if this actually needs to be exported.
// TODO(options): Choose a more accurate name for what this does.
func (o *Options) setRepoInfo() error {
	eventPath := o.GithubEventPath
	if eventPath == "" {
		return errGithubEventPathEmpty
	}

	ghClient := github.NewClient(context.Background())
	if repoInfo, err := ghClient.ParseFromFile(eventPath); err == nil &&
		o.parseFromRepoInfo(repoInfo) {
		return nil
	}

	if repoInfo, err := ghClient.ParseFromURL(o.GithubAPIURL, o.GithubRepository); err == nil &&
		o.parseFromRepoInfo(repoInfo) {
		return nil
	}

	return errGitHubRepoInfoUnavailable
}

func (o *Options) parseFromRepoInfo(repoInfo github.RepoInfo) bool {
	if repoInfo.Repo.DefaultBranch == nil &&
		repoInfo.Repo.Fork == nil &&
		repoInfo.Repo.Private == nil {
		return false
	}
	if repoInfo.Repo.Private != nil {
		o.PrivateRepoStr = strconv.FormatBool(*repoInfo.Repo.Private)
	}
	if repoInfo.Repo.Fork != nil {
		o.IsForkStr = strconv.FormatBool(*repoInfo.Repo.Fork)
	}
	if repoInfo.Repo.DefaultBranch != nil {
		o.DefaultBranch = *repoInfo.Repo.DefaultBranch
	}
	return true
}

func (o *Options) isPullRequestEvent() bool {
	return strings.HasPrefix(o.GithubEventName, pullRequestEvent)
}

func (o *Options) isDefaultBranch() bool {
	return o.GithubRef == fmt.Sprintf("refs/heads/%s", o.DefaultBranch)
}
