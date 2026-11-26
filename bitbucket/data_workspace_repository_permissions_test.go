package bitbucket

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestWorkspaceRepositoryPermissionsDataSourceFollowsCursorPages(t *testing.T) {
	pages := &cursorPagesTransport{pages: map[string]string{
		"/2.0/workspaces/ws/permissions/repositories": `{"values":[` +
			`{"type":"repository_permission","permission":"admin","user":{"type":"user","uuid":"{u1}","display_name":"Alice","account_id":"acc-1"},"repository":{"type":"repository","name":"geordi","full_name":"ws/geordi","uuid":"{r1}"}},` +
			`{"type":"repository_permission","permission":"write","user":{"type":"user","uuid":"{u2}","display_name":"Bob","account_id":"acc-2"},"repository":{"type":"repository","name":"geordi","full_name":"ws/geordi","uuid":"{r1}"}}` +
			`],"next":"https://api.bitbucket.org/2.0/workspaces/ws/permissions/repositories?page=cGVybQ"}`,
		"/2.0/workspaces/ws/permissions/repositories?page=cGVybQ": `{"values":[` +
			`{"type":"repository_permission","permission":"read","user":{"type":"user","uuid":"{u3}","display_name":"Carol","account_id":"acc-3"},"repository":{"type":"repository","name":"whee","full_name":"ws/whee","uuid":"{r2}"}}` +
			`]}`,
	}}
	rt := &queryRecordingTransport{next: pages}

	res := dataWorkspaceRepositoryPermissions()
	d := res.TestResourceData()
	for k, v := range map[string]string{"workspace": "ws", "q": `permission>"read"`} {
		if err := d.Set(k, v); err != nil {
			t.Fatal(err)
		}
	}
	if diags := res.ReadWithoutTimeout(context.Background(), d, clientsWithTransport(rt)); diags.HasError() {
		t.Fatalf("unexpected error: %v (requests: %v)", diags, pages.seen)
	}

	if d.Id() != "ws" {
		t.Errorf("got ID %q, want %q", d.Id(), "ws")
	}
	if got := d.Get("permissions.#"); got != 3 {
		t.Fatalf("got %v permissions, want 3 (requests: %v)", got, pages.seen)
	}
	for k, want := range map[string]string{
		"permissions.0.permission":           "admin",
		"permissions.0.user_uuid":            "{u1}",
		"permissions.0.user_display_name":    "Alice",
		"permissions.0.user_account_id":      "acc-1",
		"permissions.0.repository_name":      "geordi",
		"permissions.0.repository_full_name": "ws/geordi",
		"permissions.0.repository_uuid":      "{r1}",
		"permissions.2.permission":           "read",
		"permissions.2.user_account_id":      "acc-3",
		"permissions.2.repository_full_name": "ws/whee",
	} {
		if got := d.Get(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}

	if len(rt.queries) == 0 || rt.queries[0] != "q=permission%3E%22read%22" {
		t.Errorf("first request query = %q, want q=permission%%3E%%22read%%22 (all: %v)", rt.queries, rt.queries)
	}
}

func TestWorkspaceRepositoryPermissionsDataSourceOmitsEmptyQuery(t *testing.T) {
	pages := &cursorPagesTransport{pages: map[string]string{
		"/2.0/workspaces/ws/permissions/repositories": `{"values":[]}`,
	}}
	rt := &queryRecordingTransport{next: pages}

	res := dataWorkspaceRepositoryPermissions()
	d := res.TestResourceData()
	if err := d.Set("workspace", "ws"); err != nil {
		t.Fatal(err)
	}
	if diags := res.ReadWithoutTimeout(context.Background(), d, clientsWithTransport(rt)); diags.HasError() {
		t.Fatalf("unexpected error: %v (requests: %v)", diags, pages.seen)
	}
	if len(rt.queries) != 1 || rt.queries[0] != "" {
		t.Errorf("got queries %v, want a single request without a query string", rt.queries)
	}
	if got := d.Get("permissions.#"); got != 0 {
		t.Errorf("got %v permissions, want 0", got)
	}
}

func TestAccDataSourceWorkspaceRepositoryPermissions_basic(t *testing.T) {
	dataSourceName := "data.bitbucket_workspace_repository_permissions.test"

	rName := acctest.RandomWithPrefix("tf-test")

	workspace := os.Getenv("BITBUCKET_TEAM")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketWorkspaceRepositoryPermissionsConfig(workspace, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "workspace", workspace),
					resource.TestCheckResourceAttrSet(dataSourceName, "permissions.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "permissions.0.repository_uuid", "bitbucket_repository.test", "uuid"),
					resource.TestCheckResourceAttrSet(dataSourceName, "permissions.0.permission"),
					resource.TestCheckResourceAttrSet(dataSourceName, "permissions.0.user_uuid"),
				),
			},
		},
	})
}

func testAccBitbucketWorkspaceRepositoryPermissionsConfig(workspace, rName string) string {
	return fmt.Sprintf(`
resource "bitbucket_repository" "test" {
  owner = %[1]q
  name  = %[2]q
}

data "bitbucket_workspace_repository_permissions" "test" {
  workspace = %[1]q
  q         = "repository.name=\"${bitbucket_repository.test.name}\""
}
`, workspace, rName)
}
