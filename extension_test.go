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
	"github.com/spyzhov/ajson"
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
func mustBuildSpec(t *testing.T, config *Config, hook func(*gen.Graph)) *testSpecResult {
	t.Helper()

	result, err := buildSpec(config, hook)
	if err != nil {
		t.Fatal(err)
	}

	validateSpec(t, result.spec)
	return result
}

type testSpecResult struct {
	graph  *gen.Graph
	spec   *ogen.Spec
	config *Config

	ensureObj func()
	_obj      *ajson.Node
}

// json queries the resulting JSON version of the OpenAPI spec using JSONPath
// syntax. If length of results is 0, it returns nil, if length is 1, it returns
// the first result, if length is greater than 1, it returns all values as a
// []<underlying type>. Also note that integers will be returned as float64s.
//
// Refs:
//   - https://jsonpath.com/ (super helpful for debugging/testing validations,
//     but doesn't support everything ajson does).
func (s *testSpecResult) json(jsonPath string) any {
	s.ensureObj()

	results, err := s._obj.JSONPath(jsonPath)
	if err != nil {
		panic(fmt.Sprintf("failed to parse json path: %v", err))
	}

	if len(results) == 0 {
		return nil
	}

	var out []any
	var v any

	for _, n := range results {
		v, err = n.Unpack()
		if err != nil {
			panic(fmt.Sprintf("failed to unpack node: %v", err))
		}
		out = append(out, v)
	}

	if len(out) == 1 {
		return out[0]
	}
	return out
}

// buildSpec uses the shared schema cache, and invokes the extension to build the
// spec. It also invokes the provided hook on the graph before executing the extension,
// if provided. DOES NOT RUN SPEC VALIDATION.
func buildSpec(config *Config, hook func(*gen.Graph)) (*testSpecResult, error) {
	if config == nil {
		config = &Config{}
	}

	if config.Writer == nil {
		config.Writer = io.Discard
	}

	result := &testSpecResult{config: config}

	config.PreWriteHook = func(s *ogen.Spec) error {
		result.spec = s
		return nil
	}

	specMutex.Lock()
	defer specMutex.Unlock()

	ext, err := NewExtension(config)
	if err != nil {
		return nil, err
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
	result.graph, err = gen.NewGraph(gconfig, schema.Schemas...)
	if err != nil {
		return nil, err
	}

	// Same with hooks.
	var g gen.Generator = gen.GenerateFunc(func(_ *gen.Graph) error {
		return nil
	})
	for i := len(result.graph.Hooks) - 1; i >= 0; i-- {
		g = result.graph.Hooks[i](g)
	}

	err = g.Generate(result.graph)
	if err != nil {
		return nil, err
	}

	result.ensureObj = sync.OnceFunc(func() {
		var b []byte
		b, err = json.Marshal(result.spec)
		if err != nil {
			panic(fmt.Sprintf("failed to marshal spec: %v", err))
		}

		result._obj, err = ajson.Unmarshal(b)
		if err != nil {
			panic(fmt.Sprintf("failed to unmarshal spec: %v", err))
		}
	})

	return result, nil
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
