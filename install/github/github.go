// Copyright 2022 OpenSSF Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

// Package github implements GitHub authentication for the scorecard GitHub
// Action.
package github

import (
	"context"
	"fmt"
	"log"
	"net/http"

	gogh "github.com/google/go-github/v46/github"

	"github.com/ossf/scorecard-action/github"
)

// Client is a wrapper around GitHub-related functionality.
type Client struct {
	*gogh.Client
}

// New returns a new GitHub client.
func New(ctx context.Context) *Client {
	c := github.NewClient(ctx)
	hc := &http.Client{
		Transport: c.Transport(),
	}
	gh := gogh.NewClient(hc)
	client := &Client{gh}

	return client
}

// Modeled after
// https://github.com/kubernetes-sigs/release-sdk/blob/e23d2c82bbb41a007cdf019c30930e8fd2649c01/github/github.go

// GetRepositoriesByOrg // TODO(lint): Needs a comment.
func (c *Client) GetRepositoriesByOrg(
	ctx context.Context,
	owner string,
) ([]*gogh.Repository, *gogh.Response, error) {
	repos, resp, err := c.Repositories.ListByOrg(
		ctx,
		owner,
		// TODO(install): Does this need to parameterized?
		&gogh.RepositoryListByOrgOptions{
			Type: "all",
		},
	)
	if err != nil {
		return repos, resp, fmt.Errorf("getting repositories: %w", err)
	}

	return repos, resp, nil
}

// GetRepository // TODO(lint): Needs a comment.
func (c *Client) GetRepository(
	ctx context.Context,
	owner,
	repo string,
) (*gogh.Repository, *gogh.Response, error) {
	pr, resp, err := c.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return pr, resp, fmt.Errorf("getting repository: %w", err)
	}

	return pr, resp, nil
}

// GetBranch // TODO(lint): Needs a comment.
func (c *Client) GetBranch(
	ctx context.Context,
	owner,
	repo,
	branch string,
	followRedirects bool,
) (*gogh.Branch, *gogh.Response, error) {
	// TODO: Revisit logic and simplify returns, where possible.
	b, resp, err := c.Repositories.GetBranch(
		ctx,
		owner,
		repo,
		branch,
		followRedirects,
	)
	if err != nil {
		return b, resp, fmt.Errorf("getting branch: %w", err)
	}

	return b, resp, nil
}

// GetContents // TODO(lint): Needs a comment.
func (c *Client) GetContents(
	ctx context.Context,
	owner,
	repo,
	path string,
	opts *gogh.RepositoryContentGetOptions,
) (*gogh.RepositoryContent, []*gogh.RepositoryContent, *gogh.Response, error) {
	// TODO: Revisit logic and simplify returns, where possible.
	file, dir, resp, err := c.Repositories.GetContents(
		ctx,
		owner,
		repo,
		path,
		opts,
	)
	if err != nil {
		return file, dir, resp, fmt.Errorf("getting repo content: %w", err)
	}

	return file, dir, resp, nil
}

// CreateGitRef // TODO(lint): Needs a comment.
func (c *Client) CreateGitRef(
	ctx context.Context,
	owner,
	repo string,
	ref *gogh.Reference,
) (*gogh.Reference, *gogh.Response, error) {
	// TODO: Revisit logic and simplify returns, where possible.
	gRef, resp, err := c.Git.CreateRef(
		ctx,
		owner,
		repo,
		ref,
	)
	if err != nil {
		return gRef, resp, fmt.Errorf("creating git reference: %w", err)
	}

	return gRef, resp, nil
}

// CreateFile // TODO(lint): Needs a comment.
func (c *Client) CreateFile(
	ctx context.Context,
	owner,
	repo,
	path string,
	opts *gogh.RepositoryContentFileOptions,
) (*gogh.RepositoryContentResponse, *gogh.Response, error) {
	// TODO: Revisit logic and simplify returns, where possible.
	repoContentResp, resp, err := c.Repositories.CreateFile(
		ctx,
		owner,
		repo,
		path,
		opts,
	)
	if err != nil {
		return repoContentResp, resp, fmt.Errorf("creating file: %w", err)
	}

	return repoContentResp, resp, nil
}

// CreatePullRequest // TODO(lint): Needs a comment.
func (c *Client) CreatePullRequest(
	ctx context.Context,
	owner,
	repo,
	baseBranchName,
	headBranchName,
	title,
	body string,
) (*gogh.PullRequest, error) {
	newPullRequest := &gogh.NewPullRequest{
		Title:               &title,
		Head:                &headBranchName,
		Base:                &baseBranchName,
		Body:                &body,
		MaintainerCanModify: gogh.Bool(true),
	}

	pr, _, err := c.PullRequests.Create(ctx, owner, repo, newPullRequest)
	if err != nil {
		return pr, fmt.Errorf("creating pull request: %w", err)
	}

	log.Printf(
		"successfully created PR #%d for repository %s: %s",
		pr.GetNumber(),
		repo,
		pr.GetHTMLURL(),
	)

	return pr, nil
}

// CreateGitRefOptions // TODO(lint): Needs a comment.
func CreateGitRefOptions(ref string, sha *string) *gogh.Reference {
	return &gogh.Reference{
		Ref:    gogh.String(ref),
		Object: &gogh.GitObject{SHA: sha},
	}
}

// CreateRepositoryContentFileOptions // TODO(lint): Needs a comment.
func CreateRepositoryContentFileOptions(
	content []byte,
	msg, branch string,
) *gogh.RepositoryContentFileOptions {
	return &gogh.RepositoryContentFileOptions{
		Message: gogh.String(msg),
		Content: content,
		Branch:  gogh.String(branch),
	}
}

// CreateRepositoryContentGetOptions // TODO(lint): Needs a comment.
func CreateRepositoryContentGetOptions() *gogh.RepositoryContentGetOptions {
	return &gogh.RepositoryContentGetOptions{}
}
