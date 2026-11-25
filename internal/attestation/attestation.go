// Copyright 2026 Security Scorecard Authors
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

// Package attestation signs the in-toto statement the action produces.
package attestation

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/policylabs/signer"
)

var errEmptyStatement = errors.New("cannot sign an empty statement")

// stsProviderActions is the name of the signer's credential provider that
// reads the OIDC token of a GitHub Actions job.
const stsProviderActions = "actions"

// newSigner returns a signer that only signs with the identity of the
// workflow.
//
// The signer can find credentials in more places than the workflow: on a
// self-hosted runner in Google Cloud, for example, it could sign as the
// service account of the machine. Only the workflow's token is used.
func newSigner() *signer.Signer {
	s := signer.NewSigner()
	s.Options.STSProviders = []string{stsProviderActions}
	return s
}

// Sign signs an in-toto statement keylessly with Sigstore, using the job's
// OIDC identity, and returns the resulting Sigstore bundle as JSON. The
// signer obtains the identity token from GitHub Actions, so the job needs
// the id-token: write permission.
func Sign(statement []byte) ([]byte, error) {
	if len(statement) == 0 {
		return nil, errEmptyStatement
	}
	s := newSigner()
	defer s.Close()

	bndl, err := s.SignStatementBundle(statement)
	if err != nil {
		return nil, fmt.Errorf("signing statement: %w", err)
	}

	var out bytes.Buffer
	if err := s.WriteBundle(bndl, &out); err != nil {
		return nil, fmt.Errorf("encoding sigstore bundle: %w", err)
	}
	return out.Bytes(), nil
}
