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

func dataRepositoryVariables() *schema.Resource {
	return &schema.Resource{
		Description:        "Lists the Pipelines variables of a repository.",
		ReadWithoutTimeout: dataReadRepositoryVariables,

		Schema: map[string]*schema.Schema{
			"workspace": {
				Description:  "The workspace ID (slug) or UUID.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"repository": {
				Description:  "The repository slug.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"variables": pipelineVariablesListSchema("The repository's Pipelines variables, in the order the API returns them."),
		},
	}
}

// pipelineVariablesListSchema is the computed list of Pipelines variables
// shared by the repository and workspace variables data sources.
func pipelineVariablesListSchema(description string) *schema.Schema {
	return &schema.Schema{
		Description: description,
		Type:        schema.TypeList,
		Computed:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"uuid": {
					Description: "The UUID of the variable.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"key": {
					Description: "The name of the variable.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"value": {
					Description: "The value of the variable. Empty when the variable is secured, because the API does not return secured values.",
					Type:        schema.TypeString,
					Computed:    true,
					Sensitive:   true,
				},
				"secured": {
					Description: "Whether the variable is secured.",
					Type:        schema.TypeBool,
					Computed:    true,
				},
			},
		},
	}
}

func flattenPipelineVariables(vars []bitbucket.PipelineVariable) []interface{} {
	out := make([]interface{}, 0, len(vars))
	for _, v := range vars {
		value := v.Value
		if v.Secured {
			value = ""
		}
		out = append(out, map[string]interface{}{
			"uuid":    v.Uuid,
			"key":     v.Key,
			"value":   value,
			"secured": v.Secured,
		})
	}
	return out
}

func dataReadRepositoryVariables(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(Clients).genClient
	pipeApi := c.ApiClient.PipelinesApi
	client := m.(Clients).httpClient

	workspace := d.Get("workspace").(string)
	repoSlug := d.Get("repository").(string)

	var vars []bitbucket.PipelineVariable
	next := ""
	for {
		var page bitbucket.PaginatedPipelineVariables
		var res *http.Response
		var err error
		if next == "" {
			page, res, err = pipeApi.GetRepositoryPipelineVariables(c.AuthContext, workspace, repoSlug, nil)
		} else {
			res, err = client.GetNext(next, &page)
		}
		if err := handleClientError(res, err); err != nil {
			return diag.FromErr(err)
		}

		vars = append(vars, page.Values...)

		if page.Next == "" {
			break
		}
		next = page.Next
	}

	d.SetId(fmt.Sprintf("%s/%s", workspace, repoSlug))
	if err := d.Set("variables", flattenPipelineVariables(vars)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
