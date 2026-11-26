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

func dataRepositoryPermissions() *schema.Resource {
	return &schema.Resource{
		Description:        "Lists the users and groups that have explicit permissions on a repository.",
		ReadWithoutTimeout: dataReadRepositoryPermissions,

		Schema: map[string]*schema.Schema{
			"workspace": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The workspace slug or UUID.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"repository": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The repository slug or UUID.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"users": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Users with explicit permissions on the repository, in the order the API returns them.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uuid": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The user's UUID.",
						},
						"display_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The user's display name.",
						},
						"account_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The user's Atlassian account ID, if returned.",
						},
						"permission": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The permission the user has on the repository: `read`, `write` or `admin`.",
						},
					},
				},
			},
			"groups": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Groups with explicit permissions on the repository, in the order the API returns them.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"slug": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The group's slug.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The group's name.",
						},
						"permission": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The permission the group has on the repository: `read`, `write` or `admin`.",
						},
					},
				},
			},
		},
	}
}

func dataReadRepositoryPermissions(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(Clients).genClient
	repoApi := c.ApiClient.RepositoriesApi
	client := m.(Clients).httpClient

	workspace := d.Get("workspace").(string)
	repoSlug := d.Get("repository").(string)

	users := make([]interface{}, 0)
	next := ""
	for {
		var page bitbucket.PaginatedRepositoryUserPermissions
		var res *http.Response
		var err error
		if next == "" {
			page, res, err = repoApi.RepositoriesWorkspaceRepoSlugPermissionsConfigUsersGet(c.AuthContext, repoSlug, workspace, nil)
		} else {
			res, err = client.GetNext(next, &page)
		}
		if err := handleClientError(res, err); err != nil {
			return diag.FromErr(err)
		}

		for _, perm := range page.Values {
			user := map[string]interface{}{"permission": perm.Permission}
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
		var page bitbucket.PaginatedRepositoryGroupPermissions
		var res *http.Response
		var err error
		if next == "" {
			page, res, err = repoApi.RepositoriesWorkspaceRepoSlugPermissionsConfigGroupsGet(c.AuthContext, repoSlug, workspace, nil)
		} else {
			res, err = client.GetNext(next, &page)
		}
		if err := handleClientError(res, err); err != nil {
			return diag.FromErr(err)
		}

		for _, perm := range page.Values {
			group := map[string]interface{}{"permission": perm.Permission}
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

	d.SetId(fmt.Sprintf("%s/%s", workspace, repoSlug))
	if err := d.Set("users", users); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("groups", groups); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
