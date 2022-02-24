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
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sigstore/cosign/cmd/cosign/cli/rekor"
	"github.com/sigstore/cosign/pkg/cosign"
)

//not setting t.Parallel() here because we are mutating the env variables
//nolint
func Test_scorecardIsFork(t *testing.T) {
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

//not setting t.Parallel() here because we are mutating the env variables
//nolint
func Test_initalizeENVVariables(t *testing.T) {
	//nolint
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
			if tt.inputresultsfileSet {
				defer os.Unsetenv(inputpublishresults)
				os.Setenv(inputresultsfile, tt.inputresultsfile)
			} else {
				os.Unsetenv(inputresultsfile)
			}
			if tt.inputresultsFormatSet {
				defer os.Unsetenv(inputresultsformat)
				os.Setenv(inputresultsformat, tt.inputresultsFormat)
			} else {
				os.Unsetenv(inputresultsformat)
			}
			if tt.inputPublishResultsSet {
				defer os.Unsetenv(inputpublishresults)
				os.Setenv(inputpublishresults, tt.inputPublishResults)
			} else {
				os.Unsetenv(inputpublishresults)
			}
			if tt.githubEventPathSet {
				defer os.Unsetenv(githubEventPath)
				os.Setenv(githubEventPath, tt.githubEventPath)
			} else {
				os.Unsetenv(githubEventPath)
			}
			if err := initalizeENVVariables(); (err != nil) != tt.wantErr {
				t.Errorf("initalizeENVVariables() error = %v, wantErr %v %v", err, tt.wantErr, t.Name())
			}

			envvars := make(map[string]string)
			envvars[enableSarif] = "1"
			envvars[enableLicense] = "1"
			envvars[enableDangerousWorkflow] = "1"

			for k, v := range envvars {
				if os.Getenv(k) != v {
					t.Errorf("%s env var not set correctly %s", k, v)
				}
			}
		})
	}
}

//not setting t.Parallel() here because we are mutating the env variables
//nolint
func Test_updateEnvVariables(t *testing.T) {
	tests := []struct {
		name                string
		outputResultsFormat string
		isPrivateRepo       bool
		wantErr             bool
	}{
		{
			name:                "Success - private repo",
			outputResultsFormat: "json",
			isPrivateRepo:       true,
			wantErr:             false,
		},
		{
			name:                "Success - private repo - sarif",
			outputResultsFormat: "sarif",
			isPrivateRepo:       true,
			wantErr:             false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if err := updateEnvVariables(); (err != nil) != tt.wantErr {
				t.Errorf("updateEnvVariables() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.isPrivateRepo {
				if scorecardPublishResults != "false" {
					t.Errorf("scorecardPublishResults env var should be false")
				}
			}

			if !tt.wantErr && tt.outputResultsFormat == sarif {
				if _, ok := os.LookupEnv(scorecardPolicyFile); ok {
					t.Errorf("enableSarif env var should not be set")
				}
			}
		})
	}
}

//not setting t.Parallel() here because we are mutating the env variables
//nolint
func Test_updateRepoistoryInformation(t *testing.T) {
	type args struct {
		defaultBranch string
		privateRepo   bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Success - private repo",
			args: args{
				defaultBranch: "master",
				privateRepo:   true,
			},
			wantErr: false,
		},
		{
			name: "Success - public repo",
			args: args{
				defaultBranch: "master",
				privateRepo:   false,
			},
			wantErr: false,
		},
		{
			name: "Success - public repo - no default branch",
			args: args{
				defaultBranch: "",
				privateRepo:   false,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if err := updateRepositoryInformation(tt.args.privateRepo, tt.args.defaultBranch); (err != nil) != tt.wantErr {
				t.Errorf("updateRepoistoryInformation() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.args.privateRepo {
				if scorecardPrivateRepository != strconv.FormatBool(tt.args.privateRepo) {
					t.Errorf("scorecardPublishResults env var should be false")
				}
			}
			if tt.args.defaultBranch != "" {
				if scorecardDefaultBranch != fmt.Sprintf("refs/heads/%s", tt.args.defaultBranch) {
					t.Errorf("scorecardDefaultBranch env var should be %s", tt.args.defaultBranch)
				}
			}
		})
	}
}

//not setting t.Parallel() here because we are mutating the env variables
//nolint
func Test_checkIfRequiredENVSet(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Success - all required env vars set",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		envVariables := make(map[string]bool)
		envVariables[githubRepository] = true
		envVariables[githubAuthToken] = true
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantErr {
				for k := range envVariables {
					defer os.Unsetenv(k)
					if err := os.Setenv(k, "true"); err != nil {
						t.Errorf("failed to set env var %s", k)
					}
				}
			}
			if err := checkIfRequiredENVSet(); (err != nil) != tt.wantErr {
				t.Errorf("checkIfRequiredENVSet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

//nolint
func Test_gitHubEventPath(t *testing.T) {
	tests := []struct {
		name                       string
		wantErr                    bool
		shouldgitHubEventPathBeSet bool
		gitHubEventPath            string
	}{
		{
			name:                       "Success - gitHubEventPath set",
			wantErr:                    false,
			shouldgitHubEventPathBeSet: true,
			gitHubEventPath:            "./testdata/fork.json",
		},
		{
			name:                       "Success - gitHubEventPath not set",
			wantErr:                    true,
			shouldgitHubEventPathBeSet: false,
			gitHubEventPath:            "",
		},
		{
			name:                       "Success - gitHubEventPath is empty",
			wantErr:                    true,
			shouldgitHubEventPathBeSet: true,
			gitHubEventPath:            "",
		},
		{
			name:                       "Failure non-existent file",
			wantErr:                    true,
			shouldgitHubEventPathBeSet: true,
			gitHubEventPath:            "./foo.bar.json",
		},
		{
			name:                       "Failure non-existent file",
			wantErr:                    true,
			shouldgitHubEventPathBeSet: true,
			gitHubEventPath:            "./testdata/incorrect.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldgitHubEventPathBeSet {
				if err := os.Setenv(githubEventPath, tt.gitHubEventPath); err != nil {
					t.Errorf("failed to set env var %s", githubEventPath)
				}
				defer os.Unsetenv(githubEventPath)
			}
			if err := gitHubEventPath(); (err != nil) != tt.wantErr {
				t.Errorf("gitHubEventPath() error = %v, wantErr %v %v", err, tt.wantErr, tt.name)
			}
		})
	}
}

// The reason we are not using t.Parallel() here is because we are mutating the env variables
//nolint
func Test_validate(t *testing.T) {
	//nolint
	tests := []struct {
		name                   string
		wantWriter             string
		wantErr                bool
		authToken              string
		scorecardFork          bool
		gitHubEventName        string
		ref                    string
		scorecardDefaultBranch string
	}{

		{
			name:                   "scorecardFork set and failure",
			wantErr:                true,
			authToken:              "",
			scorecardFork:          true,
			gitHubEventName:        "",
			ref:                    "",
			scorecardDefaultBranch: "",
		},
		{
			name:                   "Success - scorecardFork set",
			wantErr:                false,
			authToken:              "token",
			scorecardFork:          false,
			gitHubEventName:        "",
			ref:                    "",
			scorecardDefaultBranch: "",
		},
		{
			name:                   "Success - scorecardFork set",
			wantErr:                true,
			authToken:              "token",
			scorecardFork:          true,
			gitHubEventName:        "pull_request",
			ref:                    "main",
			scorecardDefaultBranch: "main",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &bytes.Buffer{}
			if err := os.Setenv(scorecardFork, strconv.FormatBool(tt.scorecardFork)); err != nil {
				t.Errorf("failed to set env var %s", scorecardFork)
			}
			defer os.Unsetenv(scorecardFork)
			if tt.gitHubEventName != "" {
				if err := os.Setenv(githubEventName, tt.gitHubEventName); err != nil {
					t.Errorf("failed to set env var %s", githubEventName)
				}
				defer os.Unsetenv(githubEventName)
			}
			if tt.ref != "" {
				if err := os.Setenv(githubRef, tt.ref); err != nil {
					t.Errorf("failed to set env var %s", githubRef)
				}
				defer os.Unsetenv(githubRef)
			}
			if tt.scorecardDefaultBranch != "" {
				scorecardDefaultBranch = tt.scorecardDefaultBranch
			}
			if tt.authToken != "" {
				if err := os.Setenv(githubAuthToken, tt.authToken); err != nil {
					t.Errorf("failed to set env var %s", githubAuthToken)
				}
				defer os.Unsetenv(githubAuthToken)
			}
			if err := validate(writer); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_runScorecardSettings(t *testing.T) {
	t.Parallel()
	type args struct {
		githubEventName        string
		scorecardPolicyFile    string
		scorecardResultsFormat string
		scorecardBin           string
		scorecardResultsFile   string
		githubRepository       string
	}
	//nolint
	tests := []struct {
		wantErr bool
		name    string
		args    args
		want    *exec.Cmd
	}{
		{
			name: "Success - scorecardFork set",
			args: args{
				githubEventName:        "pull_request",
				scorecardPolicyFile:    "./testdata/scorecard.yaml",
				scorecardResultsFormat: "json",
				scorecardBin:           "scorecard",
				scorecardResultsFile:   "./testdata/scorecard.json",
				githubRepository:       "foo/bar",
			},
			want: &exec.Cmd{
				Path: "scorecard",
				Args: []string{
					"scorecard",
					"--policy",
					"./testdata/scorecard.yaml",
					"--results-format",
					"json",
					"--results-file",
					"./testdata/scorecard.json",
					"--repo",
					"foo/bar",
				},
			},
		},
		{
			name: "Success - scorecardFork set",
			args: args{
				githubEventName:        "pull_request",
				scorecardPolicyFile:    "./testdata/scorecard.yaml",
				scorecardResultsFormat: "json",
				scorecardBin:           "scorecard",
				scorecardResultsFile:   "./testdata/scorecard.json",
				githubRepository:       "foo/bar",
			},
			want: &exec.Cmd{
				Path: "scorecard",
				Args: []string{
					"scorecard",
					"--policy",
					"./testdata/scorecard.yaml",
					"--results-format",
					"json",
					"--results-file",
					"./testdata/scorecard.json",
					"--repo",
					"foo/bar",
				},
			},
		},
		{
			name: "Success - scorecardFork set",
			args: args{
				githubEventName:        "pull_request",
				scorecardPolicyFile:    "./testdata/scorecard.yaml",
				scorecardResultsFormat: "json",
				scorecardBin:           "scorecard",
				scorecardResultsFile:   "./testdata/scorecard.json",
				githubRepository:       "foo/bar",
			},
			want: &exec.Cmd{
				Path: "scorecard",
				Args: []string{
					"scorecard",
					"--policy",
					"./testdata/scorecard.yaml",
					"--results-format",
					"json",
					"--results-file",
					"./testdata/scorecard.json",
					"--repo",
					"foo/bar",
				},
			},
		},
		{
			name: "Success - scorecardFork set",
			args: args{
				githubEventName:        "pull_request",
				scorecardResultsFormat: "json",
				scorecardBin:           "scorecard",
				scorecardResultsFile:   "./testdata/scorecard.json",
				githubRepository:       "foo/bar",
			},
			want: &exec.Cmd{
				Path: "scorecard",
				Args: []string{
					"scorecard",
					"--results-format",
					"json",
					"--results-file",
					"./testdata/scorecard.json",
					"--repo",
					"foo/bar",
				},
			},
		},
		{
			name: "Success - scorecardFork set",
			args: args{
				githubEventName:        "pull_request",
				scorecardResultsFormat: "json",
				scorecardBin:           "scorecard",
				scorecardResultsFile:   "./testdata/scorecard.json",
				githubRepository:       "foo/bar",
			},
			want: &exec.Cmd{
				Path: "scorecard",
				Args: []string{
					"scorecard",
					"--results-format",
					"json",
					"--results-file",
					"./testdata/scorecard.json",
					"--repo",
					"foo/bar",
				},
			},
		},
		{
			name: "Success - scorecardFork set",
			args: args{
				scorecardResultsFormat: "json",
				scorecardBin:           "scorecard",
				scorecardResultsFile:   "./testdata/scorecard.json",
				githubRepository:       "foo/bar",
			},
			want: &exec.Cmd{
				Path: "scorecard",
				Args: []string{
					"scorecard",
					"--results-format",
					"json",
					"--results-file",
					"./testdata/scorecard.json",
					"--repo",
					"foo/bar",
				},
			},
		},
		{
			name: "Success - Branch protection rule",
			args: args{
				githubEventName:        "branch_protection_rule",
				scorecardResultsFormat: "json",
				scorecardBin:           "scorecard",
				scorecardResultsFile:   "./testdata/scorecard.json",
				githubRepository:       "foo/bar",
			},
			want: &exec.Cmd{
				Path: "scorecard",
				Args: []string{
					"scorecard",
					"--results-format",
					"json",
					"--results-file",
					"./testdata/scorecard.json",
					"--repo",
					"foo/bar",
				},
			},
		},
		{
			name: "Success - Branch protection rule",
			args: args{
				scorecardPolicyFile:    "./testdata/scorecard.yaml",
				githubEventName:        "branch_protection_rule",
				scorecardResultsFormat: "json",
				scorecardBin:           "scorecard",
				scorecardResultsFile:   "./testdata/scorecard.json",
				githubRepository:       "foo/bar",
			},
			want: &exec.Cmd{
				Path: "scorecard",
				Args: []string{
					"scorecard",
					"--policy",
					"./testdata/scorecard.yaml",
					"--results-format",
					"json",
					"--results-file",
					"./testdata/scorecard.json",
					"--repo",
					"foo/bar",
				},
			},
		},
		{
			name: "Want error - Branch protection rule",
			args: args{
				githubEventName:        "",
				scorecardResultsFormat: "",
				scorecardBin:           "",
				scorecardResultsFile:   "",
				githubRepository:       "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := runScorecardSettings(tt.args.githubEventName, tt.args.scorecardPolicyFile,
				tt.args.scorecardResultsFormat, tt.args.scorecardBin, tt.args.scorecardResultsFile, tt.args.githubRepository)
			if (err != nil) != tt.wantErr {
				t.Errorf("runScorecardSettings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && cmp.Equal(got.Args, tt.want.Args) {
				t.Errorf("runScorecardSettings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_signScorecardResult(t *testing.T) {
	// Sign example scorecard results file.
	scorecardResultsFile := "./testdata/scorecard-results-example.sarif"
	err := signScorecardResult(scorecardResultsFile)
	if err != nil {
		t.Errorf("signScorecardResult() error, %v", err)
		return
	}

	// Verify that the signature was created and uploaded to the Rekor tlog by looking up the payload.
	ctx := context.Background()
	rekorClient, _ := rekor.NewClient("https://rekor.sigstore.dev")
	scorecardResultData, err := ioutil.ReadFile("./testdata/scorecard-results-example.sarif")
	if err != nil {
		t.Errorf("signScorecardResult() error reading scorecard result file, %v", err)
		return
	}
	uuids, _ := cosign.FindTLogEntriesByPayload(ctx, rekorClient, scorecardResultData)

	if len(uuids) == 0 {
		t.Errorf("signScorecardResult() error finding signature in Rekor tlog, %v", err)
		return
	}
}
