# OpenSSF Dependency Analysis

This repository contains the source code for the OpenSSF Dependency Analysis project. The aim of the project is to check the security posture of a project's dependencies using the [GitHub Dependency Graph API](https://docs.github.com/en/rest/dependency-graph/dependency-review?apiVersion=2022-11-28#get-a-diff-of-the-dependencies-between-commits) and the [Security Scorecards API](https://api.securityscorecards.dev).

## Usage
The OpenSSF Dependency Analysis is a GitHub Action that can be easily incorporated into a workflow. 
The workflow can be triggered on a pull request event. 
The action will run on the latest commit on the default branch of the repository, and will create a comment on the pull request with the results of the analysis. 
An example of the comment can be found [here](https://github.com/ossf-tests/vulpy/pull/2#issuecomment-1442310469).

## Prerequisites
The actions require enabling the [GitHub Dependency](https://docs.github.com/en/code-security/supply-chain-security/understanding-your-software-supply-chain/about-dependency-review) for the repository.

### Configuration
The action can be configured using the following inputs:

- `SCORECARD_CHECKS`: This environment variable takes a file containing a list of checks to run. 
- The file should be in JSON format and follow the format provided by the [Scorecard checks documentation](https://github.com/ossf/scorecard/blob/main/docs/checks.md). For example:
```json
[
  "Binary-Artifacts",
  "Pinned-Dependencies"
] 
```

### Installation
The action can be installed by adding the following snippet to the workflow file:
```yaml
name: scorecard-dependency-analysis

on:
  pull_request:
    types: [opened, synchronize, reopened]
permissions:
  pull-requests: write # Required to create a comment on the pull request.

jobs:
  dependency-analysis:
    name: Scorecards dependency analysis
    runs-on: ubuntu-latest
    env:
      GITHUB_PR_NUMBER: ${{ github.event.number }}
      GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      GITHUB_REPOSITORY: ${{ github.repository }}
      GITHUB_REPOSITORY_OWNER: ${{ github.repository_owner }}
      GITHUB_SHA: ${{ github.sha }}
      GITHUB_ACTOR: ${{ github.actor }}


    steps:
      - name: Checkout code
        uses: actions/checkout@v2
        with:
          persist-credentials: false

      - name: Run dependency analysis
        uses: github.com/ossf/scorecard-action/dependency-analysis@main # Replace with the latest release version.
```
