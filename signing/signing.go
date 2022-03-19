package signing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/ossf/scorecard-action/entrypoint"
	"github.com/ossf/scorecard-action/options"
	sigOpts "github.com/sigstore/cosign/cmd/cosign/cli/options"
	"github.com/sigstore/cosign/cmd/cosign/cli/sign"
)

func SignScorecardResult(scorecardResultsFile string) error {
	if err := os.Setenv("COSIGN_EXPERIMENTAL", "true"); err != nil {
		return fmt.Errorf("error setting COSIGN_EXPERIMENTAL env var: %w", err)
	}

	// Prepare settings for SignBlobCmd.
	rootOpts := &sigOpts.RootOptions{Timeout: sigOpts.DefaultTimeout} // Just the timeout.

	keyOpts := sign.KeyOpts{
		FulcioURL:    sigOpts.DefaultFulcioURL,     // Signing certificate provider.
		RekorURL:     sigOpts.DefaultRekorURL,      // Transparency log.
		OIDCIssuer:   sigOpts.DefaultOIDCIssuerURL, // OIDC provider to get ID token to auth for Fulcio.
		OIDCClientID: "sigstore",
	}
	regOpts := sigOpts.RegistryOptions{} // Not necessary so we leave blank.

	// This command will use the provided OIDCIssuer to authenticate into Fulcio, which will generate the
	// signing certificate on the scorecard result. This attestation is then uploaded to the Rekor transparency log.
	// The output bytes (signature) and certificate are discarded since verification can be done with just the payload.
	if _, err := sign.SignBlobCmd(rootOpts, keyOpts, regOpts, scorecardResultsFile, true, "", ""); err != nil {
		return fmt.Errorf("error signing payload: %w", err)
	}

	return nil
}

// Calls scorecard-api to process & upload signed scorecard results.
// TODO: not sure how to test this because it requires running the entire scorecard action.
func ProcessSignature() error {
	// Get sarif output from file.
	sarifPayload, err := ioutil.ReadFile(os.Getenv(options.EnvInputResultsFile))
	if err != nil {
		return fmt.Errorf("reading scorecard sarif results from file: %v", err)
	}

	// Change output settings to json and run scorecard again.
	os.Setenv(options.EnvInputResultsFile, "results.json")
	os.Setenv(options.EnvInputResultsFormat, "json")
	actionJson, err := entrypoint.New()

	if err != nil {
		return fmt.Errorf("creating scorecard entrypoint: %v", err)
	}
	if err := actionJson.Execute(); err != nil {
		return fmt.Errorf("error during command execution: %v", err)
	}
	if err = SignScorecardResult("results.sarif"); err != nil {
		return fmt.Errorf("error signing scorecard sarif results: %v", err)
	}

	// Get json output from file.
	jsonPayload, err := ioutil.ReadFile(os.Getenv(options.EnvInputResultsFile))
	if err != nil {
		return fmt.Errorf("reading scorecard json results from file: %v", err)
	}

	// Prepare HTTP request body for scorecard-webapp-api call.
	resultsPayload := struct {
		SarifOutput string
		JsonOutput  string
	}{
		SarifOutput: string(sarifPayload),
		JsonOutput:  string(jsonPayload),
	}

	payloadBytes, err := json.Marshal(resultsPayload)
	if err != nil {
		return fmt.Errorf("reading scorecard json results from file: %v", err)
	}

	// Call scorecard-webapp-api to process and upload signature.
	url := "https://api.securityscorecards.dev/verify"
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	req.Header.Set("Repository", os.Getenv(options.EnvGithubRepository))
	req.Header.Set("Branch", os.Getenv(options.EnvGithubRef))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("executing scorecard-api call: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("http response error: %v", err)
	}

	return nil
}