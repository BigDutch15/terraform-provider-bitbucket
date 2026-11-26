---
layout: "bitbucket"
page_title: "Bitbucket: bitbucket_workspace_repository_permissions"
sidebar_current: "docs-bitbucket-data-workspace-repository-permissions"
description: |-
  Provides a data source for the repository permissions in a Bitbucket workspace
---

# bitbucket\_workspace\_repository\_permissions

Provides a way to fetch the effective repository permissions of every user across all repositories in a workspace.

Permissions returned are effective permissions: the highest level of permission the user has. They do not distinguish between direct and group permissions. Only workspace admins can read this data.

* OAuth2 Scopes: `account`
* API token permissions: `read:repository:bitbucket`

## Example Usage

```hcl
data "bitbucket_workspace_repository_permissions" "example" {
  workspace = "example"
}
```

Filter the results with a query:

```hcl
data "bitbucket_workspace_repository_permissions" "admins" {
  workspace = "example"
  q         = "permission=\"admin\""
}
```

## Argument Reference

The following arguments are supported:

* `workspace` - (Required) The workspace ID (slug) or UUID surrounded by curly braces.
* `q` - (Optional) Query string to filter the results, for example `repository.name="geordi"` or `permission>"read"`. See [filtering and sorting](https://developer.atlassian.com/cloud/bitbucket/rest/intro/#filtering).

## Attributes Reference

* `permissions` - Repository permissions in the workspace, in the order returned by the API. See [Permission](#permission) below.

### Permission

* `permission` - The effective permission level: `admin`, `write` or `read`.
* `repository_name` - The name of the repository.
* `repository_full_name` - The full name of the repository, in the form `workspace/repo_slug`.
* `repository_uuid` - The UUID of the repository.
* `user_uuid` - The UUID of the user.
* `user_display_name` - The display name of the user.
* `user_account_id` - The Atlassian account ID of the user.
