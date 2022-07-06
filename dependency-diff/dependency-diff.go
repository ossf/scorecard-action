package depdiff

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"time"

	gogh "github.com/google/go-github/v38/github"
)

// Get the depednency-diff using the GitHub Dependency Review
// (https://docs.github.com/en/rest/dependency-graph/dependency-review) API
func GetDepDiffByCommitsSHA(authToken, repoOwner string, repoName string,
	base string, head string) ([]Dependency, error) {
	client := gogh.NewClient(http.DefaultClient)
	reqURL := path.Join(
		"repos", repoOwner, repoName, "dependency-graph", "compare", base+"..."+head,
	)
	req, err := client.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("request for dependency-diff failed with %w", err)

	}
	// To specify the return type to be JSON.
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	// An access token is required in the request header to be able to use this API.
	req.Header.Set("Authorization", "token "+authToken)

	// Set a ten-seconds timeout to make sure the client can be created correctly.
	myClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := myClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get response error: %w", err)
	}
	defer resp.Body.Close()

	depDiff := []Dependency{}
	err = json.NewDecoder(resp.Body).Decode(&depDiff)
	if err != nil {
		return nil, fmt.Errorf("parse response error: %w", err)
	}
	for i := range depDiff {
		depDiff[i].IsDirect = true
	}
	return depDiff, nil
}
