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

func mergeAnnotations(t *testing.T, in gen.Annotations, annotations ...Annotation) gen.Annotations {
	t.Helper()
	ant := *decodeAnnotation(in)
	for _, a := range annotations { // nolint:gocritic
		ant, _ = ant.Merge(a).(Annotation)
	}
	in.Set(ant.Name(), ant)
	return in
}

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
	assert.Equal(
		t,
		true,
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
			assert.True(t, tt.wantErr == (err != nil))
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

func TestAnnotation_With(t *testing.T) {
	tests := []struct {
		name       string
		annotation Annotation
		want       Annotation
	}{
		{name: "no-annotation", annotation: Annotation{}, want: Annotation{}},
		{
			name:       "with-operation-summary",
			annotation: WithOperationSummary(OperationCreate, "foo"),
			want: Annotation{
				OperationSummary: map[Operation]string{
					OperationCreate: "foo",
				},
			},
		},
		{
			name:       "with-operation-description",
			annotation: WithOperationDescription(OperationCreate, "foo"),
			want: Annotation{
				OperationDescription: map[Operation]string{
					OperationCreate: "foo",
				},
			},
		},
		{
			name:       "with-additional-tags",
			annotation: WithAdditionalTags("foo", "bar"),
			want:       Annotation{AdditionalTags: []string{"foo", "bar"}},
		},
		{
			name:       "with-tags",
			annotation: WithTags("foo", "bar"),
			want:       Annotation{Tags: []string{"foo", "bar"}},
		},
		{
			name:       "with-operation-id",
			annotation: WithOperationID(OperationCreate, "foo"),
			want: Annotation{
				OperationID: map[Operation]string{
					OperationCreate: "foo",
				},
			},
		},
		{name: "with-description", annotation: WithDescription("foo"), want: Annotation{Description: "foo"}},
		{name: "with-pagination", annotation: WithPagination(true), want: Annotation{Pagination: ptr(true)}},
		{name: "with-min-items-per-page", annotation: WithMinItemsPerPage(10), want: Annotation{MinItemsPerPage: 10}},
		{name: "with-max-items-per-page", annotation: WithMaxItemsPerPage(10), want: Annotation{MaxItemsPerPage: 10}},
		{name: "with-items-per-page", annotation: WithItemsPerPage(10), want: Annotation{ItemsPerPage: 10}},
		{name: "with-eager-load", annotation: WithEagerLoad(true), want: Annotation{EagerLoad: ptr(true)}},
		{name: "with-edge-endpoint", annotation: WithEdgeEndpoint(true), want: Annotation{EdgeEndpoint: ptr(true)}},
		{
			name:       "with-filter",
			annotation: WithFilter(FilterGroupArray | FilterGroupLength),
			want:       Annotation{Filter: FilterGroupArray | FilterGroupLength},
		},
		{name: "with-handler", annotation: WithHandler(true), want: Annotation{Handler: ptr(true)}},
		{name: "with-sortable", annotation: WithSortable(true), want: Annotation{Sortable: true}},
		{name: "with-skip", annotation: WithSkip(true), want: Annotation{Skip: true}},
		{name: "with-read-only", annotation: WithReadOnly(true), want: Annotation{ReadOnly: true}},
		{
			name:       "with-example",
			annotation: WithExample(map[string]any{"foo": "bar"}),
			want:       Annotation{Example: map[string]any{"foo": "bar"}},
		},
		{
			name:       "with-schema",
			annotation: WithSchema(&ogen.Schema{Ref: "#/components/schemas/foo"}),
			want:       Annotation{Schema: &ogen.Schema{Ref: "#/components/schemas/foo"}},
		},
		{
			name:       "with-include-operations",
			annotation: WithIncludeOperations(OperationCreate, OperationRead),
			want:       Annotation{Operations: []Operation{OperationCreate, OperationRead}},
		},
		{
			name:       "with-exclude-operations",
			annotation: WithExcludeOperations(OperationCreate, OperationRead),
			want: Annotation{Operations: []Operation{
				OperationUpdate,
				OperationDelete,
				OperationList,
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.annotation)
		})
	}
}
