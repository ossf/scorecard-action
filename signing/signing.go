package signing

import (
	"fmt"
	"os"

	"github.com/sigstore/cosign/cmd/cosign/cli/options"
	"github.com/sigstore/cosign/cmd/cosign/cli/sign"
)

func signScorecardResult(scorecardResultsFile string) error {
	if err := os.Setenv("COSIGN_EXPERIMENTAL", "true"); err != nil {
		return fmt.Errorf("error setting COSIGN_EXPERIMENTAL env var: %w", err)
	}

	// Prepare settings for SignBlobCmd.
	rootOpts := &options.RootOptions{Timeout: options.DefaultTimeout} // Just the timeout.

	keyOpts := sign.KeyOpts{
		FulcioURL:    options.DefaultFulcioURL,     // Signing certificate provider.
		RekorURL:     options.DefaultRekorURL,      // Transparency log.
		OIDCIssuer:   options.DefaultOIDCIssuerURL, // OIDC provider to get ID token to auth for Fulcio.
		OIDCClientID: "sigstore",
	}
	regOpts := options.RegistryOptions{} // Not necessary so we leave blank.

	// This command will use the provided OIDCIssuer to authenticate into Fulcio, which will generate the
	// signing certificate on the scorecard result. This attestation is then uploaded to the Rekor transparency log.
	// The output bytes (signature) and certificate are discarded since verification can be done with just the payload.
	if _, err := sign.SignBlobCmd(rootOpts, keyOpts, regOpts, scorecardResultsFile, true, "", ""); err != nil {
		return fmt.Errorf("error signing payload: %w", err)
	}

	return nil
}
