# Releasing the scorecard GitHub Action

This is a draft document to describe the release process for the scorecard
GitHub Action.

(If there are improvements you'd like to see, please comment on the
[tracking issue](https://github.com/ossf/scorecard-action/issues/33) or issue a
pull request to discuss.)

- [Tracking](#tracking)
- [Preparing the release](#preparing-the-release)
  - [Validate the Action](#validate-the-action)
  - [Update the scorecard-action version](#update-the-scorecard-action-version)
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

### Validate the Action

The [postsubmits.yml](https://github.com/ossf/scorecard-action/actions/workflows/postsubmits.yml)
workflow runs on every push to `main` and weekly. It builds and pushes the
`ghcr.io/ossf/scorecard-action:latest` image, then runs Scorecard on this
repository with that image, uploading the SARIF and publishing the results.

Confirm that the `postsubmits.yml` run for the commit you plan to release
succeeded, including the Scorecard analysis job (no
`error sending scorecard results` line in its log). If there isn't one, run the
workflow manually.

Do this before updating the image tag below: once `action.yaml` points at the
new tag, `ossf/scorecard-action@main` references an image that won't exist until
the release tag is pushed.

### Update the scorecard-action version

NOTE: we have a chicken-and-egg problem where the commit to be used for the release
needs to have the image tag that only gets created *after* the commit is pushed. We
workaround that by pre-selecting and referencing the image tag instead of the SHA which isn't ideal 
but workable.

Pre-select the tag which will be used for the release. For this document, we'll use: `Tag`.

Update the image tag in [action.yaml](action.yaml) to use `Tag`.

Example:

```
runs:
  using: "docker"
  image: "docker://ghcr.io/ossf/scorecard-action:Tag"
```

Create a pull request with this change and merge into `main`.

## Drafting release notes

Start from GitHub's generated notes for the range, either by clicking
**Generate release notes** on a
[new draft release](https://github.com/ossf/scorecard-action/releases/new) or
with:

```console
gh api -X POST repos/ossf/scorecard-action/releases/generate-notes \
  -f tag_name=v100.0.0 -f previous_tag_name=<previous tag>
```

Then edit, following
[v2.4.4](https://github.com/ossf/scorecard-action/releases/tag/v2.4.4):

- Open with a sentence naming the Scorecard version the release moves to,
  linking that version's Scorecard release notes, which cover the changes to
  checks and results.
- List the action's own changes. Drop Dependabot bumps unless they change
  behavior.
- Keep **New Contributors** and the **Full Changelog** link as generated.

Save the draft and share it with the other maintainers.

## Release

### Create a tag

Locally, create a signed tag `Tag` on commitSHA `SHA`:

```console
git remote update
git checkout `SHA`
git tag -s -m "v100.0.0" v100.0.0
git push <upstream> v100.0.0
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

GitHub's starter workflow is what new users get when they add the action from
the Actions tab, so keep it on the latest release.

1. Get the commit SHA of the release tag:

   ```console
   git rev-parse v100.0.0^{commit}
   ```

1. Open a pull request in
   [actions/starter-workflows](https://github.com/actions/starter-workflows)
   that updates the action in
   [`code-scanning/scorecard.yml`](https://github.com/actions/starter-workflows/blob/main/code-scanning/scorecard.yml)
   to that SHA, keeping the version comment in step:

   ```yaml
   uses: ossf/scorecard-action@<SHA> # v100.0.0
   ```

1. Update the example workflow our README points to,
   [`ossf/scorecard`'s `scorecard-analysis.yml`](https://github.com/ossf/scorecard/blob/main/.github/workflows/scorecard-analysis.yml),
   the same way.

1. Verify on
   [GitHub Marketplace](https://github.com/marketplace/actions/ossf-scorecard-action)
   that the workflow example uses the new release.

   _NOTE: GitHub Marketplace uses the default branch as reference documentation_

## Announce

<!-- TODO(release): Provide details -->
