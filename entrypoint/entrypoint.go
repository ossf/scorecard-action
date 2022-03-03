// Copyright 2022 Security Scorecard Authors
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

package entrypoint

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ossf/scorecard-action/options"
	"github.com/ossf/scorecard/v4/cmd"
	scopts "github.com/ossf/scorecard/v4/options"
)

// Errors.
var errEmptyScorecardBin = errors.New("scorecard_bin variable is empty")

type repo struct {
	DefaultBranch string `json:"default_branch"`
	Private       bool   `json:"private"`
}

// New creates a new scorecard command which can be used as an entrypoint for
// GitHub Actions.
func New() *cobra.Command {
	opts := options.New()
	opts.Initialize()
	scOpts := opts.ScorecardOpts

	actionCmd := cmd.New(scOpts)

	actionCmd.Flags().StringVar(
		&scOpts.ResultsFile,
		"output-file",
		scOpts.ResultsFile,
		"path to output results to",
	)

	hiddenFlags := []string{
		scopts.FlagNPM,
		scopts.FlagPyPI,
		scopts.FlagRubyGems,
	}
	for _, f := range hiddenFlags {
		actionCmd.Flags().MarkHidden(f)
	}

	return actionCmd
}

// Run is the entrypoint for the action.
func Run(o *options.Options) error {
	if err := o.Initialize(); err != nil {
		return fmt.Errorf("initializing options: %w", err)
	}

	if err := o.CheckRequired(); err != nil {
		return fmt.Errorf("checking if required options are set: %w", err)
	}

	// The repository should have already been initialized, so if for whatever
	// reason it hasn't, we should exit here with an appropriate error
	if o.RepoIsSet() {
		return fmt.Errorf("repository cannot be empty") //nolint:goerr113 // TODO(lint): Fix
	}

	token := options.GetGithubToken()
	repo, err := getRepo(o.Repo(), token)
	if err != nil {
		return fmt.Errorf("getting repository information: %w", err)
	}

	err = o.SetDefaultBranch(repo.DefaultBranch)
	if err != nil {
		return fmt.Errorf("setting default branch: %w", err)
	}

	o.SetRepoVisibility(repo.Private)
	o.SetPublishResults()

	o.Print(os.Stdout)

	if err := o.Validate(os.Stderr); err != nil {
		return fmt.Errorf("validating options: %w", err)
	}

	// gets the cmd run settings
	cmd, err := getScorecardCmd(o)
	if err != nil {
		return err
	}

	cmd.Dir = options.GetGithubWorkspace()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running scorecard command: %w", err)
	}

	results, err := ioutil.ReadFile(o.ScorecardOpts.ResultsFile)
	if err != nil {
		return fmt.Errorf("reading results file: %w", err)
	}

	fmt.Println(string(results))

	return nil
}

// getRepo is a function to get the repository information.
// It is decided to not use the golang GitHub library because of the
// dependency on the github.com/google/go-github/github library
// which will in turn require other dependencies.
func getRepo(name, token string) (repo, error) {
	var r repo
	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.github.com/repos/%s", name), nil)
	if err != nil {
		return r, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Authorization", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return r, fmt.Errorf("error creating request: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return r, fmt.Errorf("error reading response body: %w", err)
	}

	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return r, fmt.Errorf("error decoding response body: %w", err)
	}

	return r, nil
}

func getScorecardCmd(o *options.Options) (*exec.Cmd, error) {
	/*
		if o.ScorecardBin == "" {
			return nil, errEmptyScorecardBin
		}
	*/
	var result exec.Cmd
	//result.Path = o.ScorecardBin

	// if pull_request
	if strings.Contains(o.GithubEventName, "pull_request") {
		// empty policy file
		if o.ScorecardOpts.PolicyFile == "" {
			result.Args = []string{
				"--local",
				".",
				"--format",
				o.ScorecardOpts.Format,
				"--show-details",
				">",
				o.ScorecardOpts.ResultsFile,
			}
			return &result, nil
		}

		result.Args = []string{
			"--local",
			".",
			"--format",
			o.ScorecardOpts.Format,
			"--policy",
			o.ScorecardOpts.PolicyFile,
			"--show-details",
			">",
			o.ScorecardOpts.ResultsFile,
		}
		return &result, nil
	}

	var enabledChecks string
	if o.GithubEventName == "branch_protection_rule" {
		enabledChecks = "--checks Branch-Protection"
	}

	if o.ScorecardOpts.PolicyFile == "" {
		result.Args = []string{
			"--repo",
			o.ScorecardOpts.Repo,
			"--format",
			o.ScorecardOpts.Format,
			enabledChecks,
			"--show-details",
			">",
			o.ScorecardOpts.ResultsFile,
		}
		return &result, nil
	}

	result.Args = []string{
		"--repo",
		o.ScorecardOpts.Repo,
		"--format",
		o.ScorecardOpts.Format,
		enabledChecks,
		"--policy",
		o.ScorecardOpts.PolicyFile,
		"--show-details",
		">",
		o.ScorecardOpts.ResultsFile,
	}

	return &result, nil
}
