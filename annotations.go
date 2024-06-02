// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package entrest

import (
	"encoding/json"
	"fmt"
	"slices"

	"entgo.io/ent/entc/gen"
	"entgo.io/ent/schema"
	"github.com/ogen-go/ogen"
)

// GetAnnotation returns the annotation on the given graph item.
func GetAnnotation[T gen.Type | gen.Field | gen.Edge](v *T) *Annotation {
	switch v := any(*v).(type) {
	case gen.Type:
		return decodeAnnotation(v.Annotations)
	case gen.Field:
		return decodeAnnotation(v.Annotations)
	case gen.Edge:
		return decodeAnnotation(v.Annotations)
	}
	panic("unreachable")
}

// decodeAnnotation decodes the decodeAnnotation from the given gen.Annotations.
func decodeAnnotation(as gen.Annotations) *Annotation {
	ant := &Annotation{}
	if as != nil && as[ant.Name()] != nil {
		if err := ant.Decode(as[ant.Name()]); err != nil {
			panic(fmt.Sprintf("failed to decode annotation: %v", err))
		}
	}
	return ant
}

var ( // Ensure that Annotation implements necessary interfaces.
	_ schema.Annotation = (*Annotation)(nil)
	_ schema.Merger     = (*Annotation)(nil)
)

type Annotation struct {
	// WARNING: if you add a new field, ensure you update the Merge method.

	// Fields that map directly to the OpenAPI schema.

	AdditionalTags       []string             `json:",omitempty"` // schema/edge.
	Tags                 []string             `json:",omitempty"` // schema/edge.
	OperationSummary     map[Operation]string `json:",omitempty"` // schema/edge.
	OperationDescription map[Operation]string `json:",omitempty"` // schema/edge.
	OperationID          map[Operation]string `json:",omitempty"` // schema/edge.
	Description          string               `json:",omitempty"` // schema/edge/field.
	Example              any                  `json:",omitempty"` // field.
	Deprecated           bool                 `json:",omitempty"` // schema/edge/field.
	Schema               *ogen.Schema         `json:",omitempty"` // field.

	// All others.

	Pagination      *bool       `json:",omitempty"` // schema/edge.
	MinItemsPerPage int         `json:",omitempty"` // schema/edge.
	MaxItemsPerPage int         `json:",omitempty"` // schema/edge.
	ItemsPerPage    int         `json:",omitempty"` // schema/edge.
	EagerLoad       *bool       `json:",omitempty"` // edge.
	Filter          Predicate   `json:",omitempty"` // schema/edge/field.
	Handler         *bool       `json:",omitempty"` // schema/edge.
	Sortable        bool        `json:",omitempty"` // field.
	Skip            bool        `json:",omitempty"` // schema/edge/field.
	ReadOnly        bool        `json:",omitempty"` // field.
	Operations      []Operation `json:",omitempty"` // schema/edge.
}

func (Annotation) Name() string {
	return "EntRest"
}

// Merge merges the given annotation into the current annotation.
func (a Annotation) Merge(o schema.Annotation) schema.Annotation { // nolint:gocyclo,cyclop
	var am Annotation

	switch o := o.(type) {
	case Annotation:
		am = o
	case *Annotation:
		am = *o
	default:
		return a
	}

	if len(am.AdditionalTags) > 0 {
		for _, t := range am.AdditionalTags {
			if !slices.Contains(a.Tags, t) {
				a.Tags = append(a.Tags, t)
			}
		}
	}
	if len(am.Tags) > 0 {
		for _, t := range am.Tags {
			if !slices.Contains(a.Tags, t) {
				a.Tags = append(a.Tags, t)
			}
		}
	}
	if len(am.OperationSummary) > 0 {
		if a.OperationSummary == nil {
			a.OperationSummary = make(map[Operation]string)
		}
		for k, v := range am.OperationSummary {
			a.OperationSummary[k] = v
		}
	}
	if len(am.OperationDescription) > 0 {
		if a.OperationDescription == nil {
			a.OperationDescription = make(map[Operation]string)
		}
		for k, v := range am.OperationDescription {
			a.OperationDescription[k] = v
		}
	}
	if len(am.OperationID) > 0 {
		if a.OperationID == nil {
			a.OperationID = make(map[Operation]string)
		}
		for k, v := range am.OperationID {
			a.OperationID[k] = v
		}
	}
	if am.Description != "" {
		a.Description = am.Description
	}
	if am.Example != nil {
		a.Example = am.Example
	}
	a.Deprecated = a.Deprecated || am.Deprecated
	if am.Schema != nil {
		a.Schema = am.Schema
	}

	if am.Pagination != nil {
		a.Pagination = am.Pagination
	}
	if am.MinItemsPerPage != 0 {
		a.MinItemsPerPage = am.MinItemsPerPage
	}
	if am.MaxItemsPerPage != 0 {
		a.MaxItemsPerPage = am.MaxItemsPerPage
	}
	if am.ItemsPerPage != 0 {
		a.ItemsPerPage = am.ItemsPerPage
	}
	if am.EagerLoad != nil {
		a.EagerLoad = am.EagerLoad
	}
	if am.Filter != 0 {
		a.Filter = am.Filter.Add(a.Filter)
	}
	if am.Handler != nil {
		a.Handler = am.Handler
	}
	a.Sortable = a.Sortable || am.Sortable
	a.Skip = a.Skip || am.Skip
	a.ReadOnly = a.ReadOnly || am.ReadOnly
	if len(am.Operations) > 0 {
		for _, op := range am.Operations {
			if !slices.Contains(a.Operations, op) {
				a.Operations = append(a.Operations, op)
			}
		}
	}

	return a
}

func (a *Annotation) Decode(o any) error {
	buf, err := json.Marshal(o)
	if err != nil {
		return err
	}
	return json.Unmarshal(buf, a)
}

// GetPagination returns the pagination annotation (or defaults).
func (a *Annotation) GetPagination(config *Config) bool {
	if a.Pagination == nil {
		return !config.DisablePagination
	}
	return *a.Pagination
}

func (a *Annotation) GetMinItemsPerPage(config *Config) int {
	if a.MinItemsPerPage == 0 {
		return config.MinItemsPerPage
	}
	return a.MinItemsPerPage
}

func (a *Annotation) GetMaxItemsPerPage(config *Config) int {
	if a.MaxItemsPerPage == 0 {
		return config.MaxItemsPerPage
	}
	return a.MaxItemsPerPage
}

func (a *Annotation) GetItemsPerPage(config *Config) int {
	if a.ItemsPerPage == 0 {
		return config.ItemsPerPage
	}
	return a.ItemsPerPage
}

// GetEagerLoad returns if the edge should be eager-loaded (or defaults).
func (a *Annotation) GetEagerLoad(config *Config) bool {
	if a.EagerLoad == nil {
		return config.DefaultEagerLoad
	}
	return *a.EagerLoad
}

// GetOperations returns the operations annotation (or defaults).
func (a *Annotation) GetOperations(config *Config) []Operation {
	if a.Operations == nil {
		return config.DefaultOperations
	}
	return a.Operations
}

func (a *Annotation) GetOperationSummary(op Operation) string {
	if a.OperationSummary == nil {
		return ""
	}
	return a.OperationSummary[op]
}

func (a *Annotation) GetOperationDescription(op Operation) string {
	if a.OperationDescription == nil {
		return ""
	}
	return a.OperationDescription[op]
}

func (a *Annotation) GetOperationID(op Operation) string {
	if a.OperationID == nil {
		return ""
	}
	return a.OperationID[op]
}

// OperationSummary provides a summary for the specified operation.
func OperationSummary(op Operation, v string) Annotation {
	return Annotation{OperationSummary: map[Operation]string{op: v}}
}

// OperationDescription provides a description for the specified operation.
func OperationDescription(op Operation, v string) Annotation {
	return Annotation{OperationDescription: map[Operation]string{op: v}}
}

// AdditionalTags adds additional tags to all operations for this schema/edge.
func AdditionalTags(v ...string) Annotation {
	return Annotation{AdditionalTags: v}
}

// Tags sets the tags for the schema/edge in the REST API. This will otherwise default
// to the schema/edge's name(s).
func Tags(v ...string) Annotation {
	return Annotation{Tags: v}
}

func OperationID(op Operation, v string) Annotation {
	return Annotation{OperationID: map[Operation]string{op: v}}
}

// Description sets the description for the schema/edge/field in the REST API. This will
// otherwise default to the schema/edge/field's description according to Ent (e.g. the
// comment).
func Description(v string) Annotation {
	return Annotation{Description: v}
}

// Pagination sets the schema to be paginated in the REST API. This is not required to be
// provided unless pagination was disabled globally.
func Pagination(v bool) Annotation {
	return Annotation{Pagination: ptr(v)}
}

// MinItemsPerPage sets an explicit minimum number of items per page for paginated calls.
func MinItemsPerPage(v int) Annotation {
	return Annotation{MinItemsPerPage: v}
}

// MaxItemsPerPage sets an explicit maximum number of items per page for paginated calls.
func MaxItemsPerPage(v int) Annotation {
	return Annotation{MaxItemsPerPage: v}
}

// ItemsPerPage sets an explicit default number of items per page for paginated calls.
func ItemsPerPage(v int) Annotation {
	return Annotation{ItemsPerPage: v}
}

// EagerLoad sets the edge to be eager-loaded in the REST API for each associated
// entity. Note that edges are not eager-loaded by default.
func EagerLoad(v bool) Annotation {
	return Annotation{EagerLoad: ptr(v)}
}

// Filter sets the field to be filterable with the provided predicate(s). When applied
// on an edge with [FilterEdge], it will include the fields associated with the edge
// that are also filterable.
//
// Example:
//
//	entrest.Filter(entrest.FilterGroupArray | entrest.FilterGroupLength) // Bundle using groups.
//	entrest.Filter(entrest.FilterEQ | entrest.FilterNEQ) // Or use individual predicates.
func Filter(v Predicate) Annotation {
	return Annotation{Filter: v}
}

// Handler sets the schema/edge to have an HTTP handler generated for it.
func Handler(v bool) Annotation {
	return Annotation{Handler: ptr(v)}
}

// Sortable sets the field to be sortable in the REST API.
func Sortable(v bool) Annotation {
	return Annotation{Sortable: v}
}

// Skip sets the schema, edge, or field to be skipped in the REST API.
func Skip(v bool) Annotation {
	return Annotation{Skip: v}
}

// ReadOnly sets the field to be read-only in the REST API. If you want to make
// a schema or edge read-only, use the Operations annotation instead.
func ReadOnly(v bool) Annotation {
	return Annotation{ReadOnly: v}
}

// Example sets the OpenAPI Specification example value for a field.
func Example(v any) Annotation {
	return Annotation{Example: v}
}

// Schema sets the OpenAPI schema for a field.
func Schema(v *ogen.Schema) Annotation {
	return Annotation{Schema: v}
}

// IncludeOperations includes the specified operations in the REST API for the
// schema. If empty, all operations are generated (unless globally disabled).
func IncludeOperations(v ...Operation) Annotation {
	return Annotation{Operations: v}
}

// ExcludeOperations excludes the specified operations in the REST API for the
// schema. If empty, all operations are generated (unless globally disabled).
func ExcludeOperations(v ...Operation) Annotation {
	var ops []Operation
	for _, o := range AllOperations {
		if !slices.Contains(v, o) {
			ops = append(ops, o)
		}
	}
	return Annotation{Operations: ops}
}
