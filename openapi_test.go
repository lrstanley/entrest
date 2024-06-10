// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package entrest

import (
	"testing"

	"github.com/ogen-go/ogen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			require.True(t, tt.wantErr == (err != nil), "unexpected error: %v", err)
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
			require.True(t, tt.wantErr == (err != nil), "unexpected error: %v", err)
			if tt.wantErr {
				return
			}
			assert.Equal(t, tt.out, tt.orig)
		})
	}
}
