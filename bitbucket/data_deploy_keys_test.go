package bitbucket

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestDeployKeysDataSourceFollowsCursorPages(t *testing.T) {
	rt := &cursorPagesTransport{pages: map[string]string{
		"/2.0/repositories/ws/repo/deploy-keys":             `{"values":[{"id":123,"key":"ssh-ed25519 AAAA","label":"ci","comment":"ci@example","created_on":"2024-01-02T03:04:05.000000+00:00","last_used":null}],"next":"https://api.bitbucket.org/2.0/repositories/ws/repo/deploy-keys?page=a2V5cw"}`,
		"/2.0/repositories/ws/repo/deploy-keys?page=a2V5cw": `{"values":[{"id":456,"key":"ssh-rsa BBBB","label":"deploy","added_on":"2024-02-03T04:05:06.000000+00:00","last_used":"2024-03-04T05:06:07.000000+00:00"}]}`,
	}}

	res := dataDeployKeys()
	d := res.TestResourceData()
	for k, v := range map[string]string{"workspace": "ws", "repository": "repo"} {
		if err := d.Set(k, v); err != nil {
			t.Fatal(err)
		}
	}
	if diags := res.ReadWithoutTimeout(context.Background(), d, clientsWithTransport(rt)); diags.HasError() {
		t.Fatalf("unexpected error: %v (requests: %v)", diags, rt.seen)
	}
	if len(rt.seen) != 2 {
		t.Fatalf("got %d requests, want 2: %v", len(rt.seen), rt.seen)
	}
	if got := d.Id(); got != "ws/repo" {
		t.Errorf("id = %q, want ws/repo", got)
	}
	if got := d.Get("deploy_keys.#"); got != 2 {
		t.Fatalf("got %v deploy keys, want 2 (requests: %v)", got, rt.seen)
	}
	for k, want := range map[string]string{
		"deploy_keys.0.key_id":     "123",
		"deploy_keys.0.key":        "ssh-ed25519 AAAA",
		"deploy_keys.0.label":      "ci",
		"deploy_keys.0.comment":    "ci@example",
		"deploy_keys.0.created_on": "2024-01-02T03:04:05.000000+00:00",
		"deploy_keys.0.last_used":  "",
		"deploy_keys.1.key_id":     "456",
		"deploy_keys.1.label":      "deploy",
		"deploy_keys.1.created_on": "2024-02-03T04:05:06.000000+00:00",
		"deploy_keys.1.last_used":  "2024-03-04T05:06:07.000000+00:00",
	} {
		if got := d.Get(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}
}

func TestDeployKeysDataSourceReturnsAPIError(t *testing.T) {
	rt := &cursorPagesTransport{pages: map[string]string{}}

	res := dataDeployKeys()
	d := res.TestResourceData()
	for k, v := range map[string]string{"workspace": "ws", "repository": "repo"} {
		if err := d.Set(k, v); err != nil {
			t.Fatal(err)
		}
	}
	if diags := res.ReadWithoutTimeout(context.Background(), d, clientsWithTransport(rt)); !diags.HasError() {
		t.Fatal("expected an error for a non-2xx response")
	}
}

func TestAccDataSourceDeployKeys_basic(t *testing.T) {
	dataSourceName := "data.bitbucket_deploy_keys.test"
	resourceName := "bitbucket_deploy_key.test"

	rName := acctest.RandomWithPrefix("tf-test")
	workspace := os.Getenv("BITBUCKET_TEAM")
	publicKey, _, err := RandSSHKeyPairSize(2048, "tf-test@example.com")
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketDeployKeysConfig(workspace, rName, publicKey),
			},
			{
				Config: testAccBitbucketDeployKeysConfig(workspace, rName, publicKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "workspace", workspace),
					resource.TestCheckResourceAttr(dataSourceName, "deploy_keys.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "deploy_keys.0.key_id", resourceName, "key_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "deploy_keys.0.label", resourceName, "label"),
					resource.TestCheckResourceAttrPair(dataSourceName, "deploy_keys.0.comment", resourceName, "comment"),
					resource.TestCheckResourceAttrSet(dataSourceName, "deploy_keys.0.key"),
					resource.TestCheckResourceAttrSet(dataSourceName, "deploy_keys.0.created_on"),
				),
			},
		},
	})
}

func testAccBitbucketDeployKeysConfig(workspace, rName, pubkey string) string {
	return testAccBitbucketDeployKeyLabelConfig(workspace, rName, pubkey, "tf-test") + fmt.Sprintf(`
data "bitbucket_deploy_keys" "test" {
  workspace  = %[1]q
  repository = bitbucket_repository.test.name

  depends_on = [bitbucket_deploy_key.test]
}
`, workspace)
}
