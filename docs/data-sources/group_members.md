---
layout: "bitbucket"
page_title: "Bitbucket: bitbucket_group_members"
sidebar_current: "docs-bitbucket-data-group-members"
description: |-
  Provides a data for Bitbucket group members
---

# bitbucket\_group\_members

Provides a way to fetch data of group members.

* OAuth2 Scopes: `account`
* API token permissions: `read:permission:bitbucket`

~> **Note:** Groups are managed through the Bitbucket REST 1.0 groups API, which has no 2.0 replacement yet. Atlassian's [v1 API deprecation notice](https://developer.atlassian.com/cloud/bitbucket/deprecation-notice-v1-apis/) keeps these endpoints available until Atlassian platform equivalents exist, with no removal date announced.

## Example Usage

```hcl
data "bitbucket_group_members" "example" {
  workspace = "example"
  slug      = "example"
}
```

## Argument Reference

The following arguments are supported:

* `workspace` - (Required) The workspace slug or UUID. A UUID (for example `data.bitbucket_workspace.example.id`) is resolved to the slug, which the 1.0 groups API requires.
* `slug` - (Required) The group's slug.

## Attributes Reference

* `members` - A list of group member uuid.
* `group_members` - A set of group member objects. See [Group Member](#group-member) below.

### Group Member

* `uuid` - User UUID.
* `username` - The Username.
* `display_name` - The User display name.
