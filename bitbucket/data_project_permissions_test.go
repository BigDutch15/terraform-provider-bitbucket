package bitbucket

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestProjectPermissionsReadFollowsCursorPages(t *testing.T) {
	rt := &cursorPagesTransport{pages: map[string]string{
		"/2.0/workspaces/ws/projects/PRJ/permissions-config/users":                `{"values":[{"permission":"admin","user":{"uuid":"{a}","display_name":"Alice","account_id":"557058:a"}}],"next":"https://api.bitbucket.org/2.0/workspaces/ws/projects/PRJ/permissions-config/users?page=dXNlcnM"}`,
		"/2.0/workspaces/ws/projects/PRJ/permissions-config/users?page=dXNlcnM":   `{"values":[{"permission":"read","user":{"uuid":"{b}","display_name":"Bob"}}]}`,
		"/2.0/workspaces/ws/projects/PRJ/permissions-config/groups":               `{"values":[{"permission":"write","group":{"slug":"devs","name":"Devs"}}],"next":"https://api.bitbucket.org/2.0/workspaces/ws/projects/PRJ/permissions-config/groups?page=Z3JvdXBz"}`,
		"/2.0/workspaces/ws/projects/PRJ/permissions-config/groups?page=Z3JvdXBz": `{"values":[{"permission":"create-repo","group":{"slug":"ops","name":"Ops"}}]}`,
	}}

	res := dataProjectPermissions()
	d := res.TestResourceData()
	if err := d.Set("workspace", "ws"); err != nil {
		t.Fatal(err)
	}
	if err := d.Set("project_key", "PRJ"); err != nil {
		t.Fatal(err)
	}
	if diags := res.ReadWithoutTimeout(context.Background(), d, clientsWithTransport(rt)); diags.HasError() {
		t.Fatalf("unexpected error: %v (requests: %v)", diags, rt.seen)
	}

	if got := d.Id(); got != "ws/PRJ" {
		t.Errorf("got ID %q, want %q", got, "ws/PRJ")
	}
	if got := d.Get("users.#"); got != 2 {
		t.Fatalf("got %v users, want 2 (requests: %v)", got, rt.seen)
	}
	if got := d.Get("groups.#"); got != 2 {
		t.Fatalf("got %v groups, want 2 (requests: %v)", got, rt.seen)
	}

	for attr, want := range map[string]string{
		"users.0.uuid":         "{a}",
		"users.0.display_name": "Alice",
		"users.0.account_id":   "557058:a",
		"users.0.permission":   "admin",
		"users.1.uuid":         "{b}",
		"users.1.account_id":   "",
		"users.1.permission":   "read",
		"groups.0.slug":        "devs",
		"groups.0.name":        "Devs",
		"groups.0.permission":  "write",
		"groups.1.slug":        "ops",
		"groups.1.permission":  "create-repo",
	} {
		if got := d.Get(attr); got != want {
			t.Errorf("%s = %q, want %q", attr, got, want)
		}
	}
}

func TestAccDataSourceProjectPermissions_basic(t *testing.T) {
	dataSourceName := "data.bitbucket_project_permissions.test"
	groupPermissionName := "bitbucket_project_group_permission.test"

	rName := acctest.RandomWithPrefix("tf-test")
	workspace := os.Getenv("BITBUCKET_TEAM")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketProjectPermissionsDataConfig(workspace, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "workspace", workspace),
					resource.TestCheckResourceAttr(dataSourceName, "project_key", "DSPRJPERM"),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "groups.*", map[string]string{
						"permission": "read",
					}),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "groups.*.slug", groupPermissionName, "group_slug"),
				),
			},
		},
	})
}

// testAccBitbucketProjectPermissionsDataConfig grants an existing workspace group
// read access instead of creating a group: creating one through the 1.0 groups
// API needs the account:write OAuth scope, which Atlassian API tokens can't be
// granted, so it fails with HTTP 405 (DrFaust92/terraform-provider-bitbucket#258).
func testAccBitbucketProjectPermissionsDataConfig(workspace, rName string) string {
	return fmt.Sprintf(`
resource "bitbucket_project" "test" {
  owner = %[1]q
  name  = %[2]q
  key   = "DSPRJPERM"
}

data "bitbucket_groups" "test" {
  workspace = %[1]q
}

resource "bitbucket_project_group_permission" "test" {
  workspace   = %[1]q
  project_key = bitbucket_project.test.key
  group_slug  = sort([for g in data.bitbucket_groups.test.groups : g.slug])[0]
  permission  = "read"
}

data "bitbucket_project_permissions" "test" {
  workspace   = %[1]q
  project_key = bitbucket_project.test.key

  depends_on = [bitbucket_project_group_permission.test]
}
`, workspace, rName)
}
