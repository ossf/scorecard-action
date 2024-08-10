// Copyright 2024 OpenSSF Scorecard Authors
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

// Package scorecard provides functionality to run Scorecard and format the results.
package scorecard

import (
	"context"
	"errors"
	"fmt"

	"github.com/ossf/scorecard-action/options"
	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/clients/githubrepo"
	"github.com/ossf/scorecard/v5/clients/localdir"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/pkg/scorecard"
)

// Run provides a wrapper around the Scorecard library's Run function, converting our options into theirs.
func Run(opts *options.Options) (scorecard.Result, error) {
	repo, err := getRepo(opts)
	if err != nil {
		return scorecard.Result{}, fmt.Errorf("unable to create repo: %w", err)
	}

	result, err := scorecard.Run(context.Background(), repo)
	if err != nil && !errors.Is(err, sce.ErrCheckRuntime) {
		return scorecard.Result{}, fmt.Errorf("scorecard had an error: %w", err)
	}
	return result, nil
}

//nolint:wrapcheck // just a helper
func getRepo(opts *options.Options) (clients.Repo, error) {
	if opts.ScorecardOpts.Local != "" {
		return localdir.MakeLocalDirRepo(opts.ScorecardOpts.Local)
	}
	return githubrepo.MakeGithubRepo(opts.ScorecardOpts.Repo)
}
