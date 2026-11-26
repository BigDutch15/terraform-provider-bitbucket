package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// deployKeyListItem is a repository access key as returned by
// GET /repositories/{workspace}/{repo_slug}/deploy-keys. The generated
// client's DeployKey model has no id field, so the raw client is used.
type deployKeyListItem struct {
	ID        int    `json:"id"`
	Key       string `json:"key"`
	Label     string `json:"label"`
	Comment   string `json:"comment"`
	CreatedOn string `json:"created_on"`
	AddedOn   string `json:"added_on"`
	LastUsed  string `json:"last_used"`
}

type paginatedDeployKeyList struct {
	Values []deployKeyListItem `json:"values"`
	Next   string              `json:"next"`
}

func dataDeployKeys() *schema.Resource {
	return &schema.Resource{
		Description:        "Lists the access (deploy) keys of a repository.",
		ReadWithoutTimeout: dataReadDeployKeys,

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
				Description:  "The repository slug.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"deploy_keys": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The repository's deploy keys, in the order the API returns them.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the deploy key.",
						},
						"key": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The public SSH key.",
						},
						"label": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The user-defined label of the deploy key.",
						},
						"comment": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The comment parsed from the key, if present.",
						},
						"created_on": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "When the deploy key was added.",
						},
						"last_used": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "When the deploy key was last used, empty if never.",
						},
					},
				},
			},
		},
	}
}

func dataReadDeployKeys(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(Clients).httpClient

	workspace := d.Get("workspace").(string)
	repo := d.Get("repository").(string)

	var keys []deployKeyListItem
	next := ""
	for {
		var page paginatedDeployKeyList
		var res *http.Response
		var err error
		if next == "" {
			res, err = client.Get(fmt.Sprintf("2.0/repositories/%s/%s/deploy-keys",
				url.PathEscape(workspace), url.PathEscape(repo)))
			if err == nil {
				err = json.NewDecoder(res.Body).Decode(&page)
				res.Body.Close()
			}
		} else {
			res, err = client.GetNext(next, &page)
		}
		if err := handleClientError(res, err); err != nil {
			return diag.Errorf("error reading Deploy Keys for %s/%s: %s", workspace, repo, err)
		}

		keys = append(keys, page.Values...)

		if page.Next == "" {
			break
		}
		next = page.Next
	}

	d.SetId(fmt.Sprintf("%s/%s", workspace, repo))
	if err := d.Set("deploy_keys", flattenDeployKeyList(keys)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func flattenDeployKeyList(keys []deployKeyListItem) []interface{} {
	tfList := make([]interface{}, 0, len(keys))
	for _, k := range keys {
		createdOn := k.CreatedOn
		if createdOn == "" {
			createdOn = k.AddedOn
		}
		tfList = append(tfList, map[string]interface{}{
			"key_id":     strconv.Itoa(k.ID),
			"key":        k.Key,
			"label":      k.Label,
			"comment":    k.Comment,
			"created_on": createdOn,
			"last_used":  k.LastUsed,
		})
	}

	return tfList
}
