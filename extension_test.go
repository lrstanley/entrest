// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package entrest

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"entgo.io/ent/entc/gen"
	"entgo.io/ent/entc/load"
	"github.com/ogen-go/ogen"
	"github.com/pb33f/libopenapi"
	validator "github.com/pb33f/libopenapi-validator"
	"github.com/spyzhov/ajson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeSpec(t *testing.T, spec *ogen.Spec, fn string) { //nolint:unused
	t.Helper()
	err := os.MkdirAll(filepath.Dir(fn), 0o750)
	if err != nil {
		panic(fmt.Sprintf("failed to create directory: %v", err))
	}

	b, err := json.MarshalIndent(spec, "", "    ")
	if err != nil {
		panic(fmt.Sprintf("failed to marshal spec: %v", err))
	}

	err = os.WriteFile(fn, b, 0o640)
	if err != nil {
		panic(fmt.Sprintf("failed to write spec: %v", err))
	}
}

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
func mustBuildSpec(t *testing.T, config *Config) *testSpecResult {
	t.Helper()

	result, err := buildSpec(t, config)
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
func buildSpec(t *testing.T, config *Config) (*testSpecResult, error) {
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

	// This is effectively the same as [entc.LoadGraph], but it caches the schema
	// so we can easily concurrently test multiple spec/graphs.
	schema := integrationSchema()
	gconfig.Schema = schema.PkgPath
	if gconfig.Package == "" {
		gconfig.Package = path.Dir(schema.PkgPath)
	}

	gconfig.Storage, err = gen.NewStorage("sql")
	if err != nil {
		return nil, err
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

	// Ensure that all operation IDs are unique, regardless of the test being ran.
	var ids []string
	for _, path := range result.spec.Paths {
		PatchOperations(path, func(_ string, op *ogen.Operation) *ogen.Operation {
			if op == nil {
				return nil
			}
			ids = append(ids, op.OperationID)
			return op
		})
	}
	slices.Sort(ids)
	assert.Len(t, ids, len(slices.Compact(ids)), "operation IDs are not unique")

	return result, nil
}

// injectAnnotations injects the provided annotations into the provided schema path.
// "Pet" means the Pet schema, "Pet.categories" means the categories edge on the Pet,
// and "Pet.some_field" means the some_field field on the Pet schema.
func injectAnnotations(t *testing.T, g *gen.Graph, schemaPath string, annotations ...Annotation) {
	t.Helper()
	parts := strings.Split(schemaPath, ".")

	for i := range g.Nodes {
		if g.Nodes[i].Name != parts[0] {
			continue
		}

		if len(parts) < 2 {
			g.Nodes[i].Annotations = mergeAnnotations(t, g.Nodes[i].Annotations, annotations...)
			return
		}

		for j := range g.Nodes[i].Fields {
			if g.Nodes[i].Fields[j].Name == parts[1] {
				g.Nodes[i].Fields[j].Annotations = mergeAnnotations(t, g.Nodes[i].Fields[j].Annotations, annotations...)
				return
			}
		}
		for j := range g.Nodes[i].Edges {
			if g.Nodes[i].Edges[j].Name == parts[1] {
				g.Nodes[i].Edges[j].Annotations = mergeAnnotations(t, g.Nodes[i].Edges[j].Annotations, annotations...)
				return
			}
		}
		t.Fatalf("failed to find field or edge with name %q in type %q", parts[0], parts[1])
	}

	t.Fatalf("failed to find type %q", parts[0])
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
	require.Empty(t, errs)

	valid, validErrs := docValidator.ValidateDocument()
	if !valid {
		for _, e := range validErrs {
			t.Errorf("spec validation failed:\n  type: %s\n  failure: %s\n  fix: %s\n", e.ValidationType, e.Message, e.HowToFix)
		}
		t.Logf("spec: %s", string(b))
		t.FailNow()
	}
}

// getPathMethods returns all of the methods that are defined on the provided path,
// within the spec. Note that it's only going to check for methods that are defined
// from testRequiredMethods.
func getPathMethods(t *testing.T, r *testSpecResult, endpoint string) (methods []string) {
	data, ok := r.json(`$.paths.` + endpoint).(map[string]any)
	if !ok {
		t.Fatalf("path %q does not exist or invalid", endpoint)
	}
	for _, method := range testRequiredMethods {
		if _, ok = data[strings.ToLower(method)]; ok {
			methods = append(methods, method)
		}
	}
	return methods
}
