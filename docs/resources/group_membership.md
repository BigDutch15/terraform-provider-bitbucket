---
layout: "bitbucket"
page_title: "Bitbucket: bitbucket_group_membership"
sidebar_current: "docs-bitbucket-resource-group-membership"
description: |-
  Provides support for setting Bitbucket Group Membership
---

# bitbucket\_group\_membership

Provides a Bitbucket group membership resource.

This allows you to manage your group membership.

* OAuth2 Scopes: `account:write`
* API token permissions: `read:permission:bitbucket` and `write:permission:bitbucket`

~> **Note:** Groups are managed through the Bitbucket REST 1.0 groups API, which has no 2.0 replacement yet. Atlassian's [v1 API deprecation notice](https://developer.atlassian.com/cloud/bitbucket/deprecation-notice-v1-apis/) keeps these endpoints available until Atlassian platform equivalents exist, with no removal date announced.

## Example Usage

```hcl
data "bitbucket_workspace" "test" {
  workspace = "example"
}

resource "bitbucket_group" "test" {
  workspace  = data.bitbucket_workspace.test.slug
  name       = "example"
}

data "bitbucket_current_user" "test" {}

resource "bitbucket_group_membership" "test" {
  workspace  = bitbucket_group.test.workspace
  group_slug = bitbucket_group.test.slug
  uuid       = data.bitbucket_current_user.test.id
}
```

## Argument Reference

The following arguments are supported:

* `workspace` - (Required) The workspace slug or UUID. A UUID (for example `data.bitbucket_workspace.example.id`) is resolved to the slug, which the 1.0 groups API requires.
* `group_slug` - (Required) The slug of the group.
* `uuid` - (Required) The member UUID to add to the group.

## Import

Group Members can be imported using their `workspace/group-slug/member-uuid` ID, e.g.

```sh
terraform import bitbucket_group_membership.group my-workspace/group-slug/member-uuid
```
