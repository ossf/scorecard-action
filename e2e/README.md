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


## Steps to add a new test case

1. Create a new repository in the `ossf-tests` organization
2. Clone this workflow https://github.com/ossf-tests/scorecard-action-non-main-branch/blob/other/.github/workflows/scorecard-analysis.yml which has the steps to create an issue if the action fails to run.
3. Update the Title on the issue creation commandline https://github.com/ossf-tests/scorecard-action-non-main-branch/blob/457e0b56ed2b2764fbfa60ce68fdbce315ae3732/.github/workflows/scorecard-analysis.yml#L67. In this example replace `Failed to run automated scorecard-action@main ossf-tests/scorecard-action-non-main-branch` with `Failed to run automated scorecard-action@main with the new org/reponame`
