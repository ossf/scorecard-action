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
package entrypoint

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ossf/scorecard-action/env"
	"github.com/ossf/scorecard-action/options"
	scopts "github.com/ossf/scorecard/v4/options"
)

//nolint:paralleltest
// Not setting t.Parallel() here because we are mutating the env variables.
func Test_RepoIsFork(t *testing.T) {
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

			got, err := options.RepoIsFork(string(data))
			if (err != nil) != tt.wantErr {
				t.Errorf("%v", err)
				t.Errorf("RepoIsFork() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RepoIsFork() = %v, want %v", got, tt.want)
			}
		})
	}
}

//nolint:paralleltest
// Not setting t.Parallel() here because we are mutating the env variables.
func TestInitializeEnvVariables(t *testing.T) {
	tests := []struct {
		opts                   *options.Options
		name                   string
		githubEventPath        string
		inputResultsFile       string
		inputPublishResults    string
		wantErr                bool
		githubEventPathSet     bool
		inputResultsFileSet    bool
		inputResultsFormatSet  bool
		inputPublishResultsSet bool
	}{
		{
			name:    "Success",
			wantErr: false,
			opts: &options.Options{
				ScorecardOpts: &scopts.Options{
					Format: "json",
				},
			},
			inputResultsFileSet:    true,
			inputResultsFile:       "./testdata/results.json",
			inputResultsFormatSet:  true,
			inputPublishResultsSet: true,
			inputPublishResults:    "true",
			githubEventPathSet:     true,
			githubEventPath:        "./testdata/fork.json",
		},
		{
			name:    "Success - no results file",
			wantErr: true,
			opts: &options.Options{
				ScorecardOpts: &scopts.Options{
					Format: "json",
				},
			},
			inputResultsFileSet:    false,
			inputResultsFile:       "",
			inputResultsFormatSet:  true,
			inputPublishResultsSet: true,
			inputPublishResults:    "true",
			githubEventPathSet:     true,
			githubEventPath:        "./testdata/fork.json",
		},
		{
			name:    "Success - no results format",
			wantErr: true,
			opts: &options.Options{
				ScorecardOpts: &scopts.Options{
					Format: "",
				},
			},
			inputResultsFileSet:    true,
			inputResultsFile:       "./testdata/results.json",
			inputResultsFormatSet:  false,
			inputPublishResultsSet: true,
			inputPublishResults:    "true",
			githubEventPathSet:     true,
			githubEventPath:        "./testdata/fork.json",
		},
		{
			name:    "Success - no publish results",
			wantErr: true,
			opts: &options.Options{
				ScorecardOpts: &scopts.Options{
					Format: "json",
				},
			},
			inputResultsFileSet:    true,
			inputResultsFile:       "./testdata/results.json",
			inputResultsFormatSet:  true,
			inputPublishResultsSet: false,
			inputPublishResults:    "",
			githubEventPathSet:     true,
			githubEventPath:        "./testdata/fork.json",
		},
		{
			name:    "Success - no github event path",
			wantErr: true,
			opts: &options.Options{
				ScorecardOpts: &scopts.Options{
					Format: "json",
				},
			},
			inputResultsFileSet:    true,
			inputResultsFile:       "./testdata/results.json",
			inputResultsFormatSet:  true,
			inputPublishResultsSet: true,
			inputPublishResults:    "true",
			githubEventPathSet:     false,
			githubEventPath:        "./testdata/fork.json",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.inputResultsFileSet {
				defer os.Unsetenv(env.InputPublishResults)
				os.Setenv(env.InputResultsFile, tt.inputResultsFile)
			} else {
				os.Unsetenv(env.InputResultsFile)
			}
			if tt.inputResultsFormatSet {
				defer os.Unsetenv(env.InputResultsFormat)
				os.Setenv(env.InputResultsFormat, tt.opts.ScorecardOpts.Format)
			} else {
				os.Unsetenv(env.InputResultsFormat)
			}
			if tt.inputPublishResultsSet {
				defer os.Unsetenv(env.InputPublishResults)
				os.Setenv(env.InputPublishResults, tt.inputPublishResults)
			} else {
				os.Unsetenv(env.InputPublishResults)
			}
			if tt.githubEventPathSet {
				defer os.Unsetenv(env.GithubEventPath)
				os.Setenv(env.GithubEventPath, tt.githubEventPath)
			} else {
				os.Unsetenv(env.GithubEventPath)
			}
			if err := tt.opts.Initialize(); (err != nil) != tt.wantErr {
				t.Errorf("options.Initialize() error = %v, wantErr %v %v", err, tt.wantErr, t.Name())
			}

			envvars := make(map[string]string)
			envvars[env.EnableSarif] = "1"
			envvars[env.EnableLicense] = "1"
			envvars[env.EnableDangerousWorkflow] = "1"

			for k, v := range envvars {
				if os.Getenv(k) != v {
					t.Errorf("%s env var not set correctly %s", k, v)
				}
			}
		})
	}
}

//nolint:paralleltest
// Not setting t.Parallel() here because we are mutating the env variables.
func TestUpdateEnvVariables(t *testing.T) {
	tests := []struct {
		opts          *options.Options
		name          string
		isPrivateRepo bool
		wantErr       bool
	}{
		{
			name: "Success - private repo",
			opts: &options.Options{
				ScorecardOpts: &scopts.Options{
					Format: "json",
				},
			},
			isPrivateRepo: true,
			wantErr:       false,
		},
		{
			name: "Success - private repo - sarif",
			opts: &options.Options{
				ScorecardOpts: &scopts.Options{
					Format: "sarif",
				},
			},
			isPrivateRepo: true,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.opts.SetRepoVisibility(tt.isPrivateRepo)
			tt.opts.SetPublishResults()
			if !tt.wantErr && tt.isPrivateRepo {
				if tt.opts.PublishResults != "false" {
					t.Errorf("scorecardPublishResults env var (%s) should be false", tt.opts.PublishResults)
				}
			}

			if !tt.wantErr && tt.opts.ScorecardOpts.Format == "sarif" {
				if _, ok := os.LookupEnv(tt.opts.ScorecardOpts.PolicyFile); ok {
					t.Errorf("envEnableSarif env var should not be set")
				}
			}
		})
	}
}

//nolint:paralleltest
// Not setting t.Parallel() here because we are mutating the env variables.
func TestUpdateRepositoryInformation(t *testing.T) {
	// Not setting t.Parallel() here because we are mutating the env variables
	type args struct {
		opts          *options.Options
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
				opts:          &options.Options{},
				defaultBranch: "master",
				privateRepo:   true,
			},
			wantErr: false,
		},
		{
			name: "Success - public repo",
			args: args{
				opts:          &options.Options{},
				defaultBranch: "master",
				privateRepo:   false,
			},
			wantErr: false,
		},
		{
			name: "Success - public repo - no default branch",
			args: args{
				opts:          &options.Options{},
				defaultBranch: "",
				privateRepo:   false,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.args.opts.SetDefaultBranch(tt.args.defaultBranch); (err != nil) != tt.wantErr {
				t.Errorf("options.SetDefaultBranch() error = %v, wantErr %v", err, tt.wantErr)
			}

			tt.args.opts.SetRepoVisibility(tt.args.privateRepo)

			if tt.args.privateRepo {
				if tt.args.opts.PrivateRepo != strconv.FormatBool(tt.args.privateRepo) {
					t.Errorf("scorecardPublishResults env var should be false")
				}
			}
			if tt.args.defaultBranch != "" {
				if tt.args.opts.DefaultBranch != fmt.Sprintf("refs/heads/%s", tt.args.defaultBranch) {
					t.Errorf("scorecardDefaultBranch env var should be %s", tt.args.defaultBranch)
				}
			}
		})
	}
}

//nolint:paralleltest
// Not setting t.Parallel() here because we are mutating the env variables.
func TestCheckRequired(t *testing.T) {
	tests := []struct {
		opts    *options.Options
		name    string
		wantErr bool
	}{
		{
			name:    "Success - all required env vars set",
			opts:    options.New(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		envVariables := make(map[string]bool)
		envVariables[env.GithubRepository] = true
		envVariables[env.GithubAuthToken] = true
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantErr {
				for k := range envVariables {
					defer os.Unsetenv(k)
					if err := os.Setenv(k, "true"); err != nil {
						t.Errorf("failed to set env var %s", k)
					}
				}
			}

			if err := tt.opts.CheckRequired(); (err != nil) != tt.wantErr {
				t.Errorf("options.CheckRequired() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

//nolint:paralleltest
// Not setting t.Parallel() here because we are mutating the env variables.
func TestGithubEventPath(t *testing.T) {
	tests := []struct {
		name                          string
		githubEventPath               string
		shouldEnvGithubEventPathBeSet bool
		wantErr                       bool
		isFork                        bool
	}{
		{
			name:                          "Success - githubEventPath set",
			wantErr:                       false,
			shouldEnvGithubEventPathBeSet: true,
			githubEventPath:               "./testdata/non-fork.json",
			isFork:                        false,
		},
		{
			name:                          "Success - githubEventPath not set",
			wantErr:                       true,
			shouldEnvGithubEventPathBeSet: false,
			githubEventPath:               "",
		},
		{
			name:                          "Success - githubEventPath is empty",
			wantErr:                       true,
			shouldEnvGithubEventPathBeSet: true,
			githubEventPath:               "",
		},
		{
			name:                          "Failure non-existent file",
			wantErr:                       true,
			shouldEnvGithubEventPathBeSet: true,
			githubEventPath:               "./foo.bar.json",
		},
		{
			name:                          "Failure non-existent file",
			wantErr:                       true,
			shouldEnvGithubEventPathBeSet: true,
			githubEventPath:               "./testdata/incorrect.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldEnvGithubEventPathBeSet {
				if err := os.Setenv(env.GithubEventPath, tt.githubEventPath); err != nil {
					t.Errorf("failed to set env var %s", env.GithubEventPath)
				}
				defer os.Unsetenv(env.GithubEventPath)
			}

			if err := options.GithubEventPath(); (err != nil) != tt.wantErr {
				t.Errorf("options.GithubEventPath() error = %v, wantErr %v %v", err, tt.wantErr, tt.name)
			}

			if tt.isFork {
				forkEnv := os.Getenv(env.ScorecardFork)
				if forkEnv != "true" {
					t.Errorf("isFork = %v, want %v %v", tt.isFork, forkEnv, tt.name)
				}
			}
		})
	}
}

//nolint:paralleltest,gocognit
// Not setting t.Parallel() here because we are mutating the env variables.
func TestValidate(t *testing.T) {
	tests := []struct {
		opts                   *options.Options
		name                   string
		wantWriter             string
		authToken              string
		gitHubEventName        string
		ref                    string
		scorecardDefaultBranch string
		wantErr                bool
		scorecardFork          bool
	}{
		{
			name:                   "scorecardFork set and failure",
			opts:                   &options.Options{},
			wantErr:                true,
			authToken:              "",
			scorecardFork:          true,
			gitHubEventName:        "",
			ref:                    "",
			scorecardDefaultBranch: "",
		},
		{
			name:                   "Success - scorecardFork set",
			opts:                   &options.Options{},
			wantErr:                false,
			authToken:              "token",
			scorecardFork:          false,
			gitHubEventName:        "",
			ref:                    "",
			scorecardDefaultBranch: "",
		},
		{
			name:                   "Success - scorecardFork set",
			opts:                   &options.Options{},
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
			if err := os.Setenv(env.ScorecardFork, strconv.FormatBool(tt.scorecardFork)); err != nil {
				t.Errorf("failed to set env var %s", env.ScorecardFork)
			}
			defer os.Unsetenv(env.ScorecardFork)
			if tt.gitHubEventName != "" {
				if err := os.Setenv(env.GithubEventName, tt.gitHubEventName); err != nil {
					t.Errorf("failed to set env var %s", env.GithubEventName)
				}
				defer os.Unsetenv(env.GithubEventName)
			}
			if tt.ref != "" {
				if err := os.Setenv(env.GithubRef, tt.ref); err != nil {
					t.Errorf("failed to set env var %s", env.GithubRef)
				}
				defer os.Unsetenv(env.GithubRef)
			}
			if tt.scorecardDefaultBranch != "" {
				tt.opts.DefaultBranch = tt.scorecardDefaultBranch
			}
			if tt.authToken != "" {
				if err := os.Setenv(env.GithubAuthToken, tt.authToken); err != nil {
					t.Errorf("failed to set env var %s", env.GithubAuthToken)
				}
				defer os.Unsetenv(env.GithubAuthToken)
			}
			if err := tt.opts.Validate(writer); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestGetScorecardCmd(t *testing.T) {
	t.Parallel()
	type args *options.Options

	//nolint
	tests := []struct {
		wantErr bool
		name    string
		args    args
		want    *exec.Cmd
	}{
		{
			name: "Success - envScorecardFork set",
			args: &options.Options{
				ScorecardOpts: &scopts.Options{
					Repo:       "foo/bar",
					Format:     "json",
					PolicyFile: "./testdata/scorecard.yaml",
				},
				GithubEventName: "pull_request",
				ScorecardBin:    "scorecard",
				ResultsFile:     "./testdata/scorecard.json",
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
			name: "Success - envScorecardFork set",
			args: &options.Options{
				ScorecardOpts: &scopts.Options{
					Repo:       "foo/bar",
					Format:     "json",
					PolicyFile: "./testdata/scorecard.yaml",
				},
				GithubEventName: "pull_request",
				ScorecardBin:    "scorecard",
				ResultsFile:     "./testdata/scorecard.json",
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
			name: "Success - envScorecardFork set",
			args: &options.Options{
				ScorecardOpts: &scopts.Options{
					Repo:       "foo/bar",
					Format:     "json",
					PolicyFile: "./testdata/scorecard.yaml",
				},
				GithubEventName: "pull_request",
				ScorecardBin:    "scorecard",
				ResultsFile:     "./testdata/scorecard.json",
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
			name: "Success - envScorecardFork set",
			args: &options.Options{
				ScorecardOpts: &scopts.Options{
					Repo:   "foo/bar",
					Format: "json",
				},
				GithubEventName: "pull_request",
				ScorecardBin:    "scorecard",
				ResultsFile:     "./testdata/scorecard.json",
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
			name: "Success - envScorecardFork set",
			args: &options.Options{
				ScorecardOpts: &scopts.Options{
					Repo:   "foo/bar",
					Format: "json",
				},
				GithubEventName: "pull_request",
				ScorecardBin:    "scorecard",
				ResultsFile:     "./testdata/scorecard.json",
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
			name: "Success - envScorecardFork set",
			args: &options.Options{
				ScorecardOpts: &scopts.Options{
					Repo:   "foo/bar",
					Format: "json",
				},
				ScorecardBin: "scorecard",
				ResultsFile:  "./testdata/scorecard.json",
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
			args: &options.Options{
				ScorecardOpts: &scopts.Options{
					Repo:   "foo/bar",
					Format: "json",
				},
				GithubEventName: "branch_protection_rule",
				ScorecardBin:    "scorecard",
				ResultsFile:     "./testdata/scorecard.json",
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
			args: &options.Options{
				ScorecardOpts: &scopts.Options{
					Repo:       "foo/bar",
					Format:     "json",
					PolicyFile: "./testdata/scorecard.yaml",
				},
				GithubEventName: "branch_protection_rule",
				ScorecardBin:    "scorecard",
				ResultsFile:     "./testdata/scorecard.json",
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
			args: &options.Options{
				ScorecardOpts: &scopts.Options{
					Repo:   "",
					Format: "",
				},
				GithubEventName: "",
				ScorecardBin:    "",
				ResultsFile:     "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := getScorecardCmd(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("getScorecardCmd() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && cmp.Equal(got.Args, tt.want.Args) {
				t.Errorf("getScorecardCmd() = %v, want %v", got, tt.want)
			}
		})
	}
}
