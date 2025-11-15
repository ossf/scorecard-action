# NOTE: Keep this in sync with go.mod for ossf/scorecard.
LDFLAGS=-X sigs.k8s.io/release-utils/version.gitVersion=v5.4.0 -X sigs.k8s.io/release-utils/version.gitCommit=80ee3ecfedf8b19ab8991713a9fdb2e7dcd7262e -w -extldflags \"-static\"

build: ## Runs go build on repo
	# Run go build and generate scorecard executable
	CGO_ENABLED=0 go build -o scorecard-action -trimpath -a -tags netgo -buildvcs=false -ldflags '$(LDFLAGS)'
