package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// EffectiveBranchingModel is the response of
// GET /repositories/{workspace}/{repo_slug}/effective-branching-model.
// The generated client models development, production and branch_types as
// untyped interface{} values, so the raw client decodes into this struct.
type EffectiveBranchingModel struct {
	Development *EffectiveBranch      `json:"development,omitempty"`
	Production  *EffectiveBranch      `json:"production,omitempty"`
	BranchTypes []EffectiveBranchType `json:"branch_types,omitempty"`
}

type EffectiveBranch struct {
	Name               string `json:"name"`
	UseMainbranch      bool   `json:"use_mainbranch"`
	BranchDoesNotExist bool   `json:"branch_does_not_exist,omitempty"`
}

type EffectiveBranchType struct {
	Kind   string `json:"kind"`
	Prefix string `json:"prefix"`
}

func dataEffectiveBranchingModel() *schema.Resource {
	branchSchema := func(kind string) *schema.Schema {
		return &schema.Schema{
			Type:        schema.TypeList,
			Computed:    true,
			Description: fmt.Sprintf("The %s branch in effect. Empty when the effective model has no %s branch.", kind, kind),
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: fmt.Sprintf("Name of the %s branch.", kind),
					},
					"use_mainbranch": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: fmt.Sprintf("Whether the repository's main branch is used as the %s branch.", kind),
					},
					"branch_does_not_exist": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: fmt.Sprintf("Whether the configured %s branch is missing from the repository.", kind),
					},
				},
			},
		}
	}

	return &schema.Resource{
		Description:        "Gets the branching model in effect for a repository, after inheritance from its project.",
		ReadWithoutTimeout: dataReadEffectiveBranchingModel,

		Schema: map[string]*schema.Schema{
			"workspace": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The workspace ID (slug) or UUID.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"repository": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The repository slug or UUID.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"development": branchSchema("development"),
			"production":  branchSchema("production"),
			"branch_type": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The active branch types, in the order the API returns them.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kind": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The kind of branch: `feature`, `bugfix`, `release` or `hotfix`.",
						},
						"prefix": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The prefix for this branch type.",
						},
					},
				},
			},
		},
	}
}

func dataReadEffectiveBranchingModel(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(Clients).httpClient

	workspace := d.Get("workspace").(string)
	repoSlug := d.Get("repository").(string)

	res, err := client.Get(fmt.Sprintf("2.0/repositories/%s/%s/effective-branching-model",
		url.PathEscape(workspace), url.PathEscape(repoSlug)))
	if err != nil {
		return diag.FromErr(err)
	}

	if res.Body == nil {
		return diag.Errorf("error reading effective branching model for %s/%s: empty response", workspace, repoSlug)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Effective Branching Model Response JSON: %v", string(body))

	var model EffectiveBranchingModel
	if err := json.Unmarshal(body, &model); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s/%s", workspace, repoSlug))
	if err := d.Set("development", flattenEffectiveBranch(model.Development)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("production", flattenEffectiveBranch(model.Production)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("branch_type", flattenEffectiveBranchTypes(model.BranchTypes)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func flattenEffectiveBranch(b *EffectiveBranch) []interface{} {
	if b == nil {
		return []interface{}{}
	}
	return []interface{}{map[string]interface{}{
		"name":                  b.Name,
		"use_mainbranch":        b.UseMainbranch,
		"branch_does_not_exist": b.BranchDoesNotExist,
	}}
}

func flattenEffectiveBranchTypes(types []EffectiveBranchType) []interface{} {
	tfList := make([]interface{}, 0, len(types))
	for _, bt := range types {
		tfList = append(tfList, map[string]interface{}{
			"kind":   bt.Kind,
			"prefix": bt.Prefix,
		})
	}
	return tfList
}
