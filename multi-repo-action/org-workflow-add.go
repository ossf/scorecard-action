package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/google/go-github/v42/github"
	"golang.org/x/oauth2"
)

// *****SET THESE PARAMETERS*****
const ORG_NAME string = "organization name"
const PAT string = "access token"

// Adds the OpenSSF Scorecard workflow to all repositores under the given organization.
func main() {
	// Get github user client.
	context := context.Background()
	tokenService := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: PAT},
	)
	tokenClient := oauth2.NewClient(context, tokenService)
	client := github.NewClient(tokenClient)

	// Get repositories under organization.
	lops := &github.RepositoryListByOrgOptions{Type: "all"}
	repos, _, err := client.Repositories.ListByOrg(context, ORG_NAME, lops)
	err_check(err, "List Org Repos Error")

	// Convert to list of repository names.
	repoNames := []string{}
	for _, repo := range repos {
		repoNames = append(repoNames, *repo.Name)
	}

	// Get yml file into byte array.
	fileContent, err := ioutil.ReadFile("scorecards.yml")
	err_check(err, "Read File Error")

	// Process each repository.
	for _, repoName := range repoNames {

		// Get head commit SHA of default branch.
		repo, _, err := client.Repositories.Get(context, ORG_NAME, repoName)
		err_check(err, "Get Repository Error")
		defaultBranch, _, err := client.Repositories.GetBranch(context, ORG_NAME, repoName, *repo.DefaultBranch, true)
		err_check(err, "Get Branch Error")
		defaultBranchSHA := defaultBranch.Commit.SHA

		// Check if scorecard file already exists in workflows folder.
		scoreFileContent, _, _, err := client.Repositories.GetContents(context, ORG_NAME, repoName, ".github/workflows/scorecards-analysis.yml", &github.RepositoryContentGetOptions{})
		if scoreFileContent != nil || err == nil {
			fmt.Println("Could not process repo", repoName, "since scorecard workflow already exists.")
			os.Exit(1)
		}

		// Check if branch scorecard already exists.
		scorecardBranch, _, err := client.Repositories.GetBranch(context, ORG_NAME, repoName, "scorecard", true)
		if scorecardBranch != nil || err == nil {
			fmt.Println("Could not process repo", repoName, "since branch scorecard already exists.")
			os.Exit(1)
		}

		// Create new branch using a reference that stores the new commit hash.
		ref := &github.Reference{
			Ref:    github.String("refs/heads/scorecard"),
			Object: &github.GitObject{SHA: defaultBranchSHA},
		}
		_, _, err = client.Git.CreateRef(context, ORG_NAME, repoName, ref)
		err_check(err, "Create Ref Error")

		// Create file in repository.
		opts := &github.RepositoryContentFileOptions{
			Message: github.String("Adding scorecard workflow"),
			Content: fileContent,
			Branch:  github.String("scorecard"),
		}
		_, _, err = client.Repositories.CreateFile(context, ORG_NAME, repoName, ".github/workflows/scorecards-analysis.yml", opts)
		err_check(err, "CreateFile Error")

		// Create Pull request.
		pr := &github.NewPullRequest{
			Title: github.String("Added Scorecard Workflow"),
			Head:  github.String("scorecard"),
			Base:  github.String(*defaultBranch.Name),
			Body:  github.String("Added the workflow for OpenSSF's Security Scorecard"),
			Draft: github.Bool(false),
		}

		_, _, err = client.PullRequests.Create(context, ORG_NAME, repoName, pr)
		err_check(err, "Pull Request Error")

		//Logging
		fmt.Println("Added scorecard workflow PR from scorecard to", *defaultBranch.Name, "branch of repo", repoName)
	}
}

func err_check(err error, msg string) {
	if err != nil {
		fmt.Println(msg, err)
	}
}
