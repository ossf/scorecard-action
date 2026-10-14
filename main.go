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
	"strconv"

	"github.com/ossf/scorecard-action/internal/publication"
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

	failOnPublishError, err := getFailOnPublishError()
	if err != nil {
		log.Fatal(err)
	}

	result, err := scorecard.Run(opts)
	if err != nil {
		log.Fatal(err)
	}

	if err := scorecard.Format(&result, opts); err != nil {
		log.Fatal(err)
	}

	var pub publication.Result

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
		pub = publication.Result{
			Attempted: true,
			Err:       s.ProcessSignature(jsonPayload, repoName, repoRef),
		}
	}

	if err := pub.Report(os.Getenv(options.EnvGithubOutput), os.Getenv(options.EnvGithubStepSummary)); err != nil {
		fmt.Printf("::warning::Unable to report the publication status: %v\n", err)
	}
	if pub.Err != nil && reportPublishError(pub, failOnPublishError) {
		os.Exit(1)
	}
}

// reportPublishError annotates a failed publish, and reports whether the step
// should fail: only when fail_on_publish_error is set and the Scorecard API
// rejected the results. Workflow commands must start a line, so this prints
// them without the log package's timestamp.
func reportPublishError(pub publication.Result, failOnPublishError bool) bool {
	if pub.Rejected() && failOnPublishError {
		fmt.Printf("::error::The Scorecard API rejected the results: %v. "+
			"Check the workflow restrictions at "+
			"https://github.com/ossf/scorecard-action#workflow-restrictions.\n", pub.Err)
		return true
	}
	hint := "If this issue persists, check the repo issues for more information."
	if pub.Rejected() {
		hint = "Set fail_on_publish_error: true to fail the step when the Scorecard API rejects results."
	}
	fmt.Printf("::warning::Unable to POST scorecard results to webapp: %v. %s\n", pub.Err, hint)
	return false
}

func getFailOnPublishError() (bool, error) {
	val := os.Getenv(options.EnvInputFailOnPublishError)
	if val == "" {
		return false, nil
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return false, fmt.Errorf("parsing %s: %w", options.EnvInputFailOnPublishError, err)
	}
	return b, nil
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
