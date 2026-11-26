---
layout: "bitbucket"
page_title: "Bitbucket: bitbucket_project_permissions"
sidebar_current: "docs-bitbucket-data-project-permissions"
description: |-
  Provides a data source for the explicit user and group permissions on a Bitbucket project
---

# bitbucket\_project\_permissions

Provides a way to list the users and groups that have explicit permissions on a project.

* OAuth2 Scopes: `project:admin`
* API token permissions: `read:project:bitbucket`

## Example Usage

```hcl
data "bitbucket_project_permissions" "example" {
  workspace   = "example"
  project_key = "PROJ"
}
```

## Argument Reference

The following arguments are supported:

* `workspace` - (Required) The workspace ID (slug) or UUID that owns the project.
* `project_key` - (Required) The key of the project.

## Attributes Reference

* `users` - Users with explicit permissions on the project, in the order the API returns them. See [Users](#users) below.
* `groups` - Groups with explicit permissions on the project, in the order the API returns them. See [Groups](#groups) below.

### Users

* `uuid` - The UUID of the user.
* `display_name` - The display name of the user.
* `account_id` - The Atlassian account ID of the user, when returned by the API.
* `permission` - The permission the user has on the project: `read`, `write`, `create-repo` or `admin`.

### Groups

* `slug` - The slug of the group.
* `name` - The name of the group.
* `permission` - The permission the group has on the project: `read`, `write`, `create-repo` or `admin`.
