build: ## Runs go build on repo
	# Run go build and generate scorecard executable. The Scorecard version it
	# reports comes from go.mod, via scripts/version-ldflags.
	version_ldflags=$$(./scripts/version-ldflags) && \
	CGO_ENABLED=0 go build -o scorecard-action -trimpath -a -tags netgo -buildvcs=false -ldflags "$$version_ldflags -w -extldflags \"-static\""

generate: ## Renders the example workflows into examples/
	# Every example comes from internal/workflowtemplate, with actions pinned
	# by internal/workflowtemplate/pins/action.yml.
	go run ./internal/workflowtemplate/cmd/generate -out examples
