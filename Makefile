# NOTE: Keep this in sync with go.mod for ossf/scorecard.
LDFLAGS=-X sigs.k8s.io/release-utils/version.gitVersion=v5.3.0 -X sigs.k8s.io/release-utils/version.gitCommit=c22063e786c11f9dd714d777a687ff7c4599b600 -w -extldflags \"-static\"

build: ## Runs go build on repo
	# Run go build and generate scorecard executable
	CGO_ENABLED=0 go build -o scorecard-action -trimpath -a -tags netgo -buildvcs=false -ldflags '$(LDFLAGS)'
