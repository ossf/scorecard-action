# Scorecard GitHub Action installer

This tool can add the
[scorecard GitHub Action](https://github.com/ossf/scorecard-action) to all
accessible repositories under a given organization. A pull request will be
created so that owners can decide whether or not they want to include the
workflow.

## Requirements

Running this tool requires a Personal Access Token (PAT) with the following scopes:

- `repo > public_repo`
- `admin:org > read:org`

Instructions on creating a personal access token can be found
[here](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token).

## Usage

```console
❯ go run cmd/installer/main.go --help

The Scorecard GitHub Action installer simplifies the installation of the
scorecard GitHub Action by creating pull requests through the command line.

Usage:
  --owner example_org [--repos <repo1,repo2,repo3>] [flags]

Flags:
  -h, --help            help for --owner
      --owner string    org/owner to install the scorecard action for
      --repos strings   repositories to install the scorecard action on
```

Another PAT should also be defined as an organization secret for
`scorecards.yml` using steps listed in
[scorecard-action](https://github.com/ossf/scorecard-action#pat-token-creation).
