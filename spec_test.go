// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package entrest

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"testing"

	"entgo.io/ent/entc/gen"
	"github.com/ogen-go/ogen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpec_ThroughSchema_TwoTypes(t *testing.T) {
	t.Parallel()
	r := mustBuildSpec(t, &Config{})

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
	r := mustBuildSpec(t, &Config{})

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

func TestSpec_HoistedEnums(t *testing.T) {
	t.Parallel()

	r := mustBuildSpec(t, &Config{
		PreGenerateHook: func(g *gen.Graph, _ *ogen.Spec) error {
			injectAnnotations(t, g, "User.type", WithFilter(FilterGroupEqualExact|FilterGroupArray))
			return nil
		},
	})

	assert.NotNil(t, r.json(`$.components.schemas.UserTypeEnum`))
	assert.Equal(t, "string", r.json(`$.components.schemas.UserTypeEnum.type`))
	assert.Contains(t, r.json(`$.components.schemas.UserTypeEnum.enum`), "USER")
	assert.Contains(t, r.json(`$.components.schemas.UserTypeEnum.enum`), "SYSTEM")

	assert.Contains(t, r.json(`$.components.schemas.User.properties.type.$ref`), "/UserTypeEnum")
	assert.Contains(t, r.json(`$.components.schemas.UserCreate.properties.type.$ref`), "/UserTypeEnum")
	assert.Contains(t, r.json(`$.components.schemas.UserUpdate.properties.type.$ref`), "/UserTypeEnum")
	assert.Contains(t, r.json(`$.components.parameters.UserTypeEQ.schema.$ref`), "/UserTypeEnum")
	assert.Contains(t, r.json(`$.components.parameters.UserTypeNEQ.schema.$ref`), "/UserTypeEnum")
	assert.Contains(t, r.json(`$.components.parameters.UserTypeIn.schema.items.$ref`), "/UserTypeEnum")
	assert.Contains(t, r.json(`$.components.parameters.UserTypeNotIn.schema.items.$ref`), "/UserTypeEnum")
}

func TestSpec_Sensitive(t *testing.T) {
	t.Parallel()

	r := mustBuildSpec(t, &Config{})

	assert.Nil(t, r.json(`$.components.schemas.User.properties.password_hashed`))
	assert.NotNil(t, r.json(`$.components.schemas.UserCreate.properties.password_hashed`))
	assert.NotNil(t, r.json(`$.components.schemas.UserUpdate.properties.password_hashed`))

	// convert all of the parameters to json, and see if the field exists anywhere in those.
	b, err := json.Marshal(r.spec.Components.Parameters)
	require.NoError(t, err)
	assert.NotContains(t, string(b), `"password_hashed"`)
}

var testRequiredMethods = []string{
	http.MethodGet,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodOptions,
	http.MethodHead,
	http.MethodTrace,
}

func TestPatchOperations(t *testing.T) {
	spec := ogen.NewSpec().AddPathItem("/test", &ogen.PathItem{
		Get:     &ogen.Operation{OperationID: http.MethodGet},
		Post:    &ogen.Operation{OperationID: http.MethodPost},
		Put:     &ogen.Operation{OperationID: http.MethodPut},
		Patch:   &ogen.Operation{OperationID: http.MethodPatch},
		Delete:  &ogen.Operation{OperationID: http.MethodDelete},
		Options: &ogen.Operation{OperationID: http.MethodOptions},
		Head:    &ogen.Operation{OperationID: http.MethodHead},
		Trace:   &ogen.Operation{OperationID: http.MethodTrace},
	})

	var foundMethods []string

	PatchOperations(spec.Paths["/test"], func(method string, op *ogen.Operation) *ogen.Operation {
		foundMethods = append(foundMethods, method)
		return op
	})

	for _, method := range testRequiredMethods {
		t.Run(method, func(t *testing.T) {
			assert.Contains(t, foundMethods, method)
		})
	}
}

func TestPatchPathItem(t *testing.T) {
	spec := ogen.NewSpec().AddPathItem("/test", &ogen.PathItem{
		Get: &ogen.Operation{
			OperationID: "getTest",
			Responses: map[string]*ogen.Response{
				"200": {Description: http.MethodGet},
			},
		},
		Post: &ogen.Operation{
			OperationID: "postTest",
			Responses: map[string]*ogen.Response{
				"200": {Description: http.MethodPost},
			},
		},
		Put: &ogen.Operation{
			OperationID: "putTest",
			Responses: map[string]*ogen.Response{
				"200": {Description: http.MethodPut},
			},
		},
		Patch: &ogen.Operation{
			OperationID: "patchTest",
			Responses: map[string]*ogen.Response{
				"200": {Description: http.MethodPatch},
			},
		},
		Delete: &ogen.Operation{
			OperationID: "deleteTest",
			Responses: map[string]*ogen.Response{
				"200": {Description: http.MethodDelete},
			},
		},
		Options: &ogen.Operation{
			OperationID: "optionsTest",
			Responses: map[string]*ogen.Response{
				"200": {Description: http.MethodOptions},
			},
		},
		Head: &ogen.Operation{
			OperationID: "headTest",
			Responses: map[string]*ogen.Response{
				"200": {Description: http.MethodHead},
			},
		},
		Trace: &ogen.Operation{
			OperationID: "traceTest",
			Responses: map[string]*ogen.Response{
				"200": {Description: http.MethodTrace},
			},
		},
	})

	var foundMethods []string

	PatchPathItem(spec.Paths["/test"], func(resp *ogen.Response) *ogen.Response {
		foundMethods = append(foundMethods, resp.Description)
		return resp
	})

	for _, method := range testRequiredMethods {
		t.Run(method, func(t *testing.T) {
			assert.Contains(t, foundMethods, method)
		})
	}
}

func TestMergeOperation(t *testing.T) {
	tests := []struct {
		name    string
		orig    *ogen.Operation
		toMerge *ogen.Operation
		out     *ogen.Operation
		overlap bool
		wantErr bool
	}{
		{
			name:    "empty",
			orig:    &ogen.Operation{},
			toMerge: &ogen.Operation{},
			out:     &ogen.Operation{},
			overlap: false,
			wantErr: false,
		},
		{
			name: "overlap-allowed",
			orig: &ogen.Operation{
				OperationID: "foo",
				Responses:   ogen.Responses{"200": {Description: "ok"}},
			},
			toMerge: &ogen.Operation{
				Responses: ogen.Responses{"200": {Description: "foo"}},
			},
			out: &ogen.Operation{
				OperationID: "foo",
				Responses:   ogen.Responses{"200": {Description: "foo"}},
			},
			overlap: true,
			wantErr: false,
		},
		{
			name: "overlap-not-allowed",
			orig: &ogen.Operation{
				Responses: ogen.Responses{"200": {Description: "ok"}},
			},
			toMerge: &ogen.Operation{
				Responses: ogen.Responses{"200": {Description: "ok"}},
			},
			overlap: false,
			wantErr: true,
		},
		{
			name: "multi-response",
			orig: &ogen.Operation{
				Description: "foo",
				Responses:   ogen.Responses{"200": {Description: "ok"}},
			},
			toMerge: &ogen.Operation{
				Responses: ogen.Responses{"201": {Description: "ok"}},
			},
			out: &ogen.Operation{
				Description: "foo",
				Responses: ogen.Responses{
					"200": {Description: "ok"},
					"201": {Description: "ok"},
				},
			},
			overlap: true,
			wantErr: false,
		},
		{
			name: "empty-to-things",
			orig: &ogen.Operation{Responses: ogen.Responses{}},
			toMerge: &ogen.Operation{
				OperationID: "foo",
				Description: "foo bar",
				Responses:   ogen.Responses{"200": {Description: "ok"}},
			},
			out: &ogen.Operation{
				OperationID: "foo",
				Description: "foo bar",
				Responses:   ogen.Responses{"200": {Description: "ok"}},
			},
			overlap: true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mergeOperation(tt.overlap, tt.orig, tt.toMerge)
			require.Equal(t, tt.wantErr, err != nil, "unexpected error: %v", err)
			if tt.wantErr {
				return
			}
			assert.Equal(t, tt.out, tt.orig)
		})
	}
}

func TestMergeSpec(t *testing.T) {
	tests := []struct {
		name    string
		orig    *ogen.Spec
		toMerge []*ogen.Spec
		out     *ogen.Spec
		overlap bool
		wantErr bool
	}{
		{
			name:    "empty",
			orig:    &ogen.Spec{},
			toMerge: []*ogen.Spec{},
			out:     &ogen.Spec{},
			overlap: false,
			wantErr: false,
		},
		{
			name: "overlap-allowed",
			orig: &ogen.Spec{
				Tags: []ogen.Tag{
					{Name: "foo", Description: "foo bar"},
				},
				Servers: []ogen.Server{{URL: "http://localhost:8080"}},
				Paths: ogen.Paths{
					"/foo": &ogen.PathItem{
						Get: &ogen.Operation{
							Summary:     "foo",
							Description: "foo bar",
							Parameters: []*ogen.Parameter{
								{Name: "foo", In: "query", Schema: &ogen.Schema{Type: "string"}},
							},
							Responses: ogen.Responses{"200": {Description: "ok"}},
						},
						Post: &ogen.Operation{
							Responses: ogen.Responses{"200": {Description: "ok"}},
							Parameters: []*ogen.Parameter{
								{Name: "foo", In: "query", Schema: &ogen.Schema{Type: "string"}},
							},
						},
					},
					"/bar": &ogen.PathItem{
						Get: &ogen.Operation{
							Responses: ogen.Responses{"200": {Description: "ok"}},
						},
						Post: &ogen.Operation{
							Responses: ogen.Responses{"200": {Description: "ok"}},
							Parameters: []*ogen.Parameter{
								{Name: "foo", In: "query", Schema: &ogen.Schema{Type: "string"}},
							},
						},
					},
				},
				Components: &ogen.Components{
					Schemas: map[string]*ogen.Schema{
						"Foo": {Description: "foo", Type: "string"},
						"Bar": {Type: "string"},
					},
				},
			},
			toMerge: []*ogen.Spec{{
				Servers: []ogen.Server{{URL: "http://foobar:8080"}},
				Tags: []ogen.Tag{
					{Name: "bar", Description: "bar baz"},
				},
				Paths: ogen.Paths{
					"/foo": &ogen.PathItem{
						Get: &ogen.Operation{
							Responses: ogen.Responses{"200": {Description: "foo"}},
						},
					},
				},
				Components: &ogen.Components{
					Schemas: map[string]*ogen.Schema{
						"Foo": {Description: "bar", Type: "number"},
					},
				},
			}},
			out: &ogen.Spec{
				Servers: []ogen.Server{{URL: "http://localhost:8080"}, {URL: "http://foobar:8080"}},
				Tags: []ogen.Tag{
					{Name: "foo", Description: "foo bar"},
					{Name: "bar", Description: "bar baz"},
				},
				Paths: ogen.Paths{
					"/foo": &ogen.PathItem{
						Get: &ogen.Operation{
							Summary:     "foo",
							Description: "foo bar",
							Parameters: []*ogen.Parameter{
								{Name: "foo", In: "query", Schema: &ogen.Schema{Type: "string"}},
							},
							Responses: ogen.Responses{"200": {Description: "foo"}},
						},
						Post: &ogen.Operation{
							Responses: ogen.Responses{"200": {Description: "ok"}},
							Parameters: []*ogen.Parameter{
								{Name: "foo", In: "query", Schema: &ogen.Schema{Type: "string"}},
							},
						},
					},
					"/bar": &ogen.PathItem{
						Get: &ogen.Operation{
							Responses: ogen.Responses{"200": {Description: "ok"}},
						},
						Post: &ogen.Operation{
							Responses: ogen.Responses{"200": {Description: "ok"}},
							Parameters: []*ogen.Parameter{
								{Name: "foo", In: "query", Schema: &ogen.Schema{Type: "string"}},
							},
						},
					},
				},
				Components: &ogen.Components{
					Schemas: map[string]*ogen.Schema{
						"Foo": {Description: "bar", Type: "number"},
						"Bar": {Type: "string"},
					},
				},
			},
			overlap: true,
			wantErr: false,
		},
		{
			name: "overlap-not-allowed",
			orig: &ogen.Spec{
				Paths: ogen.Paths{
					"/foo": &ogen.PathItem{
						Get: &ogen.Operation{
							Summary:     "foo",
							Description: "foo bar",
							Parameters: []*ogen.Parameter{
								{Name: "foo", In: "query", Schema: &ogen.Schema{Type: "string"}},
							},
							Responses: ogen.Responses{"200": {Description: "foo"}},
						},
					},
				},
			},
			toMerge: []*ogen.Spec{{
				Paths: ogen.Paths{
					"/foo": &ogen.PathItem{
						Get: &ogen.Operation{
							Responses: ogen.Responses{"200": {Description: "foo"}},
						},
					},
				},
			}},
			overlap: false,
			wantErr: true,
		},
		{
			name: "empty-to-things",
			orig: &ogen.Spec{Paths: ogen.Paths{}},
			toMerge: []*ogen.Spec{
				{
					Tags: []ogen.Tag{{Name: "foo", Description: "foo bar"}},
					Paths: ogen.Paths{
						"/foo": &ogen.PathItem{
							Get: &ogen.Operation{
								Responses: ogen.Responses{"200": {Description: "foo"}},
							},
						},
					},
				},
			},
			out: &ogen.Spec{
				Tags: []ogen.Tag{{Name: "foo", Description: "foo bar"}},
				Paths: ogen.Paths{
					"/foo": &ogen.PathItem{
						Get: &ogen.Operation{
							Responses: ogen.Responses{"200": {Description: "foo"}},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mergeSpec(tt.overlap, tt.orig, tt.toMerge...)
			require.Equal(t, tt.wantErr, err != nil, "unexpected error: %v", err)
			if tt.wantErr {
				return
			}
			assert.Equal(t, tt.out, tt.orig)
		})
	}
}
