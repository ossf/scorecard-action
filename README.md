# Scorecards' GitHub action
[![CodeQL](https://github.com/ossf/scorecard-action/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/ossf/scorecard-action/actions/workflows/codeql-analysis.yml)
[![codecov](https://codecov.io/gh/ossf/scorecard-action/branch/main/graph/badge.svg?token=MAXISWR53I)](https://codecov.io/gh/ossf/scorecard-action)
> Official GitHub Action for [OSSF Scorecards](https://github.com/ossf/scorecard).

The Scorecards GitHub Action is free for all public repositories. Private repositories are supported if they have [GitHub Advanced Security](https://docs.github.com/en/get-started/learning-about-github/about-github-advanced-security). Private repositories without GitHub Advanced Security can run Scorecards from the command line by following the [standard installation instructions](https://github.com/ossf/scorecard#using-scorecards-1).

________
[Installation](#installation) 
- [Authentication](#authentication-with-pat)
- [Workflow Setup](#workflow-setup)

[View Results](#view-results)
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

## Installation
The Scorecards Action is installed by setting up a workflow on the GitHub UI.

Note: One Scorecards check ([Branch-Protection](https://github.com/ossf/scorecard/blob/main/docs/checks.md#branch-protection)) requires authentication using a Personal Access Token (PAT). If you want all Scorecards checks to run, you will need to follow the optional Authentication step. If you don't, all checks will run except Branch-Protection.

Optional Authentication: Create a Personal Access Token (PAT) for authentication and save the token value as a repository secret; 
    
    (Note: If you have already installed Scorecards on your repository from the command line, you can reuse your existing PAT for the repository secret. If you no longer have access to the PAT, though, simply create a new one.)

**Required**: Set up the workflow via the GitHub UI - see [Workflow Setup](#workflow-setup)

### Authentication with PAT
1. [Create a Personal Access Token](https://github.com/settings/tokens/new?scopes=public_repo,read:org,read:repo_hook,read:discussion) with the following read permissions:
    - Note: `Read-only token for OSSF Scorecard Action - myorg/myrepo` (Note: replace `myorg/myrepo` with the names of your organization and repository so you can keep track of your tokens.)
    - Expiration: `No expiration`
    - Scopes: 
        * `repo > public_repo`                  Required to read [Branch-Protection](https://github.com/ossf/scorecard/blob/main/docs/checks.md#branch-protection) settings. **Warning**: for private repositories, you need scope `repo`.
        * `admin:org > read:org`                Optional: not used in current implementation.
        * `admin:repo_hook > read:repo_hook`    Optional: needed for the experimental [Webhook](https://github.com/ossf/scorecard/blob/main/docs/checks.md#webhooks) check.
        * `write:discussion > read:discussion`  Optional: not used in current implementation.

![image](/images/tokenscopes.png)

2. Copy the token value. 

3. [Create a new repository secret](https://docs.github.com/en/actions/security-guides/encrypted-secrets#creating-encrypted-secrets-for-a-repository) with the following settings:
    - Name: `SCORECARD_READ_TOKEN`
    - Value: the value of the token created in step 1 above.

4. (Optional) If you install Scorecard on a repository owned by an organization that uses [SAML SSO](https://docs.github.com/en/enterprise-cloud@latest/authentication/authenticating-with-saml-single-sign-on/about-authentication-with-saml-single-sign-on), be sure to [enable SSO](https://docs.github.com/en/enterprise-cloud@latest/authentication/authenticating-with-saml-single-sign-on/authorizing-a-personal-access-token-for-use-with-saml-single-sign-on) for your PAT token.

### Workflow Setup
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

## View Results

The workflow is preconfigured to run on every repository contribution. After making a code change, you can view a list of results by going to the Security tab and clicking "Code Scanning Alerts" (it can take a couple minutes for the run to complete and the results to show up). Click on the individual alerts for more information, including remediation instructions. You will need to click "Show more" to expand the full remediation instructions.

![image](/images/remediation.png)

### Verify Runs 
The workflow is preconfigured to run on every repository contribution. 

To verify that the Action is running successfully, click the repository's Actions tab to see the status of all recent workflow runs. This tab will also show the logs, which can help you troubleshoot if the run failed.

![image](/images/actionconfirm.png)

### Troubleshooting 
If the run has failed, the most likely reason is an authentication failure. Confirm that the Personal Access Token is saved as an encrypted secret within the same repository (see [Authentication](#authentication)). 

If you install Scorecard on a private repository with a PAT token, provide the `repo` scope. (The `repo > public_repo` scope only provides access to public repositories.)

If you install Scorecard on a repository owned by an organization that uses [SAML SSO](https://docs.github.com/en/enterprise-cloud@latest/authentication/authenticating-with-saml-single-sign-on/about-authentication-with-saml-single-sign-on) or if you see `403 Resource protected by organization SAML enforcement` in the logs, be sure to [enable SSO](https://docs.github.com/en/enterprise-cloud@latest/authentication/authenticating-with-saml-single-sign-on/authorizing-a-personal-access-token-for-use-with-saml-single-sign-on) for your PAT token (see [Authentication](#authentication)).

If the PAT is saved as an encrypted secret and the run is still failing, confirm that you have not made any changes to the workflow yaml file that affected the syntax. Review the [workflow example](#workflow-example) and reset to the default values if necessary.

## Manual Action Setup
    
If you prefer to manually set up the Scorecards GitHub Action, you will need to set up a [workflow file](https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions).

First, [create a new file](https://docs.github.com/en/repositories/working-with-files/managing-files/creating-new-files) in this location: `[yourrepo]/.github/workflows/scorecards-analysis.yml`. Then use the input values below.
 

### Inputs

| Name | Required | Description |
| ----- | -------- | ----------- |
| `result_file` | yes | The file that contains the results. |
| `result_format` | yes | The format in which to store the results [json \| sarif]. For GitHub's scanning dashboard, select `sarif`. |
| `repo_token` | yes | PAT token with read-only access. Follow [these steps](#pat-token-creation) to create it. |
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
      actions: read
      contents: read
    
    steps:
      - name: "Checkout code"
        uses: actions/checkout@ec3a7ce113134d7a93b817d10a8272cb61118579 # v2.4.0
        with:
          persist-credentials: false

      - name: "Run analysis"
        uses: ossf/scorecard-action@c1aec4ac820532bab364f02a81873c555a0ba3a1 # v1.0.4
        with:
          results_file: results.sarif
          results_format: sarif
          # (Optional) Read-only PAT token for Branch-Protection check. To create it,
          # follow the steps in https://github.com/ossf/scorecard-action#pat-token-creation.
          repo_token: ${{ secrets.SCORECARD_READ_TOKEN }}
          # Publish the results for public repositories to enable scorecard badges. For more details, see
          # https://github.com/ossf/scorecard-action#publishing-results. 
          # For private repositories, `publish_results` will automatically be set to `false`, regardless 
          # of the value entered here.
          publish_results: true

      # Upload the results as artifacts (optional). Commenting out will disable uploads of run results in SARIF
      # format to the repository Actions tab.
      - name: "Upload artifact"
        uses: actions/upload-artifact@82c141cc518b40d92cc801eee768e7aafc9c2fa2 # v2.3.1
        with:
          name: SARIF file
          path: results.sarif
          retention-days: 5
      
      # Upload the results to GitHub's code scanning dashboard.
      - name: "Upload to code-scanning"
        uses: github/codeql-action/upload-sarif@5f532563584d71fdef14ee64d17bafb34f751ce5 # v1.0.26
        with:
          sarif_file: results.sarif
```
