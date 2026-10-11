# e2e tests

The action's end-to-end tests run on a daily schedule in
[ossf-tests/scorecard-action](https://github.com/ossf-tests/scorecard-action),
across a matrix of results formats (`sarif`, `json`), with and without
publishing results, and with the default `GITHUB_TOKEN` and a separate
read-only token.

Workflow | What it runs | Status
-------- | ------------ | ------
[`scorecards-head.yml`](https://github.com/ossf-tests/scorecard-action/blob/main/.github/workflows/scorecards-head.yml) | `docker://ghcr.io/ossf/scorecard-action:latest`, the image built from this repository's `main` | [![scorecard-head](https://github.com/ossf-tests/scorecard-action/actions/workflows/scorecards-head.yml/badge.svg)](https://github.com/ossf-tests/scorecard-action/actions/workflows/scorecards-head.yml)
[`scorecards-latest-release.yml`](https://github.com/ossf-tests/scorecard-action/blob/main/.github/workflows/scorecards-latest-release.yml) | `ossf/scorecard-action@main`, which uses the image of the latest release | [![scorecard-latest-release](https://github.com/ossf-tests/scorecard-action/actions/workflows/scorecards-latest-release.yml/badge.svg)](https://github.com/ossf-tests/scorecard-action/actions/workflows/scorecards-latest-release.yml)

In this repository, [postsubmits.yml](../.github/workflows/postsubmits.yml) also
runs the `latest` image against this repository after every merge.
