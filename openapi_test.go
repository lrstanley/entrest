// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package entrest

import (
	"net/http"
	"testing"

	"github.com/ogen-go/ogen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
