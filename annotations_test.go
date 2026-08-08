// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package entrest

import (
	"testing"

	"entgo.io/ent/entc/gen"
	"github.com/ogen-go/ogen"
	"github.com/stretchr/testify/assert"
)

func TestGetAnnotation(t *testing.T) {
	// True.
	assert.Equal(
		t,
		ptr(true),
		GetAnnotation(&gen.Type{Annotations: map[string]any{
			Annotation{}.Name(): WithPagination(true),
		}}).Pagination,
	)

	// False.
	assert.Equal(
		t,
		ptr(false),
		GetAnnotation(&gen.Type{Annotations: map[string]any{
			Annotation{}.Name(): WithPagination(false),
		}}).Pagination,
	)

	// Unspecified.
	var ptrBoolNil *bool
	assert.Equal(
		t,
		ptrBoolNil,
		GetAnnotation(&gen.Type{Annotations: map[string]any{
			Annotation{}.Name(): Annotation{},
		}}).Pagination,
	)

	// Test fields.
	assert.True(
		t,
		GetAnnotation(&gen.Field{Annotations: map[string]any{
			Annotation{}.Name(): WithSortable(true),
		}}).Sortable,
	)

	// Test edges.
	assert.Equal(
		t,
		ptr(true),
		GetAnnotation(&gen.Edge{Annotations: map[string]any{
			Annotation{}.Name(): WithEagerLoad(true),
		}}).EagerLoad,
	)
}

func TestValidateAnnotation(t *testing.T) {
	tests := []struct {
		name    string
		value   *gen.Type
		wantErr bool
	}{
		{
			name:  "no-annotation",
			value: &gen.Type{Annotations: map[string]any{}},
		},
		{
			name: "valid-annotation",
			value: &gen.Type{Annotations: map[string]any{
				Annotation{}.Name(): WithPagination(true), // Type's should support pagination.
			}},
			wantErr: false,
		},
		{
			name: "valid-schema-edge-endpoint",
			value: &gen.Type{Annotations: map[string]any{
				Annotation{}.Name(): WithEdgeEndpoint(false),
			}},
			wantErr: false,
		},
		{
			name: "invalid-annotation-type-with-edge",
			value: &gen.Type{Annotations: map[string]any{
				Annotation{}.Name(): WithEagerLoad(false), // Only edges support eager loading.
			}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAnnotations(tt.value)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestAnnotation_Merge(t *testing.T) {
	tests := []struct {
		name        string
		annotations []Annotation
		want        Annotation
	}{
		{
			name: "no-annotations",
			annotations: []Annotation{
				{},
				{},
			},
			want: Annotation{},
		},
		{
			name: "overlap-single",
			annotations: []Annotation{
				WithPagination(true),
				WithPagination(false),
			},
			want: Annotation{
				Pagination: ptr(false),
			},
		},
		{
			name: "overlap-multiple",
			annotations: []Annotation{
				WithPagination(false),
				WithDescription("foo"),
				WithPagination(true),
				WithDescription("bar"),
			},
			want: Annotation{
				Pagination:  ptr(true),
				Description: "bar",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out Annotation
			for _, a := range tt.annotations {
				out, _ = out.Merge(a).(Annotation)
			}
			assert.Equal(t, tt.want, out)
		})
	}
}

func TestAnnotation_AdditionalTags(t *testing.T) {
	t.Parallel()

	r := mustBuildSpec(t, &Config{
		PreGenerateHook: func(g *gen.Graph, _ *ogen.Spec) error {
			injectAnnotations(t, g, "Pet", WithAdditionalTags("Foo"))
			injectAnnotations(t, g, "Pet.categories", WithAdditionalTags("Bar"))
			return nil
		},
	})

	assert.Contains(t, r.json(`$.paths./pets.get.tags`), "Foo")
	assert.Contains(t, r.json(`$.paths./pets.post.tags`), "Foo")
	assert.Contains(t, r.json(`$.paths./pets/{petID}.get.tags`), "Foo")
	assert.Contains(t, r.json(`$.paths./pets/{petID}.patch.tags`), "Foo")
	assert.Contains(t, r.json(`$.paths./pets/{petID}.delete.tags`), "Foo")
	assert.Contains(t, r.json(`$.paths./pets/{petID}/categories.get.tags`), "Bar")
}

func TestAnnotation_Tags(t *testing.T) {
	t.Parallel()

	r := mustBuildSpec(t, &Config{
		PreGenerateHook: func(g *gen.Graph, _ *ogen.Spec) error {
			injectAnnotations(t, g, "Pet", WithTags("Foo"))
			injectAnnotations(t, g, "Pet.categories", WithTags("Bar"))
			return nil
		},
	})

	assert.Equal(t, "Foo", r.json(`$.paths./pets.get.tags.*`))
	assert.Equal(t, "Foo", r.json(`$.paths./pets.post.tags.*`))
	assert.Equal(t, "Foo", r.json(`$.paths./pets/{petID}.get.tags.*`))
	assert.Equal(t, "Foo", r.json(`$.paths./pets/{petID}.patch.tags.*`))
	assert.Equal(t, "Foo", r.json(`$.paths./pets/{petID}.delete.tags.*`))
	assert.Equal(t, "Bar", r.json(`$.paths./pets/{petID}/categories.get.tags.*`))
}

func TestAnnotation_EdgeUpdateBulk(t *testing.T) {
	t.Parallel()

	r := mustBuildSpec(t, &Config{
		PreGenerateHook: func(g *gen.Graph, _ *ogen.Spec) error {
			injectAnnotations(t, g, "Pet.categories", WithEdgeUpdateBulk(true))
			return nil
		},
	})

	// Ensure create and update on categories have the bulk field in the right place.
	assert.NotNil(t, r.json(`$.components.schemas.PetCreate.properties.categories`))
	assert.Nil(t, r.json(`$.components.schemas.PetCreate.properties.add_categories`))
	assert.Nil(t, r.json(`$.components.schemas.PetCreate.properties.remove_categories`))
	assert.NotNil(t, r.json(`$.components.schemas.PetUpdate.properties.categories`))
	assert.NotNil(t, r.json(`$.components.schemas.PetUpdate.properties.add_categories`))
	assert.NotNil(t, r.json(`$.components.schemas.PetUpdate.properties.remove_categories`))

	// And ensure it's not on others.
	assert.NotNil(t, r.json(`$.components.schemas.PetCreate.properties.friends`))
	assert.Nil(t, r.json(`$.components.schemas.PetCreate.properties.add_friends`))
	assert.Nil(t, r.json(`$.components.schemas.PetCreate.properties.remove_friends`))
	assert.Nil(t, r.json(`$.components.schemas.PetUpdate.properties.friends`))
	assert.NotNil(t, r.json(`$.components.schemas.PetUpdate.properties.add_friends`))
	assert.NotNil(t, r.json(`$.components.schemas.PetUpdate.properties.remove_friends`))
}

func TestAnnotation_IncludeOperationsEdgeInheritance(t *testing.T) {
	t.Parallel()

	t.Run("schema-ops-inherit-to-edges", func(t *testing.T) {
		t.Parallel()
		r := mustBuildSpec(t, &Config{
			PreGenerateHook: func(g *gen.Graph, _ *ogen.Spec) error {
				// Restrict schema to create+update — edges without their own
				// operations annotation inherit this and lose GET endpoints.
				injectAnnotations(t, g, "Pet", WithIncludeOperations(OperationCreate, OperationUpdate))
				return nil
			},
		})

		assert.Nil(t, r.json(`$.paths./pets.get`))
		assert.Nil(t, r.json(`$.paths./pets/{petID}.get`))
		assert.NotNil(t, r.json(`$.paths./pets.post`))
		assert.NotNil(t, r.json(`$.paths./pets/{petID}.patch`))

		assert.Nil(t, r.json(`$.paths./pets/{petID}/owner.get`))
		assert.Nil(t, r.json(`$.paths./pets/{petID}/categories.get`))
		assert.Nil(t, r.json(`$.paths./pets/{petID}/friends.get`))
	})

	t.Run("edge-ops-override-schema", func(t *testing.T) {
		t.Parallel()
		r := mustBuildSpec(t, &Config{
			PreGenerateHook: func(g *gen.Graph, _ *ogen.Spec) error {
				injectAnnotations(t, g, "Pet", WithIncludeOperations(OperationCreate, OperationUpdate))
				// Opt owner back in for read endpoints; categories stays inherited.
				injectAnnotations(t, g, "Pet.owner", WithIncludeOperations(OperationRead, OperationUpdate))
				return nil
			},
		})

		assert.NotNil(t, r.json(`$.paths./pets/{petID}/owner.get`))
		assert.Nil(t, r.json(`$.paths./pets/{petID}/categories.get`))
		assert.Nil(t, r.json(`$.paths./pets/{petID}/friends.get`))
	})

	t.Run("per-edge-exclude", func(t *testing.T) {
		t.Parallel()
		r := mustBuildSpec(t, &Config{
			PreGenerateHook: func(g *gen.Graph, _ *ogen.Spec) error {
				injectAnnotations(t, g, "Pet.categories", WithExcludeOperations(OperationList))
				injectAnnotations(t, g, "Pet.owner", WithExcludeOperations(OperationRead))
				return nil
			},
		})

		assert.Nil(t, r.json(`$.paths./pets/{petID}/categories.get`))
		assert.Nil(t, r.json(`$.paths./pets/{petID}/owner.get`))
		assert.NotNil(t, r.json(`$.paths./pets/{petID}/friends.get`))
	})

	t.Run("edge-update-without-create", func(t *testing.T) {
		t.Parallel()
		r := mustBuildSpec(t, &Config{
			PreGenerateHook: func(g *gen.Graph, _ *ogen.Spec) error {
				// Allow updating the owner edge only (not on create).
				injectAnnotations(t, g, "Pet.owner", WithIncludeOperations(OperationUpdate, OperationRead, OperationList, OperationDelete))
				injectAnnotations(t, g, "Pet.friends", WithExcludeOperations(OperationUpdate))
				return nil
			},
		})

		assert.Nil(t, r.json(`$.components.schemas.PetCreate.properties.owner`))
		assert.NotNil(t, r.json(`$.components.schemas.PetUpdate.properties.owner`))
		assert.NotNil(t, r.json(`$.components.schemas.PetCreate.properties.friends`))
		assert.Nil(t, r.json(`$.components.schemas.PetUpdate.properties.add_friends`))
		assert.Nil(t, r.json(`$.components.schemas.PetUpdate.properties.remove_friends`))
	})

	t.Run("schema-exclude-create-omits-create-schema", func(t *testing.T) {
		t.Parallel()
		r := mustBuildSpec(t, &Config{
			PreGenerateHook: func(g *gen.Graph, _ *ogen.Spec) error {
				injectAnnotations(t, g, "Pet", WithExcludeOperations(OperationCreate, OperationDelete))
				return nil
			},
		})

		assert.Nil(t, r.json(`$.paths./pets.post`))
		assert.Nil(t, r.json(`$.paths./pets/{petID}.delete`))
		assert.Nil(t, r.json(`$.components.schemas.PetCreate`))
		assert.NotNil(t, r.json(`$.paths./pets/{petID}.patch`))
		assert.NotNil(t, r.json(`$.components.schemas.PetUpdate`))
	})
}
