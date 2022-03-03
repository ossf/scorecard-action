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
	"github.com/spf13/cobra"

	"github.com/ossf/scorecard-action/options"
	"github.com/ossf/scorecard/v4/cmd"
	scopts "github.com/ossf/scorecard/v4/options"
)

// New creates a new scorecard command which can be used as an entrypoint for
// GitHub Actions.
func New() *cobra.Command {
	opts := options.New()
	opts.Initialize()
	scOpts := opts.ScorecardOpts

	actionCmd := cmd.New(scOpts)

	actionCmd.Flags().StringVar(
		&scOpts.ResultsFile,
		"output-file",
		scOpts.ResultsFile,
		"path to output results to",
	)

	hiddenFlags := []string{
		scopts.FlagNPM,
		scopts.FlagPyPI,
		scopts.FlagRubyGems,
	}
	for _, f := range hiddenFlags {
		actionCmd.Flags().MarkHidden(f)
	}

	// Add sub-commands.
	actionCmd.AddCommand(printConfigCmd(opts))

	return actionCmd
}

func printConfigCmd(o *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use: "print-config",
		Run: func(cmd *cobra.Command, args []string) {
			o.Print()
		},
	}

	return cmd
}
