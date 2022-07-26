package dependencydiff

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/v45/github"
	"github.com/ossf/scorecard-action/options"
	"github.com/ossf/scorecard/v4/clients/githubrepo/roundtripper"
	"github.com/ossf/scorecard/v4/dependencydiff"
	"github.com/ossf/scorecard/v4/log"
	"github.com/ossf/scorecard/v4/pkg"
)

const (
	commentID int64 = 9867
)

func RunDependencyDiff(ctx context.Context) error {
	repoURI := os.Getenv(options.EnvGithubRepository)
	ownerRepo := strings.Split(repoURI, "/")
	if len(ownerRepo) != 2 {
		return fmt.Errorf("%w: repo uri", errInvalid)
	}
	// Since the event listener is set to pull requests to main, this will be the main branch reference.
	base := os.Getenv(options.EnvGithubBaseRef)
	if base == "" {
		return fmt.Errorf("%w: base ref", errEmpty)
	}
	// The head reference of the pull request source branch.
	head := os.Getenv(options.EnvGitHubHeadRef)
	if head == "" {
		return fmt.Errorf("%w: head ref", errEmpty)
	}
	// GetDependencyDiffResults will handle the error checking of checks.
	checks := strings.Split(os.Getenv(options.EnvInputChecks), ",")
	changeTypes := strings.Split(os.Getenv(options.EnvInputChangeTypes), ",")
	changeTypeMap := map[pkg.ChangeType]bool{}
	for _, ct := range changeTypes {
		key := pkg.ChangeType(ct)
		if !key.IsValid() {
			return fmt.Errorf("%w: change type", errInvalid)
		}
		changeTypeMap[key] = true
	}
	deps, err := dependencydiff.GetDependencyDiffResults(
		ctx, repoURI, base, head, checks, changeTypeMap,
	)
	if err != nil {
		return fmt.Errorf("error getting dependency-diff: %w", err)
	}
	fmt.Println(deps)
	report, err := DependencydiffResultsAsMarkdown(deps, base, head)
	if err != nil {
		return fmt.Errorf("error formatting results as markdown: %w", err)
	}
	err = writeToComment(ctx, ownerRepo[0], ownerRepo[1], report)
	if err != nil {
		return fmt.Errorf("error writting the report to comment: %w", err)
	}
	fmt.Println(*report)
	return nil
}

func writeToComment(ctx context.Context, owner, repo string, report *string) error {
	fmt.Printf("env github ref: %s", options.EnvGithubRef)
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
	logger := log.NewLogger(log.DefaultLevel)
	ghrt := roundtripper.NewTransport(ctx, logger) /* This round tripper handles the access token. */
	ghClient := github.NewClient(&http.Client{Transport: ghrt})
	// Get the issue comment in the pull request by ID.
	cmt, _, err := ghClient.Issues.GetComment(ctx, repo, owner, commentID)
	if err != nil {
		// Create a new one if the comment doesn't exist.
		_, _, err := ghClient.Issues.CreateComment(
			ctx, owner, repo, prNumber,
			&github.IssueComment{
				Body: report,
			},
		)
		if err != nil {
			return fmt.Errorf("error creating comment: %w", err)
		}
	} else {
		cmt.Body = report
		// Edit the comment.
		_, _, err = ghClient.Issues.EditComment(ctx, owner, repo, commentID, cmt)
		if err != nil {
			return fmt.Errorf("error editing comment: %w", err)
		}
	}
	return nil
}

var (
	errEmpty   = errors.New("empty")
	errInvalid = errors.New("invalid")
)

func asPointer(i int64) *int64 {
	return &i
}
