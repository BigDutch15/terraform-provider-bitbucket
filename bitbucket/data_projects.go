package bitbucket

import (
	"context"
	"net/http"

	"github.com/DrFaust92/bitbucket-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataProjects() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataReadProjects,
		Description:        "Datasource to list the projects in a workspace",

		Schema: map[string]*schema.Schema{
			"workspace": {
				Type:         schema.TypeString,
				Description:  "Workspace slug or {UUID}",
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"projects": {
				Type:        schema.TypeList,
				Description: "Projects in the workspace, in the order the API returns them",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:        schema.TypeString,
							Description: "Project key",
							Computed:    true,
						},
						"uuid": {
							Type:        schema.TypeString,
							Description: "Project UUID",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "Project name",
							Computed:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "Project description",
							Computed:    true,
						},
						"is_private": {
							Type:        schema.TypeBool,
							Description: "Project is private",
							Computed:    true,
						},
						"has_publicly_visible_repos": {
							Type:        schema.TypeBool,
							Description: "Repositories are publicly visible",
							Computed:    true,
						},
						"enforced_signed_commits": {
							Type:        schema.TypeBool,
							Description: "Whether signed commits are required for the project: true when the workspace is on a Premium plan and enforcement is on for the project or its workspace.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataReadProjects(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(Clients).genClient
	workspaceApi := c.ApiClient.WorkspacesApi
	client := m.(Clients).httpClient

	workspace := d.Get("workspace").(string)

	var projects []interface{}
	next := ""
	for {
		var page bitbucket.PaginatedProjects
		var res *http.Response
		var err error
		if next == "" {
			page, res, err = workspaceApi.WorkspacesWorkspaceProjectsGet(c.AuthContext, workspace, nil)
		} else {
			res, err = client.GetNext(next, &page)
		}
		if err := handleClientError(res, err); err != nil {
			return diag.FromErr(err)
		}

		for _, project := range page.Values {
			projects = append(projects, map[string]interface{}{
				"key":                        project.Key,
				"uuid":                       project.Uuid,
				"name":                       project.Name,
				"description":                project.Description,
				"is_private":                 project.IsPrivate,
				"has_publicly_visible_repos": project.HasPubliclyVisibleRepos,
				"enforced_signed_commits":    project.EnforcedSignedCommits,
			})
		}

		if page.Next == "" {
			break
		}
		next = page.Next
	}

	d.SetId(workspace)
	if err := d.Set("projects", projects); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
