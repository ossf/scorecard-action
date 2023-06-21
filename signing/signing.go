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

package signing

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	sigOpts "github.com/sigstore/cosign/v2/cmd/cosign/cli/options"
	"github.com/sigstore/cosign/v2/cmd/cosign/cli/sign"

	"github.com/ossf/scorecard-action/entrypoint"
	"github.com/ossf/scorecard-action/options"
)

var (
	errorEmptyToken   = errors.New("error token empty")
	errorInvalidToken = errors.New("invalid token")

	// backoff schedule for interactions with cosign/rekor and our web API.
	backoffSchedule = []time.Duration{
		1 * time.Second,
		3 * time.Second,
		10 * time.Second,
	}
)

// Signing is a signing structure.
type Signing struct {
	token string
}

// New creates a new Signing instance.
func New(token string) (*Signing, error) {
	// Set the default GITHUB_TOKEN, because it's not available by default
	// in a GitHub Action. We need it for OIDC.
	if token == "" {
		return nil, fmt.Errorf("%w", errorEmptyToken)
	}

	// Check for a workflow secret.
	if !strings.HasPrefix(token, "ghs_") {
		return nil, fmt.Errorf("%w: not a default GITHUB_TOKEN", errorInvalidToken)
	}
	if err := os.Setenv("GITHUB_TOKEN", token); err != nil {
		return nil, fmt.Errorf("error setting GITHUB_TOKEN env var: %w", err)
	}

	return &Signing{
		token: token,
	}, nil
}

// SignScorecardResult signs the results file and uploads the attestation to the Rekor transparency log.
func (s *Signing) SignScorecardResult(scorecardResultsFile string) error {
	// Prepare settings for SignBlobCmd.
	rootOpts := &sigOpts.RootOptions{Timeout: sigOpts.DefaultTimeout} // Just the timeout.
	keyOpts := sigOpts.KeyOpts{
		FulcioURL:        sigOpts.DefaultFulcioURL,     // Signing certificate provider.
		RekorURL:         sigOpts.DefaultRekorURL,      // Transparency log.
		OIDCIssuer:       sigOpts.DefaultOIDCIssuerURL, // OIDC provider to get ID token to auth for Fulcio.
		OIDCClientID:     "sigstore",
		SkipConfirmation: true, // skip cosign's privacy confirmation prompt as we run non-interactively
	}

	var err error
	for _, backoff := range backoffSchedule {
		// This command will use the provided OIDCIssuer to authenticate into Fulcio, which will generate the
		// signing certificate on the scorecard result. This attestation is then uploaded to the Rekor transparency log.
		// The output bytes (signature) and certificate are discarded since verification can be done with just the payload.
		_, err = sign.SignBlobCmd(rootOpts, keyOpts, scorecardResultsFile, true, "", "", true)
		if err == nil {
			break
		}
		log.Printf("error signing scorecard results: %v\n", err)
		log.Printf("retrying in %v...\n", backoff)
		time.Sleep(backoff)
	}

	// retries failed
	if err != nil {
		return fmt.Errorf("error signing payload: %w", err)
	}

	return nil
}

// GetJSONScorecardResults changes output settings to json and runs scorecard again.
// TODO: run scorecard only once and generate multiple formats together.
func GetJSONScorecardResults() ([]byte, error) {
	defer os.Setenv(options.EnvInputResultsFile, os.Getenv(options.EnvInputResultsFile))
	defer os.Setenv(options.EnvInputResultsFormat, os.Getenv(options.EnvInputResultsFormat))
	os.Setenv(options.EnvInputResultsFile, "results.json")
	os.Setenv(options.EnvInputResultsFormat, "json")

	actionJSON, err := entrypoint.New()
	if err != nil {
		return nil, fmt.Errorf("creating scorecard entrypoint: %w", err)
	}
	if err := actionJSON.Execute(); err != nil {
		return nil, fmt.Errorf("error during command execution: %w", err)
	}

	// Get json output data from file.
	jsonPayload, err := os.ReadFile(os.Getenv(options.EnvInputResultsFile))
	if err != nil {
		return nil, fmt.Errorf("reading scorecard json results from file: %w", err)
	}

	return jsonPayload, nil
}

// ProcessSignature calls scorecard-api to process & upload signed scorecard results.
func (s *Signing) ProcessSignature(jsonPayload []byte, repoName, repoRef string) error {
	// Prepare HTTP request body for scorecard-webapp-api call.
	// TODO: Use the `ScorecardResult` struct from `scorecard-webapp`.
	resultsPayload := struct {
		Result      string `json:"result"`
		Branch      string `json:"branch"`
		AccessToken string `json:"accessToken"`
	}{
		Result:      string(jsonPayload),
		Branch:      repoRef,
		AccessToken: s.token,
	}

	payloadBytes, err := json.Marshal(resultsPayload)
	if err != nil {
		return fmt.Errorf("marshalling json results: %w", err)
	}

	apiURL := os.Getenv(options.EnvInputInternalPublishBaseURL)
	rawURL := fmt.Sprintf("%s/projects/github.com/%s", apiURL, repoName)
	postURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("parsing Scorecard API endpoint: %w", err)
	}

	for _, backoff := range backoffSchedule {
		// Call scorecard-webapp-api to process and upload signature.
		err = postResults(postURL, payloadBytes)
		if err == nil {
			break
		}
		log.Printf("error sending scorecard results to webapp: %v\n", err)
		log.Printf("retrying in %v...\n", backoff)
		time.Sleep(backoff)
	}

	// retries failed
	if err != nil {
		return fmt.Errorf("error sending scorecard results to webapp: %w", err)
	}

	return nil
}

func postResults(endpoint *url.URL, payload []byte) error {
	req, err := http.NewRequest("POST", endpoint.String(), bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("creating HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(req.Context(), 10*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	// Execute request.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("executing scorecard-api call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("reading response body: %w", err)
		}
		return fmt.Errorf("http response %d, status: %v, error: %v", resp.StatusCode, resp.Status, string(bodyBytes)) //nolint
	}

	return nil
}
