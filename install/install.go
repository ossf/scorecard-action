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
	"golang.org/x/oauth2"
)

const (
	orgName = "organization name"
	pat     = "personal access token"
)

// RepoList leave empty to process all repos under org (optional).
var RepoList = []string{}

// Run adds the OpenSSF Scorecard workflow to all repositories under the given
// organization.
// TODO(install): Improve description.
func Run() {
	// Get github user client.
	ctx := context.Background()
	tokenService := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: pat},
	)

	tokenClient := oauth2.NewClient(ctx, tokenService)
	client := github.NewClient(tokenClient)

	// If not provided, get all repositories under organization.
	if len(RepoList) == 0 {
		lops := &github.RepositoryListByOrgOptions{Type: "all"}
		repos, _, err := client.Repositories.ListByOrg(ctx, orgName, lops)
		errCheck(err, "Error listing organization's repos.")

		// Convert to list of repository names.
		for _, repo := range repos {
			RepoList = append(RepoList, *repo.Name)
		}
	}

	// Get yml file into byte array.
	workflowContent, err := ioutil.ReadFile("scorecards-analysis.yml")
	errCheck(err, "Error reading in scorecard workflow file.")

	// Process each repository.
	for _, repoName := range RepoList {
		// Get repo metadata.
		repo, _, err := client.Repositories.Get(ctx, orgName, repoName)
		if err != nil {
			fmt.Println(
				"Skipped repo",
				repoName,
				"because it does not exist or could not be accessed.",
			)

			continue
		}

		// Get head commit SHA of default branch.
		defaultBranch, _, err := client.Repositories.GetBranch(
			ctx,
			orgName,
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
		scoreFileContent, _, _, err := client.Repositories.GetContents(
			ctx,
			orgName,
			repoName,
			".github/workflows/scorecards-analysis.yml",
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
		scorecardBranch, _, err := client.Repositories.GetBranch(
			ctx,
			orgName,
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
		_, _, err = client.Git.CreateRef(ctx, orgName, repoName, ref)
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
		_, _, err = client.Repositories.CreateFile(
			ctx,
			orgName,
			repoName,
			".github/workflows/scorecards-analysis.yml",
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
		pr := &github.NewPullRequest{
			Title: github.String("Added Scorecard Workflow"),
			Head:  github.String("scorecard"),
			Base:  github.String(*defaultBranch.Name),
			Body: github.String(
				"Added the workflow for OpenSSF's Security Scorecard",
			),
			Draft: github.Bool(false),
		}

		_, _, err = client.PullRequests.Create(ctx, orgName, repoName, pr)
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
