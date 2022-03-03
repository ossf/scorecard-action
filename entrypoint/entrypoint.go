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
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/ossf/scorecard-action/options"
	"github.com/ossf/scorecard/v4/cmd"
	scopts "github.com/ossf/scorecard/v4/options"
)

// TODO(github): Move to separate package.
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

	// Add sub-commands.
	actionCmd.AddCommand(printConfigCmd(opts))

	return actionCmd
}

func printConfigCmd(o *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use: "print-config",
		Run: func(cmd *cobra.Command, args []string) {
			o.Print()
		},
	}

	return cmd
}

// getRepo is a function to get the repository information.
// It is decided to not use the golang GitHub library because of the
// dependency on the github.com/google/go-github/github library
// which will in turn require other dependencies.
// TODO(github): Move to separate package.
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
