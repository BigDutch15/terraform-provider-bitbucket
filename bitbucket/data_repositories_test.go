package bitbucket

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// queryRecordingTransport records each request's raw query before passing it on.
type queryRecordingTransport struct {
	next    http.RoundTripper
	queries []string
}

func (q *queryRecordingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	q.queries = append(q.queries, r.URL.RawQuery)
	return q.next.RoundTrip(r)
}

func TestRepositoriesDataSourceFollowsCursorPages(t *testing.T) {
	rt := &cursorPagesTransport{pages: map[string]string{
		"/2.0/repositories/ws":             `{"values":[{"uuid":"{a}","slug":"one","full_name":"ws/one","is_private":true,"mainbranch":{"name":"main"},"project":{"key":"PRJ"},"created_on":"2026-01-02T03:04:05.123456+00:00"}],"next":"https://api.bitbucket.org/2.0/repositories/ws?page=cmVwbw"}`,
		"/2.0/repositories/ws?page=cmVwbw": `{"values":[{"uuid":"{b}","slug":"two"}]}`,
	}}

	res := dataRepositories()
	d := res.TestResourceData()
	if err := d.Set("workspace", "ws"); err != nil {
		t.Fatal(err)
	}
	if diags := res.ReadWithoutTimeout(context.Background(), d, clientsWithTransport(rt)); diags.HasError() {
		t.Fatalf("unexpected error: %v (requests: %v)", diags, rt.seen)
	}

	if got := d.Get("repositories.#"); got != 2 {
		t.Fatalf("got %v repositories, want 2 (requests: %v)", got, rt.seen)
	}
	if d.Id() != "ws" {
		t.Errorf("id = %q, want %q", d.Id(), "ws")
	}
	for attr, want := range map[string]interface{}{
		"repositories.0.uuid":        "{a}",
		"repositories.0.slug":        "one",
		"repositories.0.full_name":   "ws/one",
		"repositories.0.is_private":  true,
		"repositories.0.main_branch": "main",
		"repositories.0.project_key": "PRJ",
		"repositories.0.created_on":  "2026-01-02T03:04:05Z",
		"repositories.1.slug":        "two",
		"repositories.1.main_branch": "",
		"repositories.1.project_key": "",
	} {
		if got := d.Get(attr); got != want {
			t.Errorf("%s = %#v, want %#v", attr, got, want)
		}
	}
}

func TestRepositoriesDataSourceFiltersByProject(t *testing.T) {
	rt := &queryRecordingTransport{next: &cursorPagesTransport{pages: map[string]string{
		"/2.0/repositories/ws": `{"values":[{"uuid":"{a}","slug":"one","project":{"key":"PRJ"}}]}`,
	}}}

	res := dataRepositories()
	d := res.TestResourceData()
	for k, v := range map[string]string{"workspace": "ws", "project_key": "PRJ"} {
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
	if want := "q=project.key%3D%22PRJ%22"; rt.queries[0] != want {
		t.Errorf("query = %q, want %q", rt.queries[0], want)
	}
	if d.Id() != "ws/PRJ" {
		t.Errorf("id = %q, want %q", d.Id(), "ws/PRJ")
	}
}

func TestAccDataSourceRepositories_basic(t *testing.T) {
	dataSourceName := "data.bitbucket_repositories.test"
	resourceName := "bitbucket_repository.test"

	rName := acctest.RandomWithPrefix("tf-test")
	projectKey := acctest.RandStringFromCharSet(8, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	workspace := os.Getenv("BITBUCKET_TEAM")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketRepositoriesConfig(workspace, rName, projectKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "repositories.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "repositories.0.uuid", resourceName, "uuid"),
					resource.TestCheckResourceAttrPair(dataSourceName, "repositories.0.slug", resourceName, "slug"),
					resource.TestCheckResourceAttr(dataSourceName, "repositories.0.project_key", projectKey),
				),
			},
		},
	})
}

func testAccBitbucketRepositoriesConfig(workspace, rName, projectKey string) string {
	return fmt.Sprintf(`
resource "bitbucket_project" "test" {
  owner = %[1]q
  name  = %[2]q
  key   = %[3]q
}

resource "bitbucket_repository" "test" {
  owner       = %[1]q
  name        = %[2]q
  project_key = bitbucket_project.test.key
}

data "bitbucket_repositories" "test" {
  workspace   = %[1]q
  project_key = bitbucket_project.test.key

  depends_on = [bitbucket_repository.test]
}
`, workspace, rName, projectKey)
}
