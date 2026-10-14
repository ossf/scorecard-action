// Copyright 2026 OpenSSF Scorecard Authors
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

// Package workflowtemplate renders the Scorecard workflow that adopters copy
// into their repositories. Every distributed copy is generated from one
// embedded template, with actions pinned by the embedded pins/action.yml.
package workflowtemplate

import (
	"bufio"
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"text/template"
)

// Variant selects the permissions and publishing settings of a workflow.
type Variant string

const (
	// Public is for public repositories. It publishes results to the
	// Scorecard API.
	Public Variant = "public"
	// Private is for private repositories. It grants the extra read
	// permissions Scorecard needs there, and doesn't publish results.
	Private Variant = "private"
)

var (
	//go:embed scorecard.yml.tmpl
	templateSource string

	//go:embed pins/action.yml
	pinsSource string

	// Matches `uses: owner/repo[/path]@<full SHA>  # <version>`.
	pinPattern = regexp.MustCompile(`^\s*-?\s*uses:\s*([\w.-]+/[\w./-]+)@([0-9a-f]{40})\s+#\s*(\S+)\s*$`)

	errUnknownVariant = errors.New("unknown variant")
	errEmptyOption    = errors.New("option must not be empty")
	errInvalidPin     = errors.New("invalid pin")
	errDuplicatePin   = errors.New("duplicate pin")
	errMissingPin     = errors.New("no pin for action")
)

// Options configure a rendered workflow.
type Options struct {
	// Variant selects public or private repository settings.
	Variant Variant
	// Branch is the default branch that triggers the workflow on push.
	Branch string
	// Cron is the weekly schedule.
	Cron string
}

// Pin is an action pinned to a full commit SHA.
type Pin struct {
	Action  string
	SHA     string
	Version string
}

// String formats the pin as a `uses:` value with a version comment.
func (p Pin) String() string {
	return fmt.Sprintf("%s@%s  # %s", p.Action, p.SHA, p.Version)
}

// Pins returns the pinned actions, keyed by action.
func Pins() (map[string]Pin, error) {
	return parsePins(pinsSource)
}

// Render renders the workflow for the given options.
func Render(opts Options) ([]byte, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	pins, err := Pins()
	if err != nil {
		return nil, err
	}
	return render(templateSource, pins, opts)
}

func render(source string, pins map[string]Pin, opts Options) ([]byte, error) {
	funcs := template.FuncMap{
		"pin": func(action string) (string, error) {
			p, ok := pins[action]
			if !ok {
				return "", fmt.Errorf("%w: %s", errMissingPin, action)
			}
			return p.String(), nil
		},
	}
	// GitHub expressions use `${{ }}`, so the template uses `[[ ]]`.
	tmpl, err := template.New("scorecard.yml").Delims("[[", "]]").Funcs(funcs).Option("missingkey=error").Parse(source)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}
	data := struct {
		Options
		Private bool
	}{
		Options: opts,
		Private: opts.Variant == Private,
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("rendering template: %w", err)
	}
	return buf.Bytes(), nil
}

func (o Options) validate() error {
	switch o.Variant {
	case Public, Private:
	default:
		return fmt.Errorf("%w: %q", errUnknownVariant, o.Variant)
	}
	if o.Branch == "" {
		return fmt.Errorf("%w: Branch", errEmptyOption)
	}
	if o.Cron == "" {
		return fmt.Errorf("%w: Cron", errEmptyOption)
	}
	return nil
}

func parsePins(source string) (map[string]Pin, error) {
	pins := make(map[string]Pin)
	scanner := bufio.NewScanner(strings.NewReader(source))
	for line := 1; scanner.Scan(); line++ {
		text := scanner.Text()
		if !strings.Contains(text, "uses:") {
			continue
		}
		m := pinPattern.FindStringSubmatch(text)
		if m == nil {
			return nil, fmt.Errorf("%w on line %d: want `uses: <action>@<full SHA>  # <version>`, got %q",
				errInvalidPin, line, strings.TrimSpace(text))
		}
		p := Pin{Action: m[1], SHA: m[2], Version: m[3]}
		if _, ok := pins[p.Action]; ok {
			return nil, fmt.Errorf("%w on line %d: %s", errDuplicatePin, line, p.Action)
		}
		pins[p.Action] = p
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading pins: %w", err)
	}
	return pins, nil
}

// Example is a rendered workflow committed under examples/.
type Example struct {
	// Path is relative to the examples directory.
	Path    string
	Options Options
}

// Branch and schedule for the examples adopters copy.
const (
	exampleBranch = "main"
	exampleCron   = "15 14 * * 0"
)

// Examples are the workflows `make generate` renders into examples/.
var Examples = []Example{
	{
		Path:    "scorecard.yml",
		Options: Options{Variant: Public, Branch: exampleBranch, Cron: exampleCron},
	},
	{
		Path:    "scorecard-private.yml",
		Options: Options{Variant: Private, Branch: exampleBranch, Cron: exampleCron},
	},
}
