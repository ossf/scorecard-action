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

// Command generate renders the example Scorecard workflows.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ossf/scorecard-action/internal/workflowtemplate"
)

func main() {
	out := flag.String("out", "examples", "directory to write the rendered workflows to")
	flag.Parse()
	if err := generate(*out); err != nil {
		log.Fatal(err)
	}
}

func generate(out string) error {
	for _, ex := range workflowtemplate.Examples {
		content, err := workflowtemplate.Render(ex.Options)
		if err != nil {
			return fmt.Errorf("rendering %s: %w", ex.Path, err)
		}
		path := filepath.Join(out, ex.Path)
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			return fmt.Errorf("creating directory for %s: %w", path, err)
		}
		if err := os.WriteFile(path, content, 0o600); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}
	}
	return nil
}
