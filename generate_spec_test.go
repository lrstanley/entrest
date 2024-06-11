// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package entrest

import (
	"fmt"
	"path"
	"sync"
	"testing"

	"entgo.io/ent/entc/gen"
	"entgo.io/ent/entc/load"
	"github.com/ogen-go/ogen"
)

var (
	integrationSchema = sync.OnceValue(func() *load.SchemaSpec {
		spec, err := (&load.Config{Path: "./integration/schema", BuildFlags: nil}).Load()
		if err != nil {
			panic(fmt.Sprintf("failed to load schema: %v", err))
		}
		return spec
	})

	specMutex sync.Mutex
)

func mustBuildSpec(t *testing.T, config *Config, hook func(*gen.Graph)) (graph *gen.Graph, spec *ogen.Spec) {
	graph, spec, err := buildSpec(config, hook)
	if err != nil {
		t.Fatal(err)
	}
	return graph, spec
}

func buildSpec(config *Config, hook func(*gen.Graph)) (graph *gen.Graph, spec *ogen.Spec, err error) {
	specMutex.Lock()
	defer specMutex.Unlock()

	ext, err := NewExtension(config)
	if err != nil {
		return nil, nil, err
	}
	ext.disableSpecWrite = true

	gconfig := &gen.Config{Hooks: ext.Hooks(), Annotations: gen.Annotations{}}

	// LoadSchema doesn't configure annotations, so we have to do that manually.
	for _, a := range ext.Annotations() {
		gconfig.Annotations[a.Name()] = a
	}

	if hook != nil {
		gconfig.Hooks = append([]gen.Hook{
			func(next gen.Generator) gen.Generator {
				return gen.GenerateFunc(func(g *gen.Graph) error {
					hook(g)
					return next.Generate(g)
				})
			},
		}, gconfig.Hooks...)
	}

	// This is effectively the same as [entc.LoadGraph], but it caches the schema
	// so we can easily concurrently test multiple spec/graphs.
	schema := integrationSchema()
	gconfig.Schema = schema.PkgPath
	if gconfig.Package == "" {
		gconfig.Package = path.Dir(schema.PkgPath)
	}
	graph, err = gen.NewGraph(gconfig, schema.Schemas...)
	if err != nil {
		return nil, nil, err
	}

	// Same with hooks.
	var g gen.Generator = gen.GenerateFunc(func(_ *gen.Graph) error {
		return nil
	})
	for i := len(graph.Hooks) - 1; i >= 0; i-- {
		g = graph.Hooks[i](g)
	}

	err = g.Generate(graph)
	if err != nil {
		return nil, nil, err
	}
	return graph, ext.generatedSpec, nil
}

func modifyType(t *testing.T, g *gen.Graph, name string, cb func(t *gen.Type)) {
	t.Helper()
	for i := range g.Nodes {
		if g.Nodes[i].Name == name {
			cb(g.Nodes[i])
			return
		}
	}
	t.Fatalf("failed to find type %q", name)
}

func modifyTypeField(t *testing.T, g *gen.Graph, tname, fname string, cb func(t *gen.Field)) {
	t.Helper()
	for i := range g.Nodes {
		if g.Nodes[i].Name == tname {
			for j := range g.Nodes[i].Fields {
				if g.Nodes[i].Fields[j].Name == fname {
					cb(g.Nodes[i].Fields[j])
					return
				}
			}
			t.Fatalf("failed to find field %q in type %q", fname, tname)
		}
	}
	t.Fatalf("failed to find type %q", tname)
}

func modifyTypeEdge(t *testing.T, g *gen.Graph, tname, ename string, cb func(t *gen.Edge)) {
	t.Helper()
	for i := range g.Nodes {
		if g.Nodes[i].Name == tname {
			for j := range g.Nodes[i].Edges {
				if g.Nodes[i].Edges[j].Name == ename {
					cb(g.Nodes[i].Edges[j])
					return
				}
			}
			t.Fatalf("failed to find edge %q in type %q", ename, tname)
		}
	}
	t.Fatalf("failed to find type %q", tname)
}
