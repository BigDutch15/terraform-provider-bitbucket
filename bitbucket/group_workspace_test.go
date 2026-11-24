package bitbucket

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

// fakeGroupsAPI answers the workspace lookup and the 1.0 group endpoint,
// recording every path requested.
func fakeGroupsAPI(t *testing.T, paths *[]string) Client {
	t.Helper()
	return Client{HTTPClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		*paths = append(*paths, r.URL.Path)

		var body string
		switch r.URL.Path {
		case "/2.0/workspaces/{0bbb67de-0000-0000-0000-000000000000}":
			body = `{"uuid":"{0bbb67de-0000-0000-0000-000000000000}","slug":"my-workspace"}`
		case "/1.0/groups/my-workspace/devs":
			body = `{"name":"Devs","slug":"devs","permission":"write"}`
		default:
			return &http.Response{
				StatusCode: http.StatusMethodNotAllowed,
				Body:       io.NopCloser(strings.NewReader(`{"type":"error","error":{"message":"not allowed"}}`)),
				Request:    r,
			}, nil
		}

		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
	})}}
}

func TestGroupsWorkspaceSlug(t *testing.T) {
	var paths []string
	client := fakeGroupsAPI(t, &paths)

	got, err := groupsWorkspaceSlug(&client, "{0bbb67de-0000-0000-0000-000000000000}")
	if err != nil || got != "my-workspace" {
		t.Fatalf("UUID: got %q, %v; want my-workspace", got, err)
	}

	paths = nil
	got, err = groupsWorkspaceSlug(&client, "my-workspace")
	if err != nil || got != "my-workspace" {
		t.Fatalf("slug: got %q, %v; want my-workspace", got, err)
	}
	if len(paths) != 0 {
		t.Errorf("a slug should not trigger a lookup, got requests %v", paths)
	}

	if _, err := groupsWorkspaceSlug(&client, "{unknown}"); err == nil {
		t.Error("expected an error for a UUID the API rejects")
	}
}

func TestDataGroupAcceptsWorkspaceUUIDOrSlug(t *testing.T) {
	for _, workspace := range []string{"{0bbb67de-0000-0000-0000-000000000000}", "my-workspace"} {
		t.Run(workspace, func(t *testing.T) {
			var paths []string
			clients := Clients{httpClient: fakeGroupsAPI(t, &paths)}

			d := dataGroup().TestResourceData()
			_ = d.Set("workspace", workspace)
			_ = d.Set("slug", "devs")

			if diags := dataReadGroup(context.Background(), d, clients); diags.HasError() {
				t.Fatalf("read failed: %v (requests %v)", diags, paths)
			}
			if got := d.Get("name").(string); got != "Devs" {
				t.Errorf("name = %q, want Devs", got)
			}
			// The configured value is kept so plans don't show a diff.
			if got := d.Get("workspace").(string); got != workspace {
				t.Errorf("workspace = %q, want %q", got, workspace)
			}
		})
	}
}

func TestResourceGroupReadAcceptsWorkspaceUUID(t *testing.T) {
	var paths []string
	clients := Clients{httpClient: fakeGroupsAPI(t, &paths)}

	d := resourceGroup().TestResourceData()
	d.SetId("{0bbb67de-0000-0000-0000-000000000000}/devs")

	if diags := resourceGroupsRead(context.Background(), d, clients); diags.HasError() {
		t.Fatalf("read failed: %v (requests %v)", diags, paths)
	}
	if d.Id() == "" || d.Get("slug").(string) != "devs" {
		t.Errorf("group not read: id=%q slug=%q", d.Id(), d.Get("slug"))
	}
	if got := d.Get("workspace").(string); got != "{0bbb67de-0000-0000-0000-000000000000}" {
		t.Errorf("workspace = %q, want the UUID kept", got)
	}
}
