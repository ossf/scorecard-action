# Scorecards' GitHub action

> Official GitHub Action for [OSSF scorecard](https://github.com/ossf/scorecard).

## Installation
The simplest and quickest way to install Scorecards's GitHub action is from the [GitHub's marketplace](https://github.com/marketplace/actions/ossf-scorecard-action).

### Inputs

| Name | Required | Description |
| ----- | -------- | ----------- |
| `result_file` | yes | The file that contains the results. |
| `result_format` | yes | The format in which to store the results [json \| sarif]. For GitHub's scanning dashboard, select `sarif`. |
| `repo_token` | yes | PAT token with read-only access. Follow [these steps](#pat-token-creation) to create it. |
| `publish_results` | recommended | This will allow you to display a badge on your repository to show off your hard work (release scheduled for Q2'22). See details [here](#publishing-results).|

### PAT token creation
1. Create a PAT token [here](https://github.com/settings/tokens/new) with the following read permissions:
    - Note: `Read-only token for OSSF Scorecard Action`
    - Expiration: `No expiration`
    - Scopes: 
        * `repo > public_repo`
        * `admin:org > read:org`
        * `admin:repo_hook > read:repo_hook`
        * `write:discussion > read:discussion`
    - Create and copy the token.

2. Create a new repository secret at `https://github.com/<org>/<repo>/settings/secrets/actions/new` with the following settings:
    - Name: `SCORECARD_TOKEN`
    - Value: the value of the token created in step 1 above.

### Publishing results
The Scorecard team runs a weekly scan of public GitHub repositories in order to track 
the overall security health of the open source ecosystem. The results of the scans are publicly
available as described [here](https://github.com/ossf/scorecard#public-data).
Setting `publish_results: true` replaces the results of the team's weelky scans, 
helping us scale by cutting down on repeated workflows and GitHub API requests.
This option is needed to enable badges on the repo (release scheduled for Q2'22). 
By default, `publish_results: ${{ !github.event.repository.private }}` so not data is shared for private repositories.

### Full example

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
    
    steps:
      - name: "Checkout code"
        uses: actions/checkout@ec3a7ce113134d7a93b817d10a8272cb61118579 # v2.4.0
        with:
          persist-credentials: false

      - name: "Run analysis"
        uses: ossf/scorecard-action@59f9117686133e93b60a8f23131f87089a076e1b
        with:
          results_file: results.sarif
          results_format: sarif
          # Read-only PAT token. To create it,
          # follow the steps in https://github.com/ossf/scorecard-action#pat-token-creation.
          repo_token: ${{ secrets.SCORECARD_TOKEN }}
          # Publish the results to enable scorecard badges. For more details, see
          # https://github.com/ossf/scorecard-action#publishing-results.
          # The line below disables data sharing for private repositories.
          publish_results: ${{ !github.event.repository.private }}

      # Upload the results as artifacts (optional).
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
