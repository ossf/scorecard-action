# NOTE: Keep this in sync with go.mod for ossf/scorecard.
LDFLAGS=-X sigs.k8s.io/release-utils/version.gitVersion=v4.3.1 -X sigs.k8s.io/release-utils/version.gitCommit=70d045b9ef00e7171ce3950aca38eef6ea4d7308 -w -extldflags \"-static\"

build: ## Runs go build on repo
	# Run go build and generate scorecard executable
	CGO_ENABLED=0 go build -o scorecard-action -trimpath -a -tags netgo -ldflags '$(LDFLAGS)'