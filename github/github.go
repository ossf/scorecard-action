// Copyright OpenSSF Authors
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

package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/ossf/scorecard/v4/clients/githubrepo/roundtripper"
	sclog "github.com/ossf/scorecard/v4/log"
)

// RepoInfo is a struct for repository information.
type RepoInfo struct {
	Repo      repo `json:"repository"`
	respBytes []byte
}

type repo struct {
	/*
		https://docs.github.com/en/actions/reference/workflow-commands-for-github-actions#github_repository_is_fork

		GITHUB_REPOSITORY_IS_FORK is true if the repository is a fork.
	*/
	DefaultBranch *string `json:"default_branch"`
	Fork          *bool   `json:"fork"`
	Private       *bool   `json:"private"`
}

// Client holds a context and roundtripper for querying repo info from GitHub.
type Client struct {
	ctx context.Context
	rt  http.RoundTripper
}

// SetContext sets a context for a GitHub client.
func (c *Client) SetContext(ctx context.Context) {
	c.ctx = ctx
}

// SetTransport sets a http.RoundTripper for a GitHub client.
func (c *Client) SetTransport(rt http.RoundTripper) {
	c.rt = rt
}

// Transport returns the http.RoundTripper for a GitHub client.
func (c *Client) Transport() http.RoundTripper {
	return c.rt
}

// SetDefaultTransport sets the scorecard roundtripper for a GitHub client.
func (c *Client) SetDefaultTransport() {
	logger := sclog.NewLogger(sclog.DefaultLevel)
	rt := roundtripper.NewTransport(c.ctx, logger)
	c.rt = rt
}

// ParseFromURL is a function to get the repository information.
// It is decided to not use the golang GitHub library because of the
// dependency on the github.com/google/go-github/github library
// which will in turn require other dependencies.
func (c *Client) ParseFromURL(baseRepoURL, repoName string) (RepoInfo, error) {
	var ret RepoInfo
	baseURL, err := url.Parse(baseRepoURL)
	if err != nil {
		return ret, fmt.Errorf("parsing base repo URL: %w", err)
	}

	repoURL, err := baseURL.Parse(fmt.Sprintf("repos/%s", repoName))
	if err != nil {
		return ret, fmt.Errorf("parsing repo endpoint: %w", err)
	}

	log.Printf("getting repo info from URL: %s", repoURL.String())
	//nolint:noctx
	req, err := http.NewRequestWithContext(
		c.ctx,
		http.MethodGet,
		repoURL.String(),
		nil /*body*/)
	if err != nil {
		return ret, fmt.Errorf("error creating request: %w", err)
	}

	// authenticate the request if there is a token
	// this will lower the change of hitting the rate limit
	auth, present := os.LookupEnv("GITHUB_AUTH_TOKEN")
	if present {
		req.SetBasicAuth("x", auth)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ret, fmt.Errorf("error creating request: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return ret, fmt.Errorf("error reading response body: %w", err)
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return ret, fmt.Errorf("error reading response body: %w", err)
	}

	prettyPrintJSON(respBytes)
	ret.respBytes = respBytes
	if err := json.Unmarshal(respBytes, &ret.Repo); err != nil {
		return ret, fmt.Errorf("error decoding response body: %w", err)
	}
	return ret, nil
}

// ParseFromFile is a function to get the repository information
// from GitHub event file.
func (c *Client) ParseFromFile(filepath string) (RepoInfo, error) {
	var ret RepoInfo

	log.Printf("getting repo info from file: %s", filepath)
	repoInfo, err := os.ReadFile(filepath)
	if err != nil {
		return ret, fmt.Errorf("reading GitHub event path: %w", err)
	}

	prettyPrintJSON(repoInfo)
	if err := json.Unmarshal(repoInfo, &ret); err != nil {
		return ret, fmt.Errorf("unmarshalling repo info: %w", err)
	}

	return ret, nil
}

// NewClient returns a new Client for querying repo info from GitHub.
func NewClient(ctx context.Context) *Client {
	c := &Client{
		ctx: ctx,
	}

	if c.ctx == nil {
		c.SetContext(context.Background())
	}
	c.SetDefaultTransport()
	return c
}

func prettyPrintJSON(jsonBytes []byte) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, jsonBytes, "", ""); err != nil {
		log.Printf("%v", err)
		return
	}
	log.Println(buf.String())
}
