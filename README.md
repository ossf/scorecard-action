# Scorecards' GitHub action
[![CodeQL](https://github.com/ossf/scorecard-action/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/ossf/scorecard-action/actions/workflows/codeql-analysis.yml)
[![codecov](https://codecov.io/gh/ossf/scorecard-action/branch/main/graph/badge.svg?token=MAXISWR53I)](https://codecov.io/gh/ossf/scorecard-action)
> Official GitHub Action for [OSSF Scorecards](https://github.com/ossf/scorecard).

The Scorecards GitHub Action is free for all public repositories. Private repositories are supported if they have [GitHub Advanced Security](https://docs.github.com/en/get-started/learning-about-github/about-github-advanced-security). Private repositories without GitHub Advanced Security can run Scorecards from the command line by following the [standard installation instructions](https://github.com/ossf/scorecard#using-scorecards-1).


## Breaking changes in v2

Starting from scorecard-action:v2, `GITHUB_TOKEN` permissions or job permissions needs to include
`id-token: write` for `publish_results: true`. This is needed to access GitHub's
OIDC token which verifies the authenticity of the result when publishing it.

scorecard-action:v2 has a new requirement for the job running the ossf/scorecard-action step. The steps running in this job must belong to this approved list of GitHub actions: 
- "actions/checkout" 
- "actions/upload-artifact"
- "github/codeql-action/upload-sarif"
- "ossf/scorecard-action"

If you are using custom steps in the job, it may fail.
We understand that this is restrictive, but currently it's necessary to ensure the integrity of the results that we publish, since GitHub workflow steps run in the same environment as the job they belong to. 
If possible, we will work on making this feature more flexible so we can drop this requirement in the future.  
________
[Personal Access Token (PAT) Requirements and Risks](#personal-access-token-pat-requirements-and-risks)

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
- [Uploading Artifacts](#uploading-artifacts)
- [Workflow Example](#workflow-example)
________

The following GitHub triggers are supported: `push`, `schedule` (default branch only).

The `pull_request` and `workflow_dispatch` triggers are experimental.

Running the Scorecard action on a fork repository is not supported.

GitHub Enterprise repositories are not supported.

## Personal Access Token (PAT) Requirements and Risks

Certain features require a Personal Access Token (PAT). 

-  Public repositories need a PAT to enable the
    [Branch-Protection](https://github.com/ossf/scorecard/blob/main/docs/checks.md#branch-protection)
    check. Without a PAT, Scorecards will run all checks except the
    Branch-Protection check
-  Private repositories need a PAT to use any Scorecard Action functions

Using a PAT introduces risks, however. Due to a limitation of the GitHub
permission model, the PAT needs
[write permission to the repository](https://docs.github.com/en/developers/apps/building-oauth-apps/scopes-for-oauth-apps#available-scopes)
through the `repo` scope. **The PAT will be stored as a
[GitHub encrypted secret](https://docs.github.com/en/actions/security-guides/encrypted-secrets)
and be accessible by all the workflows and maintainers of a repository.**
This means another maintainer on your project could potentially use the token to impersonate you. If there is an exploitable bug in a workflow with write permissions, an external contributor could potentially exploit it to extract the PAT.

We recommend that you **do not use a PAT** unless you feel that the
risks introduced are outweighed by the functionalities they support. 

## Installation

### Workflow Setup (Required)
1) From your GitHub project's main page, click “Security” in the top ribbon. 

![image](/images/install01.png)

2) Click “Set up Code Scanning.” 

![image](/images/install02.png)

Note: if you have already configured other code scanning tools, your UI will look different than shown above. Instead, click "Code Scanning Alerts" on the left side of the page. 

![image](/images/installb1.png)

Then click "Add More Scanning Tools."

![image](/images/installb2.png)

3) Choose the "OSSF Scorecards supply-chain security analysis" from the list of workflows, and then click “set up this workflow.”

![image](/images/install03.png)

4) Commit the changes.

![image](/images/install04.png)

### Authentication with PAT (optional)
Create a Personal Access Token (PAT) for authentication and save the token value as a repository secret. Review [Personal Access Token (PAT) Requirements and Risks](#personal-access-token-pat-requirements-and-risks) before using a PAT.  

1. [Create a Personal Access Token](https://github.com/settings/tokens/new?scopes=public_repo,read:org,read:repo_hook,read:discussion) with the following read permissions:
    - Note: `Token for OSSF Scorecard Action - myorg/myrepo` (Note: replace `myorg/myrepo` with the names of your organization and repository so you can keep track of your tokens.)
    - Expiration: `No expiration`
    - Scopes: 
        * `repo > public_repo`                  Required to read [Branch-Protection](https://github.com/ossf/scorecard/blob/main/docs/checks.md#branch-protection) settings. **Note**: for private repositories, you need scope `repo`.
        * `admin:org > read:org`                Optional: not used in current implementation.
        * `admin:repo_hook > read:repo_hook`    Optional: needed for the experimental [Webhook](https://github.com/ossf/scorecard/blob/main/docs/checks.md#webhooks) check.
        * `write:discussion > read:discussion`  Optional: not used in current implementation.

![image](/images/tokenscopes.png)

2. Copy the token value. 

3. [Create a new repository secret](https://docs.github.com/en/actions/security-guides/encrypted-secrets#creating-encrypted-secrets-for-a-repository) with the following settings (**Warning:** [GitHub encrypted secrets](https://docs.github.com/en/actions/security-guides/encrypted-secrets) are accessible by all the workflows and maintainers of a repository.):
    - Name: `SCORECARD_TOKEN`
    - Value: the value of the token created in step 1 above.

4. (Optional) If you install Scorecard on a repository owned by an organization that uses [SAML SSO](https://docs.github.com/en/enterprise-cloud@latest/authentication/authenticating-with-saml-single-sign-on/about-authentication-with-saml-single-sign-on), be sure to [enable SSO](https://docs.github.com/en/enterprise-cloud@latest/authentication/authenticating-with-saml-single-sign-on/authorizing-a-personal-access-token-for-use-with-saml-single-sign-on) for your PAT token.

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
If the run has failed, the most likely reason is an authentication failure. If you are running Scorecards on a private repository, confirm that the Personal Access Token is saved as an encrypted secret within the same repository (see [Authentication](#authentication)). In addition, provide the `repo` scope to your PAT. (The `repo > public_repo` scope only provides access to public repositories).

If you install Scorecards on a repository owned by an organization that uses [SAML SSO](https://docs.github.com/en/enterprise-cloud@latest/authentication/authenticating-with-saml-single-sign-on/about-authentication-with-saml-single-sign-on) or if you see `403 Resource protected by organization SAML enforcement` in the logs, be sure to [enable SSO](https://docs.github.com/en/enterprise-cloud@latest/authentication/authenticating-with-saml-single-sign-on/authorizing-a-personal-access-token-for-use-with-saml-single-sign-on) for your PAT token (see [Authentication](#authentication)).

If you use a PAT saved as an encrypted secret and the run is still failing, confirm that you have not made any changes to the workflow yaml file that affected the syntax. Review the [workflow example](#workflow-example) and reset to the default values if necessary.

## Manual Action Setup
    
If you prefer to manually set up the Scorecards GitHub Action, you will need to set up a [workflow file](https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions).

First, [create a new file](https://docs.github.com/en/repositories/working-with-files/managing-files/creating-new-files) in this location: `[yourrepo]/.github/workflows/scorecards.yml`. Then use the input values below.
 

### Inputs

| Name | Required | Description |
| ----- | -------- | ----------- |
| `result_file` | yes | The file that contains the results. |
| `result_format` | yes | The format in which to store the results [json \| sarif]. For GitHub's scanning dashboard, select `sarif`. |
| `repo_token` | no | PAT token with write repository access. Follow [these steps](#authentication-with-pat) to create it. |
| `publish_results` | recommended | This will allow you to display a badge on your repository to show off your hard work (release scheduled for Q2'22). See details [here](#publishing-results).|

### Publishing Results
The Scorecard team runs a weekly scan of public GitHub repositories in order to track 
the overall security health of the open source ecosystem. The results of the scans are [publicly
available](https://github.com/ossf/scorecard#public-data).
Setting `publish_results: true` replaces the results of the team's weekly scans with your own scan results, 
helping us scale by cutting down on repeated workflows and GitHub API requests.
This option is also needed to enable badges on the repository (release scheduled for Q2'22). 

### Uploading Artifacts
The Scorecards Action uses the [artifact uploader action](https://github.com/actions/upload-artifact) to upload results in SARIF format to the Actions tab. These results are available to anybody for five days after the run to help with debugging. To disable the upload, comment out the `Upload Artifact` value in the Workflow Example. 

Note: if you disable this option, the results of the Scorecards Action run will be available only to maintainers (on the Security tab scanning dashboard). 

### Workflow Example

```yml
name: Scorecards supply-chain security
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
    name: Scorecards analysis
    runs-on: ubuntu-latest
    permissions:
      # Needed to upload the results to code-scanning dashboard.
      security-events: write
      # Used to receive a badge. (Upcoming feature)
      id-token: write
      actions: read
      contents: read

    steps:
      - name: "Checkout code"
        uses: actions/checkout@a12a3943b4bdde767164f792f33f40b04645d846 # tag=v3.0.0
        with:
          persist-credentials: false

      - name: "Run analysis"
        uses: ossf/scorecard-action@3e15ea8318eee9b333819ec77a36aca8d39df13e # tag=v1.1.1
        with:
          results_file: results.sarif
          results_format: sarif
          # (Optional) "write" PAT token. Uncomment the `repo_token` line below if:
          # - you want to enable the Branch-Protection check on a *public* repository, or
          # - you are installing Scorecards on a *private* repository
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
        uses: actions/upload-artifact@6673cd052c4cd6fcf4b4e6e60ea986c889389535 # tag=v3.0.0
        with:
          name: SARIF file
          path: results.sarif
          retention-days: 5

      # Upload the results to GitHub's code scanning dashboard.
      - name: "Upload to code-scanning"
        uses: github/codeql-action/upload-sarif@5f532563584d71fdef14ee64d17bafb34f751ce5 # tag=v1.0.26
        with:
          sarif_file: results.sarif
```
