# "Classic" Personal Access Token (PAT) Requirements and Risks
Certain features require a Personal Access Token (PAT).
We recommend you use a fine-grained token as described in [Authentication with Fine-grained PAT](/docs/authentication/fine-grained-auth-token.md).
A "classic" PAT also works, but we strongly discourage its use.

Due to a limitation of the "classic" tokens' permission model,
the PAT needs [write permission to the repository](https://docs.github.com/developers/apps/building-oauth-apps/scopes-for-oauth-apps#available-scopes) through the `repo` scope.
**The PAT will be stored as a [GitHub encrypted secret](https://docs.github.com/actions/security-guides/encrypted-secrets)
and be accessible by all of the repository's workflows and maintainers.**
This means another maintainer on your project could potentially use the token to impersonate you.
If there is an exploitable bug in a workflow with write permissions,
an external contributor could potentially exploit it to extract the PAT.

The only benefit of a "classic" PAT is that it can be set to never expire.
However, we believe this does not outweigh the significantly higher risk of "classic" PATs compared to fine-grained PATs.
