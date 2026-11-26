---
layout: "bitbucket"
page_title: "Bitbucket: bitbucket_projects"
sidebar_current: "docs-bitbucket-data-projects"
description: |-
  Provides a data source to list the projects in a Bitbucket workspace
---

# bitbucket\_projects

Provides a way to list the projects in a workspace.

* OAuth2 Scopes: `project`
* API token permissions: `read:project:bitbucket`

## Example Usage

```hcl
data "bitbucket_projects" "example" {
  workspace = "example"
}

output "project_keys" {
  value = data.bitbucket_projects.example.projects[*].key
}
```

## Argument Reference

The following arguments are supported:

* `workspace` - (Required) The workspace slug or `{UUID}`.

## Attributes Reference

* `id` - The workspace.
* `projects` - The projects in the workspace, in the order the API returns them. See [Projects](#projects) below.

### Projects

* `key` - Project key.
* `uuid` - Project UUID.
* `name` - Project name.
* `description` - Project description.
* `is_private` - Whether the project is private.
* `has_publicly_visible_repos` - Whether the project contains publicly visible repositories.
* `enforced_signed_commits` - Whether signed commits are required for the project. True when the workspace is on a Premium plan and enforcement is on for the project or its workspace.
