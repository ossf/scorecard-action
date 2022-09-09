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
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/ossf/scorecard-action/options"
	sccmd "github.com/ossf/scorecard/v4/cmd"
	sce "github.com/ossf/scorecard/v4/errors"
	scopts "github.com/ossf/scorecard/v4/options"
)

// New creates a new scorecard command which can be used as an entrypoint for
// GitHub Actions.
func New() (*cobra.Command, error) {
	opts, err := options.New()
	if err != nil {
		return nil, fmt.Errorf("creating new options: %w", err)
	}
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("validating options: %w", err)
	}
	opts.Print()

	// Adapt Scorecard CMD.
	scOpts := opts.ScorecardOpts
	actionCmd := sccmd.New(scOpts)
	actionCmd.Flags().StringVar(
		&scOpts.ResultsFile,
		"output-file",
		scOpts.ResultsFile,
		"path to output results to",
	)
	actionCmd.Flags().BoolVar(
		&opts.PublishResults,
		"publish",
		opts.PublishResults,
		"if set, results will be published (for public repositories only)",
	)

	// Adapt scorecard's PreRunE to support an output file
	// TODO(scorecard): Move this into scorecard
	var out, stdout *os.File
	actionCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		// TODO: the results file should be completed and validated by the time we get it.
		if scOpts.ResultsFile != "" {
			var err error
			resultsFilePath := fmt.Sprintf("%v/%v", opts.GithubWorkspace, scOpts.ResultsFile)
			out, err = os.Create(resultsFilePath)
			if err != nil {
				return fmt.Errorf(
					"creating output file (%s): %w",
					resultsFilePath,
					err,
				)
			}
			stdout = os.Stdout
			os.Stdout = out
			actionCmd.SetOut(out)
		}
		return nil
	}

	scorecardRunE := actionCmd.RunE
	actionCmd.RunE = func(cmd *cobra.Command, args []string) error {
		err := scorecardRunE(cmd, args)
		if errors.Is(err, sce.ErrorCheckRuntime) {
			err = nil
		}
		return err
	}

	actionCmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		if out != nil {
			if _, err = out.Seek(0, io.SeekStart); err == nil {
				//nolint:errcheck
				_, _ = io.Copy(stdout, out)
			}
			_ = out.Close()
		}
		os.Stdout = stdout
	}

	var hideErrs []error
	hiddenFlags := []string{
		scopts.FlagNPM,
		scopts.FlagPyPI,
		scopts.FlagRubyGems,
	}

	for _, f := range hiddenFlags {
		err := actionCmd.Flags().MarkHidden(f)
		if err != nil {
			hideErrs = append(hideErrs, err)
		}
	}

	if len(hideErrs) > 0 {
		return nil, fmt.Errorf(
			"%w: %+v",
			errHideFlags,
			hideErrs,
		)
	}

	// Add sub-commands.
	actionCmd.AddCommand(printConfigCmd(opts))

	return actionCmd, nil
}

func printConfigCmd(o *options.Options) *cobra.Command {
	c := &cobra.Command{
		Use: "print-config",
		Run: func(cmd *cobra.Command, args []string) {
			o.Print()
		},
	}

	return c
}

var errHideFlags = errors.New("errors occurred while trying to hide scorecard flags")
