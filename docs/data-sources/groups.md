---
layout: "bitbucket"
page_title: "Bitbucket: bitbucket_groups"
sidebar_current: "docs-bitbucket-data-groups"
description: |-
  Provides a data for Bitbucket groups
---

# bitbucket\_groups

Provides a way to fetch data of groups in a workspace.

* OAuth2 Scopes: `account`
* API token permissions: `read:permission:bitbucket`

~> **Note:** Groups are managed through the Bitbucket REST 1.0 groups API, which has no 2.0 replacement yet. Atlassian's [v1 API deprecation notice](https://developer.atlassian.com/cloud/bitbucket/deprecation-notice-v1-apis/) keeps these endpoints available until Atlassian platform equivalents exist, with no removal date announced.

## Example Usage

```hcl
data "bitbucket_groups" "example" {
  workspace = "example"
}
```

## Argument Reference

The following arguments are supported:

* `workspace` - (Required) The workspace slug or UUID. A UUID (for example `data.bitbucket_workspace.example.id`) is resolved to the slug, which the 1.0 groups API requires.

## Attributes Reference

* `groups` - The list of groups in the workspace. See Group below for structure of each element

### Group

* `name` - The name of the groups.
* `auto_add` - Whether to automatically add users the groups
* `permission` - One of `read`, `write`, and `admin`.
* `slug` - The groups's slug.
* `email_forwarding_disabled` - Whether to disable email forwarding for group.
