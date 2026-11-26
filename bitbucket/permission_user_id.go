package bitbucket

import "fmt"

// permissionUserID returns the user identifier to put in the ID of a
// bitbucket_repository_user_permission or bitbucket_project_user_permission.
//
// The permissions-config API accepts a user's UUID, Atlassian account ID or
// legacy username, but always answers with the UUID. Keying the ID on that
// UUID means later reads, deletes and imports address the user the same way,
// whichever form the configuration used.
func permissionUserID(configured, returnedUUID string) (string, error) {
	if returnedUUID == "" {
		return configured, fmt.Errorf("bitbucket did not return a UUID for user %q", configured)
	}
	return returnedUUID, nil
}

// permissionUserIDForState returns the user_id to store after a read. The
// configured value is kept as written, so a username or account ID in the
// configuration does not show a diff, and so does not force a replacement,
// against the UUID the API returns. user_id is only empty after an import,
// when the API's UUID is the only value available.
func permissionUserIDForState(current, returnedUUID string) string {
	if current == "" {
		return returnedUUID
	}
	return current
}
