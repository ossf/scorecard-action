// Copyright 2022 OpenSSF Authors
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
//
// SPDX-License-Identifier: Apache-2.0

package install

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/google/go-github/v42/github"

	scagh "github.com/ossf/scorecard-action/install/github"
)

const workflowFile = ".github/workflows/scorecards-analysis.yml"

// Options are installation options for the scorecard action.
type Options struct {
	Owner string

	// Repositories
	Repositories []string
}

// NewOptions creates a new instance of installation options.
func NewOptions() *Options {
	return &Options{}
}

// Run adds the OpenSSF Scorecard workflow to all repositories under the given
// organization.
// TODO(install): Improve description.
// TODO(install): Accept a context instead of setting one.
func Run(o *Options) {
	// Get github user client.
	ctx := context.Background()
	gh := scagh.New()
	client := gh.Client()

	// If not provided, get all repositories under organization.
	if len(o.Repositories) == 0 {
		repos, _, err := client.GetRepositoriesByOrg(ctx, o.Owner)
		errCheck(err, "Error listing organization's repos.")

		// Convert to list of repository names.
		for _, repo := range repos {
			o.Repositories = append(o.Repositories, *repo.Name)
		}
	}

	// Get yml file into byte array.
	workflowContent, err := ioutil.ReadFile("scorecards-analysis.yml")
	errCheck(err, "Error reading in scorecard workflow file.")

	// Process each repository.
	for _, repoName := range o.Repositories {
		// Get repo metadata.
		repo, _, err := client.GetRepository(ctx, o.Owner, repoName)
		if err != nil {
			fmt.Println(
				"Skipped repo",
				repoName,
				"because it does not exist or could not be accessed.",
			)

			continue
		}

		// Get head commit SHA of default branch.
		defaultBranch, _, err := client.GetBranch(
			ctx,
			o.Owner,
			repoName,
			*repo.DefaultBranch,
			true,
		)
		if err != nil {
			fmt.Println(
				"Skipped repo",
				repoName,
				"because it's default branch could not be accessed.",
			)

			continue
		}

		defaultBranchSHA := defaultBranch.Commit.SHA

		// Skip if scorecard file already exists in workflows folder.
		scoreFileContent, _, _, err := client.GetContents(
			ctx,
			o.Owner,
			repoName,
			workflowFile,
			&github.RepositoryContentGetOptions{},
		)
		if scoreFileContent != nil || err == nil {
			fmt.Println(
				"Skipped repo",
				repoName,
				"since scorecard workflow already exists.",
			)

			continue
		}

		// Skip if branch scorecard already exists.
		scorecardBranch, _, err := client.GetBranch(
			ctx,
			o.Owner,
			repoName,
			"scorecard",
			true,
		)
		if scorecardBranch != nil || err == nil {
			fmt.Println(
				"Skipped repo",
				repoName,
				"since branch scorecard already exists.",
			)

			continue
		}

		// Create new branch using a reference that stores the new commit hash.
		ref := &github.Reference{
			Ref:    github.String("refs/heads/scorecard"),
			Object: &github.GitObject{SHA: defaultBranchSHA},
		}
		_, _, err = client.CreateGitRef(ctx, o.Owner, repoName, ref)
		if err != nil {
			fmt.Println(
				"Skipped repo",
				repoName,
				"because new branch could not be created.",
			)

			continue
		}

		// Create file in repository.
		opts := &github.RepositoryContentFileOptions{
			Message: github.String("Adding scorecard workflow"),
			Content: workflowContent,
			Branch:  github.String("scorecard"),
		}
		_, _, err = client.CreateFile(
			ctx,
			o.Owner,
			repoName,
			workflowFile,
			opts,
		)
		if err != nil {
			fmt.Println(
				"Skipped repo",
				repoName,
				"because new file could not be created.",
			)

			continue
		}

		// Create Pull request.
		_, err = client.CreatePullRequest(
			ctx,
			o.Owner,
			repoName,
			*defaultBranch.Name,
			"scorecard",
			"Added Scorecard Workflow",
			"Added the workflow for OpenSSF's Security Scorecard",
		)
		if err != nil {
			fmt.Println(
				"Skipped repo",
				repoName,
				"because pull request could not be created.",
			)

			continue
		}

		// Logging.
		fmt.Println(
			"Successfully added scorecard workflow PR from scorecard to",
			*defaultBranch.Name,
			"branch of repo",
			repoName,
		)
	}
}

func errCheck(err error, msg string) {
	if err != nil {
		fmt.Println(msg, err)
	}
}
