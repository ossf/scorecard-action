# NOTE: Keep this in sync with go.mod for ossf/scorecard.
LDFLAGS=-X sigs.k8s.io/release-utils/version.gitVersion=v5.5.0 -X sigs.k8s.io/release-utils/version.gitCommit=c395761df6afe1a69e476bc60a013a94bcbc153f -w -extldflags \"-static\"

build: ## Runs go build on repo
	# Run go build and generate scorecard executable
	CGO_ENABLED=0 go build -o scorecard-action -trimpath -a -tags netgo -buildvcs=false -ldflags '$(LDFLAGS)'
