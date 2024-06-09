// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package entrest

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"

	"github.com/ogen-go/ogen"
)

func mergeOperation(overlap bool, orig, op *ogen.Operation) (*ogen.Operation, error) {
	if orig == nil {
		return op, nil
	}
	if op == nil {
		return orig, nil
	}

	orig.Tags = appendIfNotContains(orig.Tags, op.Tags)
	orig.Summary = withDefault(op.Summary, orig.Summary)
	orig.Description = withDefault(op.Description, orig.Description)
	orig.OperationID = withDefault(op.OperationID, orig.OperationID)
	orig.Deprecated = orig.Deprecated || op.Deprecated

	// Merge parameters.
	orig.Parameters = appendIfNotContainsFunc(orig.Parameters, op.Parameters, func(oldParam, newParam *ogen.Parameter) bool {
		return oldParam.Name == newParam.Name
	})

	if orig.Responses == nil {
		orig.Responses = map[string]*ogen.Response{}
	}
	err := mergeMap(overlap, orig.Responses, op.Responses)
	if err != nil {
		return nil, err
	}

	return orig, nil
}

// mergeSpec merges multiple [ogen.Spec] into a single [ogen.Spec]. See [MergeSpec] and
// [MergeSpecOverlap] for more information.
func mergeSpec(overlap bool, orig *ogen.Spec, toMerge ...*ogen.Spec) error { //nolint:gocyclo,cyclop,funlen
	var err error

	for _, spec := range toMerge {
		if spec == nil {
			continue
		}

		for _, newServer := range spec.Servers {
			if !slices.ContainsFunc(orig.Servers, func(oldServer ogen.Server) bool {
				return newServer.URL == oldServer.URL
			}) {
				orig.Servers = append(orig.Servers, newServer)
			}
		}

		orig.Servers = appendIfNotContainsFunc(orig.Servers, spec.Servers, func(oldServer, newServer ogen.Server) bool {
			return oldServer.URL == newServer.URL
		})

		if orig.Paths == nil {
			orig.Paths = ogen.Paths{}
		}

		for k, v := range spec.Paths {
			_, ok := orig.Paths[k]
			if !overlap && ok {
				return fmt.Errorf("path %q already exists in the spec", k)
			}

			if !ok {
				orig.Paths[k] = v
				continue
			}

			// Basic description stuff.
			if v.Ref != "" {
				orig.Paths[k].Ref = v.Ref
			}
			if v.Summary != "" {
				orig.Paths[k].Summary = v.Summary
			}
			if v.Description != "" {
				orig.Paths[k].Description = v.Description
			}

			orig.Paths[k].Get, err = mergeOperation(overlap, orig.Paths[k].Get, v.Get)
			if err != nil {
				return err
			}
			orig.Paths[k].Put, err = mergeOperation(overlap, orig.Paths[k].Put, v.Put)
			if err != nil {
				return err
			}
			orig.Paths[k].Post, err = mergeOperation(overlap, orig.Paths[k].Post, v.Post)
			if err != nil {
				return err
			}
			orig.Paths[k].Delete, err = mergeOperation(overlap, orig.Paths[k].Delete, v.Delete)
			if err != nil {
				return err
			}
			orig.Paths[k].Options, err = mergeOperation(overlap, orig.Paths[k].Options, v.Options)
			if err != nil {
				return err
			}
			orig.Paths[k].Head, err = mergeOperation(overlap, orig.Paths[k].Head, v.Head)
			if err != nil {
				return err
			}
			orig.Paths[k].Patch, err = mergeOperation(overlap, orig.Paths[k].Patch, v.Patch)
			if err != nil {
				return err
			}
			orig.Paths[k].Trace, err = mergeOperation(overlap, orig.Paths[k].Trace, v.Trace)
			if err != nil {
				return err
			}

			orig.Paths[k].Servers = appendIfNotContainsFunc(orig.Paths[k].Servers, v.Servers, func(oldServer, newServer ogen.Server) bool {
				return oldServer.URL == newServer.URL
			})

			orig.Paths[k].Parameters = appendIfNotContainsFunc(orig.Paths[k].Parameters, v.Parameters, func(oldParam, newParam *ogen.Parameter) bool {
				return oldParam.Name == newParam.Name
			})
		}

		if orig.Components == nil {
			orig.Components = &ogen.Components{
				Schemas:    map[string]*ogen.Schema{},
				Responses:  map[string]*ogen.Response{},
				Parameters: map[string]*ogen.Parameter{},
				Headers:    map[string]*ogen.Header{},
			}
		}

		if spec.Components != nil {
			if orig.Components.Schemas == nil {
				orig.Components.Schemas = map[string]*ogen.Schema{}
			}
			err = mergeMap(overlap, orig.Components.Schemas, spec.Components.Schemas)
			if err != nil {
				return err
			}
			if orig.Components.Responses == nil {
				orig.Components.Responses = map[string]*ogen.Response{}
			}
			err = mergeMap(overlap, orig.Components.Responses, spec.Components.Responses)
			if err != nil {
				return err
			}
			if orig.Components.Parameters == nil {
				orig.Components.Parameters = map[string]*ogen.Parameter{}
			}
			err = mergeMap(overlap, orig.Components.Parameters, spec.Components.Parameters)
			if err != nil {
				return err
			}
			if orig.Components.RequestBodies == nil {
				orig.Components.RequestBodies = map[string]*ogen.RequestBody{}
			}
			err = mergeMap(overlap, orig.Components.RequestBodies, spec.Components.RequestBodies)
			if err != nil {
				return err
			}
			if orig.Components.Headers == nil {
				orig.Components.Headers = map[string]*ogen.Header{}
			}
			err = mergeMap(overlap, orig.Components.Headers, spec.Components.Headers)
			if err != nil {
				return err
			}
			if orig.Components.SecuritySchemes == nil {
				orig.Components.SecuritySchemes = map[string]*ogen.SecurityScheme{}
			}
			err = mergeMap(overlap, orig.Components.SecuritySchemes, spec.Components.SecuritySchemes)
			if err != nil {
				return err
			}
			if orig.Components.PathItems == nil {
				orig.Components.PathItems = map[string]*ogen.PathItem{}
			}
			err = mergeMap(overlap, orig.Components.PathItems, spec.Components.PathItems)
			if err != nil {
				return err
			}
		}

		orig.Tags = appendIfNotContainsFunc(orig.Tags, spec.Tags, func(oldTag, newTag ogen.Tag) bool {
			return oldTag.Name == newTag.Name
		})
	}

	return nil
}

// MergeSpec merges multiple [ogen.Spec] into a single [ogen.Spec]. This does not cover
// all possible fields, and is not a full merge. It's a simple merge at the core-component
// level, for things like servers, paths, components, tags, etc.
func MergeSpec(orig *ogen.Spec, toMerge ...*ogen.Spec) error {
	return mergeSpec(false, orig, toMerge...)
}

// MergeSpecOverlap merges multiple [ogen.Spec] into a single [ogen.Spec], allowing for
// overlapping fields. This does not cover all possible fields, and is not a full merge.
// It's a simple merge at the core-component level, for things like servers, paths, components,
// tags, etc.
func MergeSpecOverlap(orig *ogen.Spec, toMerge ...*ogen.Spec) error {
	return mergeSpec(true, orig, toMerge...)
}

// AddOpenAPIEndpoint adds an endpoint to the OpenAPI spec that returns the OpenAPI
// spec itself, as JSON.
func AddOpenAPIEndpoint(path string) *ogen.Spec {
	return ogen.NewSpec().AddPathItem(path, ogen.NewPathItem().
		SetGet(
			ogen.NewOperation().
				SetSummary("Get OpenAPI spec").
				SetDescription("Get the OpenAPI specification for this service.").
				SetOperationID("getOpenAPI").
				SetTags([]string{"Meta"}).
				SetResponses(map[string]*ogen.Response{
					strconv.Itoa(http.StatusOK): ogen.NewResponse().
						SetDescription("OpenAPI specification was found").
						SetJSONContent(&ogen.Schema{
							Type: "object",
							AdditionalProperties: &ogen.AdditionalProperties{
								Bool: ptr(true), // https://github.com/ogen-go/ogen/issues/1221
							},
						}),
				}),
		),
	)
}
