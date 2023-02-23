package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/v46/github"
	"golang.org/x/oauth2"
)

func main() {
	vulnerabilities := ""
	result := ""

	owner := os.Getenv("GITHUB_REPOSITORY_OWNER")
	repo := os.Getenv("GITHUB_REPOSITORY")
	commitSHA := os.Getenv("GITHUB_SHA")
	token := os.Getenv("GITHUB_TOKEN")
	pr := os.Getenv("GITHUB_PR_NUMBER")
	ghUser := os.Getenv("GITHUB_ACTOR")
	if err := Validate(token, owner, repo, commitSHA, pr); err != nil {
		log.Fatal(err)
	}

	ownerRepo := strings.Split(repo, "/")
	owner = ownerRepo[0]
	repo = ownerRepo[1]
	checks, err := GetScorecardChecks()
	if err != nil {
		log.Fatal(err)
	}

	defaultBranch, err := getDefaultBranch(owner, repo, token)
	if err != nil {
		log.Fatal(err)
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)
	data, err := GetDependencyDiff(owner, repo, token, defaultBranch, commitSHA)
	if err != nil {
		log.Fatal(err)
	}

	m := make(map[string]DependencyDiff)
	for _, dep := range data {
		m[dep.SourceRepositoryURL] = dep
	}

	for k, i := range m {
		url := strings.TrimPrefix(k, "https://")
		scorecard, error := GetScore(url)
		if error != nil && len(i.Vulnerabilities) > 0 {
			sb := strings.Builder{}
			sb.WriteString(fmt.Sprintf("<details><summary>Vulnerabilties %s</summary>\n </br>", i.SourceRepositoryURL))
			sb.WriteString("<table>\n")
			sb.WriteString("<tr>\n")
			sb.WriteString("<th>Severity</th>\n")
			sb.WriteString("<th>AdvisoryGhsaId</th>\n")
			sb.WriteString("<th>AdvisorySummary</th>\n")
			sb.WriteString("<th>AdvisoryUrl</th>\n")
			sb.WriteString("</tr>\n")
			for _, v := range i.Vulnerabilities {
				sb.WriteString("<tr>\n")
				sb.WriteString(fmt.Sprintf("<td>%s</td>\n", v.Severity))
				sb.WriteString(fmt.Sprintf("<td>%s</td>\n", v.AdvisoryGhsaId))
				sb.WriteString(fmt.Sprintf("<td>%s</td>\n", v.AdvisorySummary))
				sb.WriteString(fmt.Sprintf("<td>%s</td>\n", v.AdvisoryUrl))
			}
			sb.WriteString("</table>\n")
			sb.WriteString("</details>\n")
			vulnerabilities += sb.String()
			continue
		}
		scorecard.Checks = filter(scorecard.Checks, func(check Check) bool {
			for _, c := range checks {
				if check.Name == c {
					return true
				}
			}
			return false
		})
		scorecard.Vulnerabilities = i.Vulnerabilities
		result += GitHubIssueComment(scorecard)
	}
	// convert pr to int
	prInt, err := strconv.Atoi(pr)
	if err != nil {
		log.Fatal(err)
	}
	// create or update comment
	if vulnerabilities == "" && result == "" {
		return
	}
	if err := createOrUpdateComment(client, owner, ghUser, repo, prInt, "## Scorecard Results</br>\n"+vulnerabilities+"</br>"+result); err != nil {
		log.Fatal(err)
	}
	if vulnerabilities != "" {
		// this will fail the workflow if there are any vulnerabilities
		os.Exit(1)
	}
}

// getDefaultBranch gets the default branch of the repository.
func getDefaultBranch(owner, repo, token string) (string, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	repository, _, err := client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return "", fmt.Errorf("failed to get repository: %v", err)
	}

	return repository.GetDefaultBranch(), nil
}

// Validate validates the input parameters.
func Validate(token string, owner string, repo string, commitSHA string, pr string) error {
	if token == "" {
		return fmt.Errorf("token is empty")
	}
	if owner == "" {
		return fmt.Errorf("owner is empty")
	}
	if repo == "" {
		return fmt.Errorf("repo is empty")
	}
	if commitSHA == "" {
		return fmt.Errorf("commitSHA is empty")
	}
	if pr == "" {
		return fmt.Errorf("pr is empty")
	}
	return nil
}

// createOrUpdateComment creates a new comment on the pull request or updates an existing one.
func createOrUpdateComment(client *github.Client, owner, githubUser, repo string, prNum int, commentBody string) error {
	comments, _, err := client.Issues.ListComments(context.Background(), owner, repo, prNum, &github.IssueListCommentsOptions{})
	if err != nil {
		return fmt.Errorf("failed to get comments: %v", err)
	}
	// Check if the user has already left a comment on the pull request.
	var existingComment *github.IssueComment
	for _, comment := range comments {
		if comment.GetUser().GetLogin() == githubUser {
			existingComment = comment
			break
		}
	}

	// If the user has already left a comment, update it.
	if existingComment != nil {
		existingComment.Body = &commentBody
		_, _, err = client.Issues.EditComment(context.Background(), owner, repo, *existingComment.ID, existingComment)
		if err != nil {
			return fmt.Errorf("failed to update comment: %v", err)
		}
		log.Println("Comment updated successfully!")
	} else {
		// Otherwise, create a new comment.
		newComment := &github.IssueComment{
			Body: &commentBody,
		}
		_, _, err = client.Issues.CreateComment(context.Background(), owner, repo, prNum, newComment)
		if err != nil {
			return fmt.Errorf("failed to create comment: %v", err)
		}
		log.Println("Comment created successfully!")
	}
	return nil
}

// GitHubIssueComment returns a markdown string for a GitHub issue comment.
func GitHubIssueComment(checks ScorecardResult) string {
	if checks.Repo.Name == "" {
		return ""
	}
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("<details><summary>%s - %s</summary>\n </br>", checks.Repo.Name, checks.Date))
	sb.WriteString(fmt.Sprintf(
		"<a href=https://api.securityscorecards.dev/projects/%s>https://api.securityscorecards.dev/projects/%s</a><br></br>",
		checks.Repo.Name,
		checks.Repo.Name))
	sb.WriteString("<table>\n")
	sb.WriteString("<tr><th>Check</th><th>Score</th></tr>\n")
	for _, check := range checks.Checks {
		sb.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%d</td></tr>\n", check.Name, check.Score))
	}
	sb.WriteString("</table>\n")
	if len(checks.Vulnerabilities) > 0 {
		sb.WriteString("<table>\n")
		sb.WriteString("<tr><th>Vulnerability</th><th>Severity</th><th>Summary</th></tr>\n")
		for _, vulns := range checks.Vulnerabilities {
			sb.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%s</td></tr>\n", vulns.AdvisoryUrl, vulns.Severity, vulns.AdvisorySummary))
		}
		sb.WriteString("</table>\n")
	}

	sb.WriteString("</details>")
	return sb.String()
}

// GetDependencyDiff returns the dependency diff between two commits. It returns an error if the dependency graph is not enabled.
func GetDependencyDiff(owner, repo, token, base, head string) ([]DependencyDiff, error) {
	if owner == "" {
		return nil, fmt.Errorf("owner is required")
	}
	if repo == "" {
		return nil, fmt.Errorf("repo is required")
	}
	if token == "" {
		return nil, fmt.Errorf("token is required")
	}
	resp, err := GetGitHubDependencyDiff(owner, repo, token, base, head)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// if the dependency graph is not enabled, we can't get the dependency diff
		return nil, fmt.Errorf("failed to get dependency diff, please enable dependency graph https://docs.github.com/en/code-security/supply-chain-security/understanding-your-software-supply-chain/configuring-the-dependency-graph : %v", resp.Status)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get dependency diff: %w", err)
	}

	var data []DependencyDiff
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}
	// filter out the dependencies that are not added
	var filteredData []DependencyDiff
	for _, dep := range data {
		// also if the source repo doesn't start with GitHub.com, we can ignore it
		if dep.ChangeType == "added" && dep.SourceRepositoryURL != "" && strings.HasPrefix(dep.SourceRepositoryURL, "https://github.com") {
			filteredData = append(filteredData, dep)
		}
	}
	return filteredData, nil
}

// GetGitHubDependencyDiff returns the dependency diff between two commits. It returns an error if the dependency graph is not enabled.
func GetGitHubDependencyDiff(owner string, repo string, token string, base string, head string) (*http.Response, error) {
	req, err := http.NewRequest("GET",
		fmt.Sprintf("https://api.github.com/repos/%s/%s/dependency-graph/compare/%s...%s", owner, repo, base, head), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
		// handle err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get response: %w", err)
	}
	return resp, nil
}

// filter returns a new slice containing all elements of slice that satisfy the predicate f.
func filter[T any](slice []T, f func(T) bool) []T {
	var n []T
	for _, e := range slice {
		if f(e) {
			n = append(n, e)
		}
	}
	return n
}

// GetScorecardChecks returns the list of checks to run.
// This uses the SCORECARD_CHECKS environment variable to get the path to the checks list.
func GetScorecardChecks() ([]string, error) {
	fileName := os.Getenv("SCORECARD_CHECKS")
	if fileName == "" {
		// default to critical and high severity checks
		return []string{"Dangerous-Workflow", "Binary-Artifacts", "Branch-Protection", "Code-Review", "Dependency-Update-Tool"}, nil
	}
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	decoder := json.NewDecoder(f)
	var checksFromFile []string
	err = decoder.Decode(&checksFromFile)
	if err != nil {
		return nil, err
	}
	return checksFromFile, nil
}

// GetScore returns the scorecard result for a given repository.
func GetScore(repo string) (ScorecardResult, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.securityscorecards.dev/projects/%s", repo), nil)
	if err != nil {
		return ScorecardResult{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ScorecardResult{}, fmt.Errorf("failed to get response: %w", err)
	}
	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ScorecardResult{}, fmt.Errorf("failed to read response: %w", err)
	}
	var scorecard ScorecardResult
	err = json.Unmarshal(result, &scorecard)
	if err != nil {
		return ScorecardResult{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return scorecard, nil
}
