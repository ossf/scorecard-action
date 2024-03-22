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
- [Authentication](#authentication-with-fine-grained-pat-optional)

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

[Reporting vulnerabilities](#reporting-vulnerabilities)
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

### Authentication with Fine-grained PAT (optional)
Scorecard can run successfully with the workflow's default `GITHUB_TOKEN`, which is our recommended approach.
However, Scorecard Action requires additional permissions if you use GitHub's classic [Branch Protection](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches) settings and want to see it reflected in your results.
You can read more about how to configure Scorecard Action for these cases [here](/docs/authentication/fine-grained-auth-token.md).

GitHub's new [Repository Rules](https://docs.github.com/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/about-rulesets) are accessible to Scorecard Action with the workflow's default `GITHUB_TOKEN`. 
We recommend new repositories use Repository Rules so they can be read with the default GitHub token. 
Repositories that already use classic Branch Protection and wish to see their results without an admin token should consider migrating to Repository Rules.

## View Results

The workflow is preconfigured to run on every repository contribution. After making a code change, you can view the results for the change either through the Scorecard Badge, Code Scanning Alerts or GitHub Workflow Runs.

### REST API
Starting with scorecard-action:v2, users can use a REST API to query their latest run results. This requires setting [`publish_results: true`](https://github.com/ossf/scorecard/blob/d13ba3f3355b958d5d62edc47282a2e7ed9fa7c1/.github/workflows/scorecard-analysis.yml#L39) for the action and enabling [`id-token: write`](https://github.com/ossf/scorecard/blob/d13ba3f3355b958d5d62edc47282a2e7ed9fa7c1/.github/workflows/scorecard-analysis.yml#L22) permission for the job (needed to access GitHub OIDC token). The API is available here: https://api.securityscorecards.dev.

### Scorecard Badge

Starting with scorecard-action:v2, users can add a Scorecard Badge to their README to display the latest status of their Scorecard results. This requires setting [`publish_results: true`](https://github.com/ossf/scorecard/blob/d13ba3f3355b958d5d62edc47282a2e7ed9fa7c1/.github/workflows/scorecard-analysis.yml#L39) for the action and enabling [`id-token: write`](https://github.com/ossf/scorecard/blob/d13ba3f3355b958d5d62edc47282a2e7ed9fa7c1/.github/workflows/scorecard-analysis.yml#L22) permission for the job (needed to access GitHub OIDC token). The badge is updated on every run of scorecard-action and points to the latest result. To add a badge to your README, copy and paste the below line, and replace the {owner} and {repo} parts.

```
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/{owner}/{repo}/badge)](https://securityscorecards.dev/viewer/?uri=github.com/{owner}/{repo})
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
| `repo_token` | no | PAT token with repository read access. Follow [these steps](/docs/authentication/fine-grained-auth-token.md) to create it. |
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

Note: if you disable this option, the results of the Scorecards Action run will be only available to people with write access or more. You can find the results on the Security tab scanning dashboard).

### Workflow Example

Please see our workflow from `ossf/scorecard` for an up-to-date example.
https://github.com/ossf/scorecard/blob/main/.github/workflows/scorecard-analysis.yml

## Reporting vulnerabilities

If you find a vulnerability, please report it to us!
See [SECURITY.md](./SECURITY.md) for more information.
