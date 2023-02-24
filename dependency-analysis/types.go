package main

// ScorecardResult is the result of running scorecard.
type ScorecardResult struct { //nolint:govet
	Date string `json:"date"`
	Repo struct {
		Name   string `json:"name"`
		Commit string `json:"commit"`
	} `json:"repo"`
	Scorecard struct {
		Version string `json:"version"`
		Commit  string `json:"commit"`
	} `json:"scorecard"`
	Score           float64 `json:"score"`
	Checks          []Check `json:"checks"`
	Vulnerabilities []Vulnerability
}

// Check is a single check result.
type Check struct { //nolint:govet
	Name          string   `json:"name"`
	Score         int      `json:"score,omitempty"`
	Reason        string   `json:"reason"`
	Details       []string `json:"details"`
	Documentation struct {
		Short string `json:"short"`
	} `json:"documentation"`
}

// DependencyDiff is the result of running dependency-analysis.
type DependencyDiff struct {
	ChangeType          string          `json:"change_type"`
	Manifest            string          `json:"manifest"`
	Ecosystem           string          `json:"ecosystem"`
	Name                string          `json:"name"`
	Version             string          `json:"version"`
	PackageURL          string          `json:"package_url"`
	License             string          `json:"license"`
	SourceRepositoryURL string          `json:"source_repository_url"`
	Vulnerabilities     []Vulnerability `json:"vulnerabilities"`
}

// Vulnerability is a single vulnerability.
type Vulnerability struct {
	Severity        string `json:"severity"`
	AdvisoryGHSAId  string `json:"advisory_ghsa_id"`
	AdvisorySummary string `json:"advisory_summary"`
	AdvisoryURL     string `json:"advisory_url"`
}
