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

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ossf/scorecard-action/dependencydiff"
	"github.com/ossf/scorecard-action/options"
)

const (
	EVENT_PULL_REQUEST = "pull_request"
)

func main() {
	event := os.Getenv(options.EnvGithubEventName)
	switch event {
	case EVENT_PULL_REQUEST:
		fmt.Println(event)
		// Run the dependency-diff on pull requests.
		ctx := context.Background()
		err := dependencydiff.RunDependencyDiff(ctx)
		if err != nil {
			log.Fatalf("error running dependency-diff: %v", err)
		}
	default:
		fmt.Println(event)
	}
}
