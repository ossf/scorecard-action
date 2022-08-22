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
	"log"
	"os"
	"path"

	"github.com/google/go-github/v46/github"

	scagh "github.com/ossf/scorecard-action/install/github"
	"github.com/ossf/scorecard-action/install/options"
)

const (
	commitMessage          = ".github: Add scorecard workflow"
	pullRequestBranch      = "scorecard-action-install"
	workflowBase           = ".github/workflows"
	workflowFile           = "scorecards.yml"
	workflowFileDeprecated = "scorecards-analysis.yml"
)

var (
	branchReference        = fmt.Sprintf("refs/heads/%s", pullRequestBranch)
	pullRequestDescription = `This pull request was generated using the installer tool for scorecard's GitHub Action.

To report any issues with this tool, see [here](https://github.com/ossf/scorecard-action).
`

	pullRequestTitle = commitMessage
	workflowFiles    = []string{
		path.Join(workflowBase, workflowFile),
		path.Join(workflowBase, workflowFileDeprecated),
	}
)

// Run adds the OpenSSF Scorecard workflow to all repositories under the given
// organization.
// TODO(install): Improve description.
// TODO(install): Accept a context instead of setting one.
func Run(o *options.Options) error {
	err := o.Validate()
	if err != nil {
		return fmt.Errorf("validating installation options: %w", err)
	}

	// Get github user client.
	ctx := context.Background()
	gh := scagh.New(ctx)

	// If not provided, get all repositories under organization.
	if len(o.Repositories) == 0 {
		log.Print("No repositories provided. Fetching all repositories under organization.")
		repos, _, err := gh.GetRepositoriesByOrg(ctx, o.Owner)
		if err != nil {
			return fmt.Errorf("getting repos for owner (%s): %w", o.Owner, err)
		}

		// Convert to list of repository names.
		for _, repo := range repos {
			o.Repositories = append(o.Repositories, *repo.Name)
		}
	}

	// Get yml file into byte array.
	workflowContent, err := os.ReadFile(o.ConfigPath)
	if err != nil {
		return fmt.Errorf("reading scorecard workflow file: %w", err)
	}

	// Process each repository.
	// TODO: Capture repo access errors
	for _, repoName := range o.Repositories {
		log.Printf("Processing repository: %s", repoName)
		err := processRepo(ctx, gh, o.Owner, repoName, workflowContent)
		if err != nil {
			log.Printf("processing repository: %+v", err)
		}

		log.Printf(
			"finished processing repository %s",
			repoName,
		)
	}

	return nil
}

func processRepo(
	ctx context.Context,
	gh *scagh.Client,
	owner, repoName string,
	workflowContent []byte,
) error {
	// Get repo metadata.
	log.Printf("getting repo metadata for %s", repoName)
	repo, _, err := gh.GetRepository(ctx, owner, repoName)
	if err != nil {
		return fmt.Errorf(
			"getting repository: %w",
			err,
		)
	}

	// Get head commit SHA of default branch.
	// TODO: Capture branch access errors
	defaultBranch, _, err := gh.GetBranch(
		ctx,
		owner,
		repoName,
		*repo.DefaultBranch,
		true,
	)
	if err != nil {
		return fmt.Errorf(
			"getting default branch for %s: %w",
			repoName,
			err,
		)
	}

	defaultBranchSHA := defaultBranch.Commit.SHA

	// Skip if scorecard file already exists in workflows folder.
	workflowExists := false
	for i, f := range workflowFiles {
		log.Printf(
			"checking for scorecard workflow file (%s)",
			f,
		)
		scoreFileContent, _, _, err := gh.GetContents(
			ctx,
			owner,
			repoName,
			f,
			&github.RepositoryContentGetOptions{},
		)
		if scoreFileContent != nil {
			log.Printf(
				"skipping repo (%s) since scorecard workflow already exists: %s",
				repoName,
				f,
			)

			workflowExists = true
			break
		}
		if err != nil && i == len(workflowFiles)-1 {
			log.Printf("could not find a scorecard workflow file: %+v", err)
		}
	}

	if !workflowExists {
		// Skip if branch scorecard already exists.
		scorecardBranch, _, err := gh.GetBranch(
			ctx,
			owner,
			repoName,
			pullRequestBranch,
			true,
		)
		if scorecardBranch != nil || err == nil {
			log.Printf(
				"skipping repo (%s) since the scorecard action installation branch already exists",
				repoName,
			)

			return nil
		}

		// Create new branch using a reference that stores the new commit hash.
		// TODO: Capture ref creation errors
		ref := &github.Reference{
			Ref:    github.String(branchReference),
			Object: &github.GitObject{SHA: defaultBranchSHA},
		}
		_, _, err = gh.CreateGitRef(ctx, owner, repoName, ref)
		if err != nil {
			return fmt.Errorf(
				"creating scorecard action installation branch for %s: %w",
				repoName,
				err,
			)
		}

		// Create file in repository.
		// TODO: Capture file creation errors
		opts := &github.RepositoryContentFileOptions{
			Message: github.String(commitMessage),
			Content: workflowContent,
			Branch:  github.String(pullRequestBranch),
		}
		_, _, err = gh.CreateFile(
			ctx,
			owner,
			repoName,
			workflowFile,
			opts,
		)
		if err != nil {
			return fmt.Errorf(
				"creating scorecard workflow file for %s: %w",
				repoName,
				err,
			)
		}

		// Create pull request.
		// TODO: Capture pull request creation errors
		_, err = gh.CreatePullRequest(
			ctx,
			owner,
			repoName,
			*defaultBranch.Name,
			pullRequestBranch,
			pullRequestTitle,
			pullRequestDescription,
		)
		if err != nil {
			return fmt.Errorf(
				"creating pull request for %s: %w",
				repoName,
				err,
			)
		}
	}

	return nil
}
