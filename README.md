# Scorecards' GitHub action
[![CodeQL](https://github.com/ossf/scorecard-action/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/ossf/scorecard-action/actions/workflows/codeql-analysis.yml)
[![codecov](https://codecov.io/gh/ossf/scorecard-action/branch/main/graph/badge.svg?token=MAXISWR53I)](https://codecov.io/gh/ossf/scorecard-action)
> Official GitHub Action for [OSSF Scorecards](https://github.com/ossf/scorecard).

The Scorecards GitHub Action is free for all public repositories. Private repositories are supported if they have [GitHub Advanced Security](https://docs.github.com/en/get-started/learning-about-github/about-github-advanced-security). Private repositories without GitHub Advanced Security can run Scorecards from the command line by following the [standard installation instructions](https://github.com/ossf/scorecard#using-scorecards-1).


## Breaking changes in v2

Starting from scorecard-action:v2, `GITHUB_TOKEN` permissions or job permissions needs to include
`id-token: write` for `publish_results: true`. This is needed to access GitHub's
OIDC token which verifies the authenticity of the result when publishing it. See details [here](#publishing-results)

If publishing results, scorecard-action:v2 also imposes new requirements on both the workflow and the job running the `ossf/scorecard-action` step. For full details see [here](#workflow-restrictions). 
________
[Installation](#installation)
- [Workflow Setup](#workflow-setup-required)
- [Authentication](#authentication-with-pat-optional)

[View Results](#view-results)
- [REST API](#rest-api)
- [Scorecard Badge](#scorecard-badge)
- [Code Scanning Alerts](#code-scanning-alerts)
- [Verify Runs](#verify-runs)
- [Troubleshooting](#troubleshooting)

[Manual Action Setup](#manual-action-setup)
- [Inputs](#inputs)
- [Publishing Results](#publishing-results)
- [Workflow Restrictions](#workflow-restrictions)
- [Uploading Artifacts](#uploading-artifacts)
- [Workflow Example](#workflow-example)
________

The following GitHub triggers are supported: `push`, `schedule` (default branch only).

The `pull_request` and `workflow_dispatch` triggers are experimental.

Running the Scorecard action on a fork repository is not supported.

GitHub Enterprise repositories are not supported.

## Installation

### Workflow Setup (Required)
1) From your GitHub project's main page, click “Security” in the top ribbon.

![image](/images/install01.png)

2) Select “Code scanning”.

![image](/images/install02.png)

3) Then click "Add tool".

![image](/images/install03.png)

4) Choose the "OSSF Scorecard" from the list of workflows, and then click “Configure”.

![image](/images/install04.png)

5) Commit the changes.

![image](/images/install05.png)

### Authentication with Fine-grained Personal Access Token (optional)
Create a fine-grained Personal Access Token (PAT) for authentication and save the token value as a repository secret.

1. [Create a fine-grained Personal Access Token](https://github.com/settings/personal-access-tokens/new) with the following settings:
    - Token name: `OpenSSF Scorecard Action - $USER_NAME/$REPO_NAME>` (Note: replace `$USER_NAME/$REPO_NAME` with the names of your organization and repository so you can keep track of your tokens.)
    - Expiration: Set `Custom` and then set the date to exactly a year in the future (the maximum allowed)
    - Repository Access: `Only select repositories` and select the desired repository. Alternatively, set `All repositories` if you wish to use the same token for all your repositories.
    - Repository Permissions:
        * `Administration: Read-only`: Required to read [Branch-Protection](https://github.com/ossf/scorecard/blob/main/docs/checks.md#branch-protection) settings.
        * `Metadata: Read-only` will be automatically set when you set `Administration`
        * `Webhooks: Read-only`: (Optional) needed for the experimental [Webhook](https://github.com/ossf/scorecard/blob/main/docs/checks.md#webhooks) check.
        
![image](/images/tokenscopes.png)

2. Copy the token value.

3. [Create a new repository secret](https://docs.github.com/en/actions/security-guides/encrypted-secrets#creating-encrypted-secrets-for-a-repository) with the following settings (**Warning:** [GitHub encrypted secrets](https://docs.github.com/en/actions/security-guides/encrypted-secrets) are accessible by all the workflows and maintainers of a repository.):
    - Name: `SCORECARD_TOKEN`
    - Value: the value of the token created in step 1 above.

    Note that fine-grained tokens expire after one year. You'll receive an email from GitHub when your token is about to expire, at which point you must regenerate it. Make sure to update the token string in your repository's secrets.

4. When you call the `ossf/scorecard-action` in your workflow, pass the token as `repo-token: ${{ secrets.SCORECARD_TOKEN }}`.

## View Results

The workflow is preconfigured to run on every repository contribution. After making a code change, you can view the results for the change either through the Scorecard Badge, Code Scanning Alerts or GitHub Workflow Runs.

### REST API
Starting with scorecard-action:v2, users can use a REST API to query their latest run results. This requires setting [`publish_results: true`](https://github.com/ossf/scorecard/blob/d13ba3f3355b958d5d62edc47282a2e7ed9fa7c1/.github/workflows/scorecard-analysis.yml#L39) for the action and enabling [`id-token: write`](https://github.com/ossf/scorecard/blob/d13ba3f3355b958d5d62edc47282a2e7ed9fa7c1/.github/workflows/scorecard-analysis.yml#L22) permission for the job (needed to access GitHub OIDC token). The API is available here: https://api.securityscorecards.dev.

### Scorecard Badge

Starting with scorecard-action:v2, users can add a Scorecard Badge to their README to display the latest status of their Scorecard results. This requires setting [`publish_results: true`](https://github.com/ossf/scorecard/blob/d13ba3f3355b958d5d62edc47282a2e7ed9fa7c1/.github/workflows/scorecard-analysis.yml#L39) for the action and enabling [`id-token: write`](https://github.com/ossf/scorecard/blob/d13ba3f3355b958d5d62edc47282a2e7ed9fa7c1/.github/workflows/scorecard-analysis.yml#L22) permission for the job (needed to access GitHub OIDC token). The badge is updated on every run of scorecard-action and points to the latest result. To add a badge to your README, copy and paste the below line, and replace the {owner} and {repo} parts.

```
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/{owner}/{repo}/badge)](https://api.securityscorecards.dev/projects/github.com/{owner}/{repo})
```

Once this badge is added, clicking on the badge will take users to the latest run result of Scorecard.

![image](/images/badge.png)

### Code Scanning Alerts

A list of results is accessible by going in the Security tab and clicking "Code Scanning Alerts" (it can take a couple minutes for the run to complete and the results to show up). Click on the individual alerts for more information, including remediation instructions. You will need to click "Show more" to expand the full remediation instructions.

![image](/images/remediation.png)

### Verify Runs
The workflow is preconfigured to run on every repository contribution.

To verify that the Action is running successfully, click the repository's Actions tab to see the status of all recent workflow runs. This tab will also show the logs, which can help you troubleshoot if the run failed.

![image](/images/actionconfirm.png)

### Troubleshooting
If the run has failed, the most likely reason is an authentication failure. Confirm that the Personal Access Token is saved as an encrypted secret within the same repository (see [Authentication](#authentication)). Also confirm that the PAT is still valid and hasn't expired or been revoked.

If you have a valid PAT saved as an encrypted secret and the run is still failing, confirm that you have not made any changes to the workflow yaml file that affected the syntax. Review the [workflow example](#workflow-example) and reset to the default values if necessary.

## Manual Action Setup

If you prefer to manually set up the Scorecards GitHub Action, you will need to set up a [workflow file](https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions).

First, [create a new file](https://docs.github.com/en/repositories/working-with-files/managing-files/creating-new-files) in this location: `[yourrepo]/.github/workflows/scorecards.yml`. Then use the input values below.


### Inputs

| Name | Required | Description |
| ----- | -------- | ----------- |
| `result_file` | yes | The file that contains the results. |
| `result_format` | yes | The format in which to store the results [json \| sarif]. For GitHub's scanning dashboard, select `sarif`. |
| `repo_token` | no | PAT token with write repository access. Follow [these steps](#authentication-with-pat) to create it. |
| `publish_results` | recommended | This will allow you to display a badge on your repository to show off your hard work. See details [here](#publishing-results).|

### Publishing Results
The Scorecard team runs a weekly scan of public GitHub repositories in order to track
the overall security health of the open source ecosystem. The results of the scans are [publicly
available](https://github.com/ossf/scorecard#public-data).
Setting `publish_results: true` replaces the results of the team's weekly scans with your own scan results,
helping us scale by cutting down on repeated workflows and GitHub API requests.
This option is also needed to enable badges on the repository.

### Workflow Restrictions

If [publishing results](#publishing-results), our API [enforces certain rules](https://github.com/ossf/scorecard-webapp/blob/9c2f66d5f6ff56ca4a4ac2fba6ec8dcc5379d31c/app/server/post_results.go#L184-L187) on the producing workflow, which may reject the results and cause the Scorecard Action run to fail. 
We understand that this is restrictive, but currently it's necessary to ensure the integrity of our API dataset, since GitHub workflow steps run in the same environment as the job they belong to.
If possible, we will work on making this feature more flexible so we can drop this requirement in the future.

#### Global workflow restrictions

* The workflow can't contain top level env vars or defaults.
* No workflow level write permissions.
* Only the job with `ossf/scorecard-action` can use `id-token: write` permissions.

#### Restrictions on the job containing `ossf/scorecard-action`
* No job level env vars or defaults.
* No containers or services
* The job should run on one of the [Ubuntu hosted runners](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#choosing-github-hosted-runners)
* The steps running in this job must belong to this approved list of GitHub actions.
  * "actions/checkout"
  * "actions/upload-artifact"
  * "github/codeql-action/upload-sarif"
  * "ossf/scorecard-action"
  * "step-security/harden-runner"

### Uploading Artifacts
The Scorecards Action uses the [artifact uploader action](https://github.com/actions/upload-artifact) to upload results in SARIF format to the Actions tab. These results are available to anybody for five days after the run to help with debugging. To disable the upload, comment out the `Upload Artifact` value in the Workflow Example.

Note: if you disable this option, the results of the Scorecards Action run will be available only to maintainers (on the Security tab scanning dashboard).

### Workflow Example

```yml
name: Scorecard analysis workflow
on:
  # Only the default branch is supported.
  branch_protection_rule:
  schedule:
    # Weekly on Saturdays.
    - cron: '30 1 * * 6'
  push:
    branches: [ main, master ]

# Declare default permissions as read only.
permissions: read-all

jobs:
  analysis:
    name: Scorecard analysis
    runs-on: ubuntu-latest
    permissions:
      # Needed if using Code scanning alerts
      security-events: write
      # Needed for GitHub OIDC token if publish_results is true
      id-token: write

    steps:
      - name: "Checkout code"
        uses: actions/checkout@8e5e7e5ab8b370d6c329ec480221332ada57f0ab # v3.5.2
        with:
          persist-credentials: false

      - name: "Run analysis"
        uses: ossf/scorecard-action@80e868c13c90f172d68d1f4501dee99e2479f7af # v2.1.3
        with:
          results_file: results.sarif
          results_format: sarif
          # (Optional) fine-grained personal access token. Uncomment the `repo_token` line below if:
          # - you want to enable the Branch-Protection check on a *public* repository, or
          # To create the PAT, follow the steps in https://github.com/ossf/scorecard-action#authentication-with-pat.
          # repo_token: ${{ secrets.SCORECARD_TOKEN }}

          # Publish the results for public repositories to enable scorecard badges. For more details, see
          # https://github.com/ossf/scorecard-action#publishing-results.
          # For private repositories, `publish_results` will automatically be set to `false`, regardless
          # of the value entered here.
          publish_results: true

      # Upload the results as artifacts (optional). Commenting out will disable uploads of run results in SARIF
      # format to the repository Actions tab.
      - name: "Upload artifact"
        uses: actions/upload-artifact@0b7f8abb1508181956e8e162db84b466c27e18ce # v3.1.2
        with:
          name: SARIF file
          path: results.sarif
          retention-days: 5

      # required for Code scanning alerts
      - name: "Upload SARIF results to code scanning"
        uses: github/codeql-action/upload-sarif@83f0fe6c4988d98a455712a27f0255212bba9bd4 # v2.3.6
        with:
          sarif_file: results.sarif
```
