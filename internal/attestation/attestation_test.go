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

package attestation

import (
	"errors"
	"slices"
	"testing"

	"github.com/policylabs/signer/sts"
)

// The signer must only look for credentials in GitHub Actions, whatever
// else the machine running the job could offer.
func TestNewSignerOnlyUsesTheWorkflowIdentity(t *testing.T) {
	t.Parallel()
	s := newSigner()
	defer s.Close()

	if want := []string{stsProviderActions}; !slices.Equal(s.Options.STSProviders, want) {
		t.Fatalf("STSProviders = %v, want %v", s.Options.STSProviders, want)
	}

	// The name has to be one the signer knows, or signing fails.
	providers, err := sts.Select(s.Options.STSProviders...)
	if err != nil {
		t.Fatalf("selecting the STS providers: %v", err)
	}
	if len(providers) != 1 || providers[0].Name != stsProviderActions {
		t.Fatalf("selected providers = %v, want only %q", providers, stsProviderActions)
	}
}

func TestSignEmptyStatement(t *testing.T) {
	t.Parallel()
	if _, err := Sign(nil); !errors.Is(err, errEmptyStatement) {
		t.Fatalf("Sign(nil) error = %v, want %v", err, errEmptyStatement)
	}
}
