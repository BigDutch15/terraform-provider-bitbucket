package bitbucket

import (
	"context"
	"net/http"

	"github.com/DrFaust92/bitbucket-go-client"
	"github.com/antihax/optional"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataWorkspaceRepositoryPermissions() *schema.Resource {
	return &schema.Resource{
		Description:        "Lists the effective repository permissions of every user across all repositories in a workspace.",
		ReadWithoutTimeout: dataReadWorkspaceRepositoryPermissions,

		Schema: map[string]*schema.Schema{
			"workspace": {
				Description:  "The workspace ID (slug) or UUID surrounded by curly braces.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"q": {
				Description: "Query string to filter the results, for example `repository.name=\"geordi\"` or `permission>\"read\"`. See the Bitbucket Cloud REST API filtering documentation.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"permissions": {
				Description: "Repository permissions in the workspace, in the order returned by the API.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"permission": {
							Description: "The effective permission level: `admin`, `write` or `read`.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"repository_name": {
							Description: "The name of the repository.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"repository_full_name": {
							Description: "The full name of the repository, in the form `workspace/repo_slug`.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"repository_uuid": {
							Description: "The UUID of the repository.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"user_uuid": {
							Description: "The UUID of the user.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"user_display_name": {
							Description: "The display name of the user.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"user_account_id": {
							Description: "The Atlassian account ID of the user.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataReadWorkspaceRepositoryPermissions(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(Clients).genClient
	workspaceApi := c.ApiClient.WorkspacesApi
	client := m.(Clients).httpClient

	workspace := d.Get("workspace").(string)

	options := bitbucket.WorkspacesApiWorkspacesWorkspacePermissionsRepositoriesGetOpts{}
	if q, ok := d.GetOk("q"); ok {
		options.Q = optional.NewString(q.(string))
	}

	var permissions []bitbucket.RepositoryPermission
	next := ""
	for {
		var page bitbucket.PaginatedRepositoryPermissions
		var res *http.Response
		var err error
		if next == "" {
			page, res, err = workspaceApi.WorkspacesWorkspacePermissionsRepositoriesGet(c.AuthContext, workspace, &options)
		} else {
			res, err = client.GetNext(next, &page)
		}
		if err := handleClientError(res, err); err != nil {
			return diag.FromErr(err)
		}

		permissions = append(permissions, page.Values...)

		if page.Next == "" {
			break
		}
		next = page.Next
	}

	d.SetId(workspace)
	if err := d.Set("permissions", flattenWorkspaceRepositoryPermissions(permissions)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func flattenWorkspaceRepositoryPermissions(permissions []bitbucket.RepositoryPermission) []interface{} {
	tfList := make([]interface{}, 0, len(permissions))
	for _, p := range permissions {
		item := map[string]interface{}{
			"permission": p.Permission,
		}
		if p.Repository != nil {
			item["repository_name"] = p.Repository.Name
			item["repository_full_name"] = p.Repository.FullName
			item["repository_uuid"] = p.Repository.Uuid
		}
		if p.User != nil {
			item["user_uuid"] = p.User.Uuid
			item["user_display_name"] = p.User.DisplayName
			item["user_account_id"] = p.User.AccountId
		}
		tfList = append(tfList, item)
	}
	return tfList
}
