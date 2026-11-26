---
layout: "bitbucket"
page_title: "Bitbucket: bitbucket_repository_permissions"
sidebar_current: "docs-bitbucket-data-repository-permissions"
description: |-
  Provides a data for Bitbucket Repository Permissions
---

# bitbucket\_repository\_permissions

Provides a way to fetch the users and groups that have explicit permissions on a repository. Permissions inherited from the workspace or project are not included.

* OAuth2 Scopes: `repository:admin`
* API token permissions: `read:repository:bitbucket`

## Example Usage

```hcl
data "bitbucket_repository_permissions" "example" {
  workspace  = "example"
  repository = "example"
}
```

## Argument Reference

The following arguments are supported:

* `workspace` - (Required) The workspace slug or UUID.
* `repository` - (Required) The repository slug or UUID.

## Attributes Reference

* `id` - The ID in the form `workspace/repository`.
* `users` - Users with explicit permissions on the repository, in the order the API returns them. See [Users](#users) below.
* `groups` - Groups with explicit permissions on the repository, in the order the API returns them. See [Groups](#groups) below.

### Users

* `uuid` - The user's UUID.
* `display_name` - The user's display name.
* `account_id` - The user's Atlassian account ID, if returned.
* `permission` - The permission the user has on the repository: `read`, `write` or `admin`.

### Groups

* `slug` - The group's slug.
* `name` - The group's name.
* `permission` - The permission the group has on the repository: `read`, `write` or `admin`.
