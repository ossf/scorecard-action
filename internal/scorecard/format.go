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

package scorecard

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ossf/scorecard-action/options"
	"github.com/ossf/scorecard/v5/docs/checks"
	sclog "github.com/ossf/scorecard/v5/log"
	"github.com/ossf/scorecard/v5/pkg/scorecard"
	"github.com/ossf/scorecard/v5/policy"
)

const (
	defaultScorecardPolicyFile = "/policy.yml"
)

var (
	errNoResult      = errors.New("must provide a result")
	errUnknownFormat = errors.New("unknown result format")
)

// Format provides a wrapper around the Scorecard library's various formatting functions,
// converting our options into theirs.
func Format(result *scorecard.Result, opts *options.Options) error {
	if result == nil {
		return errNoResult
	}

	// write results to both stdout and result file
	resultFile, err := os.Create(opts.GithubWorkspace + opts.InputResultsFile)
	if err != nil {
		return fmt.Errorf("creating result file: %w", err)
	}
	defer resultFile.Close()
	writer := io.MultiWriter(resultFile, os.Stdout)

	docs, err := checks.Read()
	if err != nil {
		return fmt.Errorf("read check docs: %w", err)
	}

	switch strings.ToLower(opts.InputResultsFormat) {
	// sarif is considered the default format when unset
	case "", "sarif":
		if opts.ScorecardOpts.PolicyFile == "" {
			opts.ScorecardOpts.PolicyFile = defaultScorecardPolicyFile
		}
		pol, err := policy.ParseFromFile(opts.ScorecardOpts.PolicyFile)
		if err != nil {
			return fmt.Errorf("parse policy file: %w", err)
		}
		err = result.AsSARIF(true, sclog.DefaultLevel, writer, docs, pol, opts.ScorecardOpts)
		if err != nil {
			return fmt.Errorf("format as sarif: %w", err)
		}
	case "json":
		err = result.AsJSON2(writer, docs, &scorecard.AsJSON2ResultOption{
			Details:     true,
			Annotations: false, // TODO
			LogLevel:    sclog.DefaultLevel,
		})
		if err != nil {
			return fmt.Errorf("format as JSON: %w", err)
		}
	default:
		return errUnknownFormat
	}

	return nil
}
