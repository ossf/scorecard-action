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

// Package run implements a run subcommand for the scorecard GitHub Action.
package run

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/ossf/scorecard-action/options"
	sccmd "github.com/ossf/scorecard/v4/cmd"
	scopts "github.com/ossf/scorecard/v4/options"
)

const (
	cmdUsage     = `run`
	cmdDescShort = "Run the scorecard command"
)

// New creates a new subcommand which can be used as an entrypoint for GitHub
// Actions.
func New() *cobra.Command {
	opts := options.New()
	scOpts := opts.ScorecardOpts

	// Adapt Scorecard command
	actionCmd := sccmd.New(scOpts)
	actionCmd.Use = cmdUsage
	actionCmd.Short = cmdDescShort
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
		if err := opts.Prepare(); err != nil {
			return fmt.Errorf("preparing options: %w", err)
		}
		if err := opts.Validate(); err != nil {
			return fmt.Errorf("validating options: %w", err)
		}

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

	actionCmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		if out != nil {
			if _, err := out.Seek(0, io.SeekStart); err == nil {
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
		log.Printf(
			"%+v: %+v",
			errHideFlags,
			hideErrs,
		)
	}

	// Add sub-commands.
	actionCmd.AddCommand(printConfigCmd(opts))

	return actionCmd
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
