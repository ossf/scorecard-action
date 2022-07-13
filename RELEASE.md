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

### Update the scorecard-action version

Note be the hash of the scorecard-action image (say, `CH1`) that was tagged with `Tag`. We will use this for the release.

Update the digest in [action.yaml](action.yaml) to use `CH1`.

Example:

```
runs:
  using: "docker"
  image: "docker://gcr.io/openssf/scorecard-action:CH1"
```

Create a pull request with this change and merge into `main`.

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
