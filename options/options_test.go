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

var (
	githubEventPathNonFork   = "testdata/non-fork.json"
	githubEventPathFork      = "testdata/fork.json"
	githubEventPathIncorrect = "testdata/incorrect.json"
)

/*
func TestNew(t *testing.T) {
	//nolint:paralleltest // Until/unless we consider providing a fake environment
	// to tests, running these in parallel will have unpredictable results as
	// we're mutating environment variables.
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
*/

/*
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
	tests := []struct { //nolint:govet // TODO(lint): Fix
		name               string
		fields             fields
		wantErr            bool
		githubEventPath    string
		setGithubEventPath bool
	}{
		{
			name:               "Success - non-fork",
			wantErr:            false,
			githubEventPath:    githubEventPathNonFork,
			setGithubEventPath: true,
		},
		{
			name:               "Success - fork",
			wantErr:            false,
			githubEventPath:    githubEventPathFork,
			setGithubEventPath: true,
		},
		{
			name:               "Failure - incorrect GitHub events",
			wantErr:            true,
			githubEventPath:    githubEventPathIncorrect,
			setGithubEventPath: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setGithubEventPath {
				os.Setenv(EnvGithubEventPath, tt.githubEventPath)
			}

			o, _ := New() //nolint:errcheck // TODO(lint): Fix
			t.Logf("options before initialization: %+v", o)
			optsBeforeInit := o

			if err := o.Initialize(); (err != nil) != tt.wantErr {
				t.Logf("options after initialization: %+v", o)
				t.Logf("options comparison: %s", cmp.Diff(optsBeforeInit, o))
				t.Errorf("Options.Initialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
*/
