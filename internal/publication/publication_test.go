// Copyright 2026 OpenSSF Scorecard Authors
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

package publication

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/ossf/scorecard-action/signing"
)

const noResponseSummary = "Scorecard scan succeeded. " +
	"Results were not published: no response was received from the Scorecard API."

var errTest = errors.New("test error")

func publishErr(code int) error {
	return fmt.Errorf("publishing scorecard results: %w", &signing.PublishError{StatusCode: code, Err: errTest})
}

func TestResult(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		wantStatus    string
		wantSummary   string
		result        Result
		wantPublished bool
		wantRejected  bool
	}{
		{
			name:        "not attempted",
			result:      Result{},
			wantStatus:  "skipped",
			wantSummary: "Scorecard scan succeeded. Results were not published: publishing is not enabled for this run.",
		},
		{
			name:          "published",
			result:        Result{Attempted: true},
			wantPublished: true,
			wantStatus:    "201",
			wantSummary:   "Scorecard scan succeeded. Results were published to the Scorecard API.",
		},
		{
			name:         "rejected",
			result:       Result{Attempted: true, Err: publishErr(http.StatusBadRequest)},
			wantRejected: true,
			wantStatus:   "400",
			wantSummary:  "Scorecard scan succeeded. Results were not published: the Scorecard API rejected them with HTTP 400.",
		},
		{
			name:        "rate limited",
			result:      Result{Attempted: true, Err: publishErr(http.StatusTooManyRequests)},
			wantStatus:  "429",
			wantSummary: "Scorecard scan succeeded. Results were not published: the Scorecard API returned HTTP 429.",
		},
		{
			name:        "server error",
			result:      Result{Attempted: true, Err: publishErr(http.StatusServiceUnavailable)},
			wantStatus:  "503",
			wantSummary: "Scorecard scan succeeded. Results were not published: the Scorecard API returned HTTP 503.",
		},
		{
			name:        "no response",
			result:      Result{Attempted: true, Err: publishErr(0)},
			wantStatus:  "error",
			wantSummary: noResponseSummary,
		},
		{
			name:        "not a publish error",
			result:      Result{Attempted: true, Err: errTest},
			wantStatus:  "error",
			wantSummary: noResponseSummary,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.result.Published(); got != tt.wantPublished {
				t.Errorf("Published(): got %t, wanted %t", got, tt.wantPublished)
			}
			if got := tt.result.Rejected(); got != tt.wantRejected {
				t.Errorf("Rejected(): got %t, wanted %t", got, tt.wantRejected)
			}
			if got := tt.result.Status(); got != tt.wantStatus {
				t.Errorf("Status(): got %q, wanted %q", got, tt.wantStatus)
			}
			if got := tt.result.Summary(); got != tt.wantSummary {
				t.Errorf("Summary(): got %q, wanted %q", got, tt.wantSummary)
			}
		})
	}
}

func TestReport(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	outputPath := filepath.Join(dir, "output")
	summaryPath := filepath.Join(dir, "summary")
	// GitHub creates these files; existing content must be kept.
	if err := os.WriteFile(outputPath, []byte("other=value\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	r := Result{Attempted: true, Err: publishErr(http.StatusBadRequest)}
	if err := r.Report(outputPath, summaryPath); err != nil {
		t.Fatalf("Report(): %v", err)
	}

	gotOutput, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	wantOutput := "other=value\nresults_published=false\npublication_status=400\n"
	if string(gotOutput) != wantOutput {
		t.Errorf("outputs: got %q, wanted %q", gotOutput, wantOutput)
	}

	gotSummary, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatal(err)
	}
	if want := r.Summary() + "\n"; string(gotSummary) != want {
		t.Errorf("summary: got %q, wanted %q", gotSummary, want)
	}
}

func TestReport_emptyPaths(t *testing.T) {
	t.Parallel()
	if err := (Result{}).Report("", ""); err != nil {
		t.Errorf("Report() with empty paths: %v", err)
	}
}

func TestReport_unwritable(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "missing", "output")
	if err := (Result{}).Report(missing, ""); err == nil {
		t.Error("Report() to a missing directory: got nil error, wanted an error")
	}
}
