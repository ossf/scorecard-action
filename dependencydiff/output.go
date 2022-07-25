package dependencydiff

import (
	"fmt"
	"math"
	"sort"

	docs "github.com/ossf/scorecard/v4/docs/checks"
	"github.com/ossf/scorecard/v4/pkg"
)

const (
	// negInif is "negative infinity" used for dependencydiff results ranking.
	negInf = -math.MaxFloat64
)

type scoreAndDependencyName struct {
	dependencyName string
	aggregateScore float64
}

// DependencydiffResultsAsMarkdown exports the dependencydiff results as markdown.
func DependencydiffResultsAsMarkdown(depdiffResults []pkg.DependencyCheckResult,
	base, head string) (*string, error) {

	added := map[string]pkg.DependencyCheckResult{}
	removed := map[string]pkg.DependencyCheckResult{}
	for _, d := range depdiffResults {
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
		current += scoreTag(key.aggregateScore)
		newResult := added[dName]
		current += fmt.Sprintf(
			"%s @ %s (new) ",
			newResult.Name, *newResult.Version,
		)
		if oldResult, ok := removed[dName]; ok {
			current += fmt.Sprintf(
				"~~%s @ %s (removed)~~ ",
				oldResult.Name, *oldResult.Version,
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
		current += scoreTag(key.aggregateScore)
		oldResult := removed[dName]
		current += fmt.Sprintf(
			"~~%s @ %s~~ ",
			oldResult.Name, *oldResult.Version,
		)
		results += current + "\n\n"
	}
	out := fmt.Sprintf(
		"Dependency-diffs (changes) between the BASE commit `%s` and the HEAD commit `%s`:\n\n",
		base, head,
	)
	if results == "" {
		out += fmt.Sprintln("No dependency changes found.")
	} else {
		out += fmt.Sprintln(results)
	}
	return &out, nil
}

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
				return nil, err
			}
			scoreAndName.aggregateScore = score
		}
		sortKeys = append(sortKeys, scoreAndName)
	}
	return sortKeys, nil
}

func addedTag() string {
	return fmt.Sprintf(":sparkles: **`" + "added" + "`** ")
}

func updatedTag() string {
	return fmt.Sprintf("**`" + "updated" + "`** ")
}

func removedTag() string {
	return fmt.Sprintf("~~**`" + "removed" + "`**~~ ")
}

func scoreTag(score float64) string {
	switch score {
	case negInf:
		return ""
	default:
		return fmt.Sprintf("`Score: %.1f` ", score)
	}
}
