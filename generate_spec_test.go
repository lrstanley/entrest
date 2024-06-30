// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package entrest

import (
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpec_ThroughSchema_TwoTypes(t *testing.T) {
	t.Parallel()
	r := mustBuildSpec(t, &Config{}, nil)

	// Through schemas can be a bit different than normal schemas. Primarily:
	//   - they may not have an ID field (if composite of two different IDs
	//     via field.ID() annotation).
	//   - if they don't have an ID, you cannot individually query them.
	//   - you should still be able to create them in isolation, but updates/deletes
	//     either require a where() clause (which we don't do outright), or must
	//     be deleted through the edges in which they are attached (e.g. on a
	//     Pet, we have remove_users_following, which removes the user from the
	//     list of users following the pet).

	assert.NotNil(t, r.json(`$.paths./follows.get.responses.200`))
	assert.NotNil(t, r.json(`$.paths./follows.post.responses.201`))
	assert.NotNil(t, r.json(`$.paths./pets/{petID}/followed-by.get.responses.200`))
	assert.NotNil(t, r.json(`$.paths./users/{userID}/followed-pets.get.responses.200`))
	assert.NotNil(t, r.json(`$.components.schemas.FollowRead`))
	assert.Equal(t, "object", r.json(`$.components.schemas.FollowCreate.type`))
	assert.NotNil(t, r.json(`$.components.schemas.FollowList`))
	assert.NotNil(t, r.json(`$.components.schemas.PetCreate.properties.followed_by`))
	assert.NotNil(t, r.json(`$.components.schemas.PetUpdate.properties.add_followed_by`))
	assert.NotNil(t, r.json(`$.components.schemas.PetUpdate.properties.remove_followed_by`))
	assert.ElementsMatch(t, []string{http.MethodGet, http.MethodPost}, getPathMethods(t, r, "/follows"))
	assert.ElementsMatch(t, []string{http.MethodGet}, getPathMethods(t, r, "/pets/{petID}/followed-by"))
	assert.ElementsMatch(t, []string{http.MethodGet}, getPathMethods(t, r, "/users/{userID}/followed-pets"))

	allowedPaths := []string{
		"/follows",
		"/pets/{petID}/followed-by",
		"/users/{userID}/followed-pets",
	}

	for name := range r.spec.Paths {
		if (strings.Contains(name, "follows") || strings.Contains(name, "followed")) && !slices.Contains(allowedPaths, name) {
			t.Errorf("unexpected path %q", name)
			continue
		}
	}
}

func TestSpec_ThroughSchema_OneType(t *testing.T) {
	t.Parallel()
	r := mustBuildSpec(t, &Config{}, nil)

	assert.NotNil(t, r.json(`$.components.schemas.Friendship`))
	assert.NotNil(t, r.json(`$.components.schemas.FriendshipCreate`))
	assert.NotNil(t, r.json(`$.components.schemas.FriendshipList`))
	assert.NotNil(t, r.json(`$.components.schemas.FriendshipRead`))
	assert.NotNil(t, r.json(`$.components.schemas.FriendshipUpdate`))
	assert.ElementsMatch(t, []string{http.MethodGet, http.MethodPost}, getPathMethods(t, r, "/friendships"))
	assert.ElementsMatch(t, []string{http.MethodGet, http.MethodPatch, http.MethodDelete}, getPathMethods(t, r, "/friendships/{friendshipID}"))
	assert.ElementsMatch(t, []string{http.MethodGet}, getPathMethods(t, r, "/friendships/{friendshipID}/friend"))
	assert.ElementsMatch(t, []string{http.MethodGet}, getPathMethods(t, r, "/friendships/{friendshipID}/user"))
	assert.ElementsMatch(t, []string{http.MethodGet}, getPathMethods(t, r, "/users/{userID}/friendships"))

	allowedPaths := []string{
		"/friendships",
		"/friendships/{friendshipID}",
		"/friendships/{friendshipID}/friend", // TODO: do we want this endpoint?
		"/friendships/{friendshipID}/user",   // TODO: do we want this endpoint?
		"/users/{userID}/friendships",
	}

	for name := range r.spec.Paths {
		if strings.Contains(name, "friendship") && !slices.Contains(allowedPaths, name) {
			t.Errorf("unexpected path %q", name)
			continue
		}
	}
}
