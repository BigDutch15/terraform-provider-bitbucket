package bitbucket

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// isWorkspaceUUID reports whether workspace is a Bitbucket UUID in its
// braced form, e.g. "{0bbb67de-...}".
func isWorkspaceUUID(workspace string) bool {
	return len(workspace) > 2 && strings.HasPrefix(workspace, "{") && strings.HasSuffix(workspace, "}")
}

// groupsWorkspaceSlug returns the workspace slug to use in a 1.0/groups URL.
//
// The 1.0 groups API only accepts the workspace slug, but users often pass
// the workspace UUID (for example data.bitbucket_workspace.x.id). When
// workspace is a UUID it is resolved to its slug through the 2.0 API;
// otherwise it is returned unchanged.
func groupsWorkspaceSlug(client *Client, workspace string) (string, error) {
	if !isWorkspaceUUID(workspace) {
		return workspace, nil
	}

	res, err := client.Get(fmt.Sprintf("2.0/workspaces/%s", url.PathEscape(workspace)))
	if err != nil {
		return "", fmt.Errorf("resolving workspace %s to its slug: %w", workspace, err)
	}
	defer res.Body.Close()

	var ws struct {
		Slug string `json:"slug"`
	}
	if err := json.NewDecoder(res.Body).Decode(&ws); err != nil {
		return "", fmt.Errorf("resolving workspace %s to its slug: %w", workspace, err)
	}
	if ws.Slug == "" {
		return "", fmt.Errorf("resolving workspace %s to its slug: response has no slug", workspace)
	}

	return ws.Slug, nil
}
