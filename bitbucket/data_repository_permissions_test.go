package bitbucket

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestRepositoryPermissionsDataSourceFollowsCursorPages(t *testing.T) {
	rt := &cursorPagesTransport{pages: map[string]string{
		"/2.0/repositories/ws/repo/permissions-config/users":            `{"values":[{"permission":"admin","user":{"uuid":"{a}","display_name":"Alice","account_id":"557058:a"}}],"next":"https://api.bitbucket.org/2.0/repositories/ws/repo/permissions-config/users?page=dXNy"}`,
		"/2.0/repositories/ws/repo/permissions-config/users?page=dXNy":  `{"values":[{"permission":"read","user":{"uuid":"{b}","display_name":"Bob"}}]}`,
		"/2.0/repositories/ws/repo/permissions-config/groups":           `{"values":[{"permission":"write","group":{"slug":"devs","name":"Developers"}}],"next":"https://api.bitbucket.org/2.0/repositories/ws/repo/permissions-config/groups?page=Z3Jw"}`,
		"/2.0/repositories/ws/repo/permissions-config/groups?page=Z3Jw": `{"values":[{"permission":"read","group":{"slug":"ops","name":"Operations"}}]}`,
	}}

	res := dataRepositoryPermissions()
	d := res.TestResourceData()
	for k, v := range map[string]string{"workspace": "ws", "repository": "repo"} {
		if err := d.Set(k, v); err != nil {
			t.Fatal(err)
		}
	}
	if diags := res.ReadWithoutTimeout(context.Background(), d, clientsWithTransport(rt)); diags.HasError() {
		t.Fatalf("unexpected error: %v (requests: %v)", diags, rt.seen)
	}

	if d.Id() != "ws/repo" {
		t.Fatalf("got ID %q, want ws/repo", d.Id())
	}

	want := map[string]interface{}{
		"users.#":              2,
		"users.0.uuid":         "{a}",
		"users.0.display_name": "Alice",
		"users.0.account_id":   "557058:a",
		"users.0.permission":   "admin",
		"users.1.uuid":         "{b}",
		"users.1.account_id":   "",
		"users.1.permission":   "read",
		"groups.#":             2,
		"groups.0.slug":        "devs",
		"groups.0.name":        "Developers",
		"groups.0.permission":  "write",
		"groups.1.slug":        "ops",
		"groups.1.permission":  "read",
	}
	for k, v := range want {
		if got := d.Get(k); got != v {
			t.Errorf("%s = %v, want %v (requests: %v)", k, got, v, rt.seen)
		}
	}
	if len(rt.seen) != 4 {
		t.Errorf("got %d requests, want 4: %v", len(rt.seen), rt.seen)
	}
}

func TestAccDataSourceRepositoryPermissions_basic(t *testing.T) {
	dataSourceName := "data.bitbucket_repository_permissions.test"
	rName := acctest.RandomWithPrefix("tf-test")
	workspace := os.Getenv("BITBUCKET_TEAM")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketRepositoryPermissionsConfig(workspace, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "workspace", workspace),
					resource.TestCheckResourceAttr(dataSourceName, "groups.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "groups.0.slug", "bitbucket_repository_group_permission.test", "group_slug"),
					resource.TestCheckResourceAttr(dataSourceName, "groups.0.permission", "read"),
				),
			},
		},
	})
}

// testAccBitbucketRepositoryPermissionsConfig grants an existing workspace group
// read access instead of creating a group: workspaces whose groups are managed
// in Atlassian Admin reject group creation through the 1.0 groups API (HTTP 405).
func testAccBitbucketRepositoryPermissionsConfig(workspace, rName string) string {
	return fmt.Sprintf(`
resource "bitbucket_repository" "test" {
  owner = %[1]q
  name  = %[2]q
}

data "bitbucket_groups" "test" {
  workspace = %[1]q
}

resource "bitbucket_repository_group_permission" "test" {
  workspace  = %[1]q
  repo_slug  = bitbucket_repository.test.name
  group_slug = sort([for g in data.bitbucket_groups.test.groups : g.slug])[0]
  permission = "read"
}

data "bitbucket_repository_permissions" "test" {
  workspace  = %[1]q
  repository = bitbucket_repository.test.name

  depends_on = [bitbucket_repository_group_permission.test]
}
`, workspace, rName)
}
