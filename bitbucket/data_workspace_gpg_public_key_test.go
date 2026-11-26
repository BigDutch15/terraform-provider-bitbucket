package bitbucket

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestWorkspaceGpgPublicKeyDataSourceMapsResponse(t *testing.T) {
	rt := &cursorPagesTransport{pages: map[string]string{
		"/2.0/workspaces/ws/settings/gpg/public-key": `{"values":[` +
			`{"type":"system_gpg_key","key":"-----BEGIN PGP PUBLIC KEY BLOCK-----\n\nAAAA\n-----END PGP PUBLIC KEY BLOCK-----","fingerprint":"5f379559a2ea3b74403738964b12c9852d0d55a4","created_on":"2026-01-22T20:09:51+00:00","expires_on":"2027-01-22T20:09:51+00:00"},` +
			`{"type":"system_gpg_key","key":"-----BEGIN PGP PUBLIC KEY BLOCK-----\n\nBBBB\n-----END PGP PUBLIC KEY BLOCK-----","fingerprint":"0123456789abcdef0123456789abcdef01234567","created_on":"2027-01-01T00:00:00+00:00","expires_on":"2028-01-01T00:00:00+00:00"}` +
			`]}`,
	}}

	res := dataWorkspaceGpgPublicKey()
	d := res.TestResourceData()
	if err := d.Set("workspace", "ws"); err != nil {
		t.Fatal(err)
	}
	if diags := res.ReadWithoutTimeout(context.Background(), d, clientsWithTransport(rt)); diags.HasError() {
		t.Fatalf("unexpected error: %v (requests: %v)", diags, rt.seen)
	}

	if d.Id() != "ws" {
		t.Fatalf("got ID %q, want %q", d.Id(), "ws")
	}
	if got := d.Get("keys.#"); got != 2 {
		t.Fatalf("got %v keys, want 2 (requests: %v)", got, rt.seen)
	}
	for attr, want := range map[string]string{
		"keys.0.fingerprint": "5f379559a2ea3b74403738964b12c9852d0d55a4",
		"keys.0.created_on":  "2026-01-22T20:09:51+00:00",
		"keys.0.expires_on":  "2027-01-22T20:09:51+00:00",
		"keys.1.fingerprint": "0123456789abcdef0123456789abcdef01234567",
	} {
		if got := d.Get(attr); got != want {
			t.Errorf("%s = %v, want %v", attr, got, want)
		}
	}
	if got := d.Get("keys.0.key").(string); !strings.HasPrefix(got, "-----BEGIN PGP PUBLIC KEY BLOCK-----\n") {
		t.Errorf("keys.0.key = %q, want an armored key", got)
	}
}

func TestWorkspaceGpgPublicKeyDataSourceReturnsAPIError(t *testing.T) {
	res := dataWorkspaceGpgPublicKey()
	d := res.TestResourceData()
	if err := d.Set("workspace", "ws"); err != nil {
		t.Fatal(err)
	}
	if diags := res.ReadWithoutTimeout(context.Background(), d, clientsWithTransport(notFoundTransport{})); !diags.HasError() {
		t.Fatal("expected an error for a 404 response")
	}
	if d.Id() != "" {
		t.Fatalf("ID set on error: %q", d.Id())
	}
}

func TestAccDataSourceWorkspaceGpgPublicKey_basic(t *testing.T) {
	dataSourceName := "data.bitbucket_workspace_gpg_public_key.test"

	workspace := os.Getenv("BITBUCKET_TEAM")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketWorkspaceGpgPublicKeyConfig(workspace),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "workspace", workspace),
					resource.TestCheckResourceAttrSet(dataSourceName, "keys.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "keys.0.key"),
					resource.TestCheckResourceAttrSet(dataSourceName, "keys.0.fingerprint"),
					resource.TestCheckResourceAttrSet(dataSourceName, "keys.0.created_on"),
					resource.TestCheckResourceAttrSet(dataSourceName, "keys.0.expires_on"),
				),
			},
		},
	})
}

func testAccBitbucketWorkspaceGpgPublicKeyConfig(workspace string) string {
	return fmt.Sprintf(`
data "bitbucket_workspace_gpg_public_key" "test" {
  workspace = %[1]q
}
`, workspace)
}
