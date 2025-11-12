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
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	protobundle "github.com/sigstore/protobuf-specs/gen/pb-go/bundle/v1"

	"github.com/ossf/scorecard-action/options"
	"google.golang.org/protobuf/encoding/protojson"
)

// TODO: For this test to work, fake the OIDC token retrieval with something like.
//nolint // https://github.com/sigstore/cosign/blob/286bb0c58757009e99ab7080c720b30e51d08855/cmd/cosign/cli/fulcio/fulcio_test.go

// func Test_SignScorecardResult(t *testing.T) {
// 	t.Parallel()
// 	// Generate random bytes to use as our payload. This is done because signing identical payloads twice
// 	// just creates multiple entries under it, so we are keeping this test simple and not comparing timestamps.
// 	fmt.Println("ACTIONS_ID_TOKEN_REQUEST_TOKEN:")
// 	fmt.Println(os.Getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN"))
// 	scorecardResultsFile := "./sign-random-data.txt"
// 	randomData := make([]byte, 20)
// 	if _, err := rand.Read(randomData); err != nil {
// 		t.Errorf("signScorecardResult() error generating random bytes, %v", err)
// 		return
// 	}
// 	if err := ioutil.WriteFile(scorecardResultsFile, randomData, 0o600); err != nil {
// 		t.Errorf("signScorecardResult() error writing random bytes to file, %v", err)
// 		return
// 	}

// 	// Sign example scorecard results file.
// 	err := SignScorecardResult(scorecardResultsFile)
// 	if err != nil {
// 		t.Errorf("signScorecardResult() error, %v", err)
// 		return
// 	}

// 	// Verify that the signature was created and uploaded to the Rekor tlog by looking up the payload.
// 	ctx := context.Background()
// 	rekorClient, err := rekor.NewClient(options.DefaultRekorURL)
// 	if err != nil {
// 		t.Errorf("signScorecardResult() error getting Rekor client, %v", err)
// 		return
// 	}
// 	scorecardResultData, err := ioutil.ReadFile(scorecardResultsFile)
// 	if err != nil {
// 		t.Errorf("signScorecardResult() error reading scorecard result file, %v", err)
// 		return
// 	}
// 	uuids, err := cosign.FindTLogEntriesByPayload(ctx, rekorClient, scorecardResultData)
// 	if err != nil {
// 		t.Errorf("signScorecardResult() error getting tlog entries, %v", err)
// 		return
// 	}

// 	if len(uuids) != 1 {
// 		t.Errorf("signScorecardResult() error finding signature in Rekor tlog, %v", err)
// 		return
// 	}
// }

//nolint:paralleltest // we are using t.Setenv
func TestProcessSignature(t *testing.T) {
	tests := []struct {
		name        string
		payloadPath string
		status      int
		wantErr     bool
	}{
		{
			name:        "post succeeded",
			status:      http.StatusCreated,
			payloadPath: "testdata/results.json",
			wantErr:     false,
		},
		{
			name:        "post failed",
			status:      http.StatusBadRequest,
			payloadPath: "testdata/results.json",
			wantErr:     true,
		},
	}
	// use smaller backoffs for the test so they run faster
	setBackoffs(t, []time.Duration{0, time.Millisecond, 2 * time.Millisecond})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonPayload, err := os.ReadFile(tt.payloadPath)
			if err != nil {
				t.Fatalf("Unexpected error reading testdata: %v", err)
			}
			repoName := "ossf-tests/scorecard-action"
			repoRef := "refs/heads/main"
			//nolint:gosec // dummy credentials
			accessToken := "ghs_foo"
			idToken := "foo"
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
			}))
			t.Setenv(options.EnvInputInternalPublishBaseURL, server.URL)
			t.Cleanup(server.Close)

			s, err := New(accessToken, idToken)
			if err != nil {
				t.Fatalf("Unexpected error New: %v", err)
			}
			err = s.ProcessSignature(jsonPayload, repoName, repoRef)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessSignature() error: %v, wantErr: %v", err, tt.wantErr)
			}
		})
	}
}

func Test_extractTlogIndex(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		bundlePath string
		want       int64
		wantErr    bool
	}{
		{
			name:       "valid bundle",
			bundlePath: "testdata/bundle.json",
			want:       623290401,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			contents, err := os.ReadFile(tt.bundlePath)
			var bundle protobundle.Bundle
			if err = protojson.Unmarshal(contents, &bundle); err != nil {
				t.Fatal(err)
			}
			got, err := extractTlogIndex(&bundle)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected err: %v, wantErr: %t", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("wrong tlog index: got %d, wanted %d", got, tt.want)
			}
		})
	}
}

//nolint:paralleltest // we are using t.Setenv
func TestProcessSignature_retries(t *testing.T) {
	tests := []struct {
		name          string
		nFailures     int
		wantNRequests int
		wantErr       bool
	}{
		{
			name:          "succeeds immediately",
			nFailures:     0,
			wantNRequests: 1,
			wantErr:       false,
		},
		{
			name:          "one retry",
			nFailures:     1,
			wantNRequests: 2,
			wantErr:       false,
		},
		{
			// limit corresponds to backoffs set in test body
			name:          "retry limit exceeded",
			nFailures:     4,
			wantNRequests: 3,
			wantErr:       true,
		},
	}
	// use smaller backoffs for the test so they run faster
	setBackoffs(t, []time.Duration{0, time.Millisecond, 2 * time.Millisecond})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var jsonPayload []byte
			repoName := "ossf-tests/scorecard-action"
			repoRef := "refs/heads/main"
			//nolint:gosec // dummy credentials
			accessToken := "ghs_foo"
			idToken := "foo"
			var nRequests int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nRequests++
				status := http.StatusCreated
				if tt.nFailures > 0 {
					status = http.StatusBadRequest
					tt.nFailures--
				}
				w.WriteHeader(status)
			}))
			t.Setenv(options.EnvInputInternalPublishBaseURL, server.URL)
			t.Cleanup(server.Close)

			s, err := New(accessToken, idToken)
			if err != nil {
				t.Fatalf("Unexpected error New: %v", err)
			}
			err = s.ProcessSignature(jsonPayload, repoName, repoRef)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessSignature() error: %v, wantErr: %v", err, tt.wantErr)
			}
			if nRequests != tt.wantNRequests {
				t.Errorf("ProcessSignature() made %d requests, wanted %d", nRequests, tt.wantNRequests)
			}
		})
	}
}

// temporarily sets the backoffs for a given test.
func setBackoffs(t *testing.T, newBackoffs []time.Duration) {
	t.Helper()
	old := backoffSchedule
	backoffSchedule = newBackoffs
	t.Cleanup(func() {
		backoffSchedule = old
	})
}
