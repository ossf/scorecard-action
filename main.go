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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

	// Run again to create json output to sign result.
	if os.Getenv(options.EnvInputPublishResults) == "true" {
		// Save sarif output.
		sarifPayload, err := ioutil.ReadFile(os.Getenv(options.EnvInputResultsFile))
		if err != nil {
			log.Fatalf("reading scorecard sarif results from file: %v", err)
		}

		// Change output settings to json and run scorecard again.
		os.Setenv(options.EnvInputResultsFile, "results.json")
		os.Setenv(options.EnvInputResultsFormat, "json")
		actionJson, err := entrypoint.New()
		fmt.Printf("%v", actionJson)
		if err != nil {
			log.Fatalf("creating scorecard entrypoint: %v", err)
		}
		if err := actionJson.Execute(); err != nil {
			log.Fatalf("error during command execution: %v", err)
		}
		if err = signing.SignScorecardResult("results.sarif"); err != nil {
			log.Fatalf("error signing scorecard sarif results: %v", err)
		}

		jsonPayload, err := ioutil.ReadFile(os.Getenv(options.EnvInputResultsFile))
		if err != nil {
			log.Fatalf("reading scorecard json results from file: %v", err)
		}

		// Prepare HTTP request body for scorecard-webapp-api call.
		resultsPayload := struct {
			SarifOutput string
			JsonOutput  string
		}{
			SarifOutput: string(sarifPayload),
			JsonOutput:  string(jsonPayload),
		}

		payloadBytes, err := json.Marshal(resultsPayload)
		if err != nil {
			log.Fatalf("reading scorecard json results from file: %v", err)
		}

		// Call scorecard-webapp-api to process and upload signature.
		url := "https://api.securityscorecards.dev/verify"
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
		// TODO: don't hardcode these.
		req.Header.Set("Repository", "rohankh532/scorecard-OIDC-test")
		req.Header.Set("Branch", "refs/heads/main")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("executing scorecard-api call: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			log.Fatalf("http response error: %v", err)
		}

		// For testing.
		body, err := ioutil.ReadAll(resp.Body)
		fmt.Println("response body:", string(body))
	}
}
