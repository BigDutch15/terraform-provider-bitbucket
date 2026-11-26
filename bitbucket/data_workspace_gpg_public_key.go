package bitbucket

import (
	"context"
	"encoding/json"
	"log"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// workspaceGpgPublicKeys is the response of
// GET /2.0/workspaces/{workspace}/settings/gpg/public-key. The generated
// client discards the body of this endpoint, so it is read with the raw client.
type workspaceGpgPublicKeys struct {
	Values []workspaceGpgPublicKey `json:"values"`
}

type workspaceGpgPublicKey struct {
	Key         string `json:"key"`
	Fingerprint string `json:"fingerprint"`
	CreatedOn   string `json:"created_on"`
	ExpiresOn   string `json:"expires_on"`
}

func dataWorkspaceGpgPublicKey() *schema.Resource {
	return &schema.Resource{
		Description:        "Fetches the system GPG public key(s) Bitbucket uses to sign commits made in the web interface of a workspace.",
		ReadWithoutTimeout: dataReadWorkspaceGpgPublicKey,

		Schema: map[string]*schema.Schema{
			"workspace": {
				Description:  "The workspace ID (slug) or the workspace UUID surrounded by curly braces.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"keys": {
				Description: "The system GPG public keys, in the order the API returns them. Usually one key; two during a key rotation.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Description: "The ASCII-armored GPG public key.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"fingerprint": {
							Description: "The GPG key fingerprint.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"created_on": {
							Description: "When the key was created, in ISO 8601 format.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"expires_on": {
							Description: "When the key expires, in ISO 8601 format.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataReadWorkspaceGpgPublicKey(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(Clients).httpClient

	workspace := d.Get("workspace").(string)

	res, err := client.Get("2.0/workspaces/" + url.PathEscape(workspace) + "/settings/gpg/public-key")
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()

	var keys workspaceGpgPublicKeys
	if err := json.NewDecoder(res.Body).Decode(&keys); err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Workspace GPG Public Keys Response Decoded: %#v", keys)

	d.SetId(workspace)
	d.Set("workspace", workspace)
	if err := d.Set("keys", flattenWorkspaceGpgPublicKeys(keys.Values)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func flattenWorkspaceGpgPublicKeys(keys []workspaceGpgPublicKey) []interface{} {
	tfList := make([]interface{}, 0, len(keys))
	for _, k := range keys {
		tfList = append(tfList, map[string]interface{}{
			"key":         k.Key,
			"fingerprint": k.Fingerprint,
			"created_on":  k.CreatedOn,
			"expires_on":  k.ExpiresOn,
		})
	}
	return tfList
}
