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
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ossf/scorecard-action/internal/attestation"
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

// signFunc signs an in-toto statement and returns the signed artifact.
type signFunc func(statement []byte) ([]byte, error)

// Format provides a wrapper around the Scorecard library's various formatting functions,
// converting our options into theirs.
func Format(result *scorecard.Result, opts *options.Options) error {
	return format(result, opts, attestation.Sign)
}

func format(result *scorecard.Result, opts *options.Options, sign signFunc) error {
	if result == nil {
		return errNoResult
	}

	// Produce the whole output before touching the results file: a failure
	// to format or to sign must not leave an empty file behind for a later
	// step to upload.
	output, err := render(result, opts, sign)
	if err != nil {
		return err
	}

	// write results to both stdout and result file
	resultFile, err := os.Create(filepath.Join(opts.GithubWorkspace, opts.InputResultsFile))
	if err != nil {
		return fmt.Errorf("creating result file: %w", err)
	}
	defer resultFile.Close()

	if _, err := io.MultiWriter(resultFile, os.Stdout).Write(output); err != nil {
		return fmt.Errorf("writing results: %w", err)
	}
	return nil
}

// render formats the result as the options ask and returns the bytes to
// write. The in-toto statement is signed when the options say so.
func render(result *scorecard.Result, opts *options.Options, sign signFunc) ([]byte, error) {
	docs, err := checks.Read()
	if err != nil {
		return nil, fmt.Errorf("read check docs: %w", err)
	}

	var output bytes.Buffer
	switch strings.ToLower(opts.InputResultsFormat) {
	// sarif is considered the default format when unset
	case "", "sarif":
		if opts.ScorecardOpts.PolicyFile == "" {
			opts.ScorecardOpts.PolicyFile = defaultScorecardPolicyFile
		}
		pol, err := policy.ParseFromFile(opts.ScorecardOpts.PolicyFile)
		if err != nil {
			return nil, fmt.Errorf("parse policy file: %w", err)
		}
		err = result.AsSARIF(true, sclog.DefaultLevel, &output, docs, pol, opts.ScorecardOpts)
		if err != nil {
			return nil, fmt.Errorf("format as sarif: %w", err)
		}
	case "json":
		err = result.AsJSON2(&output, docs, &scorecard.AsJSON2ResultOption{
			Details:     true,
			Annotations: false, // TODO
			LogLevel:    sclog.DefaultLevel,
		})
		if err != nil {
			return nil, fmt.Errorf("format as JSON: %w", err)
		}
	case "intoto":
		err = result.AsInToto(&output, docs, &scorecard.AsInTotoResultOption{
			AsJSON2ResultOption: scorecard.AsJSON2ResultOption{
				Details:     true,
				Annotations: false, // TODO
				LogLevel:    sclog.DefaultLevel,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("format as in-toto statement: %w", err)
		}
		if !opts.SignAttestation {
			break
		}
		// The results file holds the signed bundle; the statement is its payload.
		signed, err := sign(output.Bytes())
		switch {
		case err == nil:
			return signed, nil
		case opts.SignAttestationRequired:
			return nil, fmt.Errorf("signing in-toto statement: %w", err)
		default:
			// Nobody asked for the signature: it was attempted because
			// the job can request an OIDC token. Its failure must not
			// fail a run that works without it.
			fmt.Printf("::warning::Unable to sign the in-toto attestation: %v. "+
				"The results are written unsigned. Set sign_attestation to true to "+
				"fail the step when signing fails, or to false to skip signing.\n", err)
		}
	default:
		return nil, errUnknownFormat
	}

	return output.Bytes(), nil
}
