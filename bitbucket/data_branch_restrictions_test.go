package bitbucket

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// queryValuesRecordingTransport records each request's query string before handing
// it to the wrapped transport.
type queryValuesRecordingTransport struct {
	next    http.RoundTripper
	queries []url.Values
}

func (q *queryValuesRecordingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	q.queries = append(q.queries, r.URL.Query())
	return q.next.RoundTrip(r)
}

func TestBranchRestrictionsDataSourceFollowsCursorPages(t *testing.T) {
	rt := &cursorPagesTransport{pages: map[string]string{
		"/2.0/repositories/ws/repo/branch-restrictions": `{"values":[{"id":1,"kind":"push","branch_match_kind":"glob","pattern":"main",` +
			`"users":[{"uuid":"{u1}"}],"groups":[{"slug":"devs","workspace":{"slug":"ws","uuid":"{w}"}}]}],` +
			`"next":"https://api.bitbucket.org/2.0/repositories/ws/repo/branch-restrictions?page=YnI"}`,
		"/2.0/repositories/ws/repo/branch-restrictions?page=YnI": `{"values":[{"id":2,"kind":"require_approvals_to_merge","branch_match_kind":"branching_model",` +
			`"branch_type":"release","value":2,"pipelines_source_branches":["main"]}]}`,
	}}

	res := dataBranchRestrictions()
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
		t.Errorf("got id %q, want ws/repo", d.Id())
	}

	want := map[string]interface{}{
		"branch_restrictions.#":                           2,
		"branch_restrictions.0.id":                        "1",
		"branch_restrictions.0.kind":                      "push",
		"branch_restrictions.0.pattern":                   "main",
		"branch_restrictions.0.users.#":                   1,
		"branch_restrictions.0.users.0":                   "{u1}",
		"branch_restrictions.0.groups.#":                  1,
		"branch_restrictions.0.groups.0.owner":            "ws",
		"branch_restrictions.0.groups.0.slug":             "devs",
		"branch_restrictions.0.value":                     0,
		"branch_restrictions.1.id":                        "2",
		"branch_restrictions.1.branch_match_kind":         "branching_model",
		"branch_restrictions.1.branch_type":               "release",
		"branch_restrictions.1.value":                     2,
		"branch_restrictions.1.pipelines_source_branches": []interface{}{"main"},
	}
	for k, v := range want {
		if got := d.Get(k); fmt.Sprint(got) != fmt.Sprint(v) {
			t.Errorf("%s = %v, want %v", k, got, v)
		}
	}
}

func TestBranchRestrictionsDataSourceSendsFilters(t *testing.T) {
	rt := &queryValuesRecordingTransport{next: &cursorPagesTransport{pages: map[string]string{
		"/2.0/repositories/ws/repo/branch-restrictions": `{"values":[{"id":1,"kind":"push","branch_match_kind":"glob","pattern":"main"}]}`,
	}}}

	res := dataBranchRestrictions()
	d := res.TestResourceData()
	for k, v := range map[string]string{"workspace": "ws", "repository": "repo", "kind": "push", "pattern": "main"} {
		if err := d.Set(k, v); err != nil {
			t.Fatal(err)
		}
	}
	if diags := res.ReadWithoutTimeout(context.Background(), d, clientsWithTransport(rt)); diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if len(rt.queries) != 1 {
		t.Fatalf("got %d requests, want 1", len(rt.queries))
	}
	if got := rt.queries[0].Get("kind"); got != "push" {
		t.Errorf("kind query param = %q, want push", got)
	}
	if got := rt.queries[0].Get("pattern"); got != "main" {
		t.Errorf("pattern query param = %q, want main", got)
	}
}

func TestBranchRestrictionsDataSourceOmitsUnsetFilters(t *testing.T) {
	rt := &queryValuesRecordingTransport{next: &cursorPagesTransport{pages: map[string]string{
		"/2.0/repositories/ws/repo/branch-restrictions": `{"values":[]}`,
	}}}

	res := dataBranchRestrictions()
	d := res.TestResourceData()
	for k, v := range map[string]string{"workspace": "ws", "repository": "repo"} {
		if err := d.Set(k, v); err != nil {
			t.Fatal(err)
		}
	}
	if diags := res.ReadWithoutTimeout(context.Background(), d, clientsWithTransport(rt)); diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	for _, p := range []string{"kind", "pattern"} {
		if rt.queries[0].Has(p) {
			t.Errorf("unexpected %s query param", p)
		}
	}
}

func TestAccDataSourceBranchRestrictions_basic(t *testing.T) {
	dataSourceName := "data.bitbucket_branch_restrictions.test"
	filteredName := "data.bitbucket_branch_restrictions.filtered"
	resourceName := "bitbucket_branch_restriction.test"

	rName := acctest.RandomWithPrefix("tf-test")

	workspace := os.Getenv("BITBUCKET_TEAM")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketBranchRestrictionsDataConfig(workspace, rName),
			},
			{
				Config: testAccBitbucketBranchRestrictionsDataConfig(workspace, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "branch_restrictions.#", "2"),
					resource.TestCheckResourceAttr(filteredName, "branch_restrictions.#", "1"),
					resource.TestCheckResourceAttrPair(filteredName, "branch_restrictions.0.id", resourceName, "id"),
					resource.TestCheckResourceAttr(filteredName, "branch_restrictions.0.kind", "force"),
					resource.TestCheckResourceAttr(filteredName, "branch_restrictions.0.pattern", "master"),
					resource.TestCheckResourceAttr(filteredName, "branch_restrictions.0.branch_match_kind", "glob"),
				),
			},
		},
	})
}

func testAccBitbucketBranchRestrictionsDataConfig(workspace, rName string) string {
	return testAccBitbucketBranchRestrictionConfig(workspace, rName) + fmt.Sprintf(`
resource "bitbucket_branch_restriction" "approvals" {
  owner      = %[1]q
  repository = bitbucket_repository.test.name
  kind       = "require_approvals_to_merge"
  pattern    = "master"
  value      = 1
}

data "bitbucket_branch_restrictions" "test" {
  workspace  = %[1]q
  repository = bitbucket_repository.test.name

  depends_on = [bitbucket_branch_restriction.test, bitbucket_branch_restriction.approvals]
}

data "bitbucket_branch_restrictions" "filtered" {
  workspace  = %[1]q
  repository = bitbucket_repository.test.name
  kind       = "force"

  depends_on = [bitbucket_branch_restriction.test, bitbucket_branch_restriction.approvals]
}
`, workspace)
}
