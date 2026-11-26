---
layout: "bitbucket"
page_title: "Bitbucket: bitbucket_workspace_gpg_public_key"
sidebar_current: "docs-bitbucket-data-workspace-gpg-public-key"
description: |-
  Provides the system GPG public key(s) Bitbucket uses to sign web commits for a workspace
---

# bitbucket\_workspace\_gpg\_public\_key

Provides a way to fetch the system GPG public key(s) Bitbucket uses to sign commits made in the web interface of a workspace, for example to add them to a list of trusted signing keys. In most cases one key is returned; during a key rotation period, two keys may be returned.

* OAuth2 Scopes: `account`
* API token permissions: `read:workspace:bitbucket`

## Example Usage

```hcl
data "bitbucket_workspace_gpg_public_key" "example" {
  workspace = "example"
}

output "bitbucket_web_commit_signing_keys" {
  value = data.bitbucket_workspace_gpg_public_key.example.keys[*].key
}
```

## Argument Reference

The following arguments are supported:

* `workspace` - (Required) The workspace ID (slug) or the workspace UUID surrounded by curly braces.

## Attributes Reference

* `id` - The workspace, as given in `workspace`.
* `keys` - The system GPG public keys, in the order the API returns them. See [Keys](#keys) below.

### Keys

* `key` - The ASCII-armored GPG public key.
* `fingerprint` - The GPG key fingerprint.
* `created_on` - When the key was created, in ISO 8601 format.
* `expires_on` - When the key expires, in ISO 8601 format.
