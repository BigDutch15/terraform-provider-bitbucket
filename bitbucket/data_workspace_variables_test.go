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

func TestWorkspaceVariablesDataSourceFollowsCursorPages(t *testing.T) {
	rt := &cursorPagesTransport{pages: map[string]string{
		"/2.0/workspaces/ws/pipelines-config/variables":          `{"values":[{"uuid":"{a}","key":"PLAIN","value":"hello","secured":false}],"next":"https://api.bitbucket.org/2.0/workspaces/ws/pipelines-config/variables?page=d3M"}`,
		"/2.0/workspaces/ws/pipelines-config/variables?page=d3M": `{"values":[{"uuid":"{b}","key":"SECRET","secured":true}]}`,
	}}

	res := dataWorkspaceVariables()
	d := res.TestResourceData()
	if err := d.Set("workspace", "ws"); err != nil {
		t.Fatal(err)
	}
	if diags := res.ReadWithoutTimeout(context.Background(), d, clientsWithTransport(rt)); diags.HasError() {
		t.Fatalf("unexpected error: %v (requests: %v)", diags, rt.seen)
	}
	if got := d.Get("variables.#"); got != 2 {
		t.Fatalf("got %v variables, want 2 (requests: %v)", got, rt.seen)
	}
	if got := d.Id(); got != "ws" {
		t.Errorf("id = %q, want ws", got)
	}
	want := map[string]interface{}{
		"variables.0.key":     "PLAIN",
		"variables.0.value":   "hello",
		"variables.1.uuid":    "{b}",
		"variables.1.value":   "",
		"variables.1.secured": true,
	}
	for k, v := range want {
		if got := d.Get(k); got != v {
			t.Errorf("%s = %#v, want %#v", k, got, v)
		}
	}
}

func TestAccDataSourceWorkspaceVariables_basic(t *testing.T) {
	dataSourceName := "data.bitbucket_workspace_variables.test"

	workspace := os.Getenv("BITBUCKET_TEAM")
	// The workspace may already hold other variables, so use a unique key and
	// look for it rather than asserting the list length.
	key := strings.ReplaceAll(strings.ToUpper(acctest.RandomWithPrefix("tf_test")), "-", "_")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketWorkspaceVariablesDataConfig(workspace, key),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "workspace", workspace),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "variables.*", map[string]string{
						"key":     key,
						"value":   "plain-value",
						"secured": "false",
					}),
				),
			},
		},
	})
}

func testAccBitbucketWorkspaceVariablesDataConfig(workspace, key string) string {
	return fmt.Sprintf(`
resource "bitbucket_workspace_variable" "test" {
  key       = %[2]q
  value     = "plain-value"
  workspace = %[1]q
}

data "bitbucket_workspace_variables" "test" {
  workspace = %[1]q

  depends_on = [bitbucket_workspace_variable.test]
}
`, workspace, key)
}
