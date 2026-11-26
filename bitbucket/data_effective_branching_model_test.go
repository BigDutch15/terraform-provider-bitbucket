package bitbucket

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestEffectiveBranchingModelReadMapsResponse(t *testing.T) {
	rt := &cursorPagesTransport{pages: map[string]string{
		"/2.0/repositories/ws/repo/effective-branching-model": `{
			"type": "effective_repo_branching_model",
			"development": {"name": "develop", "use_mainbranch": false, "branch": {"type": "branch", "name": "develop"}},
			"production": {"name": "", "use_mainbranch": true, "branch_does_not_exist": true},
			"branch_types": [
				{"kind": "feature", "prefix": "feature/"},
				{"kind": "hotfix", "prefix": "hotfix/"}
			]
		}`,
	}}
	res := dataEffectiveBranchingModel()
	d := res.TestResourceData()
	d.Set("workspace", "ws")
	d.Set("repository", "repo")

	if diags := res.ReadWithoutTimeout(context.Background(), d, clientsWithTransport(rt)); diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if len(rt.seen) != 1 {
		t.Fatalf("expected one request, got %v", rt.seen)
	}
	if d.Id() != "ws/repo" {
		t.Errorf("id = %q, want ws/repo", d.Id())
	}
	for k, want := range map[string]string{
		"development.#":                       "1",
		"development.0.name":                  "develop",
		"development.0.use_mainbranch":        "false",
		"development.0.branch_does_not_exist": "false",
		"production.#":                        "1",
		"production.0.use_mainbranch":         "true",
		"production.0.branch_does_not_exist":  "true",
		"branch_type.#":                       "2",
		"branch_type.0.kind":                  "feature",
		"branch_type.0.prefix":                "feature/",
		"branch_type.1.kind":                  "hotfix",
		"branch_type.1.prefix":                "hotfix/",
	} {
		if got := d.State().Attributes[k]; got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}
}

func TestEffectiveBranchingModelReadWithoutProduction(t *testing.T) {
	rt := &cursorPagesTransport{pages: map[string]string{
		"/2.0/repositories/ws/repo/effective-branching-model": `{"development": {"name": "main", "use_mainbranch": true}, "branch_types": []}`,
	}}
	res := dataEffectiveBranchingModel()
	d := res.TestResourceData()
	d.Set("workspace", "ws")
	d.Set("repository", "repo")

	if diags := res.ReadWithoutTimeout(context.Background(), d, clientsWithTransport(rt)); diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if got := len(d.Get("production").([]interface{})); got != 0 {
		t.Errorf("production has %d items, want 0", got)
	}
	if got := len(d.Get("branch_type").([]interface{})); got != 0 {
		t.Errorf("branch_type has %d items, want 0", got)
	}
}

func TestEffectiveBranchingModelReadReturnsAPIError(t *testing.T) {
	rt := &cursorPagesTransport{pages: map[string]string{}}
	res := dataEffectiveBranchingModel()
	d := res.TestResourceData()
	d.Set("workspace", "ws")
	d.Set("repository", "repo")

	if diags := res.ReadWithoutTimeout(context.Background(), d, clientsWithTransport(rt)); !diags.HasError() {
		t.Fatal("expected an error for a non-2xx response")
	}
}

func TestAccDataSourceEffectiveBranchingModel_basic(t *testing.T) {
	dataSourceName := "data.bitbucket_effective_branching_model.test"

	rName := acctest.RandomWithPrefix("tf-test")

	workspace := os.Getenv("BITBUCKET_TEAM")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketEffectiveBranchingModelConfig(workspace, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "workspace", workspace),
					resource.TestCheckResourceAttr(dataSourceName, "development.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "development.0.use_mainbranch", "true"),
					// A new repository inherits its project's branching model, so the
					// effective model can list more branch types than the one set here.
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "branch_type.*", map[string]string{
						"kind":   "feature",
						"prefix": "feature/",
					}),
				),
			},
		},
	})
}

func testAccBitbucketEffectiveBranchingModelConfig(workspace, rName string) string {
	return fmt.Sprintf(`
resource "bitbucket_repository" "test" {
  owner = %[1]q
  name  = %[2]q
}

resource "bitbucket_branching_model" "test" {
  owner      = %[1]q
  repository = bitbucket_repository.test.name

  development {
    use_mainbranch = true
  }

  # Bitbucket keeps all four branch types, so declare them all; with fewer,
  # the plan after refresh tries to remove the ones left out.
  branch_type {
    enabled = true
    kind    = "feature"
    prefix  = "feature/"
  }

  branch_type {
    enabled = true
    kind    = "bugfix"
    prefix  = "bugfix/"
  }

  branch_type {
    enabled = true
    kind    = "hotfix"
    prefix  = "hotfix/"
  }

  branch_type {
    enabled = true
    kind    = "release"
    prefix  = "release/"
  }
}

data "bitbucket_effective_branching_model" "test" {
  workspace  = %[1]q
  repository = bitbucket_repository.test.slug

  depends_on = [bitbucket_branching_model.test]
}
`, workspace, rName)
}
