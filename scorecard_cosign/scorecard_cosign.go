package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sigstore/cosign/cmd/cosign/cli/options"
	"github.com/sigstore/cosign/cmd/cosign/cli/sign"
	_ "github.com/sigstore/cosign/cmd/cosign/cli/verify"
)

func main() {
	err := signUploadResult("payload.txt")
	if err != nil {
		fmt.Println(err)
	}

}

func signUploadResult(payload string) error {
	os.Setenv("COSIGN_EXPERIMENTAL", "true")
	ctx := context.Background()
	// Sign the data in payload.txt and generate certificate and tlog entry in Rekor.
	keyOpts := sign.KeyOpts{
		FulcioURL:    "https://fulcio.sigstore.dev",
		RekorURL:     "https://rekor.sigstore.dev",
		OIDCIssuer:   options.DefaultOIDCIssuerURL,
		OIDCClientID: "sigstore",
	}
	regOpts := options.RegistryOptions{}

	res, err := sign.SignBlobCmd(ctx, keyOpts, regOpts, "payload.txt", true, "output_signature", "output_certificate", time.Minute) //b64?
	if err != nil {
		return fmt.Errorf("error signing blob: %w", err)
	}
	fmt.Println(res)

	// (For testing) Verify signature.
	// err = verify.VerifyBlobCmd(ctx, keyOpts, "output_certificate", "", "", "output_signature", "payload.txt")
	// if err != nil {
	// 	return fmt.Errorf("error verifying blob signature: %w", err)
	// }
	return nil
}
