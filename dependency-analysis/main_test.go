// Copyright 2023 OpenSSF Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"os"
	"path"
	"reflect"
	"testing"
)

func Test_filter(t *testing.T) { //nolint:paralleltest
	type args[T any] struct { //nolint:govet
		slice []T
		f     func(T) bool
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want []T
	}
	tests := []testCase[string]{
		{
			name: "default true",
			args: args[string]{
				slice: []string{"a"},
				f:     func(s string) bool { return s == "a" },
			},
			want: []string{"a"},
		},
	}

	for _, tt := range tests { //nolint:paralleltest
		t.Run(tt.name, func(t *testing.T) {
			if got := filter(tt.args.slice, tt.args.f); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetScorecardChecks(t *testing.T) { //nolint:paralleltest
	tests := []struct { //nolint:govet
		name        string
		want        []string
		fileContent string
		wantErr     bool
	}{
		{
			name:    "default",
			want:    []string{"Dangerous-Workflow", "Binary-Artifacts", "Branch-Protection", "Code-Review", "Dependency-Update-Tool"}, //nolint:lll
			wantErr: false,
		},
		{
			name: "file with data",
			want: []string{"Binary-Artifacts", "Pinned-Dependencies"},
			fileContent: `[
  "Binary-Artifacts",
  "Pinned-Dependencies"
]`,
			wantErr: false,
		},
	}
	for _, tt := range tests { //nolint:paralleltest
		t.Run(tt.name, func(t *testing.T) {
			dir, err := os.MkdirTemp("", "scorecard-checks")
			defer os.RemoveAll(dir)
			if err != nil {
				t.Errorf("GetScorecardChecks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.fileContent != "" {
				if err := os.WriteFile(path.Join(dir, "scorecard.txt"), []byte(tt.fileContent), 0o644); err != nil { //nolint:gosec
					t.Errorf("GetScorecardChecks() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				t.Setenv("SCORECARD_CHECKS", path.Join(dir, "scorecard.txt"))
			}
			fileName := ""
			if tt.fileContent != "" {
				fileName = path.Join(dir, "scorecard.txt")
			}
			got, err := GetScorecardChecks(fileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetScorecardChecks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetScorecardChecks() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetScore(t *testing.T) { //nolint:paralleltest
	type args struct {
		repo string
	}
	tests := []struct {
		name    string
		args    args
		score   float64
		wantErr bool
	}{
		{
			name: "default",
			args: args{
				repo: "github.com/ossf/scorecard",
			},
			score:   5.0,
			wantErr: false,
		},
		{
			name: "invalid repo",
			args: args{
				repo: "github.com/ossf/invalid",
			},
			score:   0.0,
			wantErr: true,
		},
	}

	for _, tt := range tests { //nolint:paralleltest
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetScorecardResult(tt.args.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetScorecardResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Score < tt.score {
				t.Errorf("GetScorecardResult() got = %v, want %v", got, tt.score)
			}
		})
	}
}

func TestValidate(t *testing.T) { //nolint:paralleltest
	type args struct {
		token     string
		owner     string
		repo      string
		commitSHA string
		pr        string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "default",
			args: args{
				token:     "token",
				owner:     "ossf",
				repo:      "scorecard",
				commitSHA: "commitSHA",
				pr:        "1",
			},
			wantErr: false,
		},
		{
			name: "invalid token",
			args: args{
				owner:     "ossf",
				repo:      "scorecard",
				commitSHA: "commitSHA",
				pr:        "1",
			},
			wantErr: true,
		},
		{
			name: "invalid repo",
			args: args{
				owner:     "ossf",
				token:     "token",
				commitSHA: "commitSHA",
				pr:        "1",
			},
			wantErr: true,
		},
		{
			name: "invalid pr",
			args: args{
				owner:     "ossf",
				repo:      "scorecard",
				token:     "token",
				commitSHA: "commitSHA",
			},
			wantErr: true,
		},
		{
			name: "invalid commitSHA",
			args: args{
				owner: "ossf",
				repo:  "scorecard",
				token: "token",
				pr:    "1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests { //nolint:paralleltest
		t.Run(tt.name, func(t *testing.T) {
			if err := Validate(tt.args.token, tt.args.repo, tt.args.commitSHA, tt.args.pr); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
