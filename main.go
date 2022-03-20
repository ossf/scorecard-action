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

package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/ossf/scorecard-action/entrypoint"
	"github.com/ossf/scorecard-action/options"
	"github.com/ossf/scorecard-action/signing"
)

func main() {
	action, err := entrypoint.New()
	if err != nil {
		log.Fatalf("creating scorecard entrypoint: %v", err)
	}

	if err := action.Execute(); err != nil {
		log.Fatalf("error during command execution: %v", err)
	}

	// Process signature using scorecard-api.
	if os.Getenv(options.EnvInputPublishResults) == "true" {
		// Get sarif output from file.
		sarifPayload, err := ioutil.ReadFile(os.Getenv(options.EnvInputResultsFile))
		if err != nil {
			log.Fatalf("error reading from sarif output file: %v", err)
		}

		// Get json results by re-running scorecard.
		jsonPayload, err := signing.GetJsonScorecardResults()
		if err != nil {
			log.Fatalf("error generating json scorecard results: %v", err)
		}

		// Sign & upload scorecard results.
		if err := signing.ProcessSignature(sarifPayload, jsonPayload); err != nil {
			log.Fatalf("error processing signature: %v", err)
		}
	}
}
