// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package entrest

import (
	"encoding/json"
	"fmt"
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

		assert.Contains(t, r.json(`$.paths./pets/{id}/categories..responses..schema.$ref`), "/CategoryList")
		assert.Contains(t, r.json(`$.components.schemas.CategoryList.allOf.*.$ref`), "/PagedResponse")
	})

	t.Run("without-pagination", func(t *testing.T) {
		t.Parallel()
		r := mustBuildSpec(t, &Config{DisablePagination: true}, nil)
		assert.Equal(t, nil, r.json(`$.components.schemas.PetList.allOf`))
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

		assert.Contains(t, r.json(`$.paths./pets/{id}/categories..responses..schema.$ref`), "/PetCategoryList")
		assert.Contains(t, r.json(`$.components.schemas.PetCategoryList.allOf.*.$ref`), "/PagedResponse")
	})

	// Same as no-global-but-local-edge but pagination is enabled on the edges
	// underlying type.
	t.Run("no-global-but-local-edge-ref", func(t *testing.T) {
		t.Parallel()
		r := mustBuildSpec(t, &Config{DisablePagination: true}, func(g *gen.Graph) {
			injectAnnotations(t, g, "Category", WithPagination(true))
		})
		assert.Contains(t, r.json(`$.paths./pets/{id}/categories..responses..schema.$ref`), "/CategoryList")
		assert.Contains(t, r.json(`$.components.schemas.CategoryList.allOf.*.$ref`), "/PagedResponse")
	})
}

func TestConfig_ItemsPerPage(t *testing.T) {
	t.Parallel()

	base := `$.paths./pets..parameters[?(@.name == "itemsPerPage")].schema`

	t.Run("defaults", func(t *testing.T) {
		t.Parallel()
		c := &Config{}
		r := mustBuildSpec(t, c, nil)

		assert.Equal(t, float64(c.ItemsPerPage), r.json(base+`.default`))
		assert.Equal(t, float64(c.MinItemsPerPage), r.json(base+`.minimum`))
		assert.Equal(t, float64(c.MaxItemsPerPage), r.json(base+`.maximum`))
	})

	t.Run("with-specified", func(t *testing.T) {
		t.Parallel()
		c := &Config{
			MinItemsPerPage: 2,
			ItemsPerPage:    5,
			MaxItemsPerPage: 999,
		}
		r := mustBuildSpec(t, c, nil)

		assert.Equal(t, float64(c.ItemsPerPage), r.json(base+`.default`))
		assert.Equal(t, float64(c.MinItemsPerPage), r.json(base+`.minimum`))
		assert.Equal(t, float64(c.MaxItemsPerPage), r.json(base+`.maximum`))
	})
}

func TestConfig_DisableTotalCount(t *testing.T) {
	t.Parallel()

	t.Run("with-total-count", func(t *testing.T) {
		t.Parallel()
		r := mustBuildSpec(t, &Config{DisableTotalCount: false}, nil)
		assert.NotNil(t, r.json(`$.components.schemas.PagedResponse.properties.total_count`))
	})

	t.Run("without-total-count", func(t *testing.T) {
		t.Parallel()
		r := mustBuildSpec(t, &Config{DisableTotalCount: true}, nil)
		assert.Nil(t, r.json(`$.components.schemas.PagedResponse.properties.total_count`))
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
