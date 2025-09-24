# Developing

## Container images

### Building

(This presumes `docker` is installed on your workstation.)

Let's also assume the image to be built is `scorecard-action:testing`

From the root of the repository:

```shell
docker build -t scorecard-action:testing .
```

### Testing

To test the image, there are a few environment variables that will need to be
set to mimic running in a GitHub Actions environment.

First, set the GitHub authentication token:

```shell
export GITHUB_AUTH_TOKEN="<token>"
```

Now, run the container:

```shell
docker run -e INPUT_REPO_TOKEN="$GITHUB_AUTH_TOKEN" \
           -e INPUT_POLICY_FILE="/policy.yml" \
           -e INPUT_RESULTS_FORMAT="sarif" \
           -e INPUT_RESULTS_FILE="results.sarif" \
           -e INPUT_PUBLISH_RESULTS="false" \
           -e GITHUB_WORKSPACE="/" \
           -e GITHUB_REF="refs/heads/main" \
           -e GITHUB_EVENT_NAME="branch_protection_rule" \
           -e GITHUB_EVENT_PATH="/testdata/fork.json" \
           -e GITHUB_REPOSITORY="ossf/scorecard" \
           --mount type=bind,source=./options/testdata/fork.json,destination=/testdata/fork.json,readonly \
           scorecard-action:testing
```

Here is some example output, if the command has executed successfully (in
separate blocks for syntax highlighting):

```console
Event file: /testdata/fork.json
Event name: branch_protection_rule
Ref: refs/heads/main
Repository: ossf/scorecard
Fork repository: false
Private repository: false
Publication enabled: false
Format: sarif
Policy file: /policy.yml
Default branch: refs/heads/main
```

<details>

```json
{
  "id": 302670797,
  "node_id": "MDEwOlJlcG9zaXRvcnkzMDI2NzA3OTc=",
  "name": "scorecard",
  "full_name": "ossf/scorecard",
  "private": false,
  "owner": {
    "login": "ossf",
    "id": 67707773,
    "node_id": "MDEyOk9yZ2FuaXphdGlvbjY3NzA3Nzcz",
    "avatar_url": "https://avatars.githubusercontent.com/u/67707773?v=4",
    "gravatar_id": "",
    "url": "https://api.github.com/users/ossf",
    "html_url": "https://github.com/ossf",
    "followers_url": "https://api.github.com/users/ossf/followers",
    "following_url": "https://api.github.com/users/ossf/following{/other_user}",
    "gists_url": "https://api.github.com/users/ossf/gists{/gist_id}",
    "starred_url": "https://api.github.com/users/ossf/starred{/owner}{/repo}",
    "subscriptions_url": "https://api.github.com/users/ossf/subscriptions",
    "organizations_url": "https://api.github.com/users/ossf/orgs",
    "repos_url": "https://api.github.com/users/ossf/repos",
    "events_url": "https://api.github.com/users/ossf/events{/privacy}",
    "received_events_url": "https://api.github.com/users/ossf/received_events",
    "type": "Organization",
    "site_admin": false
  },
  "html_url": "https://github.com/ossf/scorecard",
  "description": "Security Scorecards - Security health metrics for Open Source",
  "fork": false,
  "url": "https://api.github.com/repos/ossf/scorecard",
  "forks_url": "https://api.github.com/repos/ossf/scorecard/forks",
  "keys_url": "https://api.github.com/repos/ossf/scorecard/keys{/key_id}",
  "collaborators_url": "https://api.github.com/repos/ossf/scorecard/collaborators{/collaborator}",
  "teams_url": "https://api.github.com/repos/ossf/scorecard/teams",
  "hooks_url": "https://api.github.com/repos/ossf/scorecard/hooks",
  "issue_events_url": "https://api.github.com/repos/ossf/scorecard/issues/events{/number}",
  "events_url": "https://api.github.com/repos/ossf/scorecard/events",
  "assignees_url": "https://api.github.com/repos/ossf/scorecard/assignees{/user}",
  "branches_url": "https://api.github.com/repos/ossf/scorecard/branches{/branch}",
  "tags_url": "https://api.github.com/repos/ossf/scorecard/tags",
  "blobs_url": "https://api.github.com/repos/ossf/scorecard/git/blobs{/sha}",
  "git_tags_url": "https://api.github.com/repos/ossf/scorecard/git/tags{/sha}",
  "git_refs_url": "https://api.github.com/repos/ossf/scorecard/git/refs{/sha}",
  "trees_url": "https://api.github.com/repos/ossf/scorecard/git/trees{/sha}",
  "statuses_url": "https://api.github.com/repos/ossf/scorecard/statuses/{sha}",
  "languages_url": "https://api.github.com/repos/ossf/scorecard/languages",
  "stargazers_url": "https://api.github.com/repos/ossf/scorecard/stargazers",
  "contributors_url": "https://api.github.com/repos/ossf/scorecard/contributors",
  "subscribers_url": "https://api.github.com/repos/ossf/scorecard/subscribers",
  "subscription_url": "https://api.github.com/repos/ossf/scorecard/subscription",
  "commits_url": "https://api.github.com/repos/ossf/scorecard/commits{/sha}",
  "git_commits_url": "https://api.github.com/repos/ossf/scorecard/git/commits{/sha}",
  "comments_url": "https://api.github.com/repos/ossf/scorecard/comments{/number}",
  "issue_comment_url": "https://api.github.com/repos/ossf/scorecard/issues/comments{/number}",
  "contents_url": "https://api.github.com/repos/ossf/scorecard/contents/{+path}",
  "compare_url": "https://api.github.com/repos/ossf/scorecard/compare/{base}...{head}",
  "merges_url": "https://api.github.com/repos/ossf/scorecard/merges",
  "archive_url": "https://api.github.com/repos/ossf/scorecard/{archive_format}{/ref}",
  "downloads_url": "https://api.github.com/repos/ossf/scorecard/downloads",
  "issues_url": "https://api.github.com/repos/ossf/scorecard/issues{/number}",
  "pulls_url": "https://api.github.com/repos/ossf/scorecard/pulls{/number}",
  "milestones_url": "https://api.github.com/repos/ossf/scorecard/milestones{/number}",
  "notifications_url": "https://api.github.com/repos/ossf/scorecard/notifications{?since,all,participating}",
  "labels_url": "https://api.github.com/repos/ossf/scorecard/labels{/name}",
  "releases_url": "https://api.github.com/repos/ossf/scorecard/releases{/id}",
  "deployments_url": "https://api.github.com/repos/ossf/scorecard/deployments",
  "created_at": "2020-10-09T14:48:27Z",
  "updated_at": "2022-05-24T15:31:47Z",
  "pushed_at": "2022-05-25T00:38:14Z",
  "git_url": "git://github.com/ossf/scorecard.git",
  "ssh_url": "git@github.com:ossf/scorecard.git",
  "clone_url": "https://github.com/ossf/scorecard.git",
  "svn_url": "https://github.com/ossf/scorecard",
  "homepage": "",
  "size": 48157,
  "stargazers_count": 2643,
  "watchers_count": 2643,
  "language": "Go",
  "has_issues": true,
  "has_projects": true,
  "has_downloads": true,
  "has_wiki": true,
  "has_pages": false,
  "forks_count": 249,
  "mirror_url": null,
  "archived": false,
  "disabled": false,
  "open_issues_count": 203,
  "license": {
    "key": "apache-2.0",
    "name": "Apache License 2.0",
    "spdx_id": "Apache-2.0",
    "url": "https://api.github.com/licenses/apache-2.0",
    "node_id": "MDc6TGljZW5zZTI="
  },
  "allow_forking": true,
  "is_template": false,
  "topics": [
    "scorecard",
    "security-scorecards"
  ],
  "visibility": "public",
  "forks": 249,
  "open_issues": 203,
  "watchers": 2643,
  "default_branch": "main",
  "permissions": {
    "admin": true,
    "maintain": true,
    "push": true,
    "triage": true,
    "pull": true
  },
  "temp_clone_token": "",
  "allow_squash_merge": true,
  "allow_merge_commit": false,
  "allow_rebase_merge": true,
  "allow_auto_merge": true,
  "delete_branch_on_merge": true,
  "allow_update_branch": false,
  "organization": {
    "login": "ossf",
    "id": 67707773,
    "node_id": "MDEyOk9yZ2FuaXphdGlvbjY3NzA3Nzcz",
    "avatar_url": "https://avatars.githubusercontent.com/u/67707773?v=4",
    "gravatar_id": "",
    "url": "https://api.github.com/users/ossf",
    "html_url": "https://github.com/ossf",
    "followers_url": "https://api.github.com/users/ossf/followers",
    "following_url": "https://api.github.com/users/ossf/following{/other_user}",
    "gists_url": "https://api.github.com/users/ossf/gists{/gist_id}",
    "starred_url": "https://api.github.com/users/ossf/starred{/owner}{/repo}",
    "subscriptions_url": "https://api.github.com/users/ossf/subscriptions",
    "organizations_url": "https://api.github.com/users/ossf/orgs",
    "repos_url": "https://api.github.com/users/ossf/repos",
    "events_url": "https://api.github.com/users/ossf/events{/privacy}",
    "received_events_url": "https://api.github.com/users/ossf/received_events",
    "type": "Organization",
    "site_admin": false
  },
  "network_count": 249,
  "subscribers_count": 50
}
{
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "version": "2.1.0",
  "runs": [
    {
      "automationDetails": {
        "id": "supply-chain/branch-protection/unknown-25 May 22 00:50 +0000"
      },
      "tool": {
        "driver": {
          "name": "Scorecard",
          "informationUri": "https://github.com/ossf/scorecard",
          "semanticVersion": "unknown",
          "rules": [
            {
              "id": "BranchProtectionID",
              "name": "Branch-Protection",
              "helpUri": "https://github.com/ossf/scorecard/blob/main/docs/checks.md#branch-protection",
              "shortDescription": {
                "text": "Branch-Protection"
              },
              "fullDescription": {
                "text": "Determines if the default and release branches are protected with GitHub's branch protection settings."
              },
              "help": {
                "text": "Determines if the default and release branches are protected with GitHub's branch protection settings.",
                "markdown": "**Remediation (click \"Show more\" below)**:\n\n- Enable branch protection settings in your source hosting provider to avoid force pushes or deletion of your important branches.\n\n- For GitHub, check out the steps [here](https://docs.github.com/en/github/administering-a-repository/managing-a-branch-protection-rule).\n\n\n\n**Severity**: High\n\n\n\n**Details**:\n\nRisk: `High` (vulnerable to intentional malicious code injection)  \n\n\n\nThis check determines whether a project's default and release branches are\n\nprotected with GitHub's [branch protection](https://docs.github.com/en/github/administering-a-repository/defining-the-mergeability-of-pull-requests/about-protected-branches) settings. \n\nBranch protection allows maintainers to define rules that enforce\n\ncertain workflows for branches, such as requiring review or passing certain\n\nstatus checks before acceptance into a main branch, or preventing rewriting of\n\npublic history.\n\n\n\nNote: The following settings queried by the Branch-Protection check require an admin token: `DismissStaleReviews`, `EnforceAdmin`, and `StrictStatusCheck`. If\n\nthe provided token does not have admin access, the check will query the branch\n\nsettings accessible to non-admins and provide results based only on these settings.\n\nEven so, we recommend using a non-admin token, which provides a thorough enough\n\nresult to meet most user needs. \n\n\n\nDifferent types of branch protection protect against different risks:\n\n\n\n  - Require code review: requires at least one reviewer, which greatly\n\n    reduces the risk that a compromised contributor can inject malicious code.\n\n    Review also increases the likelihood that an unintentional vulnerability in\n\n    a contribution will be detected and fixed before the change is accepted.\n\n\n\n  - Prevent force push: prevents use of the `--force` command on public\n\n    branches, which overwrites code irrevocably. This protection prevents the\n\n    rewriting of public history without external notice.\n\n\n\n  - Require [status checks](https://docs.github.com/en/github/collaborating-with-pull-requests/collaborating-on-repositories-with-code-quality-features/about-status-checks):\n\n    ensures that all required CI tests are met before a change is accepted. \n\n\n\nAlthough requiring code review can greatly reduce the chance that\n\nunintentional or malicious code enters the \"main\" branch, it is not feasible for\n\nall projects, such as those that don't have many active participants. For more\n\ndiscussion, see [Code Reviews](https://github.com/ossf/scorecard/blob/main/docs/checks.md#code-reviews).\n\n\n\nAdditionally, in some cases these rules will need to be suspended. For example,\n\nif a past commit includes illegal content such as child pornography, it may be\n\nnecessary to use a force push to rewrite the history rather than simply hide the\n\ncommit. \n\n\n\nThis test has tiered scoring. Each tier must be fully satisfied to achieve points at the next tier. For example, if you fulfill the Tier 3 checks but do not fulfill all the Tier 2 checks, you will not receive any points for Tier 3.\n\n\n\nNote: If Scorecard is run without an administrative access token, the requirements that specify “For administrators” are ignored.\n\n\n\nTier 1 Requirements (3/10 points):\n\n  - Prevent force push\n\n  - Prevent branch deletion\n\n  - For administrators: Include administrator for review\n\n\n\nTier 2 Requirements (6/10 points):\n\n  - Required reviewers >=1 \n\n  - For administrators: Strict status checks (require branches to be up-to-date before merging)\n\n\n\nTier 3 Requirements (8/10 points):\n\n  - Status checks defined\n\n\n\nTier 4 Requirements (9/10 points):\n\n  - Required reviewers >= 2\n\n\n\nTier 5 Requirements (10/10 points):\n\n  - For administrators: Dismiss stale reviews\n\n"
              },
              "defaultConfiguration": {
                "level": "error"
              },
              "properties": {
                "precision": "high",
                "problem.severity": "error",
                "security-severity": "7.0",
                "tags": [
                  "supply-chain",
                  "security",
                  "source-code",
                  "code-reviews"
                ]
              }
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "BranchProtectionID",
          "ruleIndex": 0,
          "message": {
            "text": "score is 8: branch protection is not maximal on development and all release branches:\nWarn: number of required reviewers is only 1 on branch 'main'\nWarn: Stale review dismissal disabled on branch 'main'\nClick Remediation section below to solve this issue"
          },
          "locations": [
            {
              "physicalLocation": {
                "region": {
                  "startLine": 1
                },
                "artifactLocation": {
                  "uri": "no file associated with this alert",
                  "uriBaseId": "%SRCROOT%"
                }
              }
            }
          ]
        }
      ]
    }
  ]
}
```

</details>
