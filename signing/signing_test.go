package signing

import (
	"context"
	"crypto/rand"
	"io/ioutil"
	"testing"

	"github.com/sigstore/cosign/cmd/cosign/cli/options"
	"github.com/sigstore/cosign/cmd/cosign/cli/rekor"
	"github.com/sigstore/cosign/pkg/cosign"
)

func Test_SignScorecardResult(t *testing.T) {
	t.Parallel()
	// Generate random bytes to use as our payload. This is done because signing identical payloads twice
	// just creates multiple entries under it, so we are keeping this test simple and not comparing timestamps.
	scorecardResultsFile := "./sign-random-data.txt"
	randomData := make([]byte, 20)
	if _, err := rand.Read(randomData); err != nil {
		t.Errorf("signScorecardResult() error generating random bytes, %v", err)
		return
	}
	if err := ioutil.WriteFile(scorecardResultsFile, randomData, 0644); err != nil {
		t.Errorf("signScorecardResult() error writing random bytes to file, %v", err)
		return
	}

	// Sign example scorecard results file.
	err := SignScorecardResult(scorecardResultsFile)
	if err != nil {
		t.Errorf("signScorecardResult() error, %v", err)
		return
	}

	// Verify that the signature was created and uploaded to the Rekor tlog by looking up the payload.
	ctx := context.Background()
	rekorClient, err := rekor.NewClient(options.DefaultRekorURL)
	if err != nil {
		t.Errorf("signScorecardResult() error getting Rekor client, %v", err)
		return
	}
	scorecardResultData, err := ioutil.ReadFile(scorecardResultsFile)
	if err != nil {
		t.Errorf("signScorecardResult() error reading scorecard result file, %v", err)
		return
	}
	uuids, err := cosign.FindTLogEntriesByPayload(ctx, rekorClient, scorecardResultData)
	if err != nil {
		t.Errorf("signScorecardResult() error getting tlog entries, %v", err)
		return
	}

	if len(uuids) != 1 {
		t.Errorf("signScorecardResult() error finding signature in Rekor tlog, %v", err)
		return
	}
}

// Test using scorecard results that have already been signed & uploaded.
func Test_processSignature(t *testing.T) {
	t.Parallel()

	sarifPayload, serr := ioutil.ReadFile("testdata/results.sarif")
	jsonPayload, jerr := ioutil.ReadFile("testdata/results.json")

	if serr != nil || jerr != nil {
		t.Errorf("Error reading testdata:, %v, %v", serr, jerr)
	}

	if err := ProcessSignature(sarifPayload, jsonPayload, "rohankh532/scorecard-OIDC-test", "refs/heads/main"); err != nil {
		t.Errorf("ProcessSignature() error:, %v", err)
		return
	}

}
