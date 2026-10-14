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

// Package publication reports whether results were published to the
// Scorecard API, as step outputs and a job summary line.
package publication

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/ossf/scorecard-action/signing"
)

const (
	// StatusSkipped is the publication status when publishing wasn't attempted.
	StatusSkipped = "skipped"
	// StatusError is the publication status when no HTTP response was received.
	StatusError = "error"

	outputResultsPublished   = "results_published"
	outputPublicationStatus  = "publication_status"
	summaryScanSucceeded     = "Scorecard scan succeeded."
	defaultOutputPermissions = 0o644
)

// Result is the outcome of publishing results to the Scorecard API.
type Result struct {
	// Err is the error returned by the publish attempt, if any.
	Err error
	// Attempted reports whether publishing was attempted.
	Attempted bool
}

// Published reports whether the results were published.
func (r Result) Published() bool {
	return r.Attempted && r.Err == nil
}

// Rejected reports whether the Scorecard API rejected the results with a
// response that retrying won't change: a 4xx other than 408 or 429.
func (r Result) Rejected() bool {
	code := r.statusCode()
	return code >= http.StatusBadRequest && code < http.StatusInternalServerError &&
		!r.publishError().Retryable()
}

// Status is the HTTP status code of the final publish attempt, StatusError if
// no response was received, or StatusSkipped if publishing wasn't attempted.
func (r Result) Status() string {
	switch {
	case !r.Attempted:
		return StatusSkipped
	case r.Err == nil:
		return strconv.Itoa(http.StatusCreated)
	case r.statusCode() != 0:
		return strconv.Itoa(r.statusCode())
	default:
		return StatusError
	}
}

// Summary is a one-line, Markdown description of the scan and publish outcome.
func (r Result) Summary() string {
	var outcome string
	switch {
	case !r.Attempted:
		outcome = "Results were not published: publishing is not enabled for this run."
	case r.Err == nil:
		outcome = "Results were published to the Scorecard API."
	case r.Rejected():
		outcome = fmt.Sprintf("Results were not published: the Scorecard API rejected them with HTTP %d.", r.statusCode())
	case r.statusCode() != 0:
		outcome = fmt.Sprintf("Results were not published: the Scorecard API returned HTTP %d.", r.statusCode())
	default:
		outcome = "Results were not published: no response was received from the Scorecard API."
	}
	return fmt.Sprintf("%s %s", summaryScanSucceeded, outcome)
}

// Report appends the step outputs to outputPath and the summary line to
// summaryPath. An empty path is skipped.
func (r Result) Report(outputPath, summaryPath string) error {
	outputs := fmt.Sprintf("%s=%t\n%s=%s\n",
		outputResultsPublished, r.Published(),
		outputPublicationStatus, r.Status(),
	)
	if err := appendToFile(outputPath, outputs); err != nil {
		return fmt.Errorf("writing step outputs: %w", err)
	}
	if err := appendToFile(summaryPath, r.Summary()+"\n"); err != nil {
		return fmt.Errorf("writing job summary: %w", err)
	}
	return nil
}

func (r Result) publishError() *signing.PublishError {
	var pubErr *signing.PublishError
	if errors.As(r.Err, &pubErr) {
		return pubErr
	}
	return nil
}

func (r Result) statusCode() int {
	if pubErr := r.publishError(); pubErr != nil {
		return pubErr.StatusCode
	}
	return 0
}

func appendToFile(path, content string) error {
	if path == "" {
		return nil
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, defaultOutputPermissions)
	if err != nil {
		return fmt.Errorf("opening %s: %w", path, err)
	}
	if _, err := f.WriteString(content); err != nil {
		f.Close()
		return fmt.Errorf("writing %s: %w", path, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("closing %s: %w", path, err)
	}
	return nil
}
