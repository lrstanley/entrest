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

func TestConfig_Spec(t *testing.T) {
	t.Parallel()

	t.Run("default-spec", func(t *testing.T) {
		t.Parallel()
		_, spec := mustBuildSpec(t, &Config{Spec: nil}, nil)
		assert.Equal(t, "EntGo Rest API", spec.Info.Title)
	})

	t.Run("with-spec", func(t *testing.T) {
		t.Parallel()
		_, spec := mustBuildSpec(t, &Config{Spec: &ogen.Spec{Info: ogen.Info{Title: "foo"}}}, nil)
		assert.Equal(t, "foo", spec.Info.Title)
	})
}

func TestConfig_DisablePagination(t *testing.T) {
	t.Parallel()

	t.Run("with-pagination", func(t *testing.T) {
		t.Parallel()
		_, spec := mustBuildSpec(t, &Config{DisablePagination: false}, nil)
		assert.Equal(t, "#/components/schemas/PagedResponse", spec.Components.Schemas["PetList"].AllOf[0].Ref)
	})

	t.Run("without-pagination", func(t *testing.T) {
		t.Parallel()
		_, spec := mustBuildSpec(t, &Config{DisablePagination: true}, nil)
		assert.Equal(t, []*ogen.Schema(nil), spec.Components.Schemas["PetList"].AllOf)
	})

	t.Run("no-global-but-local", func(t *testing.T) {
		t.Parallel()
		_, spec := mustBuildSpec(t, &Config{DisablePagination: true}, func(g *gen.Graph) {
			for _, t := range g.Nodes {
				if t.Name == "Pet" {
					t.Annotations = mergeAnnotations(t.Annotations, WithPagination(true))
				}
			}
		})
		assert.Equal(t, "#/components/schemas/PagedResponse", spec.Components.Schemas["PetList"].AllOf[0].Ref)
	})
}
