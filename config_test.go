// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package entrest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"testing"

	"entgo.io/ent/entc/gen"
	"github.com/ogen-go/ogen"
	"github.com/stretchr/testify/assert"
)

func TestEnsureIntegration(t *testing.T) {
	t.Parallel()

	// This test ensures that all of the integration test schemas don't have any
	// pre-attached annotations, which would cause the test to fail or succeed
	// in a way that is not expected.

	r := mustBuildSpec(t, &Config{}, nil)

	check := func(a *Annotation) error {
		t.Helper()
		out := map[string]any{}

		b, err := json.Marshal(a)
		if err != nil {
			return err
		}

		err = json.Unmarshal(b, &out)
		if err != nil {
			return err
		}

		if _, ok := out["Schema"]; (ok && len(out) != 1) || (!ok && len(out) != 0) {
			return fmt.Errorf("only 'Schema' annotation should exist in integration tests, but found others: %#v", out)
		}
		return nil
	}

	for _, n := range r.graph.Nodes {
		na := GetAnnotation(n)
		t.Run(n.Name, func(t *testing.T) {
			t.Parallel()
			if err := check(na); err != nil {
				t.Error(err)
			}
		})

		for _, f := range n.Fields {
			t.Run(n.Name+"/"+f.Name, func(t *testing.T) {
				if err := check(GetAnnotation(f)); err != nil {
					t.Error(err)
				}
			})
		}
		for _, e := range n.Edges {
			t.Run(n.Name+"/"+e.Name, func(t *testing.T) {
				if err := check(GetAnnotation(e)); err != nil {
					t.Error(err)
				}
			})
		}
	}
}

func TestConfig_Spec(t *testing.T) {
	t.Parallel()

	t.Run("default-spec", func(t *testing.T) {
		t.Parallel()
		r := mustBuildSpec(t, &Config{Spec: nil}, nil)
		assert.Equal(t, "EntGo Rest API", r.json(`$.info.title`))
	})

	t.Run("with-spec", func(t *testing.T) {
		t.Parallel()
		r := mustBuildSpec(t, &Config{Spec: &ogen.Spec{Info: ogen.Info{Title: "foo"}}}, nil)
		assert.Equal(t, "foo", r.json(`$.info.title`))
	})
}

func TestConfig_DisablePagination(t *testing.T) {
	t.Parallel()

	t.Run("with-pagination", func(t *testing.T) {
		t.Parallel()
		r := mustBuildSpec(t, &Config{DisablePagination: false}, nil)
		assert.Contains(t, r.json(`$.components.schemas.PetList.allOf.*.$ref`), "/PagedResponse")
	})

	t.Run("with-pagination-edge", func(t *testing.T) {
		t.Parallel()
		r := mustBuildSpec(t, &Config{DisablePagination: false}, nil)

		assert.Contains(t, r.json(`$.paths./pets/{petID}/categories..responses..schema.$ref`), "/CategoryList")
		assert.Contains(t, r.json(`$.components.schemas.CategoryList.allOf.*.$ref`), "/PagedResponse")
	})

	t.Run("without-pagination", func(t *testing.T) {
		t.Parallel()
		r := mustBuildSpec(t, &Config{DisablePagination: true}, nil)
		assert.Nil(t, r.json(`$.components.schemas.PetList.allOf`))
	})

	t.Run("no-global-but-local", func(t *testing.T) {
		t.Parallel()
		r := mustBuildSpec(t, &Config{DisablePagination: true}, func(g *gen.Graph) {
			injectAnnotations(t, g, "Pet", WithPagination(true))
		})
		assert.Contains(t, r.json(`$.components.schemas.PetList.allOf.*.$ref`), "/PagedResponse")
	})

	t.Run("no-global-but-local-edge", func(t *testing.T) {
		t.Parallel()
		r := mustBuildSpec(t, &Config{DisablePagination: true}, func(g *gen.Graph) {
			injectAnnotations(t, g, "Pet.categories", WithPagination(true))
		})

		assert.Contains(t, r.json(`$.paths./pets/{petID}/categories..responses..schema.$ref`), "/PetCategoryList")
		assert.Contains(t, r.json(`$.components.schemas.PetCategoryList.allOf.*.$ref`), "/PagedResponse")
	})

	// Same as no-global-but-local-edge but pagination is enabled on the edges
	// underlying type.
	t.Run("no-global-but-local-edge-ref", func(t *testing.T) {
		t.Parallel()
		r := mustBuildSpec(t, &Config{DisablePagination: true}, func(g *gen.Graph) {
			injectAnnotations(t, g, "Category", WithPagination(true))
		})
		assert.Contains(t, r.json(`$.paths./pets/{petID}/categories..responses..schema.$ref`), "/CategoryList")
		assert.Contains(t, r.json(`$.components.schemas.CategoryList.allOf.*.$ref`), "/PagedResponse")
	})
}

func TestConfig_ItemsPerPage(t *testing.T) {
	t.Parallel()

	base := `$.paths./pets..parameters[?(@.name == "per_page")].schema`

	t.Run("defaults", func(t *testing.T) {
		t.Parallel()
		c := &Config{}
		r := mustBuildSpec(t, c, nil)

		assert.Equal(t, float64(c.ItemsPerPage), r.json(base+`.default`))    //nolint:all
		assert.Equal(t, float64(c.MinItemsPerPage), r.json(base+`.minimum`)) //nolint:all
		assert.Equal(t, float64(c.MaxItemsPerPage), r.json(base+`.maximum`)) //nolint:all
	})

	t.Run("with-specified", func(t *testing.T) {
		t.Parallel()
		c := &Config{
			MinItemsPerPage: 2,
			ItemsPerPage:    5,
			MaxItemsPerPage: 999,
		}
		r := mustBuildSpec(t, c, nil)

		assert.Equal(t, float64(c.ItemsPerPage), r.json(base+`.default`))    //nolint:all
		assert.Equal(t, float64(c.MinItemsPerPage), r.json(base+`.minimum`)) //nolint:all
		assert.Equal(t, float64(c.MaxItemsPerPage), r.json(base+`.maximum`)) //nolint:all
	})
}

func TestConfig_DefaultEagerLoad(t *testing.T) {
	t.Parallel()

	t.Run("global-eager-load", func(t *testing.T) {
		t.Parallel()
		r := mustBuildSpec(t, &Config{DefaultEagerLoad: true}, nil)

		assert.NotNil(t, r.json(`$.components.schemas.PetRead..properties.edges`))
		assert.NotNil(t, r.json(`$.components.schemas.PetEdges.properties.owner.$ref`))
	})

	t.Run("no-global-eager-load", func(t *testing.T) {
		t.Parallel()
		r := mustBuildSpec(t, &Config{DefaultEagerLoad: false}, nil)
		assert.Nil(t, r.json(`$.components.schemas.PetRead..properties.edges`))
		assert.Nil(t, r.json(`$.components.schemas.PetEdges`))
	})

	t.Run("no-global-but-local-eager-load", func(t *testing.T) {
		t.Parallel()
		r := mustBuildSpec(t, &Config{DefaultEagerLoad: false}, func(g *gen.Graph) {
			injectAnnotations(t, g, "Pet.categories", WithEagerLoad(true))
		})
		assert.NotNil(t, r.json(`$.components.schemas.PetRead..properties.edges`))
		assert.NotNil(t, r.json(`$.components.schemas.PetEdges.properties.categories.items.$ref`))
		assert.Nil(t, r.json(`$.components.schemas.PetEdges.properties.owner`))
	})
}

func TestConfig_DisableEagerLoadNonPagedOpt(t *testing.T) {
	t.Parallel()

	t.Run("no-opt-with-eager-load", func(t *testing.T) {
		t.Parallel()
		r := mustBuildSpec(t, &Config{DisableEagerLoadNonPagedOpt: true}, func(g *gen.Graph) {
			injectAnnotations(t, g, "Pet.categories", WithEagerLoad(true))
		})

		// Validate the eager-loaded thing is still there.
		assert.NotNil(t, r.json(`$.components.schemas.PetRead..properties.edges`))
		assert.NotNil(t, r.json(`$.components.schemas.PetEdges.properties.categories.items.$ref`))
		assert.Nil(t, r.json(`$.components.schemas.PetEdges.properties.owner`))

		// The edge endpoint should point to the paged schema, despite us eager-loading
		// it.
		assert.Contains(t, r.json(`$.paths./pets/{petID}/categories..schema.$ref`), "/CategoryList")
		assert.Contains(t, r.json(`$.components.schemas.CategoryList.allOf.*.$ref`), "/PagedResponse")
	})

	t.Run("opt-with-eager-load", func(t *testing.T) {
		t.Parallel()
		r := mustBuildSpec(t, &Config{DisableEagerLoadNonPagedOpt: false}, func(g *gen.Graph) {
			injectAnnotations(t, g, "Pet.categories", WithEagerLoad(true))
		})

		// Validate the eager-loaded thing is still there.
		assert.NotNil(t, r.json(`$.components.schemas.PetRead..properties.edges`))
		assert.NotNil(t, r.json(`$.components.schemas.PetEdges.properties.categories.items.$ref`))
		assert.Nil(t, r.json(`$.components.schemas.PetEdges.properties.owner`))

		// The edge endpoint should also be non-paged, because we optimized away the
		// need for pagination, given the edge is eager-loaded.
		assert.Contains(t, r.json(`$.paths./pets/{petID}/categories..schema.$ref`), "/PetCategoryList")
		assert.Equal(t, "array", r.json(`$.components.schemas.PetCategoryList.type`))
	})

	t.Run("no-opt-no-eager-load", func(t *testing.T) {
		t.Parallel()
		r := mustBuildSpec(t, &Config{DisableEagerLoadNonPagedOpt: true}, nil)

		assert.Contains(t, r.json(`$.paths./pets/{petID}/categories..schema.$ref`), "/CategoryList")
		assert.Contains(t, r.json(`$.components.schemas.CategoryList.allOf.*.$ref`), "/PagedResponse")
	})
}

func TestConfig_DisableEagerLoadedEndpoints(t *testing.T) {
	t.Parallel()

	t.Run("endpoints-with-eager-load", func(t *testing.T) {
		t.Parallel()
		r := mustBuildSpec(t, &Config{DisableEagerLoadedEndpoints: false}, func(g *gen.Graph) {
			injectAnnotations(t, g, "Pet.categories", WithEagerLoad(true))
			injectAnnotations(t, g, "Pet.owner", WithEagerLoad(true))
		})

		assert.NotNil(t, r.json(`$.paths./pets/{petID}/categories`))
		assert.NotNil(t, r.json(`$.paths./pets/{petID}/owner`))
	})

	t.Run("no-endpoints-with-eager-load", func(t *testing.T) {
		t.Parallel()
		r := mustBuildSpec(t, &Config{DisableEagerLoadedEndpoints: true}, func(g *gen.Graph) {
			injectAnnotations(t, g, "Pet.categories", WithEagerLoad(true))
			injectAnnotations(t, g, "Pet.owner", WithEagerLoad(true))
		})

		// The endpoints should be nil, because we disabled them.
		assert.Nil(t, r.json(`$.paths./pets/{petID}/categories`))
		assert.Nil(t, r.json(`$.paths./pets/{petID}/owner`))
	})
}

func TestConfig_AddEdgesToTags(t *testing.T) {
	t.Parallel()

	t.Run("with-tags", func(t *testing.T) {
		t.Parallel()
		r := mustBuildSpec(t, &Config{AddEdgesToTags: true}, func(g *gen.Graph) {
			injectAnnotations(t, g, "Pet.categories", WithEagerLoad(true))
		})

		assert.Contains(t, r.json(`$.paths./pets..tags.*`), "Pets")
		assert.Contains(t, r.json(`$.paths./pets..tags.*`), "Categories")
		assert.NotContains(t, r.json(`$.paths./pets..tags.*`), "Users")
	})

	t.Run("no-tags", func(t *testing.T) {
		t.Parallel()
		r := mustBuildSpec(t, &Config{AddEdgesToTags: false}, func(g *gen.Graph) {
			injectAnnotations(t, g, "Pet.categories", WithEagerLoad(true))
		})

		assert.Contains(t, r.json(`$.paths./pets..tags.*`), "Pets")
		assert.NotContains(t, r.json(`$.paths./pets..tags.*`), "Categories")
		assert.NotContains(t, r.json(`$.paths./pets..tags.*`), "Users")
	})
}

func TestConfig_DefaultOperations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ops     []Operation
		wantErr bool
	}{
		{
			name: "all-operations",
			ops:  AllOperations,
		},
		{
			name:    "no-operations",
			ops:     []Operation{},
			wantErr: true,
		},
		{
			name: "create-1",
			ops:  []Operation{OperationCreate},
		},
		{
			name: "read-1",
			ops:  []Operation{OperationRead},
		},
		{
			name: "update-1",
			ops:  []Operation{OperationUpdate},
		},
		{
			name: "delete-1",
			ops:  []Operation{OperationDelete},
		},
		{
			name: "list-1",
			ops:  []Operation{OperationList},
		},
		{
			name: "create-read-2",
			ops:  []Operation{OperationCreate, OperationRead},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.wantErr {
				_, err := buildSpec(t, &Config{DefaultOperations: tt.ops}, nil)
				assert.Error(t, err)
				return
			}

			r := mustBuildSpec(t, &Config{DefaultOperations: tt.ops}, nil)

			if slices.Contains(tt.ops, OperationCreate) {
				assert.NotNil(t, r.json(`$.paths./pets.post`))
			} else {
				assert.Nil(t, r.json(`$.paths./pets.post`))
			}

			if slices.Contains(tt.ops, OperationRead) {
				assert.NotNil(t, r.json(`$.paths./pets/{petID}.get`))
				assert.NotNil(t, r.json(`$.paths./pets/{petID}/owner.get`))
			} else {
				assert.Nil(t, r.json(`$.paths./pets/{petID}.get`))
				assert.Nil(t, r.json(`$.paths./pets/{petID}/owner.get`))
			}

			if slices.Contains(tt.ops, OperationUpdate) {
				assert.NotNil(t, r.json(`$.paths./pets/{petID}.patch`))
			} else {
				assert.Nil(t, r.json(`$.paths./pets/{petID}.patch`))
			}

			if slices.Contains(tt.ops, OperationDelete) {
				assert.NotNil(t, r.json(`$.paths./pets/{petID}.delete`))
			} else {
				assert.Nil(t, r.json(`$.paths./pets/{petID}.delete`))
			}

			if slices.Contains(tt.ops, OperationList) {
				assert.NotNil(t, r.json(`$.paths./pets.get`))
				assert.NotNil(t, r.json(`$.paths./pets/{petID}/categories.get`))
			} else {
				assert.Nil(t, r.json(`$.paths./pets.get`))
				assert.Nil(t, r.json(`$.paths./pets/{petID}/categories.get`))
			}
		})
	}
}

func TestConfig_GlobalHeaders(t *testing.T) {
	t.Parallel()

	r := mustBuildSpec(t, &Config{
		GlobalRequestHeaders: map[string]*ogen.Parameter{
			"X-Request-ID": {
				Name:        "X-Request-ID",
				In:          "header",
				Description: "The request ID.",
				Schema:      ogen.String(),
				Required:    false,
			},
			"Foo-Bar": {
				Name:        "Foo-Bar",
				In:          "header",
				Description: "Foo bar.",
				Schema:      ogen.String(),
			},
		},
		GlobalResponseHeaders: map[string]*ogen.Header{
			"X-Ratelimit-Limit": {
				Description: "The number of requests allowed per hour.",
				Schema:      ogen.Int(),
			},
			"X-Ratelimit-Remaining": {
				Description: "The number of requests remaining in the current hour.",
				Schema:      ogen.Int(),
			},
		},
	}, nil)

	assert.Contains(t, r.json(`$.components.parameters`), "X-Request-ID")
	assert.Contains(t, r.json(`$.components.parameters`), "Foo-Bar")
	assert.Contains(t, r.json(`$.components.headers`), "X-Ratelimit-Limit")
	assert.Contains(t, r.json(`$.components.headers`), "X-Ratelimit-Remaining")

	assert.Contains(t, r.json(`$.paths./pets.parameters.*.$ref`), "#/components/parameters/X-Request-ID")
	assert.Contains(t, r.json(`$.paths./pets.parameters.*.$ref`), "#/components/parameters/Foo-Bar")
	assert.Contains(t, r.json(`$.paths./pets/{petID}.parameters.*.$ref`), "#/components/parameters/X-Request-ID")
	assert.Contains(t, r.json(`$.paths./pets/{petID}.parameters.*.$ref`), "#/components/parameters/Foo-Bar")
	assert.Contains(t, r.json(`$.paths./pets/{petID}/categories.parameters.*.$ref`), "#/components/parameters/X-Request-ID")
	assert.Contains(t, r.json(`$.paths./pets/{petID}/categories.parameters.*.$ref`), "#/components/parameters/Foo-Bar")

	assert.Contains(t, r.json(`$.paths./pets.get.responses.*.headers`), "X-Ratelimit-Limit")
	assert.Contains(t, r.json(`$.paths./pets.get.responses.*.headers`), "X-Ratelimit-Remaining")
	assert.Contains(t, r.json(`$.paths./pets/{petID}.get.responses.*.headers`), "X-Ratelimit-Limit")
	assert.Contains(t, r.json(`$.paths./pets/{petID}.get.responses.*.headers`), "X-Ratelimit-Remaining")
	assert.Contains(t, r.json(`$.paths./pets/{petID}/categories.get.responses.*.headers`), "X-Ratelimit-Limit")
	assert.Contains(t, r.json(`$.paths./pets/{petID}/categories.get.responses.*.headers`), "X-Ratelimit-Remaining")

	assert.Contains(t, r.json(`$.components.responses.ErrorConflict.headers`), "X-Ratelimit-Limit")
	assert.Contains(t, r.json(`$.components.responses.ErrorConflict.headers`), "X-Ratelimit-Remaining")
}

func TestConfig_GlobalErrorResponses(t *testing.T) {
	t.Parallel()

	newPathItem := func() *ogen.PathItem {
		return &ogen.PathItem{
			Get: &ogen.Operation{
				Summary:     "foo",
				Description: "foo",
				OperationID: "getFoo",
				Responses: ogen.Responses{
					"200": ogen.NewResponse().
						SetDescription("foo").
						SetJSONContent(&ogen.Schema{
							Type: "object",
							Properties: ogen.Properties{
								ogen.Property{
									Name: "foo",
									Schema: &ogen.Schema{
										Type: "string",
									},
								},
							},
							Required: []string{"foo"},
						}),
				},
			},
		}
	}

	t.Run("default", func(t *testing.T) {
		t.Parallel()
		r := mustBuildSpec(t, &Config{
			PostGenerateHook: func(_ *gen.Graph, spec *ogen.Spec) error {
				spec.Paths["/foo"] = newPathItem()
				return nil
			},
		}, nil)

		assert.NotNil(t, r.json(`$.components.responses.ErrorConflict`))
		assert.NotNil(t, r.json(`$.components.schemas.ErrorConflict`))

		// ErrorConflict only exists on POST and PATCH related endpoints.
		assert.Contains(t, r.json(`$.paths./pets.post.responses`), "409")
		assert.Contains(t, r.json(`$.paths./pets/{petID}.patch.responses`), "409")
		assert.NotContains(t, r.json(`$.paths./pets.get.responses`), "409")
		assert.NotContains(t, r.json(`$.paths./pets/{petID}.delete.responses`), "409")
		assert.NotContains(t, r.json(`$.paths./pets/{petID}.get.responses`), "409")
		assert.NotContains(t, r.json(`$.paths./pets/{petID}/owner.get.responses`), "409")

		// ErrorNotFound doesn't exist on list related endpoints.
		assert.Contains(t, r.json(`$.paths./pets/{petID}.get.responses`), "404")
		assert.Contains(t, r.json(`$.paths./pets/{petID}.patch.responses`), "404")
		assert.Contains(t, r.json(`$.paths./pets/{petID}.delete.responses`), "404")
		assert.Contains(t, r.json(`$.paths./pets/{petID}/owner.get.responses`), "404")
		assert.Contains(t, r.json(`$.paths./foo.get.responses`), "404")
		assert.NotContains(t, r.json(`$.paths./pets.get.responses`), "404")
		assert.NotContains(t, r.json(`$.paths./pets/{petID}/categories.get.responses`), "404")
	})

	t.Run("custom", func(t *testing.T) {
		t.Parallel()

		r := mustBuildSpec(t, &Config{
			GlobalErrorResponses: map[int]*ogen.Schema{
				http.StatusInternalServerError: {
					Type: "object",
					Properties: ogen.Properties{
						ogen.Property{
							Name: "foo",
							Schema: &ogen.Schema{
								Type: "string",
							},
						},
					},
					Required: []string{"foo"},
				},
			},
			PostGenerateHook: func(_ *gen.Graph, spec *ogen.Spec) error {
				spec.Paths["/foo"] = newPathItem()
				return nil
			},
		}, nil)

		assert.NotNil(t, r.json(`$.components.responses.ErrorInternalServerError`))
		assert.NotNil(t, r.json(`$.components.schemas.ErrorInternalServerError`))
		assert.Equal(t, "object", r.json(`$.components.schemas.ErrorInternalServerError.type`))
		assert.Equal(t, "string", r.json(`$.components.schemas.ErrorInternalServerError.properties.foo.type`))

		assert.Contains(t, r.json(`$.paths./pets/{petID}.get.responses`), "500")
		assert.Contains(t, r.json(`$.paths./pets/{petID}.patch.responses`), "500")
		assert.Contains(t, r.json(`$.paths./pets/{petID}.delete.responses`), "500")
		assert.Contains(t, r.json(`$.paths./pets/{petID}/owner.get.responses`), "500")
		assert.Contains(t, r.json(`$.paths./foo.get.responses`), "500")
		assert.Contains(t, r.json(`$.paths./pets.get.responses`), "500")
		assert.Contains(t, r.json(`$.paths./pets/{petID}/categories.get.responses`), "500")
	})
}

func TestConfig_AllowClientUUIDs(t *testing.T) {
	t.Parallel()

	t.Run("enabled", func(t *testing.T) {
		t.Parallel()

		r := mustBuildSpec(t, &Config{AllowClientUUIDs: true}, nil)

		assert.Equal(t, "string", r.json(`$.components.schemas.AllType.properties.id.type`))
		assert.Equal(t, "uuid", r.json(`$.components.schemas.AllType.properties.id.format`))
		assert.Equal(t, "string", r.json(`$.components.schemas.AllTypeCreate.properties.id.type`))
		assert.Equal(t, "uuid", r.json(`$.components.schemas.AllTypeCreate.properties.id.format`))
		assert.Equal(t, "string", r.json(`$.components.parameters.AllTypeID.schema.type`))
		assert.Equal(t, "uuid", r.json(`$.components.parameters.AllTypeID.schema.format`))
	})

	t.Run("disabled", func(t *testing.T) {
		t.Parallel()

		r := mustBuildSpec(t, &Config{AllowClientUUIDs: false}, nil)

		assert.Equal(t, "string", r.json(`$.components.schemas.AllType.properties.id.type`))
		assert.Equal(t, "uuid", r.json(`$.components.schemas.AllType.properties.id.format`))
		assert.Nil(t, r.json(`$.components.schemas.AllTypeCreate.properties.id`))
		assert.Equal(t, "string", r.json(`$.components.parameters.AllTypeID.schema.type`))
		assert.Equal(t, "uuid", r.json(`$.components.parameters.AllTypeID.schema.format`))
	})
}

func TestConfig_DisablePatchJSONTag(t *testing.T) {
	t.Parallel()

	t.Run("disabled", func(t *testing.T) {
		t.Parallel()

		r := mustBuildSpec(t, &Config{DisablePatchJSONTag: true}, nil)

		for _, n := range r.graph.Nodes {
			if n.Name == "Pet" {
				for _, f := range n.Fields {
					if f.Name == "name" {
						assert.Equal(t, `json:"name,omitempty"`, f.StructTag)
						return
					}
				}
			}
		}
		t.Errorf("failed to find field with name 'name'")
	})

	t.Run("enabled", func(t *testing.T) {
		t.Parallel()

		r := mustBuildSpec(t, &Config{DisablePatchJSONTag: false}, nil)

		for _, n := range r.graph.Nodes {
			if n.Name == "Pet" {
				for _, f := range n.Fields {
					if f.Name == "name" {
						assert.Equal(t, `json:"name"`, f.StructTag)
						return
					}
				}
			}
		}
		t.Errorf("failed to find field with name 'name'")
	})
}
