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
	"sort"

	"github.com/google/go-github/v45/github"
	"github.com/ossf/scorecard-action/options"
	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/pkg"
)

const (
	fakeStartLine = 1
	fakeEndLine   = 2
	msgNoResults  = "No Scorecard check results are available for this dependency, or this is a removed one."
)

func visualizeToCheckRun(ctx context.Context, ghClient *github.Client,
	owner, repo string,
	deps []pkg.DependencyCheckResult,
) error {
	headSHA := os.Getenv(options.EnvInputPullRequestHeadSHA)
	if headSHA == "" {
		return fmt.Errorf("%w: head ref", errEmpty)
	}
	annotations, err := createAnnotations(deps)
	if err != nil {
		return fmt.Errorf("error creating annotations: %w", err)
	}
	output := github.CheckRunOutput{
		Title: asPointerStr("Scorecard Action Dependency-diff check results"),
		Summary: asPointerStr(
			fmt.Sprintf(
				":sparkles: **%d** dependency-diffs (changes) found, **%d** annotations created.",
				len(deps), len(annotations),
			),
		),
		Annotations: annotations,
	}
	opts := github.CreateCheckRunOptions{
		Name:    "Scorecard Action Dependency-diff",
		HeadSHA: headSHA,
		// DetailsURL should be the integrator's site that has the full details of the check.
		// TODO (#issue number): Leave this as nil for now to make it explicit. This might be a
		// corresponding scorecard check page for a specific package once we have the security-scorecard.dev website.
		// https://github.com/google/go-github/blob/master/github/checks.go#L142
		DetailsURL: asPointerStr("https://deps.dev/"),
		Status:     asPointerStr("completed"),
		Conclusion: asPointerStr("neutral"),
		Output:     &output,
	}
	_, _, err = ghClient.Checks.CreateCheckRun(
		ctx, owner, repo, opts,
	)
	if err != nil {
		return fmt.Errorf("error creating the check run: %w", err)
	}
	return nil
}

func createAnnotations(deps []pkg.DependencyCheckResult) ([]*github.CheckRunAnnotation, error) {
	annotations := []*github.CheckRunAnnotation{}
	// Do sorting for dependencies by their aggregate scores in descending order.
	added, removed := dependencySliceToMaps(deps)
	addedSortKeys, err := getDependencySortKeys(added)
	if err != nil {
		return nil, err
	}
	removedSortKeys, err := getDependencySortKeys(removed)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(
		addedSortKeys,
		func(i, j int) bool { return addedSortKeys[i].aggregateScore > addedSortKeys[j].aggregateScore },
	)
	sort.SliceStable(
		removedSortKeys,
		func(i, j int) bool { return removedSortKeys[i].aggregateScore > removedSortKeys[j].aggregateScore },
	)
	for _, key := range addedSortKeys {
		if _, ok := added[key.dependencyName]; !ok {
			return nil, fmt.Errorf("%w: map entry", errInvalid)
		}
		dName := key.dependencyName
		results, err := annotationHelper(
			dName, added[dName].ManifestPath, added[dName].Version,
			key.aggregateScore, added[dName].ChangeType,
			added[dName].ScorecardResultWithError.ScorecardResult,
		)
		if err != nil {
			return nil, err
		}
		annotations = append(
			annotations, results...,
		)
	}
	for _, key := range removedSortKeys {
		if _, ok := removed[key.dependencyName]; !ok {
			return nil, fmt.Errorf("%w: map entry", errInvalid)
		}
		dName := key.dependencyName
		results, err := annotationHelper(
			dName, removed[dName].ManifestPath, removed[dName].Version,
			key.aggregateScore, removed[dName].ChangeType,
			removed[dName].ScorecardResultWithError.ScorecardResult,
		)
		if err != nil {
			return nil, err
		}
		annotations = append(
			annotations, results...,
		)
	}
	return annotations, nil
}

func annotationHelper(name string, manifest, version *string, aggregate float64,
	changeType *pkg.ChangeType, scorecardResult *pkg.ScorecardResult,
) ([]*github.CheckRunAnnotation, error) {
	annotations := []*github.CheckRunAnnotation{}
	if changeType == nil {
		return nil, fmt.Errorf("%w: dependency change type", errInvalid)
	}
	if *changeType != pkg.Removed && scorecardResult != nil {
		// Create annotations only for added dependencies and on a per-check basis.
		for _, c := range scorecardResult.Checks {
			a := github.CheckRunAnnotation{
				// No need for nil pointer checking since a.Path is also a pointer type.
				Path: manifest,
				// TODO (#issue number): use the actual start lines and end lines in manifest/lock file if a future
				// data source has such fields.
				StartLine:       asPointerInt(fakeStartLine), /* Fake the start line since we don't have the data for now. */
				EndLine:         asPointerInt(fakeEndLine),   /* Fake the end line since we don't have the data for now. */
				AnnotationLevel: asPointerStr("notice"),
				Message: asPointerStr(
					fmt.Sprintf(
						"Check: %s\nScore: %.1f\nReason: %s",
						c.Name, float64(c.Score), c.Reason,
					),
				),
				RawDetails: asPointerStr(fmt.Sprint(*scorecardResult)),
			}
			if changeType != nil && version != nil {
				a.Title = asPointerStr(fmt.Sprintf(
					"%s dependency: %s @ %s",
					*changeType, name, *version,
				))
			} else {
				a.Title = &name
			}
			if aggregate != checker.InconclusiveResultScore {
				a.Title = asPointerStr(
					fmt.Sprintf(
						"%s `Overall aggregate score: %.1f`", *a.Title, aggregate,
					),
				)
			}
			// Should we do this?
			// My thought is to make this a warning if the score of a check is lower than a certain value.
			if c.Score < 6.0 {
				a.AnnotationLevel = asPointerStr("warning")
			} else {
				a.AnnotationLevel = asPointerStr("notice")
			}
			annotations = append(annotations, &a)
		}
	} else {
		// Create exactly one annotation for those having a null scorecard check field.
		a := github.CheckRunAnnotation{
			Path:      manifest,
			StartLine: asPointerInt(fakeStartLine),
			EndLine:   asPointerInt(fakeEndLine),
			Message:   asPointerStr(msgNoResults),
		}
		if *changeType == pkg.Removed {
			a.AnnotationLevel = asPointerStr("notice")
		} else {
			a.AnnotationLevel = asPointerStr("warning")
		}
		if changeType != nil && version != nil {
			a.Title = asPointerStr(fmt.Sprintf(
				"%s dependency: %s @ %s",
				*changeType, name, *version,
			))
		} else {
			a.Title = &name
		}
		annotations = append(annotations, &a)
	}
	return annotations, nil
}

func asPointerStr(s string) *string {
	return &s
}

func asPointerInt(i int) *int {
	return &i
}
