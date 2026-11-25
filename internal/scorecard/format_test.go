// Copyright 2024 OpenSSF Scorecard Authors
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

package scorecard

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/ossf/scorecard-action/options"
	scopts "github.com/ossf/scorecard/v5/options"
	"github.com/ossf/scorecard/v5/pkg/scorecard"
)

const formatInToto = "intoto"

var errNoIdentity = errors.New("no identity")

// testResult returns a result the formatters accept.
func testResult() scorecard.Result {
	return scorecard.Result{
		Repo: scorecard.RepoInfo{
			Name: "github.com/foo/bar",
			// The in-toto statement's subject digest is the scanned commit,
			// which Scorecard validates as a SHA-1.
			CommitSHA: "0123456789abcdef0123456789abcdef01234567",
		},
	}
}

func TestFormatSignsInToto(t *testing.T) {
	t.Parallel()
	result := testResult()
	t.Run("signed statement is written", func(t *testing.T) {
		t.Parallel()
		opts := options.Options{
			InputResultsFile:   t.TempDir() + "/results",
			InputResultsFormat: formatInToto,
			SignAttestation:    true,
			ScorecardOpts:      &scopts.Options{},
		}
		var signed []byte
		sign := func(statement []byte) ([]byte, error) {
			if !bytes.Contains(statement, []byte(scorecard.InTotoPredicateType)) {
				t.Errorf("signer did not receive the in-toto statement")
			}
			signed = append([]byte("signed:"), statement...)
			return signed, nil
		}
		if err := format(&result, &opts, sign); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		contents, err := os.ReadFile(opts.InputResultsFile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Equal(contents, signed) {
			t.Errorf("results file does not hold the signed output")
		}
	})
	t.Run("signing error is returned when signing was asked for", func(t *testing.T) {
		t.Parallel()
		opts := options.Options{
			InputResultsFile:        t.TempDir() + "/results",
			InputResultsFormat:      formatInToto,
			SignAttestation:         true,
			SignAttestationRequired: true,
			ScorecardOpts:           &scopts.Options{},
		}
		sign := func([]byte) ([]byte, error) { return nil, errNoIdentity }
		if err := format(&result, &opts, sign); !errors.Is(err, errNoIdentity) {
			t.Fatalf("format() error = %v, want %v", err, errNoIdentity)
		}
	})
	t.Run("signing error leaves the statement unsigned when signing is automatic", func(t *testing.T) {
		t.Parallel()
		opts := options.Options{
			InputResultsFile:   t.TempDir() + "/results",
			InputResultsFormat: formatInToto,
			SignAttestation:    true,
			ScorecardOpts:      &scopts.Options{},
		}
		var statement []byte
		sign := func(s []byte) ([]byte, error) {
			statement = bytes.Clone(s)
			return []byte("half a bundle"), errNoIdentity
		}
		if err := format(&result, &opts, sign); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		contents, err := os.ReadFile(opts.InputResultsFile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(statement) == 0 || !bytes.Equal(contents, statement) {
			t.Errorf("results file does not hold the unsigned statement")
		}
	})
	t.Run("unsigned when signing is off", func(t *testing.T) {
		t.Parallel()
		opts := options.Options{
			InputResultsFile:   t.TempDir() + "/results",
			InputResultsFormat: formatInToto,
			ScorecardOpts:      &scopts.Options{},
		}
		sign := func([]byte) ([]byte, error) {
			t.Errorf("signer must not be called")
			return nil, nil
		}
		if err := format(&result, &opts, sign); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestFormat(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, format string
		pattern      []byte
	}{
		{
			name:    "default is sarif",
			format:  "",
			pattern: []byte("sarif-schema"),
		},
		{
			name:    "sarif format supported",
			format:  "sarif",
			pattern: []byte("sarif-schema"),
		},
		{
			name:   "json format supported",
			format: "json",
			// This isn't quite as strong of a guarantee, but dont expect this to change
			pattern: []byte(`"name":"github.com/foo/bar"`),
		},
		{
			name:    "intoto format supported",
			format:  "intoto",
			pattern: []byte(scorecard.InTotoPredicateType),
		},
		{
			name:    "format is case insensitive",
			format:  "SARIF",
			pattern: []byte("sarif-schema"),
		},
	}
	result := testResult()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			opts := options.Options{
				InputResultsFile:   t.TempDir() + "/results",
				InputResultsFormat: tt.format,
				ScorecardOpts: &scopts.Options{
					PolicyFile: "../../policies/template.yml",
				},
			}
			err := Format(&result, &opts)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			contents, err := os.ReadFile(opts.InputResultsFile)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !bytes.Contains(contents, tt.pattern) {
				t.Errorf("Output didn't match expected pattern (%s)", tt.pattern)
			}
		})
	}
}

// A run that fails to produce its output must not leave a results file
// behind: an empty one would be uploaded by a later step.
func TestFormatFailureLeavesNoResultsFile(t *testing.T) {
	t.Parallel()
	result := testResult()
	tests := []struct {
		wantErr error
		name    string
		opts    options.Options
	}{
		{
			name: "signing fails",
			opts: options.Options{
				InputResultsFormat:      formatInToto,
				SignAttestation:         true,
				SignAttestationRequired: true,
				ScorecardOpts:           &scopts.Options{},
			},
			wantErr: errNoIdentity,
		},
		{
			name: "unknown format",
			opts: options.Options{
				InputResultsFormat: "nope",
				ScorecardOpts:      &scopts.Options{},
			},
			wantErr: errUnknownFormat,
		},
		{
			name: "sarif policy file is missing",
			opts: options.Options{
				InputResultsFormat: "sarif",
				ScorecardOpts:      &scopts.Options{PolicyFile: "does-not-exist.yml"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.opts.InputResultsFile = t.TempDir() + "/results"
			sign := func([]byte) ([]byte, error) { return nil, errNoIdentity }

			err := format(&result, &tt.opts, sign)
			if err == nil {
				t.Fatalf("format() succeeded, want an error")
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("format() error = %v, want %v", err, tt.wantErr)
			}
			if _, err := os.Stat(tt.opts.InputResultsFile); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("results file exists after a failed run (stat error: %v)", err)
			}
		})
	}
}
