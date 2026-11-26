---
layout: "bitbucket"
page_title: "Bitbucket: bitbucket_branch_restrictions"
sidebar_current: "docs-bitbucket-data-branch-restrictions"
description: |-
  Provides a data source for the branch restrictions of a Bitbucket repository
---

# bitbucket\_branch\_restrictions

Provides a way to list the branch restrictions of a repository.

* OAuth2 Scopes: `repository:admin`
* API token permissions: `read:repository:bitbucket` and `admin:repository:bitbucket`

## Example Usage

```hcl
data "bitbucket_branch_restrictions" "example" {
  workspace  = "example"
  repository = "example"
}

# Only the push restrictions on branches matching "main"
data "bitbucket_branch_restrictions" "main_push" {
  workspace  = "example"
  repository = "example"
  kind       = "push"
  pattern    = "main"
}
```

## Argument Reference

The following arguments are supported:

* `workspace` - (Required) The workspace slug or UUID.
* `repository` - (Required) The repository slug or UUID.
* `kind` - (Optional) Only return branch restrictions of this kind, e.g. `push`. See the `kind` argument of the [`bitbucket_branch_restriction`](../resources/branch_restriction.md) resource for the possible values.
* `pattern` - (Optional) Only return branch restrictions that apply to branches matching this pattern.

## Attributes Reference

* `id` - The data source ID, `workspace/repository`.
* `branch_restrictions` - The repository's branch restrictions, in the order the API returns them. Each has:
    * `id` - The branch restriction ID.
    * `kind` - The kind of restriction.
    * `branch_match_kind` - How the restriction is matched against a branch: `glob` or `branching_model`.
    * `branch_type` - The branch type the restriction applies to when `branch_match_kind` is `branching_model`.
    * `pattern` - The branch pattern the restriction applies to when `branch_match_kind` is `glob`.
    * `value` - The kind-specific value, e.g. the number of approvals required. `0` when not set.
    * `users` - UUIDs of the users the restriction exempts or applies to.
    * `groups` - Groups the restriction exempts or applies to. Each has:
        * `owner` - The workspace slug that owns the group (the workspace UUID if no slug is returned).
        * `slug` - The group slug.
    * `pipelines_source_branches` - Source branches allowed to push back to the restricted branch from Pipelines.
