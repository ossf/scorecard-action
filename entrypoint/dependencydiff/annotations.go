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

package dependencydiff

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v45/github"
	"github.com/ossf/scorecard-action/options"
	docs "github.com/ossf/scorecard/v4/docs/checks"
	"github.com/ossf/scorecard/v4/pkg"
)

func visualizeToCheckRun(ctx context.Context, ghClient *github.Client,
	owner, repo string,
	deps []pkg.DependencyCheckResult,
) error {
	headSHA := os.Getenv(options.EnvGithubPullRequestHeadSHA)
	if headSHA == "" {
		return fmt.Errorf("%w: head ref", errEmpty)
	}
	annotations, err := createAnnotations(deps)
	if err != nil {
		return fmt.Errorf("error creating annotations: %w", err)
	}
	output := github.CheckRunOutput{
		Title:       asPointerStr("Scorecard Action Dependency-diff check results"),
		Summary:     asPointerStr("test **11111**"),
		Annotations: annotations,
	}
	opts := github.CreateCheckRunOptions{
		Name:    "Scorecard Action Dependency-diff",
		HeadSHA: headSHA,
		// DetailsURL should be the integrator's site that has the full details of the check.
		// TODO (#issue number): Leave this as nil for now to make it explicit. This might be a
		// corresponding scorecard check page for a specific package once we have the security-scorecard.dev website.
		// https://github.com/google/go-github/blob/master/github/checks.go#L142
		DetailsURL: nil,
		Status:     asPointerStr("completed"),
		Conclusion: asPointerStr("neutral"),
		Output:     &output,
	}
	_, resp, err := ghClient.Checks.CreateCheckRun(
		ctx, owner, repo, opts,
	)
	if err != nil {
		return fmt.Errorf("error creating the check run: %w", err)
	}
	fmt.Println("*************************************")
	fmt.Println(resp.StatusCode)
	fmt.Println("*************************************")
	return nil
}

func createAnnotations(deps []pkg.DependencyCheckResult) ([]*github.CheckRunAnnotation, error) {
	annotations := []*github.CheckRunAnnotation{}
	doc, err := docs.Read()
	if err != nil {
		return nil, fmt.Errorf("error getting the check doc: %w", err)
	}
	for _, d := range deps {
		a := github.CheckRunAnnotation{}
		// No need for nil pointer checking since a.Path is also a pointer type.
		a.Path = d.ManifestPath
		// We don't has a start line and an end line for a dependency-diff since the current data source
		// simply walks through the manifest/lock file and doesn't has such return fields.
		a.StartLine = asPointerInt(1) /* Fake the start line. */
		a.EndLine = asPointerInt(2)   /* Fake the end line. */
		a.AnnotationLevel = asPointerStr("notice")
		if d.ChangeType != nil && d.Version != nil {
			a.Title = asPointerStr(fmt.Sprintf(
				"%s dependency: %s @ %s",
				*d.ChangeType, d.Name, *d.Version,
			))
		} else {
			a.Title = &d.Name
		}
		a.Message = asPointerStr("No Scorecard check results for this dependency.")
		scResult := d.ScorecardResultWithError.ScorecardResult
		if scResult != nil {
			aggregateScore, err := scResult.GetAggregateScore(doc)
			if err != nil {
				return nil, fmt.Errorf("error getting the aggregate score: %w", err)
			}
			// msg := fmt.Sprintf("Scorecard check results: \n")
			msg := fmt.Sprintf("Aggregate Score: %.1f\n", aggregateScore)
			// for _, c := range scResult.Checks {
			// 	msg += fmt.Sprintf(
			// 		"Check name: %s, score: %.1f, reason: %s\n",
			// 		c.Name, float64(c.Score), c.Reason,
			// 	)
			// }
			a.Message = asPointerStr(msg)
			// a.RawDetails = asPointerStr(fmt.Sprintln(scResult))
		}
		annotations = append(annotations, &a)
		fmt.Println(*a.Message)
	}
	return annotations, nil
}

func asPointerStr(s string) *string {
	return &s
}

func asPointerInt(i int) *int {
	return &i
}
