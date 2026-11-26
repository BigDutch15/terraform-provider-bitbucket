---
layout: "bitbucket"
page_title: "Bitbucket: bitbucket_repository_variables"
sidebar_current: "docs-bitbucket-data-repository-variables"
description: |-
  Provides a data source for the Pipelines variables of a Bitbucket repository
---

# bitbucket\_repository\_variables

Provides a way to list the Pipelines variables of a repository.

* OAuth2 Scopes: `pipeline`
* API token permissions: `read:pipeline:bitbucket`

## Example Usage

```hcl
data "bitbucket_repository_variables" "example" {
  workspace  = "example"
  repository = "example"
}
```

## Argument Reference

The following arguments are supported:

* `workspace` - (Required) The workspace ID (slug) or UUID.
* `repository` - (Required) The repository slug.

## Attributes Reference

* `variables` - The repository's Pipelines variables, in the order the API returns them. See [Variable](#variable) below.

### Variable

* `uuid` - The UUID of the variable.
* `key` - The name of the variable.
* `value` - (Sensitive) The value of the variable. Empty when `secured` is `true`, because the API does not return secured values.
* `secured` - Whether the variable is secured.
