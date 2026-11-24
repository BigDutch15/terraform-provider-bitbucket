---
layout: "bitbucket"
page_title: "Bitbucket: bitbucket_group"
sidebar_current: "docs-bitbucket-resource-group"
description: |-
  Provides a Bitbucket Group
---

# bitbucket\_group

Provides a Bitbucket group resource.

This allows you to manage your groups.

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
  auto_add   = true
  permission = "read"
}
```

## Argument Reference

The following arguments are supported:

* `workspace` - (Required) The workspace slug or UUID. A UUID (for example `data.bitbucket_workspace.example.id`) is resolved to the slug, which the 1.0 groups API requires.
* `name` - (Required) The name of the group.
* `auto_add` - (Optional) Whether to automatically add users the group
* `permission` - (Optional) One of `read`, `write`, and `admin`.
* `email_forwarding_disabled` - Whether to disable email forwarding for group.

## Attributes Reference

* `slug` - The groups slug.

## Import

Groups can be imported using their `workspace/group-slug` ID, e.g.

```sh
terraform import bitbucket_group.group my-workspace/group-slug
```
