// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package entrest

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"entgo.io/ent/entc/gen"
	"entgo.io/ent/schema"
	"github.com/ogen-go/ogen"
)

// GetAnnotation returns the annotation on the given graph item.
func GetAnnotation(v any) *Annotation {
	switch v := v.(type) {
	case *gen.Type:
		return decodeAnnotation(v.Annotations)
	case *gen.Field:
		return decodeAnnotation(v.Annotations)
	case *gen.Edge:
		return decodeAnnotation(v.Annotations)
	default:
		panic(fmt.Sprintf("unsupported type %T", v))
	}
}

// decodeAnnotation decodes the entrest [Annotation] from the given gen.Annotations.
// Panics if it's unable to decode the annotation.
func decodeAnnotation(as gen.Annotations) *Annotation {
	ant := &Annotation{}
	if as != nil && as[ant.Name()] != nil {
		if err := ant.Decode(as[ant.Name()]); err != nil {
			panic(fmt.Sprintf("failed to decode annotation: %v", err))
		}
	}
	return ant
}

// ValidateAnnotations ensures that all annotations on the given graph are correctly
// attached to the right types (e.g. a field-only annotation on a schema or edge type).
func ValidateAnnotations(nodes ...*gen.Type) error {
	for _, t := range nodes {
		if err := GetAnnotation(t).getSupportedType(t.Name, "schema"); err != nil {
			return err
		}
		for _, f := range t.Fields {
			if err := GetAnnotation(f).getSupportedType(f.Name, "field"); err != nil {
				return err
			}
		}
		for _, e := range t.Edges {
			if err := GetAnnotation(e).getSupportedType(e.Name, "edge"); err != nil {
				return err
			}
		}
	}
	return nil
}

var ( // Ensure that Annotation implements necessary interfaces.
	_ schema.Annotation = (*Annotation)(nil)
	_ schema.Merger     = (*Annotation)(nil)
)

// Annotation adds configuration options for specific layers of the Ent graph.
type Annotation struct {
	// WARNING: if you add a new field, ensure you update the Merge method, the Get*
	// methods (primarily only if there are three states for an annotation field, e.g.
	// when the the field default can be changed via the global config), and if
	// necessary, the With* methods.

	// Fields that map directly to the OpenAPI schema.

	AdditionalTags       []string             `json:",omitempty" ent:"schema,edge"`
	Tags                 []string             `json:",omitempty" ent:"schema,edge"`
	OperationSummary     map[Operation]string `json:",omitempty" ent:"schema,edge"`
	OperationDescription map[Operation]string `json:",omitempty" ent:"schema,edge"`
	OperationID          map[Operation]string `json:",omitempty" ent:"schema,edge"`
	Description          string               `json:",omitempty" ent:"schema,edge,field"`
	Example              any                  `json:",omitempty" ent:"field"`
	Deprecated           bool                 `json:",omitempty" ent:"schema,edge,field"`
	Schema               *ogen.Schema         `json:",omitempty" ent:"field"`

	// All others.

	Pagination      *bool       `json:",omitempty" ent:"schema,edge"`
	MinItemsPerPage int         `json:",omitempty" ent:"schema,edge"`
	MaxItemsPerPage int         `json:",omitempty" ent:"schema,edge"`
	ItemsPerPage    int         `json:",omitempty" ent:"schema,edge"`
	EagerLoad       *bool       `json:",omitempty" ent:"edge"`
	EdgeEndpoint    *bool       `json:",omitempty" ent:"edge"`
	EdgeUpdateBulk  bool        `json:",omitempty" ent:"edge"`
	Filter          Predicate   `json:",omitempty" ent:"schema,edge,field"`
	Handler         *bool       `json:",omitempty" ent:"schema,edge"`
	Sortable        bool        `json:",omitempty" ent:"field"`
	Skip            bool        `json:",omitempty" ent:"schema,edge,field"`
	ReadOnly        bool        `json:",omitempty" ent:"field"`
	Operations      []Operation `json:",omitempty" ent:"schema,edge"`
}

// getSupportedType uses reflection to check if the annotation is supported on the
// given type, returning an error if it is not with information about the annotation
// and what type it is on.
func (a Annotation) getSupportedType(name, typ string) error {
	ant := reflect.ValueOf(a)

	for i := range ant.NumField() {
		if ant.Field(i).IsZero() || (reflect.ValueOf(ant.Field(i)).Kind() == reflect.Ptr && ant.Field(i).IsNil()) {
			continue
		}

		supported := strings.Split(ant.Type().Field(i).Tag.Get("ent"), ",")

		if !slices.Contains(supported, typ) {
			return fmt.Errorf(
				"annotation field %q is set on %q %s type, but only one of the following types is supported: %s",
				ant.Type().Field(i).Name,
				name,
				typ,
				strings.Join(supported, ", "),
			)
		}
	}
	return nil
}

func (Annotation) Name() string {
	return "Rest"
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
			if !slices.Contains(a.AdditionalTags, t) {
				a.AdditionalTags = append(a.AdditionalTags, t)
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
	if am.EdgeEndpoint != nil {
		a.EdgeEndpoint = am.EdgeEndpoint
	}
	a.EdgeUpdateBulk = a.EdgeUpdateBulk || am.EdgeUpdateBulk
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

// GetPagination returns the pagination annotation (or defaults from
// [Config.DisablePagination] and [Config.DefaultEagerLoad] depending on the
// type/edge/etc).
func (a *Annotation) GetPagination(config *Config, edge *gen.Edge) bool {
	if a.Pagination == nil {
		if config.DisablePagination {
			return false
		}
		if edge != nil {
			ea := GetAnnotation(edge)
			if ea.EagerLoad != nil && *ea.EagerLoad {
				return false
			}
			return !config.DefaultEagerLoad
		}
		return true
	}
	return *a.Pagination
}

// GetMinItemsPerPage returns the minimum number of items per page for paginated calls
// (or defaults from [Config.MinItemsPerPage]).
func (a *Annotation) GetMinItemsPerPage(config *Config) int {
	if a.MinItemsPerPage == 0 {
		return config.MinItemsPerPage
	}
	return a.MinItemsPerPage
}

// GetMaxItemsPerPage returns the maximum number of items per page for paginated calls
// (or defaults from [Config.MaxItemsPerPage]).
func (a *Annotation) GetMaxItemsPerPage(config *Config) int {
	if a.MaxItemsPerPage == 0 {
		return config.MaxItemsPerPage
	}
	return a.MaxItemsPerPage
}

// GetItemsPerPage returns the number of items per page for paginated calls
// (or defaults from [Config.ItemsPerPage]).
func (a *Annotation) GetItemsPerPage(config *Config) int {
	if a.ItemsPerPage == 0 {
		return config.ItemsPerPage
	}
	return a.ItemsPerPage
}

// GetEagerLoad returns if the edge should be eager-loaded (or defaults from
// [Config.DefaultEagerLoad]).
func (a *Annotation) GetEagerLoad(config *Config) bool {
	if a.EagerLoad == nil {
		return config.DefaultEagerLoad
	}
	return *a.EagerLoad
}

// GetEdgeEndpoint returns if the edge should have an endpoint (or defaults from
// [Config.DisableEagerLoadedEndpoints]).
func (a *Annotation) GetEdgeEndpoint(config *Config) bool {
	if a.EdgeEndpoint != nil {
		return *a.EdgeEndpoint
	}
	if config.DisableEagerLoadedEndpoints {
		// Only return false if the edge is in fact eager-loaded.
		if a.EagerLoad != nil && *a.EagerLoad {
			return false
		}
		if a.EagerLoad == nil && config.DefaultEagerLoad {
			return false
		}
	}
	return true
}

// HasOperation returns if the operation is allowed on the given annotation.
func (a *Annotation) HasOperation(config *Config, op Operation) bool {
	for _, o := range a.GetOperations(config) {
		if o == op {
			return true
		}
	}
	return false
}

// GetOperations returns the operations annotation (or defaults from
// [Config.DefaultOperations]).
func (a *Annotation) GetOperations(config *Config) []Operation {
	if a.Operations == nil {
		return config.DefaultOperations
	}
	return a.Operations
}

// GetOperationSummary returns the summary for the provided operation or an empty
// string if not configured.
func (a *Annotation) GetOperationSummary(op Operation) string {
	if a.OperationSummary == nil {
		return ""
	}
	return a.OperationSummary[op]
}

// GetOperationDescription returns the description for the provided operation or an
// empty string if not configured.
func (a *Annotation) GetOperationDescription(op Operation) string {
	if a.OperationDescription == nil {
		return ""
	}
	return a.OperationDescription[op]
}

// GetOperationID returns the operation ID for the provided operation or an empty
// string if not configured.
func (a *Annotation) GetOperationID(op Operation) string {
	if a.OperationID == nil {
		return ""
	}
	return a.OperationID[op]
}

func (a *Annotation) GetSkip(config *Config) bool {
	return a.Skip || len(a.GetOperations(config)) == 0
}

// WithOperationSummary provides a summary for the specified operation.
func WithOperationSummary(op Operation, v string) Annotation {
	return Annotation{OperationSummary: map[Operation]string{op: v}}
}

// WithOperationDescription provides a description for the specified operation.
func WithOperationDescription(op Operation, v string) Annotation {
	return Annotation{OperationDescription: map[Operation]string{op: v}}
}

// WithAdditionalTags adds additional tags to all operations for this schema/edge.
func WithAdditionalTags(v ...string) Annotation {
	return Annotation{AdditionalTags: v}
}

// WithTags sets the tags for the schema/edge in the REST API. This will otherwise default
// to the schema/edge's name(s).
func WithTags(v ...string) Annotation {
	return Annotation{Tags: v}
}

// WithOperationID provides an operation ID for the specified operation. This should be
// snake-cased and unique for the operation.
func WithOperationID(op Operation, v string) Annotation {
	return Annotation{OperationID: map[Operation]string{op: v}}
}

// WithDescription sets the description for the schema/edge/field in the REST API. This will
// otherwise default to the schema/edge/field's description according to Ent (e.g. the
// comment).
func WithDescription(v string) Annotation {
	return Annotation{Description: v}
}

// WithPagination sets the schema to be paginated in the REST API. This is not required to be
// provided unless pagination was disabled globally.
func WithPagination(v bool) Annotation {
	return Annotation{Pagination: &v}
}

// WithMinItemsPerPage sets an explicit minimum number of items per page for paginated calls.
func WithMinItemsPerPage(v int) Annotation {
	return Annotation{MinItemsPerPage: v}
}

// WithMaxItemsPerPage sets an explicit maximum number of items per page for paginated calls.
func WithMaxItemsPerPage(v int) Annotation {
	return Annotation{MaxItemsPerPage: v}
}

// WithItemsPerPage sets an explicit default number of items per page for paginated calls.
func WithItemsPerPage(v int) Annotation {
	return Annotation{ItemsPerPage: v}
}

// WithEagerLoad sets the edge to be eager-loaded in the REST API for each associated
// entity. Note that edges are not eager-loaded by default.
func WithEagerLoad(v bool) Annotation {
	return Annotation{EagerLoad: &v}
}

// WithEdgeEndpoint sets the edge to have an endpoint. If the edge is eager-loaded,
// and the global config is set to disable endpoints for edges which are also
// eager-loaded, this will default to false. Not required to be provided unless
// endpoints are disabled globally and you want to specifically enable one edge to
// have an endpoint, or want to disable an edge from having an endpoint in general.
func WithEdgeEndpoint(v bool) Annotation {
	return Annotation{EdgeEndpoint: &v}
}

// WithEdgeUpdateBulk allows the edge to be bulk updated on the entities associated with the
// edge. This is disabled by default, which will mean that you must use the "add_<field>"
// and "remove_<field>" object references to associate/disassociate entities with the edge.
// This is disabled by default due to the fact that this can lead to accidental disassociation
// of a massive number of entities, if a user doesn't happen to fully understand the
// implications of providing values to the "bulk" field, which would just be "<field>" (sets
// the non-unique edge to be set to those provided values).
func WithEdgeUpdateBulk(v bool) Annotation {
	return Annotation{EdgeUpdateBulk: v}
}

// WithFilter sets the field to be filterable with the provided predicate(s). When applied
// on an edge with [FilterEdge], it will include the fields associated with the edge
// that are also filterable.
//
// Example:
//
//	entrest.WithFilter(entrest.FilterGroupArray | entrest.FilterGroupLength) // Bundle using groups.
//	entrest.WithFilter(entrest.FilterEQ | entrest.FilterNEQ) // Or use individual predicates.
func WithFilter(v Predicate) Annotation {
	return Annotation{Filter: v}
}

// WithHandler sets the schema/edge to have an HTTP handler generated for it.
func WithHandler(v bool) Annotation {
	return Annotation{Handler: &v}
}

// WithSortable sets the field to be sortable in the REST API. Note that only types that can be
// sorted, will be sortable.
func WithSortable(v bool) Annotation {
	return Annotation{Sortable: v}
}

// WithSkip sets the schema, edge, or field to be skipped in the REST API.
func WithSkip(v bool) Annotation {
	return Annotation{Skip: v}
}

// WithReadOnly sets the field to be read-only in the REST API. If you want to make
// a schema or edge read-only, use the Operations annotation instead.
func WithReadOnly(v bool) Annotation {
	return Annotation{ReadOnly: v}
}

// WithExample sets the OpenAPI Specification example value for a field.
func WithExample(v any) Annotation {
	return Annotation{Example: v}
}

// WithSchema sets the OpenAPI schema for a field.
func WithSchema(v *ogen.Schema) Annotation {
	return Annotation{Schema: v}
}

// WithIncludeOperations includes the specified operations in the REST API for the
// schema. If empty, all operations are generated (unless globally disabled).
func WithIncludeOperations(v ...Operation) Annotation {
	return Annotation{Operations: v}
}

// WithExcludeOperations excludes the specified operations in the REST API for the
// schema. If empty, all operations are generated (unless globally disabled).
func WithExcludeOperations(v ...Operation) Annotation {
	var ops []Operation
	for _, o := range AllOperations {
		if !slices.Contains(v, o) {
			ops = append(ops, o)
		}
	}
	return Annotation{Operations: ops}
}
