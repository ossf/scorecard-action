// Copyright OpenSSF Authors
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

//nolint
package options

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ossf/scorecard/v5/options"
)

const (
	testRepo        = "good/repo"
	testResultsFile = "results.sarif"
	testToken       = "test-token"

	githubEventPathNonFork   = "testdata/non-fork.json"
	githubEventPathFork      = "testdata/fork.json"
	githubEventPathIncorrect = "testdata/incorrect.json"
	githubEventPathBadPath   = "testdata/bad-path.json"
	githubEventPathBadData   = "testdata/bad-data.json"
	githubEventPathPublic    = "testdata/public.json"
)

func TestNew(t *testing.T) {
	type fields struct {
		EnableSarif bool
		Format      string
		PolicyFile  string
		ResultsFile string
		Commit      string
		LogLevel    string
		Repo        string
		Local       string
		ChecksToRun []string
		ShowDetails bool
		FileMode    string
	}
	tests := []struct {
		name             string
		githubEventPath  string
		githubEventName  string
		githubRef        string
		repo             string
		resultsFile      string
		resultsFormat    string
		publishResults   string
		fileMode         string
		want             fields
		unsetResultsPath bool
		unsetToken       bool
		wantErr          bool
	}{
		{
			name:            "SuccessFormatSARIF",
			githubEventPath: githubEventPathNonFork,
			githubEventName: pushEvent,
			githubRef:       "refs/heads/main",
			repo:            testRepo,
			resultsFormat:   "sarif",
			resultsFile:     testResultsFile,
			fileMode:        options.FileModeArchive,
			want: fields{
				EnableSarif: true,
				Format:      formatSarif,
				PolicyFile:  defaultScorecardPolicyFile,
				ResultsFile: testResultsFile,
				Commit:      options.DefaultCommit,
				LogLevel:    options.DefaultLogLevel,
				Repo:        testRepo,
				ShowDetails: true,
				FileMode:    options.FileModeArchive,
			},
			wantErr: false,
		},
		{
			name:            "SuccessFormatJSON",
			githubEventPath: githubEventPathNonFork,
			githubEventName: pushEvent,
			githubRef:       "refs/heads/main",
			repo:            testRepo,
			resultsFormat:   "json",
			resultsFile:     testResultsFile,
			fileMode:        options.FileModeArchive,
			want: fields{
				EnableSarif: true,
				Format:      options.FormatJSON,
				ResultsFile: testResultsFile,
				Commit:      options.DefaultCommit,
				LogLevel:    options.DefaultLogLevel,
				Repo:        testRepo,
				ShowDetails: true,
				FileMode:    options.FileModeArchive,
			},
			wantErr: false,
		},
		{
			name:            "SuccessFileModeGit",
			githubEventPath: githubEventPathNonFork,
			githubEventName: pushEvent,
			githubRef:       "refs/heads/main",
			repo:            testRepo,
			resultsFormat:   "sarif",
			resultsFile:     testResultsFile,
			fileMode:        options.FileModeGit,
			want: fields{
				EnableSarif: true,
				Format:      formatSarif,
				PolicyFile:  defaultScorecardPolicyFile,
				ResultsFile: testResultsFile,
				Commit:      options.DefaultCommit,
				LogLevel:    options.DefaultLogLevel,
				Repo:        testRepo,
				ShowDetails: true,
				FileMode:    options.FileModeGit,
			},
			wantErr: false,
		},
		{
			name:            "SuccessPullRequest",
			githubEventPath: githubEventPathNonFork,
			githubEventName: pullRequestEvent,
			githubRef:       "refs/heads/pr-branch",
			repo:            testRepo,
			resultsFormat:   "json",
			resultsFile:     testResultsFile,
			fileMode:        options.FileModeArchive,
			want: fields{
				EnableSarif: true,
				Format:      options.FormatJSON,
				ResultsFile: testResultsFile,
				Commit:      options.DefaultCommit,
				LogLevel:    options.DefaultLogLevel,
				Local:       ".",
				ShowDetails: true,
				FileMode:    options.FileModeArchive,
			},
			wantErr: false,
		},
		{
			name:            "SuccessBranchProtectionEvent",
			githubEventPath: githubEventPathNonFork,
			githubEventName: branchProtectionEvent,
			githubRef:       "refs/heads/main",
			repo:            testRepo,
			resultsFormat:   "json",
			resultsFile:     testResultsFile,
			fileMode:        options.FileModeArchive,
			want: fields{
				EnableSarif: true,
				Format:      options.FormatJSON,
				ResultsFile: testResultsFile,
				Commit:      options.DefaultCommit,
				LogLevel:    options.DefaultLogLevel,
				Repo:        testRepo,
				ShowDetails: true,
				FileMode:    options.FileModeArchive,
			},
			wantErr: false,
		},
		{
			name:            "FailureTokenIsNotSet",
			githubEventPath: githubEventPathNonFork,
			githubEventName: pushEvent,
			githubRef:       "refs/heads/main",
			repo:            testRepo,
			resultsFormat:   "sarif",
			resultsFile:     testResultsFile,
			fileMode:        options.FileModeArchive,
			want: fields{
				EnableSarif: true,
				Format:      formatSarif,
				PolicyFile:  defaultScorecardPolicyFile,
				ResultsFile: testResultsFile,
				Commit:      options.DefaultCommit,
				LogLevel:    options.DefaultLogLevel,
				Repo:        testRepo,
				ShowDetails: true,
				FileMode:    options.FileModeArchive,
			},
			unsetToken: true,
			wantErr:    true,
		},
		{
			name:            "FailureResultsPathNotSet",
			githubEventPath: githubEventPathNonFork,
			githubEventName: pushEvent,
			githubRef:       "refs/heads/main",
			fileMode:        options.FileModeArchive,
			want: fields{
				EnableSarif: true,
				Format:      formatSarif,
				PolicyFile:  defaultScorecardPolicyFile,
				Commit:      options.DefaultCommit,
				LogLevel:    options.DefaultLogLevel,
				ShowDetails: true,
				FileMode:    options.FileModeArchive,
			},
			unsetResultsPath: true,
			wantErr:          true,
		},
		{
			name:            "FailureResultsPathEmpty",
			githubEventPath: githubEventPathNonFork,
			githubEventName: pushEvent,
			githubRef:       "refs/heads/main",
			resultsFile:     "",
			fileMode:        options.FileModeArchive,
			want: fields{
				EnableSarif: true,
				Format:      formatSarif,
				PolicyFile:  defaultScorecardPolicyFile,
				ResultsFile: "",
				Commit:      options.DefaultCommit,
				LogLevel:    options.DefaultLogLevel,
				ShowDetails: true,
				FileMode:    options.FileModeArchive,
			},
			wantErr: true,
		},
		{
			name:            "FailureBranchIsntMain",
			githubEventPath: githubEventPathNonFork,
			githubEventName: pushEvent,
			githubRef:       "refs/heads/other-branch",
			repo:            testRepo,
			resultsFormat:   "sarif",
			resultsFile:     testResultsFile,
			fileMode:        options.FileModeArchive,
			want: fields{
				EnableSarif: true,
				Format:      formatSarif,
				PolicyFile:  defaultScorecardPolicyFile,
				ResultsFile: testResultsFile,
				Commit:      options.DefaultCommit,
				LogLevel:    options.DefaultLogLevel,
				Repo:        testRepo,
				ShowDetails: true,
				FileMode:    options.FileModeArchive,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(EnvGithubAuthToken, testToken)
			defer os.Unsetenv(EnvGithubAuthToken)

			os.Setenv(EnvInputRepoToken, "token-value-123456")
			defer os.Unsetenv(EnvInputRepoToken)

			if tt.unsetToken {
				os.Unsetenv(EnvGithubAuthToken)
				os.Unsetenv(EnvInputRepoToken)
			}

			os.Setenv(EnvGithubEventPath, tt.githubEventPath)
			defer os.Unsetenv(EnvGithubEventPath)

			os.Setenv(EnvGithubEventName, tt.githubEventName)
			defer os.Unsetenv(EnvGithubEventName)

			os.Setenv(EnvGithubRef, tt.githubRef)
			defer os.Unsetenv(EnvGithubRef)

			os.Setenv(EnvGithubRepository, tt.repo)
			defer os.Unsetenv(EnvGithubRepository)

			os.Setenv(EnvInputResultsFormat, tt.resultsFormat)
			defer os.Unsetenv(EnvInputResultsFormat)

			t.Setenv(EnvInputFileMode, tt.fileMode)

			if tt.unsetResultsPath {
				os.Unsetenv(EnvInputResultsFile)
			} else {
				os.Setenv(EnvInputResultsFile, tt.resultsFile)
				defer os.Unsetenv(EnvInputResultsFile)
			}

			opts, err := New()
			scOpts := *opts.ScorecardOpts
			got := fields{
				EnableSarif: scOpts.EnableSarif,
				Format:      scOpts.Format,
				PolicyFile:  scOpts.PolicyFile,
				ResultsFile: scOpts.ResultsFile,
				Commit:      scOpts.Commit,
				LogLevel:    scOpts.LogLevel,
				Repo:        scOpts.Repo,
				Local:       scOpts.Local,
				ChecksToRun: scOpts.ChecksToRun,
				ShowDetails: scOpts.ShowDetails,
				FileMode:    opts.InputFileMode,
			}

			if err != nil {
				t.Fatalf("New(): %v", err)
			}
			if !cmp.Equal(tt.want, got) {
				t.Errorf("New(): -want, +got:\n%s", cmp.Diff(tt.want, got))
			}

			if err := opts.Validate(); (err != nil) != tt.wantErr {
				for _, e := range os.Environ() {
					t.Log(e)
				}
				t.Errorf("Validate() error = %+v, wantErr %+v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestSetRepoInfo(t *testing.T) {
	type fields struct {
		ScorecardOpts           *options.Options
		EnabledChecks           string
		EnableLicense           string
		EnableDangerousWorkflow string
		GithubEventName         string
		GithubEventPath         string
		GithubRef               string
		GithubRepository        string
		GithubWorkspace         string
		DefaultBranch           string
		IsForkStr               string
		PrivateRepoStr          string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Success",
			fields: fields{
				GithubEventPath: githubEventPathNonFork,
			},
			wantErr: false,
		},
		{
			name:    "FailureNoFieldsSet",
			wantErr: true,
		},
		{
			name: "FailureBadEventPath",
			fields: fields{
				GithubEventPath: githubEventPathBadPath,
			},
			wantErr: true,
		},
		{
			name: "FailureBadEventData",
			fields: fields{
				GithubEventPath: githubEventPathBadData,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Options{
				ScorecardOpts:           tt.fields.ScorecardOpts,
				EnabledChecks:           tt.fields.EnabledChecks,
				EnableLicense:           tt.fields.EnableLicense,
				EnableDangerousWorkflow: tt.fields.EnableDangerousWorkflow,
				GithubEventName:         tt.fields.GithubEventName,
				GithubEventPath:         tt.fields.GithubEventPath,
				GithubRef:               tt.fields.GithubRef,
				GithubRepository:        tt.fields.GithubRepository,
				GithubWorkspace:         tt.fields.GithubWorkspace,
				DefaultBranch:           tt.fields.DefaultBranch,
				IsForkStr:               tt.fields.IsForkStr,
				PrivateRepoStr:          tt.fields.PrivateRepoStr,
			}
			if err := o.setRepoInfo(); (err != nil) != tt.wantErr {
				t.Errorf("Options.setRepoInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPrint(t *testing.T) {
	type fields struct {
		ScorecardOpts           *options.Options
		EnabledChecks           string
		EnableLicense           string
		EnableDangerousWorkflow string
		GithubEventName         string
		GithubEventPath         string
		GithubRef               string
		GithubRepository        string
		GithubWorkspace         string
		DefaultBranch           string
		IsForkStr               string
		PrivateRepoStr          string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "Success",
			fields: fields{
				ScorecardOpts: options.New(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Options{
				ScorecardOpts:           tt.fields.ScorecardOpts,
				EnabledChecks:           tt.fields.EnabledChecks,
				EnableLicense:           tt.fields.EnableLicense,
				EnableDangerousWorkflow: tt.fields.EnableDangerousWorkflow,
				GithubEventName:         tt.fields.GithubEventName,
				GithubEventPath:         tt.fields.GithubEventPath,
				GithubRef:               tt.fields.GithubRef,
				GithubRepository:        tt.fields.GithubRepository,
				GithubWorkspace:         tt.fields.GithubWorkspace,
				DefaultBranch:           tt.fields.DefaultBranch,
				IsForkStr:               tt.fields.IsForkStr,
				PrivateRepoStr:          tt.fields.PrivateRepoStr,
			}
			o.Print()
		})
	}
}

func TestSetPublishResults(t *testing.T) {
	tests := []struct {
		name        string
		privateRepo string
		userInput   bool
		want        bool
	}{
		{
			name: "DefaultNoInput",
			want: false,
		},
		{
			name:        "InputTruePrivateRepo",
			privateRepo: "true",
			userInput:   true,
			want:        false,
		},
		{
			name:        "InvalidValueForPrivateRepo",
			privateRepo: "invalid-value",
			want:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &Options{
				ScorecardOpts: options.New(),
			}
			opts.PrivateRepoStr = tt.privateRepo

			opts.setPublishResults()
			got := opts.PublishResults

			if !cmp.Equal(tt.want, got) {
				t.Errorf("New(): -want, +got:\n%s", cmp.Diff(tt.want, got))
			}
		})
	}
}
