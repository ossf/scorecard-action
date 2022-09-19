# NOTE: Keep this in sync with go.mod for ossf/scorecard.
LDFLAGS=-X sigs.k8s.io/release-utils/version.gitVersion=v4.6.1-0.20220919161004 -X sigs.k8s.io/release-utils/version.gitCommit=9f67c4ead1163fceae6931e892634c3b12d86e0a -w -extldflags \"-static\"

build: ## Runs go build on repo
	# Run go build and generate scorecard executable
	CGO_ENABLED=0 go build -o scorecard-action -trimpath -a -tags netgo -ldflags '$(LDFLAGS)'
