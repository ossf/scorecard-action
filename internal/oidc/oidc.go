// Copyright 2025 OpenSSF Authors
// Copyright 2021 The Sigstore Authors.
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

// Package oidc provides functionality to get an OIDC token from github.
package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	envRequestURL   = "ACTIONS_ID_TOKEN_REQUEST_URL"
	envRequestToken = "ACTIONS_ID_TOKEN_REQUEST_TOKEN"
)

func RequestToken(ctx context.Context) (string, error) {
	url := os.Getenv(envRequestURL) + "&audience=sigstore"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	// May be replaced by a different client if we hit HTTP_1_1_REQUIRED.
	client := http.DefaultClient

	// Retry up to 3 times.
	for i := 0; ; i++ {
		req.Header.Add("Authorization", "bearer "+os.Getenv(envRequestToken))
		resp, err := client.Do(req)
		if err != nil {
			if i == 2 {
				return "", err
			}

			// This error isn't exposed by net/http, and retrying this with the
			// DefaultClient will fail because it will just use HTTP2 again.
			// I don't know why go doesn't do this for us.
			if strings.Contains(err.Error(), "HTTP_1_1_REQUIRED") {
				http1transport := http.DefaultTransport.(*http.Transport).Clone()
				http1transport.ForceAttemptHTTP2 = false

				client = &http.Client{
					Transport: http1transport,
				}
			}

			fmt.Fprintf(os.Stderr, "error fetching GitHub OIDC token (will retry): %v\n", err)
			time.Sleep(time.Second)
			continue
		}
		defer resp.Body.Close()

		var payload struct {
			Value string `json:"value"`
		}
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(&payload); err != nil {
			return "", err
		}
		return payload.Value, nil
	}
}
