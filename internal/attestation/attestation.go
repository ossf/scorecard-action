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

// Sign signs an in-toto statement keylessly with Sigstore, using the job's
// OIDC identity, and returns the resulting Sigstore bundle as JSON. The
// signer obtains the identity token from the runner's ambient credentials,
// so the job needs the id-token: write permission.
func Sign(statement []byte) ([]byte, error) {
	if len(statement) == 0 {
		return nil, errEmptyStatement
	}
	s := signer.NewSigner()
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
