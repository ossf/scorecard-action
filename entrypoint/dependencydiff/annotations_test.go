// Copyright 2022 Security Scorecard Authors
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

package dependencydiff

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/google/go-github/v45/github"
	"github.com/ossf/scorecard-action/options"
	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/pkg"
)

func Test_visualizeToCheckRun(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name    string
		wantErr bool
		deps    []pkg.DependencyCheckResult
	}{

		{
			name:    "error creating the check run",
			wantErr: true,
			deps: []pkg.DependencyCheckResult{
				{
					Name:             "dep_a",
					SourceRepository: asPointerStr("repo_a"),
					ChangeType:       asPointerChangeType(pkg.ChangeType("added")),
					Version:          asPointerStr("0.8.0"),
					Ecosystem:        asPointerStr("PyPI"),
					ScorecardResultWithError: pkg.ScorecardResultWithError{
						ScorecardResult: &pkg.ScorecardResult{
							Checks: []checker.CheckResult{
								{
									Name:  "Maintained",
									Score: 10,
								},
								{
									Name:  "Vulnerabilities",
									Score: 10,
								},
							},
						},
					},
				},
				{
					Name:             "dep_b",
					SourceRepository: asPointerStr("repo_b"),
					ChangeType:       asPointerChangeType(pkg.ChangeType("removed")),
					Version:          asPointerStr("4.7.2.a1"),
					Ecosystem:        asPointerStr("PyPI"),
					ScorecardResultWithError: pkg.ScorecardResultWithError{
						ScorecardResult: &pkg.ScorecardResult{
							Checks: []checker.CheckResult{
								{
									Name:  "Security-Policy",
									Score: 8,
								},
								{
									Name:  "SAST",
									Score: 10,
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			os.Setenv(options.EnvInputPullRequestHeadSHA, "fake_head_sha")
			err := visualizeToCheckRun(
				context.Background(),
				github.NewClient(http.DefaultClient),
				"fake_owner",
				"fake_repo",
				tt.deps,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("want error:%v, got: %v", tt.wantErr, err)
			}
		})
	}
}
