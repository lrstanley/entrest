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
	ReadOnly             bool                 `json:",omitempty" ent:"field"`

	// All others.

	Pagination      *bool       `json:",omitempty" ent:"schema,edge"`
	MinItemsPerPage int         `json:",omitempty" ent:"schema,edge"`
	MaxItemsPerPage int         `json:",omitempty" ent:"schema,edge"`
	ItemsPerPage    int         `json:",omitempty" ent:"schema,edge"`
	EagerLoad       *bool       `json:",omitempty" ent:"edge"`
	EagerLoadLimit  *int        `json:",omitempty" ent:"edge"`
	EdgeEndpoint    *bool       `json:",omitempty" ent:"edge"`
	EdgeUpdateBulk  bool        `json:",omitempty" ent:"edge"`
	Filter          Predicate   `json:",omitempty" ent:"schema,edge,field"`
	FilterGroup     string      `json:",omitempty" ent:"edge,field"`
	DisableHandler  bool        `json:",omitempty" ent:"schema,edge"`
	Sortable        bool        `json:",omitempty" ent:"field"`
	DefaultSort     *string     `json:",omitempty" ent:"schema"`
	DefaultOrder    *SortOrder  `json:",omitempty" ent:"schema"`
	Skip            bool        `json:",omitempty" ent:"schema,edge,field"`
	AllowClientIDs  *bool       `json:",omitempty" ent:"schema"`
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
	if am.EagerLoadLimit != nil {
		a.EagerLoadLimit = am.EagerLoadLimit
	}
	if am.EdgeEndpoint != nil {
		a.EdgeEndpoint = am.EdgeEndpoint
	}
	a.EdgeUpdateBulk = a.EdgeUpdateBulk || am.EdgeUpdateBulk
	if am.Filter != 0 {
		a.Filter = am.Filter.Add(a.Filter)
	}
	if am.FilterGroup != "" {
		a.FilterGroup = am.FilterGroup
	}
	a.DisableHandler = a.DisableHandler || am.DisableHandler
	a.Sortable = a.Sortable || am.Sortable
	if am.DefaultSort != nil {
		a.DefaultSort = am.DefaultSort
	}
	if am.DefaultOrder != nil {
		a.DefaultOrder = am.DefaultOrder
	}
	a.Skip = a.Skip || am.Skip
	if am.AllowClientIDs != nil {
		a.AllowClientIDs = am.AllowClientIDs
	}
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
			if !config.DisableEagerLoadNonPagedOpt {
				if (ea.EagerLoad != nil && *ea.EagerLoad) || config.DefaultEagerLoad {
					return false
				}
			}
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

// GetEagerLoadLimit returns the limit for the max number of entities to eager-load for the
// edge (or defaults from [Config.EagerLoadLimit]).
func (a *Annotation) GetEagerLoadLimit(config *Config) int {
	if a.EagerLoadLimit == nil || *a.EagerLoadLimit == 0 {
		return config.EagerLoadLimit
	}
	return *a.EagerLoadLimit
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

// GetDefaultSort returns the default sort field for the schema in the REST API.
// If one was not previously specified, but the type has an ID field, it will default
// to "id".
func (a *Annotation) GetDefaultSort(hasID bool) string {
	if a.DefaultSort == nil {
		if hasID {
			return "id"
		}
		return ""
	}
	return *a.DefaultSort
}

// GetDefaultOrder returns the default sorting order for the schema in the REST API,
// defaulting to ascending if not specified.
func (a *Annotation) GetDefaultOrder() SortOrder {
	if a.DefaultOrder == nil {
		return OrderAsc
	}
	return *a.DefaultOrder
}

func (a *Annotation) GetSkip(config *Config) bool {
	return a.Skip || len(a.GetOperations(config)) == 0
}

func (a *Annotation) GetAllowClientIDs(config *Config) bool {
	if a.AllowClientIDs == nil {
		return config.AllowClientIDs
	}
	return *a.AllowClientIDs
}

// WithOperationSummary provides a summary for the specified operation. This should be
// a short summary of what the operation does.
func WithOperationSummary(op Operation, v string) Annotation {
	return Annotation{OperationSummary: map[Operation]string{op: v}}
}

// WithOperationDescription provides a description for the specified operation. This
// should be a verbose explanation of the operation behavior. CommonMark syntax MAY be
// used for rich text representation.
func WithOperationDescription(op Operation, v string) Annotation {
	return Annotation{OperationDescription: map[Operation]string{op: v}}
}

// WithAdditionalTags adds additional tags to all operations for this schema/edge. Tags
// can be used for logical grouping of operations by resources or any other qualifier.
func WithAdditionalTags(v ...string) Annotation {
	return Annotation{AdditionalTags: v}
}

// Sets the tags for all operations for this schema/edge. This will otherwise default
// to the schema/edge's name(s). Tags can be used for logical grouping of operations by
// resources or any other qualifier.
func WithTags(v ...string) Annotation {
	return Annotation{Tags: v}
}

// WithOperationID provides an operation ID for the specified operation. This should be
// snake-cased and MUST BE UNIQUE for the operation.
func WithOperationID(op Operation, v string) Annotation {
	return Annotation{OperationID: map[Operation]string{op: v}}
}

// WithDescription sets the description for the schema/edge/field in the REST API. This will
// otherwise default to the schema/edge/field's description according to Ent (e.g. the
// comment). It's recommended to use the field comment rather than setting this annotation
// when possible.
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
// entity. Note that edges are not eager-loaded by default. Eager-loading, when enabled,
// means that the configured edge is always fetched when the parent entity is fetched
// (only covering the first level, it does not recurse).
func WithEagerLoad(v bool) Annotation {
	return Annotation{EagerLoad: &v}
}

// WithEagerLoadLimit sets the limit for the max number of entities to eager-load for the
// edge. There is a global default limit for eager-loading, which can be set via
// [Config.EagerLoadLimit]. Defaults to 1000, and the limit can be disabled by setting
// the value to -1.
func WithEagerLoadLimit(v int) Annotation {
	if v == 0 {
		return Annotation{}
	}
	return Annotation{EagerLoadLimit: &v}
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

// WithFilterGroup adds the field to a group of other fields that are filtered together. Note that
// only common filter options across all of the groups will be supported. The goal of this is to group
// common fields that would be searched together. This allows slightly more advanced logical operations,
// like the following:
//   - and(type.eq==FOO, or(field1.ihas==bar, field2.ihas==bar, field3.ihas==bar))
//
// You can use [WithFilterGroup] on edges to also allow any matching groups on the edge to be included
// in this filter group. The group name must match when used on the edge if you want that edge to be
// included in that group.
func WithFilterGroup(name string) Annotation {
	return Annotation{FilterGroup: name}
}

// WithHandler sets the schema/edge to have an HTTP handler generated for it. Unless a schema/edge
// is skipped or has the specific operation disabled, an HTTP handler/endpoint will be generated for
// it by default. This does not prevent the endpoint from being created within the spec, rather only
// prevents the handler from being mounted. The handler functions will still be generated in case
// you want to build upon them.
func WithHandler(v bool) Annotation {
	return Annotation{DisableHandler: !v}
}

// WithSortable sets the field to be sortable in the REST API. Note that only types that can be
// sorted, will be sortable.
func WithSortable(v bool) Annotation {
	return Annotation{Sortable: v}
}

// WithDefaultSort sets the default sort field for the schema in the REST API. If not specified,
// will default to the "id" field (if it exists on the schema/edge). The provided field must exist
// on the schema, otherwise codegen will fail. You may provide any of the typical fields shown for
// the "sort" field in the OpenAPI specification for this schema. E.g. "id", "created_at",
// "someedge.count" (<edge>.<edge-field>), etc.
//
// Note that this will also change the way eager-loaded edges which are based on this schema are
// sorted. This is currently the only way to sort eager-loaded data.
func WithDefaultSort(field string) Annotation {
	return Annotation{DefaultSort: &field}
}

// WithDefaultOrder sets the default sorting order for the schema in the REST API. If not specified,
// will default to ASC.
//
// Note that this will also change the way eager-loaded edges which are based on this schema are
// sorted. This is currently the only way to sort eager-loaded data.
func WithDefaultOrder(v SortOrder) Annotation {
	return Annotation{DefaultOrder: &v}
}

// WithSkip sets the schema, edge, or field to be skipped in the REST API. Primarily useful if an entire
// schema shouldn't be queryable, or if there is a sensitive field that should never be returned (but
// sensitive isn't set on the field for some reason).
func WithSkip(v bool) Annotation {
	return Annotation{Skip: v}
}

// WithAllowClientIDs sets the schema to allow the client to provide the ID field
// as part of a CREATE payload. This is beneficial to allow the client to supply
// UUIDs as primary keys (for idempotency), or when your ID field is a username, for
// example. This is not required if [Config.AllowClientIDs] is enabled.
//
// SECURITY NOTE: allowing requests to include the ID field is not recommended, unless
// you add necessary validation (permissions) or disallow resources from being deleted.
// Otherwise, you may allow an attacker to spoof a previously deleted resource, leading
// to takeover attack vectors.
func WithAllowClientIDs(v bool) Annotation {
	return Annotation{AllowClientIDs: &v}
}

// WithReadOnly sets the field to be read-only in the REST API. If you want to make
// a schema or edge read-only, use the Operations annotation instead.
func WithReadOnly(v bool) Annotation {
	return Annotation{ReadOnly: v}
}

// WithExample sets the OpenAPI Specification example value for a field. This is recommended if it's
// not obvious what the fields purpose is, or what the format could be. Many OpenAPI documentation
// browsers will use this information as an example value within the POST/PATCH body.
func WithExample(v any) Annotation {
	return Annotation{Example: v}
}

// WithDeprecated sets the field to be deprecated in the REST API.
func WithDeprecated(v bool) Annotation {
	return Annotation{Deprecated: v}
}

// WithSchema sets the OpenAPI schema for a field. This is required for any fields which
// are JSON based, or don't have a pre-defined ent type for the field.
//
// You can use [SchemaObjectAny] for an object with any properties (effectively "any").
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
