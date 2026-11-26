package bitbucket

import (
	"context"
	"fmt"
	"net/http"

	"github.com/DrFaust92/bitbucket-go-client"
	"github.com/antihax/optional"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataRepositories() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataReadRepositories,
		Description:        "Lists the repositories in a workspace, optionally only those in one project.",
		Schema: map[string]*schema.Schema{
			"workspace": {
				Type:         schema.TypeString,
				Description:  "Workspace slug or UUID.",
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"project_key": {
				Type:         schema.TypeString,
				Description:  "Only list repositories in the project with this key.",
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"repositories": {
				Type:        schema.TypeList,
				Description: "Repositories the caller can see, in the order Bitbucket returns them.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uuid": {
							Type:        schema.TypeString,
							Description: "Repository UUID.",
							Computed:    true,
						},
						"slug": {
							Type:        schema.TypeString,
							Description: "Repository slug.",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "Repository name.",
							Computed:    true,
						},
						"full_name": {
							Type:        schema.TypeString,
							Description: "Workspace slug and repository slug, e.g. `workspace/repo`.",
							Computed:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "Repository description.",
							Computed:    true,
						},
						"is_private": {
							Type:        schema.TypeBool,
							Description: "Whether the repository is private.",
							Computed:    true,
						},
						"fork_policy": {
							Type:        schema.TypeString,
							Description: "Fork policy: `allow_forks`, `no_public_forks` or `no_forks`.",
							Computed:    true,
						},
						"language": {
							Type:        schema.TypeString,
							Description: "Repository language.",
							Computed:    true,
						},
						"scm": {
							Type:        schema.TypeString,
							Description: "Repository SCM, always `git`.",
							Computed:    true,
						},
						"has_wiki": {
							Type:        schema.TypeBool,
							Description: "Whether the wiki is enabled.",
							Computed:    true,
						},
						"has_issues": {
							Type:        schema.TypeBool,
							Description: "Whether the issue tracker is enabled.",
							Computed:    true,
						},
						"main_branch": {
							Type:        schema.TypeString,
							Description: "Main branch name.",
							Computed:    true,
						},
						"project_key": {
							Type:        schema.TypeString,
							Description: "Key of the project the repository belongs to.",
							Computed:    true,
						},
						"enforced_signed_commits": {
							Type:        schema.TypeBool,
							Description: "Whether signed commits are required for the repository.",
							Computed:    true,
						},
						"created_on": {
							Type:        schema.TypeString,
							Description: "When the repository was created (RFC 3339).",
							Computed:    true,
						},
						"updated_on": {
							Type:        schema.TypeString,
							Description: "When the repository was last updated (RFC 3339).",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataReadRepositories(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(Clients).genClient
	client := m.(Clients).httpClient
	repoApi := c.ApiClient.RepositoriesApi

	workspace := d.Get("workspace").(string)
	projectKey := d.Get("project_key").(string)

	opts := &bitbucket.RepositoriesApiRepositoriesWorkspaceGetOpts{}
	id := workspace
	if projectKey != "" {
		opts.Q = optional.NewString(fmt.Sprintf("project.key=%q", projectKey))
		id = fmt.Sprintf("%s/%s", workspace, projectKey)
	}

	var repos []interface{}
	next := ""
	for {
		var page bitbucket.PaginatedRepositories
		var res *http.Response
		var err error
		if next == "" {
			page, res, err = repoApi.RepositoriesWorkspaceGet(c.AuthContext, workspace, opts)
		} else {
			res, err = client.GetNext(next, &page)
		}
		if err := handleClientError(res, err); err != nil {
			return diag.FromErr(err)
		}

		for i := range page.Values {
			repos = append(repos, flattenRepositoriesItem(&page.Values[i]))
		}

		if page.Next == "" {
			break
		}
		next = page.Next
	}

	d.SetId(id)
	if err := d.Set("repositories", repos); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func flattenRepositoriesItem(r *bitbucket.Repository) map[string]interface{} {
	item := map[string]interface{}{
		"uuid":                    r.Uuid,
		"slug":                    r.Slug,
		"name":                    r.Name,
		"full_name":               r.FullName,
		"description":             r.Description,
		"is_private":              r.IsPrivate,
		"fork_policy":             r.ForkPolicy,
		"language":                r.Language,
		"scm":                     r.Scm,
		"has_wiki":                r.HasWiki,
		"has_issues":              r.HasIssues,
		"enforced_signed_commits": r.EnforcedSignedCommits,
		"created_on":              formatTime(r.CreatedOn),
		"updated_on":              formatTime(r.UpdatedOn),
		"main_branch":             "",
		"project_key":             "",
	}
	if r.Mainbranch != nil {
		item["main_branch"] = r.Mainbranch.Name
	}
	if r.Project != nil {
		item["project_key"] = r.Project.Key
	}
	return item
}
