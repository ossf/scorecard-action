package depdiff

import (
	"fmt"
)

// SprintDependencyDiffToMarkDown analyzes the dependency-diff fetched from the GitHub Dependency
// Review API, then parse them and return as a markdown string.
func SprintDependencyDiffToMarkDown(deps []Dependency) string {
	// Divide fetched depdendencies into added, updated, and removed ones.
	added, updated, removed :=
		map[string]Dependency{}, map[string]Dependency{}, map[string]Dependency{}
	for _, d := range deps {
		switch d.ChangeType {
		case Added:
			added[d.Name] = d
		case Removed:
			removed[d.Name] = d
		}
	}
	results := ""
	for dName, d := range added {
		// If a dependency name in the added map is also in the removed map,
		// then it's an updated dependency.
		if _, ok := removed[dName]; ok {
			updated[dName] = d
		} else {
			// Otherwise, it's an added dependency.
			current := changeTypeTag(Added)
			// Add the vulnerble tag for added dependencies if vuln found.
			if len(d.Vulnerabilities) != 0 {
				current += fmt.Sprintf(vulnTag(d))
			}
			current += fmt.Sprintf(
				"%s: %s @ %s\n\n",
				d.Ecosystem, d.Name, d.Version,
			)
			results += current
		}
	}
	for dName, d := range updated {
		current := changeTypeTag(Updated)
		// Add the vulnerble tag for updated dependencies if vuln found.
		if len(d.Vulnerabilities) != 0 {
			current += fmt.Sprintf(vulnTag(d))
		}
		current += fmt.Sprintf(
			" %s: %s @ %s (**new**) (bumped from %s: %s @ %s)\n\n",
			d.Ecosystem, d.Name, d.Version,
			added[dName].Ecosystem, added[dName].Name, added[dName].Version,
		)
		results += current
	}
	for dName, d := range removed {
		// If a dependency name in the removed map is not in the added map,
		// then it's a removed dependency.
		if _, ok := added[dName]; !ok {
			// We don't care vulnerbailities found in the removed dependencies.
			current := fmt.Sprintf(
				changeTypeTag(Removed)+" %s: %s @ %s\n\n",
				d.Ecosystem, d.Name, d.Version,
			)
			results += current + "\n\n"
		}
	}
	if results == "" {
		return fmt.Sprintln("No dependencies changed")
	} else {
		return results
	}
}

// changeTypeTag generates the change type markdown label for the dependency change type.
func changeTypeTag(ct ChangeType) string {
	switch ct {
	case Added, Updated:
		return fmt.Sprintf("**`"+"%s"+"`**", ct)
	case Removed:
		return fmt.Sprintf("~~**`"+"%s"+"`**~~", ct)
	default:
		return ""
	}
}

// vulnTag generates the vulnerable markdown label with a vulnerability reference URL.
func vulnTag(d Dependency) string {
	// TODO: which URL we should give as the vulnerability reference?
	result := fmt.Sprintf("[**`"+"vulnerable"+"`**](%s) ", d.SrcRepoURL)
	return result
}
