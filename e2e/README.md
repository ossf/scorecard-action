# e2e tests

The `e2e` tests for the action is run by running the action every day on a cron for different use cases. The action that run points to `@main` which helps in catching issues sooner.

If these actions fails to run these actions would create an issue in the repository using https://github.com/naveensrinivasan/Create-GitHub-Issue

The actions primarily run out of https://github.com/ossf-tests organization.

## Status

| Testcase | Repository | Status.  |
| -------- | --------   | -------- |
| Fork     | https://github.com/ossf-tests/scorecard-action       | [![Fork](https://github.com/ossf-tests/scorecard-action/actions/workflows/scorecards.yml/badge.svg)](https://github.com/ossf-tests/scorecard-action/actions/workflows/scorecards.yml)     |
| Non-main-branch    | https://github.com/ossf-tests/scorecard-action-non-main-branch       | [![non-main-branch](https://github.com/ossf-tests/scorecard-action-non-main-branch/actions/workflows/scorecard-analysis.yml/badge.svg?branch=other)](https://github.com/ossf-tests/scorecard-action-non-main-branch/actions/workflows/scorecard-analysis.yml) |
|Private repository|https://github.com/test-organization-ls/scorecard-action-private-repo-tests| [![Scorecards supply-chain security](https://github.com/test-organization-ls/scorecard-action-private-repo-tests/actions/workflows/scorecard.yml/badge.svg)](https://github.com/test-organization-ls/scorecard-action-private-repo-tests/actions/workflows/scorecard.yml) |

| Fork-golang-staging    | https://github.com/ossf-tests/scorecard-action       |[![Scorecards supply-chain security](https://github.com/ossf-tests/scorecard-action/actions/workflows/scorecards-golang.yml/badge.svg)](https://github.com/ossf-tests/scorecard-action/actions/workflows/scorecards-golang.yml) 
| Non-main-branch-golang-staging    | https://github.com/ossf-tests/scorecard-action-non-main-branch       | [![Scorecards supply-chain security golang](https://github.com/ossf-tests/scorecard-action-non-main-branch/actions/workflows/scorecard-golang.yml/badge.svg)](https://github.com/ossf-tests/scorecard-action-non-main-branch/actions/workflows/scorecard-golang.yml)
|Private repository-golang-staging|https://github.com/test-organization-ls/scorecard-action-private-repo-tests|[![Scorecards supply-chain security golang](https://github.com/test-organization-ls/scorecard-action-private-repo-tests/actions/workflows/scorecards-golang.yml/badge.svg)](https://github.com/test-organization-ls/scorecard-action-private-repo-tests/actions/workflows/scorecards-golang.yml) 


## Diff between golang-staging branch and main

- Here is the sarif results diff between main and golang-staging. There are few text diffs https://github.com/ossf-tests/scorecard-action-results/pull/1/files. The PR is for golang run results. The `main` branch has the `scorecard-action` `main` branch run results.

## Steps to add a new test case

1. Create a new repository in the `ossf-tests` organization
2. Clone this workflow https://github.com/ossf-tests/scorecard-action-non-main-branch/blob/other/.github/workflows/scorecard-analysis.yml which has the steps to create an issue if the action fails to run. If the action fails it should create an issue like this https://github.com/ossf/scorecard-action/issues/147
