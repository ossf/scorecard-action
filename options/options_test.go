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
	"testing"

	"github.com/ossf/scorecard/v4/options"
)

const (
	testRepo = "good/repo"

	githubEventPathNonFork   = "testdata/non-fork.json"
	githubEventPathFork      = "testdata/fork.json"
	githubEventPathIncorrect = "testdata/incorrect.json"
	githubEventPathBadPath   = "testdata/bad-path.json"
	githubEventPathBadData   = "testdata/bad-data.json"
)

func TestInitialize(t *testing.T) {
	type fields struct {
		ScorecardOpts           *options.Options
		EnabledChecks           string
		InputResultsFile        string
		InputResultsFormat      string
		InputPublishResults     string
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
				InputResultsFile:        tt.fields.InputResultsFile,
				InputResultsFormat:      tt.fields.InputResultsFormat,
				InputPublishResults:     tt.fields.InputPublishResults,
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
			if err := o.Initialize(); (err != nil) != tt.wantErr {
				t.Errorf("Options.Initialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
