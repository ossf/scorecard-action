package main

import (
	"context"
	"fmt"

	"github.com/google/go-github/v42/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
)

//Currently incomplete

func main_test() {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetUsersByUsername,
			github.User{
				Name: github.String("myuser"),
			},
		),
		mock.WithRequestMatch(
			mock.GetOrgsReposByOrg,
			[]github.Repository{
				{
					Name: github.String("test-repo-1"),
				},
			},
		),
		mock.WithRequestMatch(
			mock.GetReposByOwnerByRepo,
			github.Repository{
				DefaultBranch: github.String("main"),
			},
		),
		mock.WithRequestMatch(
			mock.GetReposBranchesByOwnerByRepo,
			github.Branch{
				Commit: &github.RepositoryCommit{
					SHA: github.String("abcd"),
				},
			},
		),
	)
	c := github.NewClient(mockedHTTPClient)

	ctx := context.Background()

	user, _, _ := c.Users.Get(ctx, "myuser")

	repos, _, _ := c.Repositories.ListByOrg(ctx, *user.Name, nil)

	fmt.Println(repos)
}
