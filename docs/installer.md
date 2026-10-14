# Installing the workflow across repositories

The installer adds the Scorecard workflow to many repositories at once. For
each repository, it opens a pull request that adds
`.github/workflows/scorecard.yml`.

The workflow comes from the [example workflows](/examples), built into the
installer:

- Public repositories get [`scorecard.yml`](/examples/scorecard.yml).
- Private repositories get [`scorecard-private.yml`](/examples/scorecard-private.yml).
- The workflow runs on pushes to the repository's default branch, and weekly.

## Requirements

- Go, at the version in [`go.mod`](/go.mod).
- A GitHub token in `GITHUB_AUTH_TOKEN` (`GITHUB_TOKEN`, `GH_TOKEN` and
  `GH_AUTH_TOKEN` also work). It needs to create branches, workflow files and
  pull requests in the target repositories. With a fine-grained token, grant
  read and write access to **Contents**, **Pull requests** and **Workflows**.

## Usage

Run the installer from a release tag, so it adds that release's workflow:

```shell
git clone --depth 1 --branch <release tag> https://github.com/ossf/scorecard-action
cd scorecard-action
export GITHUB_AUTH_TOKEN=<token>
go run ./cmd/installer --owner <org> --repos <repo1>,<repo2>
```

Don't use `go run github.com/ossf/scorecard-action/cmd/installer@latest`. The
module path has no `/v2` suffix, so Go resolves `@latest` to v1.1.2, not the
latest release.

- `--owner` (required): the organization or user that owns the repositories.
- `--repos`: comma-separated repository names. Without it, the installer uses
  the first 30 repositories the token can see under `--owner`.

## What it skips

The installer leaves a repository unchanged when:

- it already has `.github/workflows/scorecard.yml`, `scorecards.yml` or
  `scorecards-analysis.yml`, or
- a `scorecard-action-install` branch already exists.

Each pull request comes from that `scorecard-action-install` branch. Review and
merge it like any other change.
