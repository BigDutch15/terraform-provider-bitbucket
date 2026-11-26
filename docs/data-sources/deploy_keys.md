---
layout: "bitbucket"
page_title: "Bitbucket: bitbucket_deploy_keys"
sidebar_current: "docs-bitbucket-data-deploy-keys"
description: |-
  Provides a data source for the access (deploy) keys of a Bitbucket repository
---

# bitbucket\_deploy\_keys

Provides a way to list the access (deploy) keys of a repository.

* OAuth2 Scopes: `repository` and `repository:admin`
* API token permissions: `admin:repository:bitbucket`

## Example Usage

```hcl
data "bitbucket_deploy_keys" "example" {
  workspace  = "example"
  repository = "example"
}
```

## Argument Reference

The following arguments are supported:

* `workspace` - (Required) The workspace slug or UUID.
* `repository` - (Required) The repository slug.

## Attributes Reference

* `id` - The ID of the data source, `<workspace>/<repository>`.
* `deploy_keys` - The repository's deploy keys, in the order the API returns them. See [Deploy Keys](#deploy-keys) below.

### Deploy Keys

* `key_id` - The ID of the deploy key.
* `key` - The public SSH key.
* `label` - The user-defined label of the deploy key.
* `comment` - The comment parsed from the key, if present.
* `created_on` - When the deploy key was added.
* `last_used` - When the deploy key was last used. Empty if it has never been used.
