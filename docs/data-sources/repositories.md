---
layout: "bitbucket"
page_title: "Bitbucket: bitbucket_repositories"
sidebar_current: "docs-bitbucket-data-repositories"
description: |-
  Lists the repositories in a workspace
---

# bitbucket\_repositories

Lists the repositories in a workspace that the caller can see, optionally only those in one project.

* OAuth2 Scopes: `repository`
* API token permissions: `read:repository:bitbucket`

## Example Usage

```hcl
# All repositories in a workspace
data "bitbucket_repositories" "all" {
  workspace = "example"
}

# Repositories in one project
data "bitbucket_repositories" "platform" {
  workspace   = "example"
  project_key = "PLAT"
}

output "platform_repo_slugs" {
  value = data.bitbucket_repositories.platform.repositories[*].slug
}
```

## Argument Reference

The following arguments are supported:

* `workspace` - (Required) The workspace slug or UUID.
* `project_key` - (Optional) Only list repositories in the project with this key.

## Attributes Reference

* `repositories` - The repositories, in the order Bitbucket returns them. Each has:
    * `uuid` - Repository UUID.
    * `slug` - Repository slug.
    * `name` - Repository name.
    * `full_name` - Workspace slug and repository slug, e.g. `workspace/repo`.
    * `description` - Repository description.
    * `is_private` - Whether the repository is private.
    * `fork_policy` - `allow_forks`, `no_public_forks` or `no_forks`.
    * `language` - Repository language.
    * `scm` - Repository SCM, always `git`.
    * `has_wiki` - Whether the wiki is enabled.
    * `has_issues` - Whether the issue tracker is enabled.
    * `main_branch` - Main branch name.
    * `project_key` - Key of the project the repository belongs to.
    * `enforced_signed_commits` - Whether signed commits are required for the repository.
    * `created_on` - When the repository was created (RFC 3339).
    * `updated_on` - When the repository was last updated (RFC 3339).
