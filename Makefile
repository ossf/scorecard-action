build: ## Runs go build on repo
	# Run go build and generate scorecard executable. The Scorecard version it
	# reports comes from go.mod, via scripts/version-ldflags.
	version_ldflags=$$(./scripts/version-ldflags) && \
	CGO_ENABLED=0 go build -o scorecard-action -trimpath -a -tags netgo -buildvcs=false -ldflags "$$version_ldflags -w -extldflags \"-static\""
