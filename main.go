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
	"log"
	"os"

	"github.com/ossf/scorecard-action/entrypoint"
	"github.com/ossf/scorecard-action/options"
	"github.com/ossf/scorecard-action/signing"
)

func main() {
	triggerEventName := os.Getenv("GITHUB_EVENT_NAME")
	if triggerEventName == "pull_request_target" {
		log.Fatalf("pull_request_target trigger is not supported for security reasons" +
			"see https://securitylab.github.com/research/github-actions-preventing-pwn-requests/")
	}

	action, err := entrypoint.New()
	if err != nil {
		log.Fatalf("creating scorecard entrypoint: %v", err)
	}

	if err := action.Execute(); err != nil {
		log.Fatalf("error during command execution: %v", err)
	}

	if os.Getenv(options.EnvInputPublishResults) == "true" &&
		// `pull_request` do not have the necessary `token-id: write` permissions.
		triggerEventName != "pull_request" {
		// Get json results by re-running scorecard.
		jsonPayload, err := signing.GetJSONScorecardResults()
		if err != nil {
			log.Fatalf("error generating json scorecard results: %v", err)
		}

		// Sign json results.
		// Always use the default GitHub token, never a PAT.
		accessToken := os.Getenv(options.EnvInputInternalRepoToken)
		s, err := signing.New(accessToken)
		if err != nil {
			log.Fatalf("error SigningNew: %v", err)
		}
		if err = s.SignScorecardResult("results.json"); err != nil {
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
