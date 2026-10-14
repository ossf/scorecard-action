# Scorecard's GitHub action
[![CodeQL](https://github.com/ossf/scorecard-action/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/ossf/scorecard-action/actions/workflows/codeql-analysis.yml)
[![codecov](https://codecov.io/gh/ossf/scorecard-action/branch/main/graph/badge.svg?token=MAXISWR53I)](https://codecov.io/gh/ossf/scorecard-action)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/ossf/scorecard-action/badge)](https://scorecard.dev/viewer/?uri=github.com/ossf/scorecard-action)

> Official GitHub Action for [OSSF Scorecard](https://github.com/ossf/scorecard).

The Scorecard GitHub Action is free for all public repositories. Private repositories are supported if they have [GitHub Advanced Security](https://docs.github.com/en/get-started/learning-about-github/about-github-advanced-security). Private repositories without GitHub Advanced Security can run Scorecard from the command line by following the [standard installation instructions](https://github.com/ossf/scorecard#using-scorecard-1).


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
1. From your GitHub project's main page, click “Security” in the top ribbon.

   ![image](/images/install01.png)

2. Select “Code scanning”.

   ![image](/images/install02.png)

3. Depending on whether the repository already has a scanning tool configured, you will see "Add tool" or "Configure scanning tool.

   a. If you see "Add tool" click it and skip step `b`.

   ![image](/images/install03.png)

   If you see "Configure scanning tool" proceed to step `b`:

   ![image](/images/configurescantool.png)

   b. After clicking "Configure scanning tool", click "Explore workflows" underneath the "Other tools" section:

   ![image](/images/exploreworkflow.png)

5. Choose the "OSSF Scorecard" from the list of workflows below, and then click “Configure”. (To find it faster type "OSSF" in searchbox.)

   ![image](/images/searchingossf.png)

6. Commit the changes. (Your button might say "Commit changes..." instead of "Start commit", it does the same thing.)

   ![image](/images/install05.png)

To add the workflow to many repositories at once, use the [installer](/docs/installer.md).

### Authentication with Fine-grained PAT (optional)
Scorecard can run successfully with the workflow's default `GITHUB_TOKEN`, which is our recommended approach.
However, Scorecard Action requires additional permissions if you use GitHub's classic [Branch Protection](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches) settings and want to see it reflected in your results.
You can read more about how to configure Scorecard Action for these cases [here](/docs/authentication/fine-grained-auth-token.md).

GitHub's new [Repository Rules](https://docs.github.com/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/about-rulesets) are accessible to Scorecard Action with the workflow's default `GITHUB_TOKEN`. 
We recommend new repositories use Repository Rules so they can be read with the default GitHub token. 
Repositories that already use classic Branch Protection and wish to see their results without an admin token should consider migrating to Repository Rules.

### Additional permissions for private repositories

When running Scorecard Action on **private repositories** with the default `GITHUB_TOKEN`, add these **job-level permissions** so Scorecard can query commits and detect configured SAST tools. Without them you may see errors like:

> `Resource not accessible by integration` (e.g., during GraphQL ListCommits)

```yaml
jobs:
  analysis:
    runs-on: ubuntu-latest
    permissions:
      # Required when publishing results (badge / API / code scanning)
      security-events: write
      id-token: write
      # Recommended reads for private repos to avoid GraphQL/SAST gaps
      contents: read
      issues: read
      pull-requests: read
      checks: read
```

## View Results

The workflow is preconfigured to run on every repository contribution. After making a code change, you can view the results for the change either through the Scorecard Badge, Code Scanning Alerts or GitHub Workflow Runs.

### REST API
Starting with scorecard-action:v2, users can use a REST API to query their latest run results. This requires setting [`publish_results: true`](https://github.com/ossf/scorecard/blob/d13ba3f3355b958d5d62edc47282a2e7ed9fa7c1/.github/workflows/scorecard-analysis.yml#L39) for the action and enabling [`id-token: write`](https://github.com/ossf/scorecard/blob/d13ba3f3355b958d5d62edc47282a2e7ed9fa7c1/.github/workflows/scorecard-analysis.yml#L22) permission for the job (needed to access GitHub OIDC token). The API is available here: https://api.scorecard.dev.

### Scorecard Badge

Starting with scorecard-action:v2, users can add a Scorecard Badge to their README to display the latest status of their Scorecard results. This requires setting [`publish_results: true`](https://github.com/ossf/scorecard/blob/d13ba3f3355b958d5d62edc47282a2e7ed9fa7c1/.github/workflows/scorecard-analysis.yml#L39) for the action and enabling [`id-token: write`](https://github.com/ossf/scorecard/blob/d13ba3f3355b958d5d62edc47282a2e7ed9fa7c1/.github/workflows/scorecard-analysis.yml#L22) permission for the job (needed to access GitHub OIDC token). The badge is updated on every run of scorecard-action and points to the latest result. To add a badge to your README, copy and paste the below line, and replace the {owner} and {repo} parts.

```
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/{owner}/{repo}/badge)](https://scorecard.dev/viewer/?uri=github.com/{owner}/{repo})
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

If the run succeeded but your results didn't update, check the job summary and the run's annotations.
A warning that the Scorecard API rejected the results usually means the workflow breaks one of the [workflow restrictions](#workflow-restrictions); the annotation includes the API's error message.
See [When publishing fails](#when-publishing-fails) to make a rejection fail the run.

## Manual Action Setup

If you prefer to manually set up the Scorecard GitHub Action, you will need to set up a [workflow file](https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions).

First, [create a new file](https://docs.github.com/en/repositories/working-with-files/managing-files/creating-new-files) in this location: `[yourrepo]/.github/workflows/scorecard.yml`. Then use the input values below.

### Inputs

| Name | Required | Description |
| ----- | -------- | ----------- |
| `results_file` | yes | The file that contains the results. |
| `results_format` | yes | The format in which to store the results [json \| sarif \| intoto]. For GitHub's scanning dashboard, select `sarif`. Select `intoto` to wrap the JSON results in an [in-toto attestation statement](https://github.com/in-toto/attestation/blob/main/spec/v1/statement.md). |
| `repo_token` | no | PAT token with repository read access. Follow [these steps](/docs/authentication/fine-grained-auth-token.md) to create it. |
| `publish_results` | recommended | This will allow you to display a badge on your repository to show off your hard work. See details [here](#publishing-results). |
| `file_mode` | no | The method to fetch files from the repository: `archive` or `git` (default `archive`). |
| `fail_on_publish_error` | no | Fail the step when the Scorecard API rejects published results (default `false`). See [Publishing Results](#publishing-results). |

### Outputs

| Name | Description |
| ---- | ----------- |
| `results_published` | `true` if the results were published to the Scorecard API, otherwise `false`. |
| `publication_status` | The HTTP status code of the final publish attempt, `error` if no response was received, or `skipped` if publishing was not attempted. |

### Publishing Results

The Scorecard team runs a weekly scan of public GitHub repositories in order to track
the overall security health of the open source ecosystem. The results of the scans are [publicly
available](https://github.com/ossf/scorecard#public-data).
Setting `publish_results: true` replaces the results of the team's weekly scans with your own scan results,
helping us scale by cutting down on repeated workflows and GitHub API requests.
This option is also needed to enable badges on the repository.

#### When publishing fails

The action retries publishing on network errors and on HTTP 408, 429 and 5xx responses.
It attempts other 4xx responses, such as a [workflow restriction](#workflow-restrictions) rejection, only once, because retrying won't change the result.

By default, a failed publish is reported as a warning annotation, and the step still succeeds.
The job summary always states whether the scan succeeded and whether the results were published.

Set `fail_on_publish_error: true` to fail the step when the Scorecard API rejects the results.
Network errors, 408, 429 and 5xx responses still produce only a warning, so an outage of the Scorecard API doesn't fail your workflow.
The results file is written before publishing, so add `if: always()` to the steps that upload it:

```yaml
      - name: Run analysis
        uses: ossf/scorecard-action@<full commit SHA> # v2.6.0 or later
        with:
          results_file: results.sarif
          results_format: sarif
          publish_results: true
          fail_on_publish_error: true

      - name: Upload artifact
        if: always()
        uses: actions/upload-artifact@<full commit SHA>
        with:
          name: SARIF file
          path: results.sarif
          retention-days: 5

      - name: Upload to code-scanning
        if: always()
        uses: github/codeql-action/upload-sarif@<full commit SHA>
        with:
          sarif_file: results.sarif
```

#### Using the outputs

The [`results_published` and `publication_status` outputs](#outputs) let later jobs act on the outcome.
When publishing, the [workflow restrictions](#workflow-restrictions) don't allow `run` steps in the Scorecard job, so read the outputs from a separate job:

```yaml
jobs:
  analysis:
    # ...
    outputs:
      results-published: ${{ steps.scorecard.outputs.results_published }}
      publication-status: ${{ steps.scorecard.outputs.publication_status }}
    steps:
      - name: Run analysis
        id: scorecard
        uses: ossf/scorecard-action@<full commit SHA> # v2.6.0 or later
        # ...

  report:
    needs: analysis
    if: always()
    runs-on: ubuntu-latest
    permissions: {}
    steps:
      - name: Report publication status
        env:
          RESULTS_PUBLISHED: ${{ needs.analysis.outputs.results-published }}
          PUBLICATION_STATUS: ${{ needs.analysis.outputs.publication-status }}
        run: |
          echo "Published: $RESULTS_PUBLISHED (status: $PUBLICATION_STATUS)"
```

### Workflow Restrictions

If [publishing results](#publishing-results), our API [enforces certain rules](https://github.com/ossf/scorecard-webapp/blob/9c2f66d5f6ff56ca4a4ac2fba6ec8dcc5379d31c/app/server/post_results.go#L184-L187) on the producing workflow, which may reject the results. With `fail_on_publish_error: true`, a rejection fails the Scorecard Action step. 
We understand that this is restrictive, but currently it's necessary to ensure the integrity of our API dataset, since GitHub workflow steps run in the same environment as the job they belong to.
If possible, we will work on making this feature more flexible so we can drop this requirement in the future.

#### Global workflow restrictions

* The workflow can't contain top level env vars or defaults.
* No workflow level write permissions.
* Only the job with `ossf/scorecard-action` can use `id-token: write` permissions.

#### Restrictions on the job containing `ossf/scorecard-action`
* No job level env vars or defaults
* No containers or services
* The job should run on one of the [Ubuntu hosted runners](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#choosing-github-hosted-runners).
* The steps running in this job must belong to this approved list of GitHub actions:
  * "actions/checkout"
  * "actions/upload-artifact"
  * "github/codeql-action/upload-sarif"
  * "ossf/scorecard-action"
  * "step-security/harden-runner"

### Uploading Artifacts
The Scorecard Action uses the [artifact uploader action](https://github.com/actions/upload-artifact) to upload results in SARIF format to the Actions tab. These results are available to anybody for five days after the run to help with debugging. To disable the upload, comment out the `Upload Artifact` value in the Workflow Example.

Note: if you disable this option, the results of the Scorecard Action run will be only available to people with write access or more. You can find the results on the Security tab scanning dashboard).

### Workflow Linters

The [example workflows](/examples) set `cache-mode: none`, which blocks all cache access.
GitHub Actions added the `cache-mode` key in September 2026, so linters released before then may not recognize it:

- actionlint v1.7.12 and earlier report `unexpected key "cache-mode"`.
  On those versions, `actionlint -ignore 'unexpected key "cache-mode"'` suppresses only this finding; other unknown keys are still reported.
- actionlint.kjanat.dev v1.17.0 and later, zizmor, and yamllint accept it.

### Workflow Example

Please see our workflow from `ossf/scorecard` for an up-to-date example.
https://github.com/ossf/scorecard/blob/main/.github/workflows/scorecard-analysis.yml

## Reporting vulnerabilities

If you find a vulnerability, please report it to us!
See [SECURITY.md](./SECURITY.md) for more information.
