# Authentication with Fine-grained PAT (optional)

For repositories that want to detect their classic Branch Protection rules, or webhooks, we suggest you create a fine-grained Personal Access Token (PAT) that Scorecard may use for authentication.

1. [Create a fine-grained Personal Access Token](https://github.com/settings/personal-access-tokens/new) with the following settings:
    - Token name: `OpenSSF Scorecard Action - $USER_NAME/$REPO_NAME>`
      (Note: replace `$USER_NAME/$REPO_NAME` with the names of your organization and repository so you can keep track of your tokens.)
    - Expiration: Set `Custom` and then set the date to exactly a year in the future (the maximum allowed)
    - Repository Access: `Only select repositories` and select the desired repository. 
      Alternatively, set `All repositories` if you wish to use the same token for all your repositories.
    - Repository Permissions:
        * `Administration: Read-only`: Required to read [Branch-Protection](https://github.com/ossf/scorecard/blob/main/docs/checks.md#branch-protection) settings.
        * `Metadata: Read-only` will be automatically set when you set `Administration`
        * `Webhooks: Read-only`: (Optional) required for the experimental [Webhook](https://github.com/ossf/scorecard/blob/main/docs/checks.md#webhooks) check.

    **Disclaimer:** Scorecard uses these permissions solely to learn about the project's branch protection rules and webhooks.
    However, the token can read many of the project's settings
    (for a full list, see the queries marked `(read)` in [GitHub's documentation](https://docs.github.com/en/rest/overview/permissions-required-for-fine-grained-personal-access-tokens?apiVersion=2022-11-28#administration)).

    "Classic" tokens with `repo` scope also work.
    However, these carry significantly higher risks compared to fine-grained PATs
    (see ["Classic" Personal Access Token (PAT) Requirements and Risks](/docs/authentication/classic-token.md))
    and are therefore strongly discouraged.

    ![image](/images/tokenscopes.png)

2. Copy the token value.

3. [Create a new repository secret](https://docs.github.com/en/actions/security-guides/encrypted-secrets#creating-encrypted-secrets-for-a-repository) with the following settings (**Warning:** [GitHub encrypted secrets](https://docs.github.com/en/actions/security-guides/encrypted-secrets) are accessible by all the workflows and maintainers of a repository.):
    - Name: `SCORECARD_TOKEN`
    - Value: the value of the token created in step 1 above.

    Note that fine-grained tokens expire after one year. You'll receive an email from GitHub when your token is about to expire, at which point you must regenerate it. Make sure to update the token string in your repository's secrets.

4. When you call the `ossf/scorecard-action` in your workflow, pass the token as `repo_token: ${{ secrets.SCORECARD_TOKEN }}`.