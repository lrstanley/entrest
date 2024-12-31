package main

import (
	"context"
	"database/sql"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/lrstanley/entrest/_examples/kitchensink/internal/database/ent"
	"github.com/lrstanley/entrest/_examples/kitchensink/internal/database/ent/enttest"
	"github.com/lrstanley/entrest/_examples/kitchensink/internal/database/ent/migrate"
	"github.com/lrstanley/entrest/_examples/kitchensink/internal/database/ent/pet"
	"github.com/lrstanley/entrest/_examples/kitchensink/internal/database/ent/rest"
	"github.com/lrstanley/entrest/_examples/kitchensink/internal/database/ent/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"modernc.org/sqlite"
)

var sqlRegister sync.Once

func newClient(t *testing.T) *ent.Client {
	t.Helper()

	sqlRegister.Do(func() {
		sql.Register("sqlite3", &sqlite.Driver{})
	})

	opts := []enttest.Option{
		enttest.WithOptions(ent.Log(t.Log)),
		enttest.WithMigrateOptions(
			migrate.WithDropColumn(true),
			migrate.WithDropIndex(true),
			migrate.WithGlobalUniqueID(true),
			migrate.WithForeignKeys(true),
		),
	}

	db := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_pragma=foreign_keys(1)", opts...)
	return db
}

func newRestServer(t *testing.T, cfg *rest.ServerConfig) (ctx context.Context, db *ent.Client, s *enttest.TestServer) {
	t.Helper()
	ctx = context.Background()
	db = newClient(t)
	s = enttest.NewServer(t, db, cfg)
	return ctx, db, s
}

func newUser(db *ent.Client) *ent.UserCreate {
	first := gofakeit.FirstName()
	last := gofakeit.LastName()

	return db.User.Create().
		SetName(first + " " + last).
		SetEmail(first + "." + last + "@example.com").
		SetEnabled(true).
		SetType(user.TypeUser).
		SetPasswordHashed(gofakeit.Password(true, true, true, true, true, 15)) // Not actually used.
}

func newPet(db *ent.Client) *ent.PetCreate {
	types := []pet.Type{
		pet.TypeDog,
		pet.TypeCat,
		pet.TypeBird,
		pet.TypeFish,
		pet.TypeAmphibian,
		pet.TypeReptile,
		pet.TypeOther,
	}
	return db.Pet.Create().
		SetName(gofakeit.PetName()).
		SetNicknames([]string{gofakeit.PetName(), gofakeit.PetName()}).
		SetAge(gofakeit.Number(1, 15)).
		SetType(types[gofakeit.Number(0, len(types)-1)])
}

func newCategory(db *ent.Client) *ent.CategoryCreate {
	return db.Category.Create().
		SetName(gofakeit.UUID()).
		SetReadonly(gofakeit.UUID())
}

func TestHandler_Get(t *testing.T) {
	t.Parallel()

	ctx, db, s := newRestServer(t, nil)
	t.Cleanup(func() { db.Close() })

	pet1 := newPet(db).SaveX(ctx)
	user1 := newUser(db).AddFollowedPets(pet1).SaveX(ctx)
	followedPets := user1.QueryFollowedPets().AllX(ctx)

	resp := enttest.Request[ent.User](
		ctx, s,
		http.MethodGet,
		"/users/"+user1.ID.String(),
		http.NoBody,
	).Must(t)

	assert.Equal(t, http.StatusOK, resp.Data.Code)
	assert.Equal(t, user1.ID, resp.Value.ID)
	require.Len(t, followedPets, 1)
	assert.Equal(t, pet1.ID, followedPets[0].ID)

	// Also validate that 404's work correctly.
	respPet := enttest.Request[ent.Pet](
		ctx, s,
		http.MethodGet,
		"/pets/123",
		http.NoBody,
	)
	assert.Equal(t, http.StatusNotFound, respPet.Data.Code)
	assert.Nil(t, respPet.Value)

	// And if using UUIDs or similar, we'd actually get a 400 if we passed in an
	// invalid UUID (but not specific info if masking is enabled).
	resp = enttest.Request[ent.User](
		ctx, s,
		http.MethodGet,
		"/users/123",
		http.NoBody,
	)
	assert.Equal(t, http.StatusBadRequest, resp.Data.Code)
	assert.Nil(t, resp.Value)
}

func TestHandler_GetEdge(t *testing.T) {
	t.Parallel()

	ctx, db, s := newRestServer(t, nil)
	t.Cleanup(func() { db.Close() })

	user1 := newUser(db).SaveX(ctx)
	user2 := newUser(db).AddFriends(user1).SaveX(ctx)

	resp := enttest.Request[rest.PagedResponse[ent.User]](
		ctx, s,
		http.MethodGet,
		"/users/"+user1.ID.String()+"/friends",
		http.NoBody,
	).Must(t)

	assert.Equal(t, http.StatusOK, resp.Data.Code)
	require.Len(t, resp.Value.Content, 1)
	assert.Equal(t, user2.ID, resp.Value.Content[0].ID)

	// Also validate that 404's work correctly.
	respPet := enttest.Request[rest.PagedResponse[ent.User]](
		ctx, s,
		http.MethodGet,
		"/pets/123/friends",
		http.NoBody,
	)

	assert.Equal(t, http.StatusNotFound, respPet.Data.Code)
	assert.Empty(t, respPet.Value.Content)
	assert.Equal(t, 0, respPet.Value.TotalCount)
	assert.Equal(t, 1, respPet.Value.LastPage)

	// And similar when invalid ID is used, just 400.

	resp = enttest.Request[rest.PagedResponse[ent.User]](
		ctx, s,
		http.MethodGet,
		"/users/123/friends",
		http.NoBody,
	)

	assert.Equal(t, http.StatusBadRequest, resp.Data.Code)
}

func TestHandler_List(t *testing.T) {
	ctx, db, s := newRestServer(t, nil)
	t.Cleanup(func() { db.Close() })

	// We want to be able to test multiple pages, so create 1 page worth plus some change,
	// so we're always sure there's more than 1 page without any filtering.
	totalUsers := rest.UserPageConfig.ItemsPerPage + 5
	sortableUser1 := newUser(db).SetName("bcdefg").SaveX(ctx)
	sortableUser2 := newUser(db).SetName("abcdef").SaveX(ctx)
	currentUserCount := db.User.Query().CountX(ctx)
	createdTime := time.Now()

	// Sleep before and after so we can test created_at filtering.
	users := db.User.CreateBulk(enttest.Multiple(newUser, db, totalUsers-currentUserCount)...).SaveX(ctx)

	esc := url.QueryEscape

	tests := []struct {
		name               string
		uri                string
		expectedPage       int         // What is the current page?
		expectedIsLastPage bool        // Is supposed to be the last page?
		expectedStatus     int         // Expected status code.
		expectedCount      int         // Expected number of returned items.
		expectedTotalCount int         // Expected total number of items.
		expectedUsers      []*ent.User // Subset of expected users in the response.
		mustUsers          bool        // If true, the users field must be exact, not subset.
		mustUsersOrder     bool        // If true, the users field must be in the same order.
	}{
		{
			name:               "default",
			uri:                "/users",
			expectedPage:       1,
			expectedStatus:     http.StatusOK,
			expectedCount:      rest.UserPageConfig.ItemsPerPage,
			expectedTotalCount: totalUsers,
		},
		{
			name:               "default-pretty",
			uri:                "/users?pretty=true",
			expectedPage:       1,
			expectedStatus:     http.StatusOK,
			expectedCount:      rest.UserPageConfig.ItemsPerPage,
			expectedTotalCount: totalUsers,
		},
		{
			name:               "page-2",
			uri:                "/users?page=2",
			expectedPage:       2,
			expectedIsLastPage: true,
			expectedStatus:     http.StatusOK,
			expectedCount:      5,
			expectedTotalCount: totalUsers,
		},
		{
			name:           "page-out-of-bounds",
			uri:            "/users?page=3",
			expectedPage:   3,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "negative-page",
			uri:            "/users?page=-1",
			expectedPage:   1,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:               "more-items",
			uri:                "/users?per_page=" + strconv.Itoa(rest.UserPageConfig.MaxItemsPerPage),
			expectedPage:       1,
			expectedIsLastPage: true,
			expectedStatus:     http.StatusOK,
			expectedCount:      totalUsers,
			expectedTotalCount: totalUsers,
		},
		{
			name:           "too-many-requested-items",
			uri:            "/users?per_page=" + strconv.Itoa(rest.UserPageConfig.MaxItemsPerPage+1),
			expectedPage:   1,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "zero-items-per-page",
			uri:            "/users?per_page=0",
			expectedPage:   1,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "negative-items-per-page",
			uri:            "/users?per_page=-1",
			expectedPage:   1,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:               "predicate-equal",
			uri:                "/users?name.eq=" + esc(users[0].Name),
			expectedPage:       1,
			expectedIsLastPage: true,
			expectedStatus:     http.StatusOK,
			expectedCount:      1,
			expectedTotalCount: 1,
			expectedUsers:      []*ent.User{users[0]},
			mustUsers:          true,
		},
		{
			name:               "predicate-equal-case-insensitive",
			uri:                "/users?name.ieq=" + esc(strings.ToUpper(users[0].Name)),
			expectedPage:       1,
			expectedIsLastPage: true,
			expectedStatus:     http.StatusOK,
			expectedCount:      1,
			expectedTotalCount: 1,
			expectedUsers:      []*ent.User{users[0]},
			mustUsers:          true,
		},
		{
			name:               "predicate-not-equal",
			uri:                "/users?name.neq=" + esc(users[0].Name),
			expectedPage:       1,
			expectedIsLastPage: false,
			expectedStatus:     http.StatusOK,
			expectedCount:      rest.UserPageConfig.ItemsPerPage,
			expectedTotalCount: totalUsers - 1,
		},
		{
			name:               "predicate-in",
			uri:                "/users?name.in=" + esc(users[0].Name) + "&name.in=" + esc(users[1].Name),
			expectedPage:       1,
			expectedIsLastPage: true,
			expectedStatus:     http.StatusOK,
			expectedCount:      2,
			expectedTotalCount: 2,
			expectedUsers:      []*ent.User{users[0], users[1]},
			mustUsers:          true,
		},
		{
			name:               "predicate-not-in",
			uri:                "/users?name.notIn=" + esc(users[0].Name) + "&name.notIn=" + esc(users[1].Name),
			expectedPage:       1,
			expectedIsLastPage: false,
			expectedStatus:     http.StatusOK,
			expectedCount:      rest.UserPageConfig.ItemsPerPage,
			expectedTotalCount: totalUsers - 2,
		},
		{
			name:               "predicate-AND-equal-match",
			uri:                "/users?name.eq=" + esc(users[0].Name) + "&email.eq=" + esc(*users[0].Email) + "&filter_op=and",
			expectedPage:       1,
			expectedIsLastPage: true,
			expectedStatus:     http.StatusOK,
			expectedCount:      1,
			expectedTotalCount: 1,
			expectedUsers:      []*ent.User{users[0]},
			mustUsers:          true,
		},
		{
			name:               "predicate-AND-equal-not-match",
			uri:                "/users?name.eq=" + esc(users[0].Name) + "&email.eq=" + esc(*users[1].Email) + "&filter_op=and",
			expectedPage:       1,
			expectedIsLastPage: true,
			expectedStatus:     http.StatusNotFound,
			expectedCount:      0,
			expectedTotalCount: 0,
		},
		{
			name:               "predicate-OR-equal-match",
			uri:                "/users?name.eq=" + esc(users[0].Name) + "&email.eq=" + esc(*users[1].Email) + "&filter_op=or",
			expectedPage:       1,
			expectedIsLastPage: true,
			expectedStatus:     http.StatusOK,
			expectedCount:      2,
			expectedTotalCount: 2,
			expectedUsers:      []*ent.User{users[0], users[1]},
			mustUsers:          true,
		},
		{
			name:               "predicate-SORT-ASC-in",
			uri:                "/users?name.in=" + esc(sortableUser1.Name) + "&name.in=" + esc(sortableUser2.Name) + "&sort=name&order=asc",
			expectedPage:       1,
			expectedIsLastPage: true,
			expectedStatus:     http.StatusOK,
			expectedCount:      2,
			expectedTotalCount: 2,
			expectedUsers:      []*ent.User{sortableUser2, sortableUser1},
			mustUsers:          true,
			mustUsersOrder:     true,
		},
		{
			name:               "predicate-SORT-DESC-in",
			uri:                "/users?name.in=" + esc(sortableUser1.Name) + "&name.in=" + esc(sortableUser2.Name) + "&sort=name&order=desc",
			expectedPage:       1,
			expectedIsLastPage: true,
			expectedStatus:     http.StatusOK,
			expectedCount:      2,
			expectedTotalCount: 2,
			expectedUsers:      []*ent.User{sortableUser1, sortableUser2},
			mustUsers:          true,
			mustUsersOrder:     true,
		},
		{
			name:               "predicate-updated-at-lt",
			uri:                "/users?updatedAt.lt=" + esc(createdTime.Format(time.RFC3339Nano)),
			expectedPage:       1,
			expectedIsLastPage: true,
			expectedStatus:     http.StatusOK,
			expectedCount:      currentUserCount,
			expectedTotalCount: currentUserCount,
		},
		{
			name:               "predicate-updated-at-gt",
			uri:                "/users?updatedAt.gt=" + esc(createdTime.Format(time.RFC3339Nano)),
			expectedPage:       1,
			expectedIsLastPage: false,
			expectedStatus:     http.StatusOK,
			expectedCount:      rest.UserPageConfig.ItemsPerPage,
			expectedTotalCount: totalUsers - currentUserCount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := enttest.Request[rest.PagedResponse[ent.User]](ctx, s, http.MethodGet, tt.uri, nil)

			require.Equal(t, tt.expectedStatus, resp.Data.Code, "expected status %d, got %d", tt.expectedStatus, resp.Data.Code)

			// Don't do any further checks if the status code is non-ok (since we're expecting this).
			if resp.Data.Code == http.StatusNoContent || resp.Data.Code < 200 || resp.Data.Code >= 300 {
				return
			}

			assert.Equal(t, tt.expectedTotalCount, resp.Value.TotalCount)
			assert.Len(t, resp.Value.Content, tt.expectedCount)
			assert.Equal(t, tt.expectedIsLastPage, resp.Value.IsLastPage)

			if len(tt.expectedUsers) < 1 {
				return
			}

			var needIDs []uuid.UUID
			for _, u := range tt.expectedUsers {
				needIDs = append(needIDs, u.ID)
			}

			var responseIDs []uuid.UUID
			for _, u := range resp.Value.Content {
				responseIDs = append(responseIDs, u.ID)
			}

			if tt.mustUsers {
				if tt.mustUsersOrder {
					assert.Equal(t, needIDs, responseIDs, "expected users to be in the same order & have the same users")
				} else {
					assert.ElementsMatch(t, responseIDs, needIDs, "expected users to be exact")
				}
			} else {
				assert.Subset(t, responseIDs, needIDs, "expected users to be subset")
			}
		})
	}
}

func TestHandler_Create(t *testing.T) {
	ctx, db, s := newRestServer(t, nil)
	t.Cleanup(func() { db.Close() })

	user1 := newUser(db).SaveX(ctx)

	data := map[string]any{
		"name":      gofakeit.FirstName(),
		"nicknames": []string{gofakeit.FirstName(), gofakeit.FirstName()},
		"age":       gofakeit.Number(1, 20),
		"type":      pet.TypeDog,
		"owner":     user1.ID,
	}

	resp := enttest.Request[ent.Pet](ctx, s, http.MethodPost, "/pets", data).Must(t)

	// Ensure the pet in the DB is the same as the one we created, as well as the response body.
	pet1 := rest.EagerLoadPet(db.Pet.Query().Where(pet.ID(resp.Value.ID))).OnlyX(ctx)

	assert.Equal(t, http.StatusCreated, resp.Data.Code)
	assert.Equal(t, data["name"], pet1.Name)
	assert.Equal(t, data["nicknames"], pet1.Nicknames)
	assert.Equal(t, data["age"], pet1.Age)
	assert.Equal(t, data["type"], pet1.Type)
	assert.Equal(t, user1.ID, pet1.Edges.Owner.ID)
}

func TestHandler_Update(t *testing.T) {
	ctx, db, s := newRestServer(t, nil)
	t.Cleanup(func() { db.Close() })

	categories := db.Category.CreateBulk(enttest.Multiple(newCategory, db, 10)...).SaveX(ctx)
	user1 := newUser(db).SaveX(ctx)
	pet1 := newPet(db).SetOwner(user1).AddCategories(categories[0]).SaveX(ctx)

	data := map[string]any{
		"name":           gofakeit.Regex("^[a-z][a-z-]{10,40}$"),
		"age":            25,
		"type":           pet.TypeCat,
		"add_categories": []int{categories[1].ID},
	}

	resp := enttest.Request[ent.Pet](ctx, s, http.MethodPatch, "/pets/"+strconv.Itoa(pet1.ID), data).Must(t)

	assert.Equal(t, http.StatusOK, resp.Data.Code)
	assert.Equal(t, data["name"], resp.Value.Name)
	assert.Equal(t, data["age"], resp.Value.Age)
	assert.Equal(t, data["type"], resp.Value.Type)
	require.Len(t, resp.Value.Edges.Categories, 2)
	assert.ElementsMatch(t, []int{categories[0].ID, categories[1].ID}, []int{resp.Value.Edges.Categories[0].ID, resp.Value.Edges.Categories[1].ID})

	// Bulk update, which should only be enabled on the pet->categories edge.
	data = map[string]any{
		"categories": []int{categories[2].ID, categories[3].ID},
	}

	resp = enttest.Request[ent.Pet](ctx, s, http.MethodPatch, "/pets/"+strconv.Itoa(pet1.ID), data).Must(t)

	assert.Equal(t, http.StatusOK, resp.Data.Code)
	assert.Len(t, resp.Value.Edges.Categories, 2)
	assert.ElementsMatch(t, []int{categories[2].ID, categories[3].ID}, []int{resp.Value.Edges.Categories[0].ID, resp.Value.Edges.Categories[1].ID})

	// Now try just "remove_categories".
	data = map[string]any{
		"remove_categories": []int{categories[3].ID},
	}

	resp = enttest.Request[ent.Pet](ctx, s, http.MethodPatch, "/pets/"+strconv.Itoa(pet1.ID), data).Must(t)

	assert.Equal(t, http.StatusOK, resp.Data.Code)
	assert.Len(t, resp.Value.Edges.Categories, 1)
	assert.Equal(t, categories[2].ID, resp.Value.Edges.Categories[0].ID)
}

func TestHandler_Delete(t *testing.T) {
	ctx, db, s := newRestServer(t, nil)
	t.Cleanup(func() { db.Close() })

	pet1 := newPet(db).SaveX(ctx)

	resp := enttest.Request[string](ctx, s, http.MethodDelete, "/pets/"+strconv.Itoa(pet1.ID), nil).Must(t)

	assert.Equal(t, http.StatusNoContent, resp.Data.Code)

	resp = enttest.Request[string](ctx, s, http.MethodDelete, "/pets/"+strconv.Itoa(pet1.ID), nil)
	require.NotNil(t, resp.Error)
	assert.Equal(t, http.StatusNotFound, resp.Data.Code)
}

func TestHandler_SortRandom(t *testing.T) {
	ctx, db, s := newRestServer(t, nil)
	t.Cleanup(func() { db.Close() })

	db.Pet.CreateBulk(enttest.Multiple(newPet, db, 100)...).ExecX(ctx)

	var results [][]*ent.Pet

	for range 50 {
		resp := enttest.Request[rest.PagedResponse[ent.Pet]](ctx, s, http.MethodGet, "/pets?page=1&per_page=100&sort=random", nil)
		require.Equal(t, http.StatusOK, resp.Data.Code)
		require.Len(t, resp.Value.Content, 100)
		results = append(results, resp.Value.Content)
	}

	// Ensure that all results are different.
	for i := range results {
		if i == 0 {
			continue
		}
		assert.NotEqual(t, results[i-1], results[i])
	}
}

func TestHandler_StrictMutate(t *testing.T) {
	ctx, db, s := newRestServer(t, nil)
	t.Cleanup(func() { db.Close() })

	data := map[string]any{
		"name":        "foo",
		"nonexistent": "bar",
	}

	// POST/create.
	resp := enttest.Request[ent.Pet](ctx, s, http.MethodPost, "/pets", data)
	require.Equal(t, http.StatusBadRequest, resp.Data.Code)
	require.NotNil(t, resp.Error)
	assert.Equal(t, http.StatusBadRequest, resp.Error.Code)

	// PATCH/update.
	pet1 := newPet(db).SaveX(ctx)
	resp = enttest.Request[ent.Pet](ctx, s, http.MethodPatch, "/pets/"+strconv.Itoa(pet1.ID), data)
	require.Equal(t, http.StatusBadRequest, resp.Data.Code)
	require.NotNil(t, resp.Error)
	assert.Equal(t, http.StatusBadRequest, resp.Error.Code)
}

func TestHandler_Valuer(t *testing.T) {
	ctx, db, s := newRestServer(t, nil)
	t.Cleanup(func() { db.Close() })

	data := map[string]any{
		"name":            "John Smith",
		"email":           "john.smith@example.com",
		"enabled":         true,
		"type":            user.TypeUser,
		"password_hashed": gofakeit.Password(true, true, true, true, true, 15),
		"profile_url":     "https://test.example.com/some-path",
	}

	resp := enttest.Request[ent.User](ctx, s, http.MethodPost, "/users", data).Must(t)
	assert.Equal(t, http.StatusCreated, resp.Data.Code)
	assert.Equal(t, data["profile_url"], resp.Value.ProfileURL.String())
}

func TestHandler_DefaultSort(t *testing.T) {
	ctx, db, s := newRestServer(t, nil)
	t.Cleanup(func() { db.Close() })

	pet1 := newPet(db).SetName("cdefgh").SaveX(ctx)
	pet2 := newPet(db).SetName("bcdefg").SaveX(ctx)
	pet3 := newPet(db).SetName("abcdef").SaveX(ctx)

	// First fetch pets with no sorting in the query. Theoretically, default is ID & ASC (1, 2, 3),
	// but we override this to be name & ASC (3, 2, 1).
	resp1 := enttest.Request[rest.PagedResponse[ent.Pet]](ctx, s, http.MethodGet, "/pets", nil).Must(t)

	assert.Equal(t, http.StatusOK, resp1.Data.Code)
	require.Len(t, resp1.Value.Content, 3)
	assert.Equal(t, pet3.Name, resp1.Value.Content[0].Name)
	assert.Equal(t, pet2.Name, resp1.Value.Content[1].Name)
	assert.Equal(t, pet1.Name, resp1.Value.Content[2].Name)

	// Now fetch another entity type, and see if the eager-loaded edges also have the appropriate
	// default sorting.
	user1 := newUser(db).AddPets(pet1, pet2, pet3).SaveX(ctx)
	resp2 := enttest.Request[ent.User](ctx, s, http.MethodGet, "/users/"+user1.ID.String(), nil).Must(t)

	assert.Equal(t, http.StatusOK, resp2.Data.Code)
	require.Len(t, resp2.Value.Edges.Pets, 3)
	assert.Equal(t, pet3.Name, resp2.Value.Edges.Pets[0].Name)
	assert.Equal(t, pet2.Name, resp2.Value.Edges.Pets[1].Name)
	assert.Equal(t, pet1.Name, resp2.Value.Edges.Pets[2].Name)
}

func TestHandler_EagerLoadLimit(t *testing.T) {
	ctx, db, s := newRestServer(t, nil)
	t.Cleanup(func() { db.Close() })

	// Ensure that the limit is respected when eager-loading.
	pet1 := newPet(db).SaveX(ctx)

	db.Category.CreateBulk(enttest.Multiple(func(db *ent.Client) *ent.CategoryCreate {
		return newCategory(db).AddPets(pet1)
	}, db, 1020)...).ExecX(ctx)

	resp1 := enttest.Request[ent.Pet](ctx, s, http.MethodGet, "/pets/"+strconv.Itoa(pet1.ID), nil).Must(t)
	require.Len(t, resp1.Value.Edges.Categories, 1000)

	// And when we hit something which has no limit...
	user1 := newUser(db).SaveX(ctx)

	db.Pet.CreateBulk(enttest.Multiple(func(db *ent.Client) *ent.PetCreate {
		return newPet(db).SetOwner(user1)
	}, db, 1020)...).ExecX(ctx)

	resp2 := enttest.Request[ent.User](ctx, s, http.MethodGet, "/users/"+user1.ID.String(), nil).Must(t)
	require.Len(t, resp2.Value.Edges.Pets, 1020)
}

func TestHandler_FilterGroups(t *testing.T) {
	ctx, db, s := newRestServer(t, nil)
	t.Cleanup(func() { db.Close() })

	db.User.CreateBulk(enttest.Multiple(newUser, db, 100)...).ExecX(ctx)

	user1 := newUser(db).
		SetName("123username").
		SetDescription("456description").
		SetEmail("789email@example.com").
		SetEnabled(true).
		SaveX(ctx)

	newUser(db).
		SetName("123username").
		SetDescription("456description").
		SetEmail("789email@example.com").
		SetEnabled(false).
		ExecX(ctx)

	// Search for three different distinct fields using the 1 filter group, and make
	// sure we only (and always) get user1 back. The logic for this query should be:
	//   - and(enabled.eq==true, or(name.ihas=="123user", description.ihas=="123user", email.ihas=="123user"))
	//
	// enabled.eq should always be AND'd with the OR'd fields of the filter group.

	tests := []string{"123user", "456desc", "789email"}

	for _, test := range tests {
		resp := enttest.Request[rest.PagedResponse[ent.User]](ctx, s, http.MethodGet, "/users?enabled.eq=true&search.ihas="+test, nil).Must(t)
		require.Equal(t, http.StatusOK, resp.Data.Code)
		require.Len(t, resp.Value.Content, 1)
		assert.Equal(t, user1.ID, resp.Value.Content[0].ID)
	}
}
