// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package entrest

import (
	"testing"

	"entgo.io/ent/entc/gen"
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

	r := mustBuildSpec(t, &Config{}, func(g *gen.Graph) {
		injectAnnotations(t, g, "Pet", WithAdditionalTags("Foo"))
		injectAnnotations(t, g, "Pet.categories", WithAdditionalTags("Bar"))
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

	r := mustBuildSpec(t, &Config{}, func(g *gen.Graph) {
		injectAnnotations(t, g, "Pet", WithTags("Foo"))
		injectAnnotations(t, g, "Pet.categories", WithTags("Bar"))
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

	r := mustBuildSpec(t, &Config{}, func(g *gen.Graph) {
		injectAnnotations(t, g, "Pet.categories", WithEdgeUpdateBulk(true))
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
