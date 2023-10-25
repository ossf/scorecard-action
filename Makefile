# NOTE: Keep this in sync with go.mod for ossf/scorecard.
LDFLAGS=-X sigs.k8s.io/release-utils/version.gitVersion=v4.13.1 -X sigs.k8s.io/release-utils/version.gitCommit=49c0eed3a423f00c872b5c3c9f1bbca9e8aae799 -w -extldflags \"-static\"

build: ## Runs go build on repo
	# Run go build and generate scorecard executable
	CGO_ENABLED=0 go build -o scorecard-action -trimpath -a -tags netgo -ldflags '$(LDFLAGS)'
