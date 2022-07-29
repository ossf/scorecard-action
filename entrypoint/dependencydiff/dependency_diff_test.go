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
	"errors"
	"testing"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/pkg"
)

func Test_dependencydiffResultsAsMarkdown(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name            string
		deps            []pkg.DependencyCheckResult
		wantEmptyResult bool
		wantErr         bool
	}{

		{
			name:            "valid output markdown",
			wantErr:         false,
			wantEmptyResult: false,
			deps: []pkg.DependencyCheckResult{
				{
					Name:             "dep_a",
					SourceRepository: asPointerStr("repo_a"),
					ChangeType:       asPointerChangeType(pkg.ChangeType("added")),
					Version:          asPointerStr("1.1.0"),
					Ecosystem:        asPointerStr("Go"),
					ScorecardResultWithError: pkg.ScorecardResultWithError{
						ScorecardResult: &pkg.ScorecardResult{
							Checks: []checker.CheckResult{
								{
									Name:  "Fuzzing",
									Score: 5,
								},
								{
									Name:  "License",
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
					Version:          asPointerStr("2.6.1"),
					Ecosystem:        asPointerStr("Go"),
					ScorecardResultWithError: pkg.ScorecardResultWithError{
						ScorecardResult: &pkg.ScorecardResult{
							Checks: []checker.CheckResult{
								{
									Name:  "Fuzzing",
									Score: 10,
								},
								{
									Name:  "License",
									Score: 10,
								},
							},
						},
					},
				},
				{
					Name:       "dep_c",
					ChangeType: asPointerChangeType(pkg.ChangeType("removed")),
					Version:    asPointerStr("v35"),
					Ecosystem:  asPointerStr("Go"),
					ScorecardResultWithError: pkg.ScorecardResultWithError{
						Error: errInvalid,
					},
				},
				{
					// The removed old version of dep_a, used for the update tag testing.
					Name:             "dep_a",
					SourceRepository: asPointerStr("repo_a"),
					ChangeType:       asPointerChangeType(pkg.ChangeType("removed")),
					Version:          asPointerStr("1.0.8"),
					Ecosystem:        asPointerStr("Go"),
				},
				{
					// The removed old version of dep_a, used for the update tag testing.
					Name:             "dep_d",
					SourceRepository: asPointerStr("repo_d"),
					ChangeType:       asPointerChangeType(pkg.ChangeType("added")),
					Version:          asPointerStr("5.6.5"),
					Ecosystem:        asPointerStr("Go"),
					ScorecardResultWithError: pkg.ScorecardResultWithError{
						ScorecardResult: &pkg.ScorecardResult{
							Checks: []checker.CheckResult{
								{
									Name:  "Fuzzing",
									Score: 0,
								},
								{
									Name:  "License",
									Score: 2,
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
			result, err := dependencydiffResultsAsMarkdown(tt.deps, "base", "head")
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr: %v, got: %v", tt.wantErr, err)
			}
			if (result != nil) == tt.wantEmptyResult {
				// Return an error if got a non-empty result but we want an empty result.
				t.Errorf("wantEmptyResult: %v, got: %v", tt.wantEmptyResult, *result)
			}
		})
	}
}

func Test_writeToComment(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name      string
		errWanted error
	}{

		{
			name:      "empty github reference",
			errWanted: errEmpty,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := writeToComment(
				context.Background(),
				nil,
				"owner", "repo",
				asPointerStr("test report"),
			)
			if !errors.Is(err, tt.errWanted) {
				t.Errorf("error wanted: %v, got %v", tt.errWanted, err)
			}
		})
	}
}
