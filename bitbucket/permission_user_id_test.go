package bitbucket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const fakePermissionUserUUID = "{0a1b2c3d-1111-2222-3333-444455556666}"

// fakeUserPermissionsAPI is an in-memory permissions-config API for users on
// repositories and projects. Like Bitbucket, it accepts a user's UUID (in any
// case), account ID or legacy username, and always answers with the UUID.
type fakeUserPermissionsAPI struct {
	*httptest.Server
	mu          sync.Mutex
	permissions map[string]string // by collection path + "/" + uuid
	puts        int
}

var userPermissionPath = regexp.MustCompile(`^(/2\.0/(?:repositories/[^/]+/[^/]+|workspaces/[^/]+/projects/[^/]+)/permissions-config/users)/([^/]+)$`)

func newFakeUserPermissionsAPI(t *testing.T) *fakeUserPermissionsAPI {
	t.Helper()
	f := &fakeUserPermissionsAPI{permissions: map[string]string{}}
	f.Server = httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(f.Close)
	for _, k := range []string{"BITBUCKET_USERNAME", "BITBUCKET_PASSWORD", "BITBUCKET_OAUTH_TOKEN", "BITBUCKET_ACCESS_TOKEN", "BITBUCKET_OAUTH_CLIENT_ID", "BITBUCKET_OAUTH_CLIENT_SECRET", "BITBUCKET_API_URL"} {
		t.Setenv(k, "")
	}
	return f
}

func (f *fakeUserPermissionsAPI) resolveUser(id string) (string, bool) {
	switch {
	case strings.EqualFold(id, fakePermissionUserUUID), id == "jdoe", id == "557058:abcdef":
		return fakePermissionUserUUID, true
	}
	return "", false
}

func (f *fakeUserPermissionsAPI) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	writeJSON := func(status int, v interface{}) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(v)
	}
	notFound := func() {
		writeJSON(http.StatusNotFound, map[string]interface{}{"type": "error", "error": map[string]string{"message": "not found"}})
	}

	match := userPermissionPath.FindStringSubmatch(r.URL.Path)
	if match == nil {
		notFound()
		return
	}
	uuid, ok := f.resolveUser(match[2])
	if !ok {
		notFound()
		return
	}
	key := match[1] + "/" + uuid
	respond := func(permission string) {
		writeJSON(http.StatusOK, map[string]interface{}{
			"permission": permission,
			"user":       map[string]string{"type": "user", "uuid": uuid},
		})
	}

	switch r.Method {
	case http.MethodPut:
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		f.permissions[key] = body["permission"]
		f.puts++
		respond(body["permission"])
	case http.MethodGet:
		permission, ok := f.permissions[key]
		if !ok {
			notFound()
			return
		}
		respond(permission)
	case http.MethodDelete:
		if _, ok := f.permissions[key]; !ok {
			notFound()
			return
		}
		delete(f.permissions, key)
		w.WriteHeader(http.StatusNoContent)
	default:
		notFound()
	}
}

func (f *fakeUserPermissionsAPI) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.permissions)
}

type userPermissionResourceCase struct {
	resourceType string
	parentArg    string // repo_slug or project_key
	parentValue  string
	permissions  [2]string // a create value and an update value
}

var userPermissionResourceCases = []userPermissionResourceCase{
	{"bitbucket_repository_user_permission", "repo_slug", "repo", [2]string{"read", "write"}},
	{"bitbucket_project_user_permission", "project_key", "KEY", [2]string{"read", "create-repo"}},
}

func (c userPermissionResourceCase) config(apiURL, userID, permission string) string {
	return testProviderConfig(apiURL, "") + fmt.Sprintf(`
resource %q "test" {
  workspace  = "ws"
  %s = %q
  user_id    = %q
  permission = %q
}
`, c.resourceType, c.parentArg, c.parentValue, userID, permission)
}

// A user_id given as a legacy username, an account ID or a UUID in upper case
// is accepted, keyed on the UUID Bitbucket returns, and kept in state as
// written so that later plans are empty. Previously a username created the
// permission and then failed the apply (repository) or forced a replacement
// on every plan (project). See BigDutch15/terraform-provider-bitbucket#15.
func TestUserPermission_userIDForms(t *testing.T) {
	for _, c := range userPermissionResourceCases {
		for _, userID := range []string{"jdoe", "557058:abcdef", strings.ToUpper(fakePermissionUserUUID)} {
			t.Run(c.resourceType+"/"+userID, func(t *testing.T) {
				fake := newFakeUserPermissionsAPI(t)
				name := c.resourceType + ".test"
				wantID := fmt.Sprintf("ws:%s:%s", c.parentValue, fakePermissionUserUUID)

				resource.UnitTest(t, resource.TestCase{
					ProviderFactories: testAccProviderFactories,
					CheckDestroy: func(*terraform.State) error {
						if n := fake.count(); n != 0 {
							return fmt.Errorf("%d permissions left after destroy", n)
						}
						return nil
					},
					Steps: []resource.TestStep{
						{
							Config: c.config(fake.URL, userID, c.permissions[0]),
							Check: resource.ComposeTestCheckFunc(
								resource.TestCheckResourceAttr(name, "id", wantID),
								resource.TestCheckResourceAttr(name, "user_id", userID),
								resource.TestCheckResourceAttr(name, "permission", c.permissions[0]),
							),
						},
						{
							// No diff between the configured user_id and the UUID.
							Config:   c.config(fake.URL, userID, c.permissions[0]),
							PlanOnly: true,
						},
						{
							// The permission updates in place.
							Config: c.config(fake.URL, userID, c.permissions[1]),
							Check: resource.ComposeTestCheckFunc(
								resource.TestCheckResourceAttr(name, "id", wantID),
								resource.TestCheckResourceAttr(name, "permission", c.permissions[1]),
								func(*terraform.State) error {
									if fake.puts != 2 {
										return fmt.Errorf("got %d PUT requests, want 2", fake.puts)
									}
									return nil
								},
							),
						},
						{
							// An import has only the ID to go on, so user_id is the UUID.
							ResourceName:            name,
							ImportState:             true,
							ImportStateVerify:       true,
							ImportStateVerifyIgnore: []string{"user_id"},
							ImportStateCheck: func(states []*terraform.InstanceState) error {
								if got := states[0].Attributes["user_id"]; got != fakePermissionUserUUID {
									return fmt.Errorf("imported user_id = %q, want %q", got, fakePermissionUserUUID)
								}
								return nil
							},
						},
					},
				})
			})
		}
	}
}

func TestPermissionUserID(t *testing.T) {
	got, err := permissionUserID("jdoe", fakePermissionUserUUID)
	if err != nil || got != fakePermissionUserUUID {
		t.Errorf("permissionUserID(jdoe, uuid) = %q, %v; want %q, nil", got, err, fakePermissionUserUUID)
	}

	got, err = permissionUserID("jdoe", "")
	if err == nil || got != "jdoe" {
		t.Errorf("permissionUserID(jdoe, \"\") = %q, %v; want jdoe and an error", got, err)
	}
}

func TestPermissionUserIDForState(t *testing.T) {
	cases := []struct{ current, want string }{
		{"", fakePermissionUserUUID},
		{"jdoe", "jdoe"},
		{strings.ToUpper(fakePermissionUserUUID), strings.ToUpper(fakePermissionUserUUID)},
		{fakePermissionUserUUID, fakePermissionUserUUID},
	}
	for _, c := range cases {
		if got := permissionUserIDForState(c.current, fakePermissionUserUUID); got != c.want {
			t.Errorf("permissionUserIDForState(%q) = %q, want %q", c.current, got, c.want)
		}
	}
}
