package bitbucket

import (
	"context"
	"net/http"

	"github.com/DrFaust92/bitbucket-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataWorkspaceVariables() *schema.Resource {
	return &schema.Resource{
		Description:        "Lists the Pipelines variables of a workspace.",
		ReadWithoutTimeout: dataReadWorkspaceVariables,

		Schema: map[string]*schema.Schema{
			"workspace": {
				Description:  "The workspace ID (slug) or UUID.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"variables": pipelineVariablesListSchema("The workspace's Pipelines variables, in the order the API returns them."),
		},
	}
}

func dataReadWorkspaceVariables(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(Clients).genClient
	pipeApi := c.ApiClient.PipelinesApi
	client := m.(Clients).httpClient

	workspace := d.Get("workspace").(string)

	var vars []bitbucket.PipelineVariable
	next := ""
	for {
		var page bitbucket.PaginatedPipelineVariables
		var res *http.Response
		var err error
		if next == "" {
			page, res, err = pipeApi.GetPipelineVariablesForWorkspace(c.AuthContext, workspace, nil)
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

	d.SetId(workspace)
	if err := d.Set("variables", flattenPipelineVariables(vars)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
