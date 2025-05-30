# NOTE: Keep this in sync with go.mod for ossf/scorecard.
LDFLAGS=-X sigs.k8s.io/release-utils/version.gitVersion=v5.2.1 -X sigs.k8s.io/release-utils/version.gitCommit=ab2f6e92482462fe66246d9e32f642855a691dc1 -w -extldflags \"-static\"

build: ## Runs go build on repo
	# Run go build and generate scorecard executable
	CGO_ENABLED=0 go build -o scorecard-action -trimpath -a -tags netgo -buildvcs=false -ldflags '$(LDFLAGS)'
