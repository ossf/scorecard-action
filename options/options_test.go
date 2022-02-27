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

package options

import (
	"os"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ossf/scorecard-action/env"
	scopts "github.com/ossf/scorecard/v4/options"
)

//nolint:paralleltest // Until/unless we consider providing a fake environment
// to tests, running these in parallel will have unpredictable results as
// we're mutating environment variables.
func TestNew(t *testing.T) {
	tests := []struct {
		want *Options
		name string
	}{
		{
			name: "Success",
			want: &Options{
				ScorecardOpts: &scopts.Options{
					PolicyFile: defaultScorecardPolicyFile,
				},
				ScorecardBin: defaultScorecardBin,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(); !cmp.Equal(got, tt.want) {
				t.Errorf("New() = %v, want %v: %v", got, tt.want, cmp.Diff(got, tt.want))
			}
		})
	}
}

//nolint:paralleltest // Until/unless we consider providing a fake environment
// to tests, running these in parallel will have unpredictable results as
// we're mutating environment variables.
func TestOptionsInitialize(t *testing.T) {
	type fields struct {
		ScorecardOpts   *scopts.Options
		GithubEventName string
		ScorecardBin    string
		DefaultBranch   string
		PrivateRepo     string
		PublishResults  string
		ResultsFile     string
	}
	tests := []struct {
		name                 string
		fields               fields
		wantErr              bool
		setEnvResultsFile    bool
		setEnvResultsFormat  bool
		setEnvPrivateRepo    bool
		setEnvPublishResults bool
		isPrivateRepo        bool
	}{
		{
			name: "Success",
			fields: fields{
				ScorecardOpts: &scopts.Options{
					PolicyFile: defaultScorecardPolicyFile,
				},
				ScorecardBin: defaultScorecardBin,
			},
			wantErr:              false,
			setEnvResultsFile:    true,
			setEnvResultsFormat:  true,
			setEnvPrivateRepo:    true,
			setEnvPublishResults: true,
			isPrivateRepo:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnvResultsFile {
				os.Setenv(env.InputResultsFile, "results-file")
			}
			if tt.setEnvResultsFormat {
				os.Setenv(env.InputResultsFormat, "sarif")
			}
			if tt.setEnvPrivateRepo {
				os.Setenv(env.ScorecardPrivateRepo, strconv.FormatBool(tt.isPrivateRepo))
			}
			if tt.setEnvPublishResults {
				os.Setenv(env.InputPublishResults, strconv.FormatBool(!tt.isPrivateRepo))
			}

			o := &Options{
				ScorecardOpts:   tt.fields.ScorecardOpts,
				GithubEventName: tt.fields.GithubEventName,
				ScorecardBin:    tt.fields.ScorecardBin,
				DefaultBranch:   tt.fields.DefaultBranch,
				PrivateRepo:     tt.fields.PrivateRepo,
				PublishResults:  tt.fields.PublishResults,
				ResultsFile:     tt.fields.ResultsFile,
			}
			if err := o.Initialize(); (err != nil) != tt.wantErr {
				t.Errorf("Options.Initialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
