---
layout: "bitbucket"
page_title: "Bitbucket: bitbucket_workspace_variables"
sidebar_current: "docs-bitbucket-data-workspace-variables"
description: |-
  Provides a data source for the Pipelines variables of a Bitbucket workspace
---

# bitbucket\_workspace\_variables

Provides a way to list the Pipelines variables of a workspace.

* OAuth2 Scopes: `pipeline`
* API token permissions: `read:pipeline:bitbucket`

## Example Usage

```hcl
data "bitbucket_workspace_variables" "example" {
  workspace = "example"
}
```

## Argument Reference

The following arguments are supported:

* `workspace` - (Required) The workspace ID (slug) or UUID.

## Attributes Reference

* `variables` - The workspace's Pipelines variables, in the order the API returns them. See [Variable](#variable) below.

### Variable

* `uuid` - The UUID of the variable.
* `key` - The name of the variable.
* `value` - (Sensitive) The value of the variable. Empty when `secured` is `true`, because the API does not return secured values.
* `secured` - Whether the variable is secured.
