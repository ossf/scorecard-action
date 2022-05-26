# Releasing the scorecard GitHub Action

This is a draft document to describe the release process for the scorecard
GitHub Action.

(If there are improvements you'd like to see, please comment on the
[tracking issue](https://github.com/ossf/scorecard-action/issues/33) or issue a
pull request to discuss.)

- [Tracking](#tracking)
- [Preparing the release](#preparing-the-release)
  - [Update the scorecard version](#update-the-scorecard-version)
- [Drafting release notes](#drafting-release-notes)
- [Release](#release)
  - [Create a tag](#create-a-tag)
  - [Create a GitHub release](#create-a-github-release)
- [Update the starter workflow](#update-the-starter-workflow)
- [Announce](#announce)

## Tracking

As the first task, a Release Manager should open a tracking issue for the
release.

We don't currently have a template for releasing, but the following
[issue](https://github.com/ossf/scorecard-action/issues/97) is a good example
to draw inspiration from.

We're not striving for perfection with the template, but the tracking issue
will serve as a reference point to aggregate feedback, so try your best to be
as descriptive as possible.

## Preparing the release

This section covers changes that need to be issued as a pull request and should
be merged before releasing the scorecard GitHub Action.

### Update the scorecard version

_NOTE: As the scorecard GitHub Action is based on scorecard, you may want to publish a new release of scorecard to ensure the next release of the GitHub Action has the most up-to-date functionality. This is not strictly required. The only requirement is that we use a stable scorecard version which is at or above the current version used for this action._

For the rest of document, let `CH1` be the hash of the scorecard image you
intend to use for this release.

See [here](https://github.com/orgs/ossf/packages?repo_name=scorecard) for
scorecard images.

(We'll use `0bc9576b3efbda7b38febbf0a1e1b9546894f9650aaead9ccb5edc7dade86552`
as `CH1` in any examples below.)

Now that you have `CH1`, update the digest in the [Dockerfile](Dockerfile) to use `CH1`.

Example:

```Dockerfile
FROM gcr.io/openssf/scorecard:v100.0.0@sha256:0bc9576b3efbda7b38febbf0a1e1b9546894f9650aaead9ccb5edc7dade86552 as base
```

Create a pull request with this change.

Once the PR is merged, note the GitHub commit hash.
We'll refer to this as `GH2` below.

## Drafting release notes

<!-- TODO(release): Provide details -->

## Release

### Create a tag

Locally, create a signed tag based on `GH2`:

```console
git remote update
git checkout `GH2`
git tag -s -m "v100.0.0" v100.0.0
git push <upstream> --tags
```

### Create a GitHub release

Create a
[GitHub release](https://github.com/ossf/scorecard-action/releases/new) using
the tag you've just created.

Release title: `<tag>`

The release notes will be the notes you drafted in the previous step.

Ensure the following fields are up to date:

- Security contact email
- Primary Category
- Another Category — optional

Click `Publish release`.

## Update the starter workflow

1. Open a pull request in the
[starter workflows repo](https://github.com/actions/starter-workflows/tree/main/code-scanning/scorecards.yml)
to update the action's digest to `GH2`.

1. Update our documentation's example workflow to use `GH2`.

1. Verify on
   [GitHub Marketplace](https://github.com/marketplace/actions/ossf-scorecard-action)
   that the workflow example contains `GH2`.

   _NOTE: GitHub Marketplace uses the default branch as reference documentation_

## Announce

<!-- TODO(release): Provide details -->
