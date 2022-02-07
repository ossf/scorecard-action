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
package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func Test_scorecardIsFork(t *testing.T) {
	t.Parallel()
	type args struct {
		ghEventPath string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:    "No event data",
			want:    false,
			wantErr: true,
		},
		{
			name: "Fork event",
			args: args{
				ghEventPath: "./testdata/fork.json",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Non fork event",
			args: args{
				ghEventPath: "./testdata/non-fork.json",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "incorrect event",
			args: args{
				ghEventPath: "./testdata/incorrect.json",
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var data []byte
			var err error
			if tt.args.ghEventPath != "" {
				data, err = ioutil.ReadFile(tt.args.ghEventPath)
				if err != nil {
					t.Errorf("Failed to open test data: %v", err)
				}
			}

			got, err := scorecardIsFork(string(data))
			if (err != nil) != tt.wantErr {
				t.Errorf("%v", err)
				t.Errorf("scorecardIsFork() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("scorecardIsFork() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_initalizeENVVariables(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                   string
		wantErr                bool
		inputresultsfileSet    bool
		inputresultsfile       string
		inputresultsFormatSet  bool
		inputresultsFormat     string
		inputPublishResultsSet bool
		inputPublishResults    string
		githubEventPathSet     bool
		githubEventPath        string
	}{
		{
			name:                   "Success",
			wantErr:                false,
			inputresultsfileSet:    true,
			inputresultsfile:       "./testdata/results.json",
			inputresultsFormatSet:  true,
			inputresultsFormat:     "json",
			inputPublishResultsSet: true,
			inputPublishResults:    "true",
			githubEventPathSet:     true,
			githubEventPath:        "./testdata/fork.json",
		},
		{
			name:                   "Success - no results file",
			wantErr:                true,
			inputresultsfileSet:    false,
			inputresultsfile:       "",
			inputresultsFormatSet:  true,
			inputresultsFormat:     "json",
			inputPublishResultsSet: true,
			inputPublishResults:    "true",
			githubEventPathSet:     true,
			githubEventPath:        "./testdata/fork.json",
		},
		{
			name:                   "Success - no results format",
			wantErr:                true,
			inputresultsfileSet:    true,
			inputresultsfile:       "./testdata/results.json",
			inputresultsFormatSet:  false,
			inputresultsFormat:     "",
			inputPublishResultsSet: true,
			inputPublishResults:    "true",
			githubEventPathSet:     true,
			githubEventPath:        "./testdata/fork.json",
		},
		{
			name:                   "Success - no publish results",
			wantErr:                true,
			inputresultsfileSet:    true,
			inputresultsfile:       "./testdata/results.json",
			inputresultsFormatSet:  true,
			inputresultsFormat:     "json",
			inputPublishResultsSet: false,
			inputPublishResults:    "",
			githubEventPathSet:     true,
			githubEventPath:        "./testdata/fork.json",
		},
		{
			name:                   "Success - no github event path",
			wantErr:                true,
			inputresultsfileSet:    true,
			inputresultsfile:       "./testdata/results.json",
			inputresultsFormatSet:  true,
			inputresultsFormat:     "json",
			inputPublishResultsSet: true,
			inputPublishResults:    "true",
			githubEventPathSet:     false,
			githubEventPath:        "./testdata/fork.json",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.inputresultsfileSet {
				os.Setenv("INPUT_RESULTS_FILE", tt.inputresultsfile)
			} else {
				os.Unsetenv("INPUT_RESULTS_FILE")
			}
			if tt.inputresultsFormatSet {
				os.Setenv("INPUT_RESULTS_FORMAT", tt.inputresultsFormat)
			} else {
				os.Unsetenv("INPUT_RESULTS_FORMAT")
			}
			if tt.inputPublishResultsSet {
				os.Setenv("INPUT_PUBLISH_RESULTS", tt.inputPublishResults)
			} else {
				os.Unsetenv("INPUT_PUBLISH_RESULTS")
			}
			if tt.githubEventPathSet {
				os.Setenv("GITHUB_EVENT_PATH", tt.githubEventPath)
			} else {
				os.Unsetenv("GITHUB_EVENT_PATH")
			}
			if err := initalizeENVVariables(); (err != nil) != tt.wantErr {
				t.Errorf("initalizeENVVariables() error = %v, wantErr %v", err, tt.wantErr)
			}

			envvars := make(map[string]string)
			envvars["ENABLE_SARIF"] = "1"
			envvars["ENABLE_LICENSE"] = "1"
			envvars["ENABLE_DANGEROUS_WORKFLOW"] = "1"
			envvars["SCORECARD_POLICY_FILE"] = "./policy.yml"
			envvars["SCORECARD_BIN"] = "/scorecard"
			envvars["ENABLED_CHECKS"] = ""

			for k, v := range envvars {
				if os.Getenv(k) != v {
					t.Errorf("%s env var not set correctly", k)
				}
			}
		})
	}
}
