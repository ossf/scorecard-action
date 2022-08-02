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
	"os"
	"testing"

	"github.com/ossf/scorecard-action/options"
)

func Test_entrypoint(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name      string
		errWanted error
		repoURI   string
		base      string
		head      string
	}{

		{
			name:      "error invalid repo uri",
			errWanted: errInvalid,
		},
		{
			name:      "error empty base ref",
			errWanted: errInvalid,
			repoURI:   "fake_owner/fake_repo",
		},
		{
			name:      "error empty head uri",
			errWanted: errEmpty,
			repoURI:   "fake_owner/fake_repo",
			base:      "fake_base",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			os.Setenv(options.EnvGithubRepository, tt.repoURI)
			os.Setenv(options.EnvGithubBaseRef, tt.base)
			os.Setenv(options.EnvGitHubHeadRef, tt.head)
			err := New(context.Background())
			if errors.Is(tt.errWanted, err) {
				t.Errorf("want err: %v, got: %v", tt.errWanted, err)
			}
		})
	}
}
