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

func dataBranchRestrictions() *schema.Resource {
	return &schema.Resource{
		Description:        "Lists the branch restrictions of a repository.",
		ReadWithoutTimeout: dataReadBranchRestrictions,

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
			"kind": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Only return branch restrictions of this kind, e.g. `push`.",
			},
			"pattern": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Only return branch restrictions that apply to branches matching this pattern.",
			},
			"branch_restrictions": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The repository's branch restrictions, in the order the API returns them.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The branch restriction ID.",
						},
						"kind": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The kind of restriction.",
						},
						"branch_match_kind": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "How the restriction is matched against a branch: `glob` or `branching_model`.",
						},
						"branch_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The branch type the restriction applies to when `branch_match_kind` is `branching_model`.",
						},
						"pattern": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The branch pattern the restriction applies to when `branch_match_kind` is `glob`.",
						},
						"value": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The kind-specific value, e.g. the number of approvals required. 0 when not set.",
						},
						"users": {
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
							Description: "UUIDs of the users the restriction exempts or applies to.",
						},
						"groups": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Groups the restriction exempts or applies to.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"owner": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The workspace slug that owns the group (the workspace UUID if no slug is returned).",
									},
									"slug": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The group slug.",
									},
								},
							},
						},
						"pipelines_source_branches": {
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
							Description: "Source branches allowed to push back to the restricted branch from Pipelines.",
						},
					},
				},
			},
		},
	}
}

func dataReadBranchRestrictions(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(Clients).genClient
	brApi := c.ApiClient.BranchRestrictionsApi
	client := m.(Clients).httpClient

	workspace := d.Get("workspace").(string)
	repoSlug := d.Get("repository").(string)

	opts := &bitbucket.BranchRestrictionsApiRepositoriesWorkspaceRepoSlugBranchRestrictionsGetOpts{}
	if v, ok := d.GetOk("kind"); ok {
		opts.Kind = optional.NewString(v.(string))
	}
	if v, ok := d.GetOk("pattern"); ok {
		opts.Pattern = optional.NewString(v.(string))
	}

	var restrictions []bitbucket.Branchrestriction
	next := ""
	for {
		var page bitbucket.PaginatedBranchrestrictions
		var res *http.Response
		var err error
		if next == "" {
			page, res, err = brApi.RepositoriesWorkspaceRepoSlugBranchRestrictionsGet(c.AuthContext, repoSlug, workspace, opts)
		} else {
			res, err = client.GetNext(next, &page)
		}
		if err := handleClientError(res, err); err != nil {
			return diag.FromErr(err)
		}

		restrictions = append(restrictions, page.Values...)

		if page.Next == "" {
			break
		}
		next = page.Next
	}

	d.SetId(fmt.Sprintf("%s/%s", workspace, repoSlug))
	if err := d.Set("branch_restrictions", flattenBranchRestrictionList(restrictions)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting branch_restrictions: %w", err))
	}

	return nil
}

// flattenBranchRestrictionList maps branch restrictions to the
// branch_restrictions list, reusing the resource's user and group flatteners.
func flattenBranchRestrictionList(restrictions []bitbucket.Branchrestriction) []interface{} {
	flattened := make([]interface{}, 0, len(restrictions))
	for _, br := range restrictions {
		sourceBranches := br.PipelinesSourceBranches
		if sourceBranches == nil {
			sourceBranches = []string{}
		}
		flattened = append(flattened, map[string]interface{}{
			"id":                        fmt.Sprintf("%d", br.Id),
			"kind":                      br.Kind,
			"branch_match_kind":         br.BranchMatchKind,
			"branch_type":               br.BranchType,
			"pattern":                   br.Pattern,
			"value":                     int(br.Value),
			"users":                     flattenBranchRestrictionUsers(br.Users),
			"groups":                    flattenBranchRestrictionGroups(br.Groups, nil),
			"pipelines_source_branches": sourceBranches,
		})
	}
	return flattened
}
