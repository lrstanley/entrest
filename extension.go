// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package entrest

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"github.com/ogen-go/ogen"
)

// Ensure that Extension implements the entc.Extension interface.
var _ entc.Extension = (*Extension)(nil)

type Extension struct {
	entc.DefaultExtension

	config *Config
}

func NewExtension(config *Config) (*Extension, error) {
	if config == nil {
		config = &Config{}
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &Extension{config: config}, nil
}

func (e *Extension) Hooks() []gen.Hook {
	return []gen.Hook{
		func(next gen.Generator) gen.Generator {
			return gen.GenerateFunc(func(g *gen.Graph) error {
				if e.config.PatchJSONTag {
					err := e.patchJSONTag(g)
					if err != nil {
						return err
					}
				}
				return next.Generate(g)
			})
		},
		func(next gen.Generator) gen.Generator {
			return gen.GenerateFunc(func(g *gen.Graph) error {
				err := e.Generate(g)
				if err != nil {
					return err
				}
				return next.Generate(g)
			})
		},
	}
}

func (e *Extension) Generate(g *gen.Graph) error {
	// Validate all annotations first.
	err := ValidateAnnotations(g)
	if err != nil {
		return fmt.Errorf("failed to validate annotations: %w", err)
	}

	spec := e.config.Spec
	if spec == nil {
		spec = ogen.NewSpec()
	}

	// If they weren't provided, set some defaults which are required by OpenAPI,
	// as well as most code-generators.
	if spec.OpenAPI == "" {
		spec.OpenAPI = OpenAPIVersion
	}
	if spec.Info.Title == "" {
		spec.Info.Title = "EntGo Generated REST API"
	}
	if spec.Info.Version == "" {
		spec.Info.Version = "1.0.0"
	}

	var specs []*ogen.Spec
	var tspec *ogen.Spec

	for _, t := range g.Nodes {
		ta := GetAnnotation(t)

		if ta.Skip {
			continue
		}

		for _, op := range ta.GetOperations(e.config) {
			tspec, err = GenSpecType(t, op)
			if err != nil {
				panic(err)
			}
			specs = append(specs, tspec)
		}

		for _, edge := range t.Edges {
			ea := GetAnnotation(edge)
			if ea.Skip {
				continue
			}

			ops := ta.GetOperations(e.config)

			if edge.Unique && slices.Contains(ops, OperationRead) {
				tspec, err = GenSpecEdge(t, edge, OperationRead)
			}
			if !edge.Unique && slices.Contains(ops, OperationList) {
				tspec, err = GenSpecEdge(t, edge, OperationList)
			}

			if err != nil {
				panic(err)
			}
			specs = append(specs, tspec)
		}
	}

	err = MergeSpecOverlap(spec, specs...)
	if err != nil {
		panic(err)
	}

	// TODO: ensure errors are added before headers are added.

	addGlobalRequestHeaders(spec, e.config.GlobalRequestHeaders)
	addGlobalResponseHeaders(spec, e.config.GlobalResponseHeaders)

	f, err := os.OpenFile("openapi.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "    ")
	err = enc.Encode(spec)
	if err != nil {
		panic(err)
	}

	return nil
}

func (e *Extension) Annotations() []entc.Annotation {
	return []entc.Annotation{e.config}
}

func (e *Extension) patchJSONTag(g *gen.Graph) error {
	for _, node := range g.Nodes {
		for _, field := range node.Fields {
			if field.StructTag == `json:"-"` {
				continue
			}
			field.StructTag = fmt.Sprintf("json:%q", field.Name)
		}
	}
	return nil
}
