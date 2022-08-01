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
	"math"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/google/go-github/v45/github"
	"github.com/ossf/scorecard-action/options"

	"github.com/ossf/scorecard/v4/pkg"
)

const (
	// negInif is "negative infinity" used for dependencydiff results ranking.
	negInf float64 = -math.MaxFloat64

	// commentMarkdownID uses a markdown comment syntax to uniquely identify our generated comment.
	commentMarkdownID string = "<!-- scorecard action dependency-diff-as-a-comment -->"
)

type scoreAndDependencyName struct {
	dependencyName string
	aggregateScore float64
}

func writeToComment(ctx context.Context, ghClient *github.Client, owner, repo string, report *string) error {
	if report == nil {
		// The markdown comment report should not be nil if there's no error.
		return fmt.Errorf("report %w", errShouldNotBeNil)
	}
	ref := os.Getenv(options.EnvGithubRef)
	splitted := strings.Split(ref, "/")
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#pull_request
	// For a pull request-triggred workflow, the env GITHUB_REF has the following format:
	// refs/pull/:prNumber/merge.
	if len(splitted) != 4 {
		return fmt.Errorf("%w: github ref", errEmpty)
	}
	prNumber, err := strconv.Atoi(splitted[2])
	if err != nil {
		return fmt.Errorf("error converting str pr number to int: %w", err)
	}

	// Generate our report comment.
	reportComment := github.IssueComment{
		Body: asPointerStr(
			commentBodyWithMarkdownID(commentMarkdownID, *report),
		),
	}
	// Get the current comments in the pull request.
	comments, _, err := ghClient.Issues.ListComments(
		ctx, owner, repo, prNumber, nil,
	)
	if len(comments) != 0 {
		// Iterate the list of comments and find ours.
		for _, comment := range comments {
			if comment.Body == nil || comment.ID == nil {
				continue
			}
			if strings.Contains(*comment.Body, commentMarkdownID) {
				// Found our previous left comment.
				_, _, err := ghClient.Issues.EditComment(
					ctx, owner, repo, *comment.ID, &reportComment,
				)
				if err != nil {
					return fmt.Errorf("error updating comment: %w", err)
				}
				// Directly return nil if we have successfully updated the comment.
				return nil
			}
		}
	}
	// If it still hasn't returned until here, meaning either (1) we don't find our comments
	// in the list of comments, or (2) the list of comments is an empty one.
	// We create and leave a new comment there.
	_, _, err = ghClient.Issues.CreateComment(
		ctx, owner, repo, prNumber,
		&reportComment,
	)
	if err != nil {
		return fmt.Errorf("error creating comment: %w", err)
	}
	return nil
}

// dependencydiffResultsAsMarkdown exports the dependencydiff results as markdown.
func dependencydiffResultsAsMarkdown(depdiffResults []pkg.DependencyCheckResult,
	base, head string) (*string, error) {

	added, removed := dependencySliceToMaps(depdiffResults)
	// Sort dependencies by their aggregate scores in descending orders.
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
	results := ""
	for _, key := range addedSortKeys {
		dName := key.dependencyName
		if _, ok := added[dName]; !ok {
			continue
		}
		current := addedTag()
		if _, ok := removed[dName]; ok {
			// Dependency in the added map also found in the removed map, indicating an updated one.
			current += updatedTag()
		}
		newResult := added[dName]
		if newResult.Ecosystem != nil && newResult.Version != nil {
			found, err := entryExists(*newResult.Ecosystem, newResult.Name, *newResult.Version)
			if err != nil {
				return nil, err
			}
			if found {
				current += depsDevTag(*newResult.Ecosystem, newResult.Name)
			}
		}
		current += scoreTag(key.aggregateScore)
		current += packageAsMarkdown(
			newResult.Name, newResult.Version, newResult.SourceRepository, newResult.ChangeType,
		)
		if oldResult, ok := removed[dName]; ok {
			current += packageAsMarkdown(
				oldResult.Name, oldResult.Version, oldResult.SourceRepository, oldResult.ChangeType,
			)
		}
		results += current + "\n\n"
	}
	for _, key := range removedSortKeys {
		dName := key.dependencyName
		if _, ok := added[dName]; ok {
			// Skip updated ones.
			continue
		}
		if _, ok := removed[dName]; !ok {
			continue
		}
		current := removedTag()
		oldResult := removed[dName]
		if oldResult.Ecosystem != nil && oldResult.Version != nil {
			found, err := entryExists(*oldResult.Ecosystem, oldResult.Name, *oldResult.Version)
			if err != nil {
				return nil, err
			}
			if found {
				current += depsDevTag(*oldResult.Ecosystem, oldResult.Name)
			}
		}
		current += scoreTag(key.aggregateScore)
		current += packageAsMarkdown(
			oldResult.Name, oldResult.Version, oldResult.SourceRepository, oldResult.ChangeType,
		)
		results += current + "\n\n"
	}
	// TODO (#772):
	out := "# [Scorecard Action](https://github.com/ossf/scorecard-action) Dependency-diff Report\n\n"
	out += fmt.Sprintf(
		"Dependency-diffs (changes) between the **BASE** `%s` and the **HEAD** `%s`:\n\n",
		base, head,
	)
	if results == "" {
		out += fmt.Sprintln("No dependency changes found.")
	} else {
		out += fmt.Sprintln(results)
	}
	out += experimentalFeature()
	return &out, nil
}

func packageAsMarkdown(name string, version, srcRepo *string, changeType *pkg.ChangeType,
) string {
	result := ""
	result += fmt.Sprintf(" %s", name)
	if srcRepo != nil {
		result = "[" + result + "]" + "(" + *srcRepo + ")"
	}
	if version != nil {
		result += fmt.Sprintf(" @ %s", *version)
	}
	if *changeType == pkg.Removed {
		result = " ~~" + strings.Trim(result, " ") + "~~ "
	}
	return result
}

func experimentalFeature() string {
	result := "> This is an experimental feature of the [Scorecard Action](https://github.com/ossf/scorecard-action). " +
		"The [scores](https://github.com/ossf/scorecard#scoring) are aggregate scores calculated by the checks specified in the workflow file. " +
		"Please refer to [Scorecard Checks](https://github.com/ossf/scorecard#scorecard-checks) for more details. " +
		"Please also see the corresponding [deps.dev](https://deps.dev/) tag for a more comprehensive view of your dependencies."
	return result
}

func depsDevTag(system, name string) string {
	url := fmt.Sprintf(
		"https://deps.dev/%s/%s",
		url.PathEscape(strings.ToLower(system)),
		url.PathEscape(strings.ToLower(name)),
	)
	return fmt.Sprintf(" **[deps.dev](%s)** ", url)
}

func commentBodyWithMarkdownID(id, report string) string {
	return fmt.Sprintf("%s\n%s", id, report)
}
