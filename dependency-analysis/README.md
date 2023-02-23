# OpenSSF Dependency Analysis

This repository contains the source code for the OpenSSF Dependency Analysis project.

## Overview
The OpenSSF Dependency Analysis project is to check the security posture of a project's dependencies. 
It uses  https://docs.github.com/en/rest/dependency-graph/dependency-review?apiVersion=2022-11-28#get-a-diff-of-the-dependencies-between-commits
to get the dependencies of a project and then uses https://api.securityscorecards.dev to get the security posture of the dependencies.
https://github.com/ossf/scorecard-action/issues/1070

## Usage
The project is a GitHub Action that can be used in a workflow. The workflow can be triggered on a push or pull request event.

This will run the action on the latest commit on the default branch of the repository and will create a comment on the pull request with the results of the analysis.

Something like this: https://github.com/ossf-tests/vulpy/pull/2#issuecomment-1442310469