---
layout: "bitbucket"
page_title: "Bitbucket: bitbucket_effective_branching_model"
sidebar_current: "docs-bitbucket-data-effective-branching-model"
description: |-
  Provides a data for the effective Bitbucket branching model of a repository
---

# bitbucket\_effective\_branching\_model

Provides a way to fetch the branching model in effect for a repository. This is the repository's own branching model, or the one inherited from its project when the repository does not override it.

* OAuth2 Scopes: `repository`
* API token permissions: `read:repository:bitbucket`

## Example Usage

```hcl
data "bitbucket_effective_branching_model" "example" {
  workspace  = "example"
  repository = "example"
}
```

## Argument Reference

The following arguments are supported:

* `workspace` - (Required) The workspace ID (slug) or UUID.
* `repository` - (Required) The repository slug or UUID.

## Attributes Reference

* `id` - The ID, in the form `workspace/repository`.
* `development` - The development branch in effect. See [Branch](#branch) below.
* `production` - The production branch in effect. Empty when the effective model has no production branch. See [Branch](#branch) below.
* `branch_type` - The active branch types, in the order the API returns them. See [Branch Type](#branch-type) below.

### Branch

* `name` - Name of the branch.
* `use_mainbranch` - Whether the repository's main branch is used.
* `branch_does_not_exist` - Whether the configured branch is missing from the repository. Useful for detecting a fallback to the main branch when a repository inherits its project's branching model.

### Branch Type

* `kind` - The kind of branch: `feature`, `bugfix`, `release` or `hotfix`.
* `prefix` - The prefix for this branch type.
