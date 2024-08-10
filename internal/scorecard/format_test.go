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
	"os"
	"testing"

	"github.com/ossf/scorecard-action/options"
	scopts "github.com/ossf/scorecard/v5/options"
	"github.com/ossf/scorecard/v5/pkg/scorecard"
)

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
			name:    "format is case insensitive",
			format:  "SARIF",
			pattern: []byte("sarif-schema"),
		},
	}
	result := scorecard.Result{
		Repo: scorecard.RepoInfo{
			Name: "github.com/foo/bar",
		},
	}
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
