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

package workflowtemplate

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

var usesPattern = regexp.MustCompile(`(?m)^\s*(?:-\s*)?uses:\s*(\S+)`)

func TestRender_variants(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		opts        Options
		wantLines   []string
		unwantLines []string
	}{
		{
			name: "public",
			opts: Options{Variant: Public, Branch: exampleBranch, Cron: exampleCron},
			wantLines: []string{
				"    - cron: '15 14 * * 0'",
				"    branches: ['main']",
				"      id-token: write  # Sign results published to api.scorecard.dev.",
				"          publish_results: true",
			},
			unwantLines: []string{
				"      contents: read  # Read the code of a private repository.",
				"          publish_results: false",
			},
		},
		{
			name: "private",
			opts: Options{Variant: Private, Branch: "trunk", Cron: "0 0 * * 1"},
			wantLines: []string{
				"    - cron: '0 0 * * 1'",
				"    branches: ['trunk']",
				"      contents: read  # Read the code of a private repository.",
				"      actions: read  # Read the workflow runs of a private repository.",
				"          publish_results: false",
			},
			unwantLines: []string{
				"      id-token: write  # Sign results published to api.scorecard.dev.",
				"          publish_results: true",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := Render(tt.opts)
			if err != nil {
				t.Fatalf("Render(): %v", err)
			}
			lines := strings.Split(string(got), "\n")
			for _, want := range tt.wantLines {
				if !contains(lines, want) {
					t.Errorf("Render() is missing line %q", want)
				}
			}
			for _, unwant := range tt.unwantLines {
				if contains(lines, unwant) {
					t.Errorf("Render() has unwanted line %q", unwant)
				}
			}
		})
	}
}

func TestRender_pinsEveryAction(t *testing.T) {
	t.Parallel()
	pins, err := Pins()
	if err != nil {
		t.Fatalf("Pins(): %v", err)
	}
	got, err := Render(Options{Variant: Public, Branch: exampleBranch, Cron: exampleCron})
	if err != nil {
		t.Fatalf("Render(): %v", err)
	}
	used := make(map[string]bool)
	for _, m := range usesPattern.FindAllStringSubmatch(string(got), -1) {
		action, sha, _ := strings.Cut(m[1], "@")
		p, ok := pins[action]
		if !ok || sha != p.SHA {
			t.Errorf("Render() uses %s, which isn't the pinned %s", m[1], p)
		}
		used[action] = true
	}
	for action := range pins {
		if !used[action] {
			t.Errorf("pin %s isn't used by the template", action)
		}
	}
}

func TestRender_invalidOptions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		wantErr error
		name    string
		opts    Options
	}{
		{
			name:    "unknown variant",
			opts:    Options{Variant: "internal", Branch: exampleBranch, Cron: exampleCron},
			wantErr: errUnknownVariant,
		},
		{
			name:    "no branch",
			opts:    Options{Variant: Public, Cron: exampleCron},
			wantErr: errEmptyOption,
		},
		{
			name:    "no cron",
			opts:    Options{Variant: Public, Branch: exampleBranch},
			wantErr: errEmptyOption,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if _, err := Render(tt.opts); !errors.Is(err, tt.wantErr) {
				t.Errorf("Render(): got error %v, wanted %v", err, tt.wantErr)
			}
		})
	}
}

func TestRender_missingPin(t *testing.T) {
	t.Parallel()
	_, err := render(`uses: [[ pin "actions/unknown" ]]`, map[string]Pin{}, Options{})
	if !errors.Is(err, errMissingPin) {
		t.Errorf("render(): got error %v, wanted %v", err, errMissingPin)
	}
}

func TestParsePins(t *testing.T) {
	t.Parallel()
	const sha = "3d3c42e5aac5ba805825da76410c181273ba90b1"
	tests := []struct {
		wantErr error
		want    map[string]Pin
		name    string
		source  string
	}{
		{
			name:   "valid",
			source: "steps:\n  - uses: actions/checkout@" + sha + "  # v7.0.1\n    with:\n",
			want:   map[string]Pin{"actions/checkout": {Action: "actions/checkout", SHA: sha, Version: "v7.0.1"}},
		},
		{
			name:   "action with a path",
			source: "  - uses: github/codeql-action/upload-sarif@" + sha + " # v4.38.3\n",
			want: map[string]Pin{
				"github/codeql-action/upload-sarif": {Action: "github/codeql-action/upload-sarif", SHA: sha, Version: "v4.38.3"},
			},
		},
		{
			name:    "tag instead of SHA",
			source:  "  - uses: actions/checkout@v7  # v7\n",
			wantErr: errInvalidPin,
		},
		{
			name:    "short SHA",
			source:  "  - uses: actions/checkout@3d3c42e  # v7.0.1\n",
			wantErr: errInvalidPin,
		},
		{
			name:    "no version comment",
			source:  "  - uses: actions/checkout@" + sha + "\n",
			wantErr: errInvalidPin,
		},
		{
			name:    "duplicate",
			source:  "  - uses: actions/checkout@" + sha + "  # v7.0.1\n  - uses: actions/checkout@" + sha + "  # v7.0.1\n",
			wantErr: errDuplicatePin,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parsePins(tt.source)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("parsePins(): got error %v, wanted %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("parsePins(): got %v, wanted %v", got, tt.want)
			}
			for action, want := range tt.want {
				if got[action] != want {
					t.Errorf("parsePins()[%s]: got %+v, wanted %+v", action, got[action], want)
				}
			}
		})
	}
}

// TestExamples fails when examples/ is stale. Run `make generate` to fix it.
func TestExamples(t *testing.T) {
	t.Parallel()
	for _, ex := range Examples {
		t.Run(ex.Path, func(t *testing.T) {
			t.Parallel()
			want, err := Render(ex.Options)
			if err != nil {
				t.Fatalf("Render(): %v", err)
			}
			got, err := os.ReadFile(filepath.Join("..", "..", "examples", ex.Path))
			if err != nil {
				t.Fatalf("reading example: %v", err)
			}
			if !bytes.Equal(got, want) {
				t.Errorf("examples/%s is stale; run `make generate`", ex.Path)
			}
		})
	}
}

// TestExamples_pinnedToFullSHA checks the committed files directly, so it
// holds even if the template stops using the pin function.
func TestExamples_pinnedToFullSHA(t *testing.T) {
	t.Parallel()
	fullSHA := regexp.MustCompile(`^[\w.-]+/[\w./-]+@[0-9a-f]{40}$`)
	paths, err := filepath.Glob(filepath.Join("..", "..", "examples", "*.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("found no examples to check")
	}
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		matches := usesPattern.FindAllStringSubmatch(string(content), -1)
		if len(matches) == 0 {
			t.Errorf("%s has no uses: lines", path)
		}
		for _, m := range matches {
			if !fullSHA.MatchString(m[1]) {
				t.Errorf("%s: %s isn't pinned to a full commit SHA", path, m[1])
			}
		}
	}
}

// TestExamples_linted fails if the actionlint step in examples.yml, which
// takes an explicit file list, doesn't lint exactly the files in examples/.
func TestExamples_linted(t *testing.T) {
	t.Parallel()
	content, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "examples.yml"))
	if err != nil {
		t.Fatal(err)
	}
	var workflow struct {
		Jobs map[string]struct {
			Steps []struct {
				With map[string]string `yaml:"with"`
				Uses string            `yaml:"uses"`
			} `yaml:"steps"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal(content, &workflow); err != nil {
		t.Fatalf("parsing examples.yml: %v", err)
	}
	var linted []string
	for _, step := range workflow.Jobs["actionlint"].Steps {
		if strings.HasPrefix(step.Uses, "kjanat/actionlint@") {
			linted = strings.Fields(step.With["files"])
		}
	}
	if len(linted) == 0 {
		t.Fatal("examples.yml has no actionlint step with files")
	}

	paths, err := filepath.Glob(filepath.Join("..", "..", "examples", "*.yml"))
	if err != nil {
		t.Fatal(err)
	}
	var want []string
	for _, path := range paths {
		want = append(want, filepath.ToSlash(filepath.Join("examples", filepath.Base(path))))
	}
	sort.Strings(linted)
	sort.Strings(want)
	if !slices.Equal(linted, want) {
		t.Errorf("examples.yml lints %v with actionlint, but examples/ has %v", linted, want)
	}
}

func contains(lines []string, want string) bool {
	for _, line := range lines {
		if line == want {
			return true
		}
	}
	return false
}
