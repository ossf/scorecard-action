# What

e2e Scorecard action tests for differences in functionality between Scorecard
action implemented in Bash and the updated version implemented using Golang.
These e2e tests will be used until the release of Scorecard Golang action after
which these tests will be modified to run regular e2e testing.

# Setup

For testing functionality difference between the 2 implementations, we need a
setup which can invoke these implementations through a GitHub Action on the same
repo/commitSHA. We achieve this by:

1.  The 2 implementations are built using 2 separate Dockerfiles. `./Dockerfile`
    for Bash and `./Dockerfile.golang` for Golang.
2.  A CloudBuild trigger uses `./cloudbuild.yaml` to continuously build and
    generate the Golang Docker image. This also helps reduce run time during the
    actual GitHub Action run. The generated Docker image is tagged
    `scorecard-action:latest`.
3.  Bash implementation at `HEAD` is invoked by referencing: `uses:
    ossf/scorecard-action@main` in a GitHub workflow file.
4.  The same repository invokes Golang implementation by referencing: `uses:
    gcr.io/openssf/scorecard-action:latest`
5.  The artifact (SARIF file) produced by these 2 implementations are diff-ed to
    verify functional similarity. This step is not yet automated and is largely
    manual.

# e2e tests

The `e2e` tests for the action is run by running the action every day on a cron
for different use cases. The action that run points to `@main` which helps in
catching issues sooner.

If these actions fails to run these actions would create an issue in the
repository using https://github.com/naveensrinivasan/Create-GitHub-Issue

The actions primarily run out of https://github.com/ossf-tests organization.

## Status

Testcase           | Action | Repository                                                                  | Status.
------------------ | ------ | --------------------------------------------------------------------------- | -------
Fork               | Bash   | https://github.com/ossf-tests/scorecard-action                              | [![Fork](https://github.com/ossf-tests/scorecard-action/actions/workflows/scorecards-bash.yml/badge.svg)](https://github.com/ossf-tests/scorecard-action/actions/workflows/scorecards-bash.yml)
Fork               | Golang | https://github.com/ossf-tests/scorecard-action                              | [![Fork](https://github.com/ossf-tests/scorecard-action/actions/workflows/scorecards-golang.yml/badge.svg)](https://github.com/ossf-tests/scorecard-action/actions/workflows/scorecards-golang.yml)
Non-main-branch    | Bash   | https://github.com/ossf-tests/scorecard-action-non-main-branch              | [![non-main-branch](https://github.com/ossf-tests/scorecard-action-non-main-branch/actions/workflows/scorecards-bash.yml/badge.svg)](https://github.com/ossf-tests/scorecard-action-non-main-branch/actions/workflows/scorecards-bash.yml)
Non-main-branch    | Golang | https://github.com/ossf-tests/scorecard-action-non-main-branch              | [![non-main-branch](https://github.com/ossf-tests/scorecard-action-non-main-branch/actions/workflows/scorecards-golang.yml/badge.svg)](https://github.com/ossf-tests/scorecard-action-non-main-branch/actions/workflows/scorecards-golang.yml)
Private repository | Bash   | https://github.com/test-organization-ls/scorecard-action-private-repo-tests | [![Scorecards supply-chain security](https://github.com/test-organization-ls/scorecard-action-private-repo-tests/actions/workflows/scorecard.yml/badge.svg)](https://github.com/test-organization-ls/scorecard-action-private-repo-tests/actions/workflows/scorecard.yml)
Private repository | Golang | https://github.com/test-organization-ls/scorecard-action-private-repo-tests | [![Scorecards supply-chain security](https://github.com/test-organization-ls/scorecard-action-private-repo-tests/actions/workflows/scorecards-golang.yml/badge.svg)](https://github.com/test-organization-ls/scorecard-action-private-repo-tests/actions/workflows/scorecards-golang.yml)

## Diff between golang-staging branch and main

-   Here is the sarif results diff between main and golang-staging. There are
    few text diffs
    https://github.com/ossf-tests/scorecard-action-results/pull/1/files. The PR
    is for golang run results. The `main` branch has the `scorecard-action`
    `main` branch run results.

## Steps to add a new test case

1.  Create a new repository in the `ossf-tests` organization
2.  Clone this workflow
    https://github.com/ossf-tests/scorecard-action-non-main-branch/blob/other/.github/workflows/scorecard-analysis.yml
    which has the steps to create an issue if the action fails to run. If the
    action fails it should create an issue like this
    https://github.com/ossf/scorecard-action/issues/147
