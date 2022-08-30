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

package cli

import (
	"github.com/spf13/cobra"

	"github.com/ossf/scorecard-action/cli/run"
)

// New creates a new scorecard-action root command.
func New() *cobra.Command {
	cmd := &cobra.Command{
		// TODO(cmd): Improve action command usage/description
		Use:               "scorecard-action",
		Short:             "scorecard-action",
		DisableAutoGenTag: true,
		SilenceUsage:      true, // Don't show usage on errors
	}

	// Add sub-commands.
	runCmd, _ := run.New()
	cmd.AddCommand(runCmd)

	return cmd
}
