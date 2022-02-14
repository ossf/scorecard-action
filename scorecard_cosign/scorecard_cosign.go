package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sigstore/cosign/cmd/cosign/cli/options"
	"github.com/sigstore/cosign/cmd/cosign/cli/sign"
)

var keyPass = []byte("hello")

var passFunc = func(_ bool) ([]byte, error) {
	return keyPass, nil
}

func main() {
	testData := "Hello World"
	fmt.Println(testData)

	os.Setenv("COSIGN_EXPERIMENTAL", "true")
	ctx := context.Background()

	// client, err := rekor.NewClient("rekor.sigstore.dev")
	// err_check(err, "Rekor new client error.")

	// //Generate key
	// key, err := cosign.GenerateKeyPair(passFunc)
	// err_check(err, "Generate key error.")
	// os.WriteFile("signKey", key.PrivateBytes, 0600)

	keyOpts := sign.KeyOpts{
		// KeyRef:    "signKey",
		PassFunc:  passFunc,
		FulcioURL: "https://fulcio.sigstore.dev",
		RekorURL:  "https://rekor.sigstore.dev",
	}
	regOpts := options.RegistryOptions{}
	res, err := sign.SignBlobCmd(ctx, keyOpts, regOpts, "payload.txt", true, "output_signature", "output_certificate", time.Minute) //b64?
	err_check(err, "Sign blob error.")
	fmt.Println(res)

	// //GetRekorPub retrieves the rekor public key from the embedded or cached TUF root.
	// //If expired, makes a network call to retrieve the updated target.
	// cosign.GetRekorPub()

	// //TLogUpload will upload the signature, public key and payload to the transparency log.
	// cosign.TLogUpload()

	// //VerifyImageSignature verifies a signature
	// cosign.VerifyImageSignature()

}

func err_check(err error, msg string) {
	if err != nil {
		fmt.Println(msg, err)
		os.Exit(1)
	}
}
