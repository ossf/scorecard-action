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

package dependencydiff

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v45/github"
	"github.com/ossf/scorecard-action/options"
	"github.com/ossf/scorecard/v4/clients/githubrepo/roundtripper"
	"github.com/ossf/scorecard/v4/dependencydiff"
	"github.com/ossf/scorecard/v4/log"
	"github.com/ossf/scorecard/v4/pkg"
)

// New creates a new instance running the scorecard dependency-diff mode
// used as an entrypoint for GitHub Actions.
func New(ctx context.Context) error {
	repoURI := os.Getenv(options.EnvGithubRepository)
	ownerRepo := strings.Split(repoURI, "/")
	if len(ownerRepo) != 2 {
		return fmt.Errorf("%w: repo uri", errInvalid)
	}
	// Since the event listener is set to pull requests to main, this will be the main branch reference.
	base := os.Getenv(options.EnvGithubBaseRef)
	if base == "" {
		return fmt.Errorf("%w: base ref", errEmpty)
	}
	// The head reference of the pull request source branch.
	head := os.Getenv(options.EnvGitHubHeadRef)
	if head == "" {
		return fmt.Errorf("%w: head ref", errEmpty)
	}
	// GetDependencyDiffResults will handle the error checking of checks.
	checks := strings.Split(os.Getenv(options.EnvInputChecks), ",")
	changeTypes := strings.Split(os.Getenv(options.EnvInputChangeTypes), ",")
	changeTypeMap := map[pkg.ChangeType]bool{}
	for _, ct := range changeTypes {
		key := pkg.ChangeType(ct)
		if !key.IsValid() {
			return fmt.Errorf("%w: change type", errInvalid)
		}
		changeTypeMap[key] = true
	}
	deps, err := dependencydiff.GetDependencyDiffResults(
		ctx, repoURI, base, head, checks, changeTypeMap,
	)
	if err != nil {
		return fmt.Errorf("error getting dependency-diff: %w", err)
	}

	// Generate a markdown string using the dependency-diffs and write it to the pull request comment.
	report, err := dependencydiffResultsAsMarkdown(deps, base, head)
	if err != nil {
		return fmt.Errorf("error formatting results as markdown: %w", err)
	}
	logger := log.NewLogger(log.DefaultLevel)
	ghrt := roundtripper.NewTransport(ctx, logger) /* This round tripper handles the access token. */
	ghClient := github.NewClient(&http.Client{Transport: ghrt})
	err = writeToComment(ctx, ghClient, ownerRepo[0], ownerRepo[1], report)
	if err != nil {
		return fmt.Errorf("error writting the report to comment: %w", err)
	}

	return nil
}
