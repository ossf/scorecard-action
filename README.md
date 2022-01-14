# Scorecards' GitHub action
[![CodeQL](https://github.com/ossf/scorecard-action/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/ossf/scorecard-action/actions/workflows/codeql-analysis.yml)
> Official GitHub Action for [OSSF Scorecards](https://github.com/ossf/scorecard).

The Scorecards GitHub Action is free for all public repositories. Private repositories are supported if they have [GitHub Advanced Security](https://docs.github.com/en/get-started/learning-about-github/about-github-advanced-security). Private repositories without GitHub Advanced Security can run Scorecards from the command line by following the [standard installation instructions](https://github.com/ossf/scorecard#using-scorecards-1).

________
[Installation](#installation) 
- [Authentication](#authentication)
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

## Installation
To install the Scorecards GitHub Action, you need to:

1) Create a Personal Access Token (PAT) for authentication and save the token value as a repository secret; 
    
    (Note: If you have already installed Scorecards on your repository from the command line, you can reuse your existing PAT for the repository secret. If you no longer have access to the PAT, though, simply create a new one.)
    
3) Set up the workflow via the GitHub UI

### Authentication
1. [Create a Personal Access Token](https://github.com/settings/tokens/new) with the following read permissions:
    - Note: `Read-only token for OSSF Scorecard Action - myorg/myrepo` (Note: replace `myorg/myrepo` with the names of your organization and repository so you can keep track of your tokens.)
    - Expiration: `No expiration`
    - Scopes: 
        * `repo > public_repo`
        * `admin:org > read:org`
        * `admin:repo_hook > read:repo_hook`
        * `write:discussion > read:discussion`

![image](/images/tokenscopes.png)
     
2. Copy the token value. 

3. [Create a new repository secret](https://docs.github.com/en/actions/security-guides/encrypted-secrets#creating-encrypted-secrets-for-a-repository) with the following settings:
    - Name: `SCORECARD_READ_TOKEN`
    - Value: the value of the token created in step 1 above.

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

To view a list of results from each Scorecards Action run, go to the Security tab and click "Code Scanning Alerts." Click on the individual alerts for more information, including remediation instructions. You will need to click "Show more" to expand the full remediation instructions.

![image](/images/remediation.png)

### Verify Runs 
The workflow is preconfigured to run on every repository contribution. 

To verify that the Action is running successfully, click the repository's Actions tab to see the status of all recent workflow runs. 

![image](/images/actionconfirm.png)

### Troubleshooting 
If the run has failed, the most likely reason is an authentication failure. Confirm that the Personal Access Token is saved as an encrypted secret within the same repository (see [Authentication](#authentication)). 

If the PAT is saved as an encrypted secret and the run is still failing, confirm that you have not made any changes to the workflow yaml file that affected the syntax. Review the [workflow example](#workflow-example) and reset to the default values if necessary.  

## Manual Action Setup
    
If you prefer to manually set up the Scorecards GitHub Action, use the following values in your [workflow file](https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions). 

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
        uses: ossf/scorecard-action@f5a7da46837397de5331ea22ce0099e2bfe265d0 # v1.0.1
        with:
          results_file: results.sarif
          results_format: sarif
          # Read-only PAT token. To create it,
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
