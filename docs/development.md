# Developing

## Testing container images

```shell
docker run -e GITHUB_REF=refs/heads/main \
           -e GITHUB_EVENT_NAME=branch_protection_rule \
           -e INPUT_RESULTS_FORMAT=sarif \
           -e INPUT_RESULTS_FILE=results.sarif \
           -e GITHUB_WORKSPACE=/ \
           -e INPUT_POLICY_FILE="/policy.yml" \
           -e INPUT_REPO_TOKEN=$GITHUB_AUTH_TOKEN \
           -e GITHUB_REPOSITORY="ossf/scorecard" \
           laurentsimon/scorecard-action:latest
```
