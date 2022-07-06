package depdiff

// ChangeType is the change type (added, updated, removed) of a dependency.
type ChangeType string

const (
	Added   ChangeType = "added"
	Updated ChangeType = "updated"
	Removed ChangeType = "removed"
)

// IsValid determines if a ChangeType is valid.
func (ct *ChangeType) IsValid() bool {
	switch *ct {
	case Added, Updated, Removed:
		return true
	default:
		return false
	}
}

// Dependency is a dependency diff in a code commit.
type Dependency struct {
	// IsDirect suggests if the dependency is a direct dependency of a code commit.
	IsDirect bool

	// ChangeType indicates whether the dependency is added or removed.
	ChangeType ChangeType `json:"change_type"`

	// ManifestFileName is the name of the manifest file of the dependency, such as go.mod for Go.
	ManifestFileName string `json:"manifest"`

	// Ecosystem is the name of the package management system, such as NPM, GO, PYPI.
	Ecosystem string `json:"ecosystem" bigquery:"System"`

	// Name is the name of the dependency.
	Name string `json:"name" bigquery:"Name"`

	// Version is the package version of the dependency.
	Version string `json:"version" bigquery:"Version"`

	// AggregateScore is the Scorecard aggregate score (0-10) of the dependency.
	AggregateScore float32

	// Package URL is a short link for a package.
	PackageURL string `json:"package_url"`

	// License is the license of the dependency.
	License string `json:"license"`

	// SrcRepoURL is the source repository URL of the dependency.
	SrcRepoURL string `json:"source_repository_url"`

	// Vulnerabilities is a list of Vulnerability.
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`

	// Dependencies is the list of dependencies of the current direct dependency,
	// i.e. indirect (transitive) dependencies.
	// TODO: this is not a version-zero property, and will be used to analyze transitive
	// dependencies in future versions.
	Dependencies []Dependency `json:"dependencies"`
}
