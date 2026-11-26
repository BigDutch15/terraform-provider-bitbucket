package bitbucket

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestProjectsReadFollowsCursorPages(t *testing.T) {
	rt := &cursorPagesTransport{pages: map[string]string{
		"/2.0/workspaces/ws/projects":             `{"values":[{"key":"AAA","uuid":"{a}","name":"Alpha","description":"first","is_private":true}],"next":"https://api.bitbucket.org/2.0/workspaces/ws/projects?page=bmV4dA"}`,
		"/2.0/workspaces/ws/projects?page=bmV4dA": `{"values":[{"key":"BBB","uuid":"{b}","name":"Beta","has_publicly_visible_repos":true}]}`,
	}}

	res := dataProjects()
	d := res.TestResourceData()
	if err := d.Set("workspace", "ws"); err != nil {
		t.Fatal(err)
	}
	if diags := res.ReadWithoutTimeout(context.Background(), d, clientsWithTransport(rt)); diags.HasError() {
		t.Fatalf("unexpected error: %v (requests: %v)", diags, rt.seen)
	}
	if len(rt.seen) != 2 {
		t.Fatalf("got %d requests, want 2: %v", len(rt.seen), rt.seen)
	}
	if got := d.Get("projects.#"); got != 2 {
		t.Fatalf("got %v projects, want 2 (requests: %v)", got, rt.seen)
	}
	if d.Id() != "ws" {
		t.Errorf("id = %q, want %q", d.Id(), "ws")
	}
	for attr, want := range map[string]interface{}{
		"projects.0.key":                        "AAA",
		"projects.0.uuid":                       "{a}",
		"projects.0.name":                       "Alpha",
		"projects.0.description":                "first",
		"projects.0.is_private":                 true,
		"projects.0.has_publicly_visible_repos": false,
		"projects.1.key":                        "BBB",
		"projects.1.name":                       "Beta",
		"projects.1.is_private":                 false,
		"projects.1.has_publicly_visible_repos": true,
	} {
		if got := d.Get(attr); got != want {
			t.Errorf("%s = %v, want %v", attr, got, want)
		}
	}
}

func TestAccDataSourceProjects_basic(t *testing.T) {
	dataSourceName := "data.bitbucket_projects.test"
	resourceName := "bitbucket_project.test"
	workspace := os.Getenv("BITBUCKET_TEAM")
	rName := acctest.RandomWithPrefix("tf-test")
	rKey := fmt.Sprintf("TF%s", strings.ToUpper(acctest.RandString(8)))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketProjectsConfig(workspace, rName, rKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "workspace", workspace),
					resource.TestCheckResourceAttrSet(dataSourceName, "projects.#"),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "projects.*", map[string]string{
						"key":  rKey,
						"name": rName,
					}),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "projects.*.uuid", resourceName, "uuid"),
				),
			},
		},
	})
}

func testAccBitbucketProjectsConfig(workspace, rName, rKey string) string {
	return fmt.Sprintf(`
resource "bitbucket_project" "test" {
  owner = %[1]q
  name  = %[2]q
  key   = %[3]q
}

data "bitbucket_projects" "test" {
  workspace = %[1]q

  depends_on = [bitbucket_project.test]
}
`, workspace, rName, rKey)
}
