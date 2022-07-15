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

### Update the scorecard-action version

NOTE: we have a chicken-and-egg problem where the commit to be used for the release
needs to have the image tag that only gets created *after* the commit is pushed. We
workaround that by pre-selecting and referencing the image tag instead of the SHA which isn't ideal 
but workable.

Pre-select the tag which will be used for the release. For this document, we'll use: `Tag`.

Update the image tage in [action.yaml](action.yaml) to use `Tag`.

Example:

```
runs:
  using: "docker"
  image: "docker://gcr.io/openssf/scorecard-action:Tag"
```

Create a pull request with this change and merge into `main`.

## Drafting release notes

<!-- TODO(release): Provide details -->

## Release

### Create a tag

Locally, create a signed tag `Tag` on commitSHA `SHA`:

```console
git remote update
git checkout `SHA`
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
