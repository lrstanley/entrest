// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package entrest

import (
	"encoding/json"
	"fmt"
	"io"
	"path"
	"sync"
	"testing"

	"entgo.io/ent/entc/gen"
	"entgo.io/ent/entc/load"
	"github.com/ogen-go/ogen"
	"github.com/pb33f/libopenapi"
	validator "github.com/pb33f/libopenapi-validator"
	"github.com/stretchr/testify/require"
)

// Setup helpers/initialization for all integration related tests.
var (
	integrationSchema = sync.OnceValue(func() *load.SchemaSpec {
		spec, err := (&load.Config{Path: "./testdata/schema", BuildFlags: nil}).Load()
		if err != nil {
			panic(fmt.Sprintf("failed to load schema: %v", err))
		}
		return spec
	})

	specMutex sync.Mutex
)

// mustBuildSpec is like buildSpec, but it fails if the extension execution fails.
// Also runs spec validation, which will fail is the spec has any errors.
func mustBuildSpec(t *testing.T, config *Config, hook func(*gen.Graph)) (graph *gen.Graph, spec *ogen.Spec) {
	t.Helper()
	graph, spec, err := buildSpec(config, hook)
	if err != nil {
		t.Fatal(err)
	}
	validateSpec(t, spec)
	return graph, spec
}

// buildSpec uses the shared schema cache, and invokes the extension to build the
// spec. It also invokes the provided hook on the graph before executing the extension,
// if provided. DOES NOT RUN SPEC VALIDATION.
func buildSpec(config *Config, hook func(*gen.Graph)) (graph *gen.Graph, spec *ogen.Spec, err error) {
	if config == nil {
		config = &Config{}
	}

	if config.Writer == nil {
		config.Writer = io.Discard
	}

	config.PreWriteHook = func(s *ogen.Spec) error {
		spec = s
		return nil
	}

	specMutex.Lock()
	defer specMutex.Unlock()

	ext, err := NewExtension(config)
	if err != nil {
		return nil, nil, err
	}

	gconfig := &gen.Config{Hooks: ext.Hooks(), Annotations: gen.Annotations{}}

	// LoadGraph doesn't configure annotations, so we have to do that manually.
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
	return graph, spec, nil
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

func mergeAnnotations(t *testing.T, in gen.Annotations, annotations ...Annotation) gen.Annotations {
	t.Helper()
	ant := *decodeAnnotation(in)
	for _, a := range annotations { // nolint:gocritic
		ant, _ = ant.Merge(a).(Annotation)
	}
	in.Set(ant.Name(), ant)
	return in
}

func validateSpec(t *testing.T, spec *ogen.Spec) {
	t.Helper()

	b, err := json.Marshal(spec)
	require.NoError(t, err)

	doc, err := libopenapi.NewDocument(b)
	require.NoError(t, err)

	docValidator, errs := validator.NewValidator(doc)
	if len(errs) > 0 {
		t.Logf("spec: %s", string(b))
	}
	require.Len(t, errs, 0)

	valid, validErrs := docValidator.ValidateDocument()
	if !valid {
		for _, e := range validErrs {
			t.Errorf("spec validation failed:\n  type: %s\n  failure: %s\n  fix: %s\n", e.ValidationType, e.Message, e.HowToFix)
		}
		t.Logf("spec: %s", string(b))
		t.FailNow()
	}
}
