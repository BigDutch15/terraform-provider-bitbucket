package bitbucket

import (
	"context"
	"fmt"
	"net/http"

	"github.com/DrFaust92/bitbucket-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataProjectPermissions() *schema.Resource {
	return &schema.Resource{
		Description:        "Lists the users and groups that have explicit permissions on a project.",
		ReadWithoutTimeout: dataReadProjectPermissions,

		Schema: map[string]*schema.Schema{
			"workspace": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The workspace ID (slug) or UUID that owns the project.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"project_key": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The key of the project.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"users": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Users with explicit permissions on the project, in the order the API returns them.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uuid": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The UUID of the user.",
						},
						"display_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The display name of the user.",
						},
						"account_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The Atlassian account ID of the user, when returned by the API.",
						},
						"permission": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The permission the user has on the project: `read`, `write`, `create-repo` or `admin`.",
						},
					},
				},
			},
			"groups": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Groups with explicit permissions on the project, in the order the API returns them.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"slug": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The slug of the group.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the group.",
						},
						"permission": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The permission the group has on the project: `read`, `write`, `create-repo` or `admin`.",
						},
					},
				},
			},
		},
	}
}

func dataReadProjectPermissions(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(Clients).genClient
	projectsApi := c.ApiClient.ProjectsApi
	client := m.(Clients).httpClient

	workspace := d.Get("workspace").(string)
	projectKey := d.Get("project_key").(string)

	users := make([]interface{}, 0)
	next := ""
	for {
		var page bitbucket.PaginatedProjectUserPermissions
		var res *http.Response
		var err error
		if next == "" {
			page, res, err = projectsApi.WorkspacesWorkspaceProjectsProjectKeyPermissionsConfigUsersGet(c.AuthContext, projectKey, workspace, nil)
		} else {
			res, err = client.GetNext(next, &page)
		}
		if err := handleClientError(res, err); err != nil {
			return diag.FromErr(err)
		}

		for _, perm := range page.Values {
			user := map[string]interface{}{
				"uuid":         "",
				"display_name": "",
				"account_id":   "",
				"permission":   perm.Permission,
			}
			if perm.User != nil {
				user["uuid"] = perm.User.Uuid
				user["display_name"] = perm.User.DisplayName
				user["account_id"] = perm.User.AccountId
			}
			users = append(users, user)
		}

		if page.Next == "" {
			break
		}
		next = page.Next
	}

	groups := make([]interface{}, 0)
	next = ""
	for {
		var page bitbucket.PaginatedProjectGroupPermissions
		var res *http.Response
		var err error
		if next == "" {
			page, res, err = projectsApi.WorkspacesWorkspaceProjectsProjectKeyPermissionsConfigGroupsGet(c.AuthContext, projectKey, workspace, nil)
		} else {
			res, err = client.GetNext(next, &page)
		}
		if err := handleClientError(res, err); err != nil {
			return diag.FromErr(err)
		}

		for _, perm := range page.Values {
			group := map[string]interface{}{
				"slug":       "",
				"name":       "",
				"permission": perm.Permission,
			}
			if perm.Group != nil {
				group["slug"] = perm.Group.Slug
				group["name"] = perm.Group.Name
			}
			groups = append(groups, group)
		}

		if page.Next == "" {
			break
		}
		next = page.Next
	}

	d.SetId(fmt.Sprintf("%s/%s", workspace, projectKey))
	if err := d.Set("users", users); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("groups", groups); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
