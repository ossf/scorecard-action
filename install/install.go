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
	pullRequestDescription = "Added the workflow for OpenSSF's Security Scorecard"
	pullRequestTitle       = commitMessage

	workflowFiles = []string{
		path.Join(workflowBase, workflowFile),
		path.Join(workflowBase, workflowFileDeprecated),
	}
)

// Run adds the OpenSSF Scorecard workflow to all repositories under the given
// organization.
// TODO(install): Improve description.
// TODO(install): Accept a context instead of setting one.
// TODO(lint): cognitive complexity 31 of func `Run` is high (> 30) (gocognit).
//
// TODO(lint): cognitive complexity 31 of func `Run` is high (> 30) (gocognit).
//
//nolint:gocognit
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
		// Get repo metadata.
		repo, _, err := gh.GetRepository(ctx, o.Owner, repoName)
		if err != nil {
			log.Printf(
				"skipped repo (%s) because it does not exist or could not be accessed: %+v",
				repoName,
				err,
			)

			continue
		}

		// Get head commit SHA of default branch.
		// TODO: Capture branch access errors
		defaultBranch, _, err := gh.GetBranch(
			ctx,
			o.Owner,
			repoName,
			*repo.DefaultBranch,
			true,
		)
		if err != nil {
			log.Printf(
				"skipped repo (%s) because its default branch could not be accessed: %+v",
				repoName,
				err,
			)

			continue
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
				o.Owner,
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
				log.Printf("checking for scorecard workflow file: %+v", err)
			}
		}

		if !workflowExists {
			// Skip if branch scorecard already exists.
			scorecardBranch, _, err := gh.GetBranch(
				ctx,
				o.Owner,
				repoName,
				pullRequestBranch,
				true,
			)
			if scorecardBranch != nil || err == nil {
				log.Printf(
					"skipped repo (%s) since the scorecard branch already exists",
					repoName,
				)

				continue
			}

			// Create new branch using a reference that stores the new commit hash.
			// TODO: Capture ref creation errors
			ref := &github.Reference{
				Ref:    github.String(branchReference),
				Object: &github.GitObject{SHA: defaultBranchSHA},
			}
			_, _, err = gh.CreateGitRef(ctx, o.Owner, repoName, ref)
			if err != nil {
				log.Printf(
					"skipped repo (%s) because new branch could not be created: %+v",
					repoName,
					err,
				)

				continue
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
				o.Owner,
				repoName,
				workflowFile,
				opts,
			)
			if err != nil {
				log.Printf(
					"skipped repo (%s) because new file could not be created: %+v",
					repoName,
					err,
				)

				continue
			}

			// Create pull request.
			// TODO: Capture pull request creation errors
			_, err = gh.CreatePullRequest(
				ctx,
				o.Owner,
				repoName,
				*defaultBranch.Name,
				pullRequestBranch,
				pullRequestTitle,
				pullRequestDescription,
			)
			if err != nil {
				log.Printf(
					"skipped repo (%s) because pull request could not be created: %+v",
					repoName,
					err,
				)

				continue
			}

			log.Printf(
				"Created a pull request to add the scorecard workflow to %s",
				repoName,
			)
		}
	}

	return nil
}
