# NOTE: Keep this in sync with go.mod for ossf/scorecard.
LDFLAGS=-X sigs.k8s.io/release-utils/version.gitVersion=v4.11.0 -X sigs.k8s.io/release-utils/version.gitCommit=4edb07802fdad892fa8d10f8fd47666b6ccc27c9 -w -extldflags \"-static\"

build: ## Runs go build on repo
	# Run go build and generate scorecard executable
	CGO_ENABLED=0 go build -o scorecard-action -trimpath -a -tags netgo -ldflags '$(LDFLAGS)'
