// Copyright 2026 OpenSSF Authors
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

package install

import (
	"strings"
	"testing"

	gogh "github.com/google/go-github/v46/github"
)

func Test_renderWorkflow(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		repo        *gogh.Repository
		wantLines   []string
		unwantLines []string
	}{
		{
			name: "public repository",
			repo: &gogh.Repository{Private: gogh.Bool(false), DefaultBranch: gogh.String("trunk")},
			wantLines: []string{
				"    branches: ['trunk']",
				"          publish_results: true",
			},
			unwantLines: []string{"          publish_results: false"},
		},
		{
			name: "private repository",
			repo: &gogh.Repository{Private: gogh.Bool(true), DefaultBranch: gogh.String("main")},
			wantLines: []string{
				"    branches: ['main']",
				"          publish_results: false",
			},
			unwantLines: []string{"          publish_results: true"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := renderWorkflow(tt.repo)
			if err != nil {
				t.Fatalf("renderWorkflow(): %v", err)
			}
			lines := strings.Split(string(got), "\n")
			for _, want := range tt.wantLines {
				if !hasLine(lines, want) {
					t.Errorf("renderWorkflow() is missing line %q", want)
				}
			}
			for _, unwant := range tt.unwantLines {
				if hasLine(lines, unwant) {
					t.Errorf("renderWorkflow() has unwanted line %q", unwant)
				}
			}
			if strings.Contains(string(got), "$default-branch") || strings.Contains(string(got), "$cron-weekly") {
				t.Error("renderWorkflow() left a starter-workflows placeholder unfilled")
			}
		})
	}
}

func Test_renderWorkflow_noDefaultBranch(t *testing.T) {
	t.Parallel()
	if _, err := renderWorkflow(&gogh.Repository{}); err == nil {
		t.Error("renderWorkflow() with no default branch: got nil error, wanted an error")
	}
}

func hasLine(lines []string, want string) bool {
	for _, line := range lines {
		if line == want {
			return true
		}
	}
	return false
}
