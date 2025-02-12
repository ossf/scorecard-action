# NOTE: Keep this in sync with go.mod for ossf/scorecard.
LDFLAGS=-X sigs.k8s.io/release-utils/version.gitVersion=v5.1.0 -X sigs.k8s.io/release-utils/version.gitCommit=b0143fc57d8d38748990027266de715052806f4b -w -extldflags \"-static\"

build: ## Runs go build on repo
	# Run go build and generate scorecard executable
	CGO_ENABLED=0 go build -o scorecard-action -trimpath -a -tags netgo -ldflags '$(LDFLAGS)'
