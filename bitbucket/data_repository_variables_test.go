package bitbucket

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestRepositoryVariablesDataSourceFollowsCursorPages(t *testing.T) {
	rt := &cursorPagesTransport{pages: map[string]string{
		"/2.0/repositories/ws/repo/pipelines_config/variables":           `{"values":[{"uuid":"{a}","key":"PLAIN","value":"hello","secured":false}],"next":"https://api.bitbucket.org/2.0/repositories/ws/repo/pipelines_config/variables?page=dmFy"}`,
		"/2.0/repositories/ws/repo/pipelines_config/variables?page=dmFy": `{"values":[{"uuid":"{b}","key":"SECRET","secured":true}]}`,
	}}

	res := dataRepositoryVariables()
	d := res.TestResourceData()
	for k, v := range map[string]string{"workspace": "ws", "repository": "repo"} {
		if err := d.Set(k, v); err != nil {
			t.Fatal(err)
		}
	}
	if diags := res.ReadWithoutTimeout(context.Background(), d, clientsWithTransport(rt)); diags.HasError() {
		t.Fatalf("unexpected error: %v (requests: %v)", diags, rt.seen)
	}
	if got := d.Get("variables.#"); got != 2 {
		t.Fatalf("got %v variables, want 2 (requests: %v)", got, rt.seen)
	}
	if got := d.Id(); got != "ws/repo" {
		t.Errorf("id = %q, want ws/repo", got)
	}
	want := map[string]interface{}{
		"variables.0.uuid":    "{a}",
		"variables.0.key":     "PLAIN",
		"variables.0.value":   "hello",
		"variables.0.secured": false,
		"variables.1.uuid":    "{b}",
		"variables.1.key":     "SECRET",
		"variables.1.value":   "",
		"variables.1.secured": true,
	}
	for k, v := range want {
		if got := d.Get(k); got != v {
			t.Errorf("%s = %#v, want %#v", k, got, v)
		}
	}
}

func TestAccDataSourceRepositoryVariables_basic(t *testing.T) {
	dataSourceName := "data.bitbucket_repository_variables.test"

	rName := acctest.RandomWithPrefix("tf-test")
	workspace := os.Getenv("BITBUCKET_TEAM")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketRepositoryVariablesDataConfig(workspace, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "workspace", workspace),
					resource.TestCheckResourceAttr(dataSourceName, "repository", rName),
					resource.TestCheckResourceAttr(dataSourceName, "variables.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "variables.*", map[string]string{
						"key":     "PLAIN",
						"value":   "plain-value",
						"secured": "false",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "variables.*", map[string]string{
						"key":     "SECRET",
						"value":   "",
						"secured": "true",
					}),
				),
			},
		},
	})
}

func testAccBitbucketRepositoryVariablesDataConfig(workspace, rName string) string {
	return fmt.Sprintf(`
resource "bitbucket_repository" "test" {
  owner = %[1]q
  name  = %[2]q
}

resource "bitbucket_repository_variable" "plain" {
  key        = "PLAIN"
  value      = "plain-value"
  repository = bitbucket_repository.test.id
}

resource "bitbucket_repository_variable" "secret" {
  key        = "SECRET"
  value      = "secret-value"
  secured    = true
  repository = bitbucket_repository.test.id
}

data "bitbucket_repository_variables" "test" {
  workspace  = %[1]q
  repository = bitbucket_repository.test.slug

  depends_on = [
    bitbucket_repository_variable.plain,
    bitbucket_repository_variable.secret,
  ]
}
`, workspace, rName)
}
