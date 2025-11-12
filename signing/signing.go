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

// Package signing provides functionality to sign and upload results to the Scorecard API.
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

	protobundle "github.com/sigstore/protobuf-specs/gen/pb-go/bundle/v1"
	"github.com/sigstore/sigstore-go/pkg/root"
	"github.com/sigstore/sigstore-go/pkg/sign"
	"github.com/sigstore/sigstore-go/pkg/tuf"
	"github.com/sigstore/sigstore-go/pkg/util"
	"github.com/theupdateframework/go-tuf/v2/metadata/fetcher"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/ossf/scorecard-action/options"
)

var (
	errorEmptyToken   = errors.New("error token empty")
	errorInvalidToken = errors.New("invalid token")
	errNoTlogEntries  = errors.New("no rekor tlog entries")

	// backoff schedule for interactions with cosign/rekor and our web API.
	backoffSchedule = []time.Duration{
		1 * time.Second,
		3 * time.Second,
		10 * time.Second,
	}
)

// Signing is a signing structure.
type Signing struct {
	token          string // github token used to fetch workflow contents
	idToken        string // oidc token used to sign results
	rekorTlogIndex int64
	bundle         string
}

// New creates a new Signing instance.
func New(token, idToken string) (*Signing, error) {
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
		token:   token,
		idToken: idToken,
	}, nil
}

// SignResult signs the results file and uploads the attestation to the Rekor transparency log.
func (s *Signing) SignResult(result []byte) error {
	content := sign.PlainData{
		Data: result,
	}
	keypair, err := sign.NewEphemeralKeypair(nil)
	if err != nil {
		return fmt.Errorf("generating ephemeral keypair: %w", err)
	}
	opts, err := getBundleOptions(s.idToken)
	if err != nil {
		return fmt.Errorf("getting bundle options: %w", err)
	}
	bundle, err := sign.Bundle(&content, keypair, opts)
	if err != nil {
		return fmt.Errorf("creating sigstore bundle: %w", err)
	}

	rekorTlogIndex, err := extractTlogIndex(bundle)
	if err != nil {
		return err
	}
	s.rekorTlogIndex = rekorTlogIndex

	bundleJSON, err := protojson.Marshal(bundle)
	if err != nil {
		return fmt.Errorf("marshalling bundle to JSON: %v", err)
	}
	s.bundle = string(bundleJSON)
	log.Println(s.bundle)

	return nil
}

// ProcessSignature calls scorecard-api to process & upload signed scorecard results.
func (s *Signing) ProcessSignature(jsonPayload []byte, repoName, repoRef string) error {
	// Prepare HTTP request body for scorecard-webapp-api call.
	// TODO: Use the `ScorecardResult` struct from `scorecard-webapp`.
	resultsPayload := struct {
		Result      string `json:"result"`
		Branch      string `json:"branch"`
		AccessToken string `json:"accessToken"`
		TlogIndex   int64  `json:"tlogIndex"`
		Bundle      string `json:"bundle"`
	}{
		Result:      string(jsonPayload),
		Branch:      repoRef,
		AccessToken: s.token,
		TlogIndex:   s.rekorTlogIndex,
		Bundle:      s.bundle,
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

func extractTlogIndex(bundle *protobundle.Bundle) (int64, error) {
	// we only submit to one rekor log, so grab the first (and only) tlog entry
	for _, entry := range bundle.GetVerificationMaterial().GetTlogEntries() {
		return entry.GetLogIndex(), nil
	}
	return 0, errNoTlogEntries
}

func getBundleOptions(idToken string) (sign.BundleOptions, error) {
	var opts sign.BundleOptions
	fetcher := fetcher.NewDefaultFetcher()
	fetcher.SetHTTPUserAgent(util.ConstructUserAgent())

	tufOptions := &tuf.Options{
		Root:              tuf.DefaultRoot(),
		RepositoryBaseURL: tuf.DefaultMirror,
		Fetcher:           fetcher,
	}
	tufClient, err := tuf.New(tufOptions)
	if err != nil {
		return sign.BundleOptions{}, fmt.Errorf("creating tuf client: %w", err)
	}

	opts.TrustedRoot, err = root.GetTrustedRoot(tufClient)
	if err != nil {
		return sign.BundleOptions{}, fmt.Errorf("getting tuf root: %w", err)
	}
	signingConfig, err := root.GetSigningConfig(tufClient)
	if err != nil {
		return sign.BundleOptions{}, fmt.Errorf("getting signing config: %w", err)
	}
	now := time.Now()
	fulcioService, err := root.SelectService(signingConfig.FulcioCertificateAuthorityURLs(), sign.FulcioAPIVersions, now)
	if err != nil {
		return sign.BundleOptions{}, fmt.Errorf("getting fulcio config: %w", err)
	}
	fulcioOpts := &sign.FulcioOptions{
		BaseURL: fulcioService.URL,
		Timeout: time.Duration(30 * time.Second),
		Retries: 3,
	}
	opts.CertificateProvider = sign.NewFulcio(fulcioOpts)
	opts.CertificateProviderOptions = &sign.CertificateProviderOptions{
		IDToken: idToken,
	}
	tsaServices, err := root.SelectServices(signingConfig.TimestampAuthorityURLs(), signingConfig.TimestampAuthorityURLsConfig(), sign.TimestampAuthorityAPIVersions, now)
	if err != nil {
		return sign.BundleOptions{}, fmt.Errorf("getting TSA config: %w", err)
	}
	for _, tsaService := range tsaServices {
		tsaOpts := &sign.TimestampAuthorityOptions{
			URL:     tsaService.URL,
			Timeout: time.Duration(30 * time.Second),
			Retries: 3,
		}
		opts.TimestampAuthorities = append(opts.TimestampAuthorities, sign.NewTimestampAuthority(tsaOpts))
	}
	forceRekorV1 := []uint32{1}
	rekorService, err := root.SelectService(signingConfig.RekorLogURLs(), forceRekorV1, now)
	if err != nil {
		return sign.BundleOptions{}, fmt.Errorf("getting rekor config: %w", err)
	}
	rekorOpts := &sign.RekorOptions{
		BaseURL: rekorService.URL,
		Timeout: time.Duration(90 * time.Second),
		Retries: 3,
		Version: rekorService.MajorAPIVersion,
	}
	opts.TransparencyLogs = append(opts.TransparencyLogs, sign.NewRekor(rekorOpts))
	return opts, nil
}
