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
	"fmt"
	"net/http"
	"net/url"

	docs "github.com/ossf/scorecard/v4/docs/checks"
	"github.com/ossf/scorecard/v4/pkg"
)

func getDependencySortKeys(dcMap map[string]pkg.DependencyCheckResult,
) ([]scoreAndDependencyName, error) {
	sortKeys := []scoreAndDependencyName{}
	doc, err := docs.Read()
	if err != nil {
		return nil, fmt.Errorf("error reading docs: %w", err)
	}
	for k := range dcMap {
		scoreAndName := scoreAndDependencyName{
			dependencyName: dcMap[k].Name,
			aggregateScore: negInf,
			// Since this struct is for sorting, the dependency having a score of negative infinite
			// will be put to the very last, unless its agregate score is not empty.
		}
		scResults := dcMap[k].ScorecardResultWithError.ScorecardResult
		if scResults != nil {
			score, err := scResults.GetAggregateScore(doc)
			if err != nil {
				return nil, fmt.Errorf("error getting the aggregate score: %w", err)
			}
			scoreAndName.aggregateScore = score
		}
		sortKeys = append(sortKeys, scoreAndName)
	}
	return sortKeys, nil
}

func addedTag() string {
	return fmt.Sprintf(" :sparkles: **`" + "added" + "`** ")
}

func updatedTag() string {
	return fmt.Sprintf(" **`" + "updated" + "`** ")
}

func removedTag() string {
	return fmt.Sprintf(" ~~**`" + "removed" + "`**~~ ")
}

func scoreTag(score float64) string {
	switch score {
	case negInf:
		return ""
	default:
		return fmt.Sprintf("`Score: %.1f` ", score)
	}
}

// Convert the dependency-diff check result slice to two maps: added and removed, for added and removed dependencies respectively.
func dependencySliceToMaps(deps []pkg.DependencyCheckResult) (map[string]pkg.DependencyCheckResult,
	map[string]pkg.DependencyCheckResult) {
	added := map[string]pkg.DependencyCheckResult{}
	removed := map[string]pkg.DependencyCheckResult{}
	for _, d := range deps {
		if d.ChangeType != nil {
			switch *d.ChangeType {
			case pkg.Added:
				added[d.Name] = d
			case pkg.Removed:
				removed[d.Name] = d
			case pkg.Updated:
				// Do nothing, for now.
				// The current data source GitHub Dependency Review won't give the updated dependencies,
				// so we need to find them manually by checking the added/removed maps.
			}
		}
	}
	return added, removed
}

func entryExists(system, name, version string) (bool, error) {
	url := fmt.Sprintf(
		"https://deps.dev/_/s/%s/p/%s/v/%s",
		url.PathEscape(system),
		url.PathEscape(name),
		url.PathEscape(version),
	)
	resp, err := http.Get(url)
	if err != nil {
		return false, fmt.Errorf("error requesting deps.dec/_: %w", err)
	}
	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	default:
		return false, nil
	}
}

func asPointerChangeType(ct pkg.ChangeType) *pkg.ChangeType {
	return &ct
}

func asPointerStr(s string) *string {
	return &s
}
