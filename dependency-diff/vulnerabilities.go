package depdiff

import (
	"time"
)

// SeverityLevel is the level of severity of a vulnerability.
type SeverityLevel string

const (
	Critical SeverityLevel = "CRITICAL"
	High     SeverityLevel = "HIGH"
	Medium   SeverityLevel = "MEDIUM"
	Moderate SeverityLevel = "MODERATE"
	Low      SeverityLevel = "LOW"
	None     SeverityLevel = "NONE"
	Unknown  SeverityLevel = "UNKNOWN"
)

// IsValid determines if a SeverityLevel is valid.
func (sl *SeverityLevel) IsValid() bool {
	switch *sl {
	case Critical, High, Medium, Moderate, Low, None, Unknown:
		return true
	default:
		return false
	}
}

// Source is an authoritative source of a vulnerability.
type Source string

const (
	GHSA Source = "GHSA"
	NSWG Source = "NSWG"
	OSV  Source = "OSV"
)

// IsValid determines if a Source is valid.
func (src *Source) IsValid() bool {
	switch *src {
	case GHSA, NSWG, OSV:
		return true
	default:
		return false
	}
}

// Vulnerability is a vulnerability of a dependency.
type Vulnerability struct {
	// Source is the source of the vulnerability.
	Source string `bigquery:"Source"`

	// ID is the identifier of the vulnerability.
	ID string `json:"advisory_ghsa_id" bigquery:"SourceID"`

	// SourceURL is the source URL of the vulnerability.
	SourceURL string `json:"advisory_url" bigquery:"SourceURL"`

	// Title is the text summary of the vulnerability.
	Title string `json:"advisory_summary" bigquery:"Title"`

	// Description is a long text paragraph of the vulnerability.
	Description string `json:"description" bigquery:"Description"`

	// Score is the score of a vulnerability given by an authoritative src.
	// TODO: this is not a version-zero property and will be included in future versions.
	// Score bigquery.NullFloat64 `bigquery:"Score"`

	// Severity is a enum type of the severity level of the vulnerability.
	Severity SeverityLevel `json:"bq_severity" bigquery:"Severity"`

	// GitHubSeverity is the severity level determined by GitHub.
	GitHubSeverity SeverityLevel `json:"github_severity" bigquery:"GitHubSeverity"`

	// ReferenceURLs include all URLs that are related to the vulnerability.
	ReferenceURLs []string `json:"reference_urls" bigquery:"ReferenceURLs"`

	// DisclosedTime is the time when the vulenrability is publicly disclosed.
	DisclosedTime time.Time `json:"disclosed_time" bigquery:"Disclosed"`
}
