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

// Command scorecard-action is the entrypoint for the Scorecard GitHub Action.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ossf/scorecard-action/internal/scorecard"
	"github.com/ossf/scorecard-action/options"
	"github.com/ossf/scorecard-action/signing"
)

func main() {
	triggerEventName := os.Getenv("GITHUB_EVENT_NAME")
	if triggerEventName == "pull_request_target" {
		log.Fatalf("pull_request_target trigger is not supported for security reasons" +
			"see https://securitylab.github.com/research/github-actions-preventing-pwn-requests/")
	}

	opts, err := getOpts()
	if err != nil {
		log.Fatal(err)
	}
	opts.Print()

	result, err := scorecard.Run(opts)
	if err != nil {
		log.Fatal(err)
	}

	if err := scorecard.Format(&result, opts); err != nil {
		log.Fatal(err)
	}

	// `pull_request` does not have the necessary `token-id: write` permissions.
	//
	//nolint:nestif // trying to keep the refactor simpler
	if os.Getenv(options.EnvInputPublishResults) == "true" && triggerEventName != "pull_request" {
		// if we don't already have the results as JSON, generate them
		if opts.InputResultsFormat != "json" {
			opts.InputResultsFormat = "json"
			opts.InputResultsFile = "results.json"
			err = scorecard.Format(&result, opts)
			if err != nil {
				log.Fatal(err)
			}
		}

		resultFile := filepath.Join(opts.GithubWorkspace, opts.InputResultsFile)
		jsonPayload, err := os.ReadFile(resultFile)
		if err != nil {
			log.Fatalf("reading json scorecard results: %v", err)
		}

		// Sign json results.
		// Always use the default GitHub token, never a PAT.
		accessToken := os.Getenv(options.EnvInputInternalRepoToken)
		s, err := signing.New(accessToken)
		if err != nil {
			log.Fatalf("error SigningNew: %v", err)
		}
		// TODO: does it matter if this is hardcoded as results.json or not?
		if err = s.SignScorecardResult(resultFile); err != nil {
			log.Fatalf("error signing scorecard json results: %v", err)
		}

		// Processes json results.
		repoName := os.Getenv(options.EnvGithubRepository)
		repoRef := os.Getenv(options.EnvGithubRef)
		if err := s.ProcessSignature(jsonPayload, repoName, repoRef); err != nil {
			log.Fatalf("error processing signature: %v", err)
		}
	}
}

func getOpts() (*options.Options, error) {
	opts, err := options.New()
	if err != nil {
		return nil, fmt.Errorf("creating new options: %w", err)
	}
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("validating options: %w", err)
	}
	return opts, nil
}
