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

func TestSpec_ThroughSchemas(t *testing.T) {
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
	assert.NotNil(t, r.json(`$.paths./pets/{id}/followed-by.get.responses.200`))
	assert.NotNil(t, r.json(`$.paths./users/{id}/followed-pets.get.responses.200`))
	assert.NotNil(t, r.json(`$.components.schemas.FollowRead`))
	assert.Equal(t, "object", r.json(`$.components.schemas.FollowCreate.type`))
	assert.NotNil(t, r.json(`$.components.schemas.FollowList`))
	assert.NotNil(t, r.json(`$.components.schemas.PetCreate.properties.followed_by`))
	assert.NotNil(t, r.json(`$.components.schemas.PetUpdate.properties.followed_by`))
	assert.NotNil(t, r.json(`$.components.schemas.PetUpdate.properties.add_followed_by`))
	assert.NotNil(t, r.json(`$.components.schemas.PetUpdate.properties.remove_followed_by`))
	assert.ElementsMatch(t, []string{http.MethodGet, http.MethodPost}, getPathMethods(t, r, "/follows"))
	assert.ElementsMatch(t, []string{http.MethodGet}, getPathMethods(t, r, "/pets/{id}/followed-by"))
	assert.ElementsMatch(t, []string{http.MethodGet}, getPathMethods(t, r, "/users/{id}/followed-pets"))

	allowedPaths := []string{
		"/follows",
		"/pets/{id}/followed-by",
		"/users/{id}/followed-pets",
	}

	for name := range r.spec.Paths {
		if (strings.Contains(name, "follows") || strings.Contains(name, "followed")) && !slices.Contains(allowedPaths, name) {
			t.Errorf("unexpected path %q", name)
			continue
		}
	}
}