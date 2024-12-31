// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

//nolint:goconst
package entrest

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"entgo.io/ent/entc/gen"
	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/jsonschema"
)

const eagerLoadDepthMessage = "If the entity has eager-loaded edges, the depth of when those will be loaded is limited to a depth of 1 (entity -> edge, not entity -> edge -> edge -> etc)."

func addPagination(spec *ogen.Spec, _ *Config) {
	if spec.Components == nil {
		spec.Components = &ogen.Components{}
	}

	if spec.Components.Parameters == nil {
		spec.Components.Parameters = make(map[string]*ogen.Parameter)
	}

	if _, ok := spec.Components.Schemas["Page"]; !ok {
		spec.Components.Parameters["Page"] = &ogen.Parameter{
			Name:        "page",
			In:          "query",
			Description: "The page number to retrieve.",
			Schema: ogen.Int().
				SetMinimum(ptr(int64(1))).
				SetDefault(json.RawMessage(`1`)),
		}
	}

	if spec.Components.Schemas == nil {
		spec.Components.Schemas = make(map[string]*ogen.Schema)
	}

	if _, ok := spec.Components.Schemas["PagedResponse"]; ok {
		return
	}

	pagedSchema := &ogen.Schema{
		Type: "object",
		Properties: ogen.Properties{
			{
				Name: "page",
				Schema: &ogen.Schema{
					Type:        "integer",
					Description: "Page which the results are associated with.",
					Example:     jsonschema.RawValue(`1`),
					Minimum:     ogen.Int().SetMinimum(ptr(int64(1))).Minimum,
				},
			},
			{
				Name: "last_page",
				Schema: &ogen.Schema{
					Type:        "integer",
					Description: "The number of the last page of results.",
					Example:     jsonschema.RawValue(`3`),
					Minimum:     ogen.Int().SetMinimum(ptr(int64(1))).Minimum,
				},
			},
			{
				Name: "is_last_page",
				Schema: &ogen.Schema{
					Type:        "boolean",
					Description: "If true, the current results are the last page of results.",
					Example:     jsonschema.RawValue(`false`),
				},
			},
			{
				Name: "total_count",
				Schema: &ogen.Schema{
					Type:        "integer",
					Description: "The total number of results based on the provided query.",
					Example:     jsonschema.RawValue(`123`),
					Minimum:     ogen.Int().SetMinimum(ptr(int64(0))).Minimum,
				},
			},
		},
		Required: []string{"page", "last_page", "is_last_page", "total_count"},
	}

	spec.Components.Schemas["PagedResponse"] = pagedSchema
}

func newBaseSpec(_ *Config) *ogen.Spec {
	spec := &ogen.Spec{
		Paths: ogen.Paths{},
		Components: &ogen.Components{
			Schemas: map[string]*ogen.Schema{
				"FilterOperation": {
					Description: "Specifies how to combine multiple filters.",
					Type:        "string",
					Enum:        sliceToRawMessage([]string{"and", "or"}),
					Default:     jsonschema.RawValue(`"and"`),
				},
			},
			Parameters: map[string]*ogen.Parameter{
				"PrettyResponse": {
					Name:        "pretty",
					In:          "query",
					Description: "If set to true, any JSON response will be indented.",
					Schema:      ogen.Bool(),
				},
				"FilterOperation": {
					Name:        "filter_op",
					In:          "query",
					Description: "Filter operation to use.",
					Schema:      &ogen.Schema{Ref: "#/components/schemas/FilterOperation"},
				},
			},
		},
		Tags: []ogen.Tag{
			{
				Name:        "Meta",
				Description: "Includes various endpoints for meta information about the service, like the OpenAPI spec, version, health, etc.",
			},
		},
	}

	return spec
}

// GetSpecType generates an independent spec for the given type, which should encapsulate
// all schemas, parameters, components and paths for the provided type that can then be
// merged into another spec.
func GetSpecType(t *gen.Type, op Operation) (*ogen.Spec, error) { // nolint:funlen,gocyclo,cyclop
	cfg := GetConfig(t.Config)
	ta := GetAnnotation(t)

	entityName := Singularize(t.Name)

	spec := newBaseSpec(cfg)
	spec.Tags = append(spec.Tags, ogen.Tag{
		Name:        Pluralize(t.Name),
		Description: ta.Description,
	})

	if op != OperationList && op != OperationCreate {
		idSchema, err := GetSchemaField(t.ID)
		if err != nil {
			return nil, err
		}

		spec.Components.Parameters[Singularize(t.Name)+"ID"] = &ogen.Parameter{
			Name:        CamelCase(Singularize(t.Name)) + "ID",
			In:          "path",
			Description: fmt.Sprintf("The ID of the %s to act upon.", entityName),
			Required:    true,
			Schema:      idSchema,
		}
	}

	for k, v := range GetSchemaType(t, op, nil) {
		spec.Components.Schemas[k] = v
	}

	switch op {
	case OperationCreate:
		oper := &ogen.Operation{
			Tags: sliceCompact(sliceOr(ta.Tags, append([]string{Pluralize(t.Name)}, ta.AdditionalTags...))),
			Summary: cmp.Or(
				ta.GetOperationSummary(op),
				"Create a new "+CamelCase(entityName),
			),
			Description: cmp.Or(
				ta.GetOperationDescription(op),
				fmt.Sprintf("Create a new %s entity. %s", entityName, eagerLoadDepthMessage),
			),
			OperationID: GetOperationIDName(op, t, nil),
			Deprecated:  ta.Deprecated,
			RequestBody: ogen.NewRequestBody().
				SetRequired(true).
				SetJSONContent(&ogen.Schema{Ref: "#/components/schemas/" + entityName + "Create"}),
			Responses: ogen.Responses{
				strconv.Itoa(http.StatusCreated): ogen.NewResponse().
					SetDescription(fmt.Sprintf("The created %s entity.", entityName)).
					SetJSONContent(&ogen.Schema{Ref: "#/components/schemas/" + entityName + "Read"}),
			},
		}

		spec.Paths[GetPathName(op, t, nil, true)] = &ogen.PathItem{
			Summary:     oper.Summary,
			Description: oper.Description,
			Post:        oper,
			Parameters: []*ogen.Parameter{
				{Ref: "#/components/parameters/PrettyResponse"},
			},
		}
	case OperationUpdate:
		oper := &ogen.Operation{
			Tags: sliceCompact(sliceOr(ta.Tags, append([]string{Pluralize(t.Name)}, ta.AdditionalTags...))),
			Summary: cmp.Or(
				ta.GetOperationSummary(op),
				"Update a "+CamelCase(entityName),
			),
			Description: cmp.Or(
				ta.GetOperationDescription(op),
				fmt.Sprintf("Update an existing %s entity. %s", entityName, eagerLoadDepthMessage),
			),
			OperationID: GetOperationIDName(op, t, nil),
			Deprecated:  ta.Deprecated,
			Parameters:  []*ogen.Parameter{},
			RequestBody: ogen.NewRequestBody().
				SetRequired(true).
				SetJSONContent(&ogen.Schema{Ref: "#/components/schemas/" + entityName + "Update"}),
			Responses: ogen.Responses{
				strconv.Itoa(http.StatusOK): ogen.NewResponse().
					SetDescription(fmt.Sprintf("The update %s entity.", entityName)).
					SetJSONContent(&ogen.Schema{Ref: "#/components/schemas/" + entityName + "Read"}),
			},
		}

		spec.Paths[GetPathName(op, t, nil, true)] = &ogen.PathItem{
			Summary:     fmt.Sprintf("Operate on a single %s entity", entityName),
			Description: fmt.Sprintf("Operate on a single %s entity by its ID.", entityName),
			Patch:       oper,
			Parameters: []*ogen.Parameter{
				{Ref: "#/components/parameters/PrettyResponse"},
				{Ref: "#/components/parameters/" + Singularize(t.Name) + "ID"},
			},
		}
	case OperationRead:
		oper := &ogen.Operation{
			Tags: sliceCompact(sliceOr(ta.Tags, append([]string{Pluralize(t.Name)}, ta.AdditionalTags...))),
			Summary: cmp.Or(
				ta.GetOperationSummary(op),
				"Retrieve a "+CamelCase(entityName),
			),
			Description: cmp.Or(
				ta.GetOperationDescription(op),
				fmt.Sprintf("Retrieve a single %s entity by its ID. %s", entityName, eagerLoadDepthMessage),
			),
			OperationID: GetOperationIDName(op, t, nil),
			Deprecated:  ta.Deprecated,
			Parameters:  []*ogen.Parameter{},
			Responses: ogen.Responses{
				strconv.Itoa(http.StatusOK): ogen.NewResponse().
					SetDescription(fmt.Sprintf("The requested %s entity.", entityName)).
					SetJSONContent(&ogen.Schema{Ref: "#/components/schemas/" + entityName + "Read"}),
			},
		}

		if cfg.AddEdgesToTags {
			oper.Tags = append(oper.Tags, edgesToTags(cfg, t)...)
		}

		spec.Paths[GetPathName(op, t, nil, true)] = &ogen.PathItem{
			Summary:     fmt.Sprintf("Operate on a single %s entity", entityName),
			Description: fmt.Sprintf("Operate on a single %s entity by its ID.", entityName),
			Get:         oper,
			Parameters: []*ogen.Parameter{
				{Ref: "#/components/parameters/PrettyResponse"},
				{Ref: "#/components/parameters/" + Singularize(t.Name) + "ID"},
			},
		}
	case OperationList:
		oper := &ogen.Operation{
			Tags: sliceCompact(sliceOr(ta.Tags, append([]string{Pluralize(t.Name)}, ta.AdditionalTags...))),
			Summary: cmp.Or(
				ta.GetOperationSummary(op),
				"List "+CamelCase(Pluralize(t.Name)),
			),
			Description: cmp.Or(
				ta.GetOperationDescription(op),
				fmt.Sprintf("List %s entities (including pagination, filtering, sorting, etc). %s", entityName, eagerLoadDepthMessage),
			),
			OperationID: GetOperationIDName(op, t, nil),
			Deprecated:  ta.Deprecated,
			Parameters:  []*ogen.Parameter{},
			Responses: ogen.Responses{
				strconv.Itoa(http.StatusOK): ogen.NewResponse().
					SetDescription(fmt.Sprintf("The requested %s.", entityName)).
					SetJSONContent(ogen.NewSchema().SetRef("#/components/schemas/" + entityName + "List")),
			},
		}

		if ta.GetPagination(cfg, nil) {
			addPagination(spec, cfg)

			oper.Parameters = append(
				oper.Parameters,
				&ogen.Parameter{Ref: "#/components/parameters/Page"},
				&ogen.Parameter{
					Name:        "per_page",
					In:          "query",
					Description: "The number of entities to retrieve per page.",
					Schema: ogen.Int().
						SetMinimum(ptr(int64(ta.GetMinItemsPerPage(cfg)))).
						SetMaximum(ptr(int64(ta.GetMaxItemsPerPage(cfg)))).
						SetDefault(json.RawMessage(strconv.Itoa(ta.GetItemsPerPage(cfg)))),
				},
			)
		}

		if sortable := GetSortableFields(t, nil); len(sortable) > 1 {
			sortParam := &ogen.Parameter{
				Name:        "sort",
				In:          "query",
				Description: "Sort entity results by the given field.",
				Schema:      &ogen.Schema{Ref: "#/components/schemas/" + addSortableFields(spec, t, sortable)},
			}
			if v := ta.GetDefaultSort(t.ID != nil); v != "" {
				sortParam.Schema = sortParam.Schema.SetDefault(json.RawMessage(fmt.Sprintf("%q", v)))
			}
			orderParam := &ogen.Parameter{
				Name:        "order",
				In:          "query",
				Description: "Order the results in ascending or descending order.",
				Schema: &ogen.Schema{
					Type:    "string",
					Enum:    sliceToRawMessage([]string{"asc", "desc"}),
					Default: ogen.Default(json.RawMessage(fmt.Sprintf("%q", ta.GetDefaultOrder()))),
				},
			}
			oper.Parameters = append(oper.Parameters, sortParam, orderParam)
		}

		if filters := GetFilterableFields(t, nil); len(filters) > 0 {
			oper.Parameters = append(oper.Parameters, &ogen.Parameter{Ref: "#/components/parameters/FilterOperation"})

			for _, f := range filters {
				name := f.ComponentName()
				spec.Components.Parameters[name] = f.Parameter()
				oper.Parameters = append(oper.Parameters, &ogen.Parameter{Ref: "#/components/parameters/" + name})
			}
		}

		if groups := GetFilterGroups(t, nil); len(groups) > 0 {
			for _, g := range groups {
				for _, op := range g.Operations {
					name := g.ComponentName(op)
					spec.Components.Parameters[name] = g.Parameter(op)
					oper.Parameters = append(oper.Parameters, &ogen.Parameter{Ref: "#/components/parameters/" + name})
				}
			}
		}

		if cfg.AddEdgesToTags {
			oper.Tags = append(oper.Tags, edgesToTags(cfg, t)...)
		}

		spec.Paths[GetPathName(op, t, nil, true)] = &ogen.PathItem{
			Summary:     oper.Summary,
			Description: oper.Description,
			Get:         oper,
			Parameters: []*ogen.Parameter{
				{Ref: "#/components/parameters/PrettyResponse"},
			},
		}
	case OperationDelete:
		oper := &ogen.Operation{
			Tags: sliceCompact(sliceOr(ta.Tags, append([]string{Pluralize(t.Name)}, ta.AdditionalTags...))),
			Summary: cmp.Or(
				ta.GetOperationSummary(op),
				"Delete a "+CamelCase(entityName),
			),
			Description: cmp.Or(
				ta.GetOperationDescription(op),
				fmt.Sprintf("Delete a single %s entity by its ID.", entityName),
			),
			OperationID: GetOperationIDName(op, t, nil),
			Deprecated:  ta.Deprecated,
			Parameters:  []*ogen.Parameter{},
			Responses: ogen.Responses{
				strconv.Itoa(http.StatusNoContent): ogen.NewResponse().
					SetDescription(fmt.Sprintf("The requested %s entity.", entityName)),
			},
		}

		spec.Paths[GetPathName(op, t, nil, true)] = &ogen.PathItem{
			Summary:     fmt.Sprintf("Operate on a single %s entity", entityName),
			Description: fmt.Sprintf("Operate on a single %s entity by its ID.", entityName),
			Delete:      oper,
			Parameters: []*ogen.Parameter{
				{Ref: "#/components/parameters/" + Singularize(t.Name) + "ID"},
			},
		}
	default:
		panic(fmt.Sprintf("unsupported operation %q", op))
	}

	return spec, nil
}

// GetSpecEdge generates an independent spec for the given edge, which should encapsulate
// all schemas, parameters, components and paths for the provided edge that can then be
// merged into another spec.
func GetSpecEdge(t *gen.Type, e *gen.Edge, op Operation) (*ogen.Spec, error) { // nolint:funlen,gocyclo,cyclop
	cfg := GetConfig(t.Config)
	ta := GetAnnotation(t)
	ea := GetAnnotation(e)
	ra := GetAnnotation(e.Type)

	if ea.GetSkip(cfg) {
		return nil, errors.New("edge is skipped")
	}

	if !ea.GetEdgeEndpoint(cfg) {
		return nil, errors.New("edge has endpoint disabled or edge is eager-loaded with global config to disable endpoints for edges which are also eager-loaded")
	}

	rootEntityName := Singularize(t.Name)
	refEntityName := Singularize(e.Type.Name)
	entityName := Singularize(PascalCase(e.Name))

	spec := newBaseSpec(cfg)
	spec.Tags = append(
		spec.Tags,
		ogen.Tag{
			Name:        Pluralize(t.Name),
			Description: ta.Description,
		},
		ogen.Tag{
			Name:        Pluralize(e.Type.Name),
			Description: ra.Description,
		},
	)

	idSchema, err := GetSchemaField(t.ID)
	if err != nil {
		return nil, err
	}

	spec.Components.Parameters[Singularize(t.Name)+"ID"] = &ogen.Parameter{
		Name:        CamelCase(Singularize(t.Name)) + "ID",
		In:          "path",
		Description: fmt.Sprintf("The ID of the %s to act upon.", rootEntityName),
		Required:    true,
		Schema:      idSchema,
	}

	for k, v := range GetSchemaType(t, op, e) {
		spec.Components.Schemas[k] = v
	}

	switch op {
	case OperationRead: // Unique.
		if !e.Unique {
			return nil, errors.New("edge is not unique")
		}

		oper := &ogen.Operation{
			Tags: sliceCompact(sliceOr(ea.Tags, append([]string{Pluralize(t.Name), Pluralize(e.Type.Name)}, ea.AdditionalTags...))),
			Summary: cmp.Or(
				ea.GetOperationSummary(op),
				e.Comment(),
				fmt.Sprintf("Get a %s associated %s", Pluralize(CamelCase(t.Name)), CamelCase(e.Name)),
			),
			Description: cmp.Or(
				ea.GetOperationDescription(op),
				fmt.Sprintf(
					"Get a %s associated %s (%s entity type). %s",
					Pluralize(CamelCase(t.Name)),
					CamelCase(e.Name),
					refEntityName,
					eagerLoadDepthMessage,
				),
			),
			OperationID: GetOperationIDName(op, t, e),
			Deprecated:  ta.Deprecated || ea.Deprecated || ra.Deprecated,
			Parameters:  []*ogen.Parameter{},
			Responses: ogen.Responses{
				strconv.Itoa(http.StatusOK): ogen.NewResponse().
					SetDescription(fmt.Sprintf("The requested %s entity.", CamelCase(e.Name))).
					SetJSONContent(&ogen.Schema{Ref: "#/components/schemas/" + refEntityName + "Read"}),
			},
		}

		spec.Paths[GetPathName(op, t, e, true)] = &ogen.PathItem{
			Summary:     oper.Summary,     // Will probably always be the same.
			Description: oper.Description, // Will probably always be the same.
			Get:         oper,
			Parameters: []*ogen.Parameter{
				{Ref: "#/components/parameters/PrettyResponse"},
				{Ref: "#/components/parameters/" + Singularize(t.Name) + "ID"},
			},
		}
	case OperationList: // Not unique.
		if e.Unique {
			return nil, errors.New("edge is unique")
		}

		oper := &ogen.Operation{
			Tags: sliceCompact(sliceOr(ea.Tags, append([]string{Pluralize(t.Name), Pluralize(e.Type.Name)}, ea.AdditionalTags...))),
			Summary: cmp.Or(
				ea.GetOperationSummary(op),
				e.Comment(),
				fmt.Sprintf("List a %s associated %s", Pluralize(CamelCase(t.Name)), Pluralize(CamelCase(e.Name))),
			),
			Description: cmp.Or(
				ea.GetOperationDescription(op),
				fmt.Sprintf(
					"List a %s associated %s (%s entity type). %s",
					Pluralize(CamelCase(t.Name)),
					Pluralize(CamelCase(e.Name)),
					refEntityName,
					eagerLoadDepthMessage,
				),
			),
			OperationID: GetOperationIDName(op, t, e),
			Deprecated:  ta.Deprecated || ea.Deprecated || ra.Deprecated,
			Parameters:  []*ogen.Parameter{},
			Responses: ogen.Responses{
				strconv.Itoa(http.StatusOK): ogen.NewResponse().
					SetDescription(fmt.Sprintf("The requested %s.", Pluralize(CamelCase(e.Name)))).
					SetJSONContent(ogen.NewSchema().SetRef("#/components/schemas/" + refEntityName + "List")),
			},
		}

		code := strconv.Itoa(http.StatusOK)

		if ea.GetPagination(cfg, e) || ra.GetPagination(cfg, e) {
			addPagination(spec, cfg)
			oper.Parameters = append(oper.Parameters,
				&ogen.Parameter{Ref: "#/components/parameters/Page"},
				&ogen.Parameter{
					Name:        "per_page",
					In:          "query",
					Description: "The number of entities to retrieve per page.",
					Schema: ogen.Int().
						SetMinimum(ptr(int64(ta.GetMinItemsPerPage(cfg)))).
						SetMaximum(ptr(int64(ta.GetMaxItemsPerPage(cfg)))).
						SetDefault(json.RawMessage(strconv.Itoa(ta.GetItemsPerPage(cfg)))),
				},
			)

			// If edge pagination is enabled, but edge type is not paginated, we cannot re-use
			// the paginated schema from the edge type.
			if !ra.GetPagination(cfg, e) {
				oper.Responses[code] = oper.Responses[code].SetJSONContent(&ogen.Schema{
					Ref: "#/components/schemas/" + rootEntityName + entityName + "List",
				})
			}
		} else if !cfg.DisableEagerLoadNonPagedOpt {
			// We're setting a specific schema for the edge response because the edge isn't
			// paginated, but the underlying edge type schema is.
			oper.Responses[code] = oper.Responses[code].SetJSONContent(&ogen.Schema{
				Ref: "#/components/schemas/" + rootEntityName + entityName + "List",
			})
		}

		if sortable := GetSortableFields(e.Type, nil); len(sortable) > 1 {
			sortParam := &ogen.Parameter{
				Name:        "sort",
				In:          "query",
				Description: "Sort entity results by the given field.",
				Schema:      &ogen.Schema{Ref: "#/components/schemas/" + addSortableFields(spec, e.Type, sortable)},
			}
			if v := ra.GetDefaultSort(t.ID != nil && (e == nil || e.Field() == nil)); v != "" {
				sortParam.Schema = sortParam.Schema.SetDefault(json.RawMessage(fmt.Sprintf("%q", v)))
			}
			orderParam := &ogen.Parameter{
				Name:        "order",
				In:          "query",
				Description: "Order the results in ascending or descending order.",
				Schema: &ogen.Schema{
					Type:    "string",
					Enum:    sliceToRawMessage([]string{"asc", "desc"}),
					Default: ogen.Default(json.RawMessage(fmt.Sprintf("%q", ra.GetDefaultOrder()))),
				},
			}
			oper.Parameters = append(oper.Parameters, sortParam, orderParam)
		}

		if filters := GetFilterableFields(e.Type, nil); len(filters) > 0 {
			oper.Parameters = append(oper.Parameters, &ogen.Parameter{Ref: "#/components/parameters/FilterOperation"})

			for _, f := range filters {
				name := f.ComponentName()
				spec.Components.Parameters[name] = f.Parameter()
				oper.Parameters = append(oper.Parameters, &ogen.Parameter{Ref: "#/components/parameters/" + name})
			}
		}

		if groups := GetFilterGroups(e.Type, nil); len(groups) > 0 {
			for _, g := range groups {
				for _, op := range g.Operations {
					name := g.ComponentName(op)
					spec.Components.Parameters[name] = g.Parameter(op)
					oper.Parameters = append(oper.Parameters, &ogen.Parameter{Ref: "#/components/parameters/" + name})
				}
			}
		}

		if cfg.AddEdgesToTags {
			oper.Tags = append(oper.Tags, edgesToTags(cfg, e.Type)...)
		}

		spec.Paths[GetPathName(op, t, e, true)] = &ogen.PathItem{
			Summary:     oper.Summary,
			Description: oper.Description,
			Get:         oper,
			Parameters: []*ogen.Parameter{
				{Ref: "#/components/parameters/PrettyResponse"},
				{Ref: "#/components/parameters/" + Singularize(t.Name) + "ID"},
			},
		}
	default:
		panic(fmt.Sprintf("unsupported operation %q", op))
	}

	return spec, nil
}

// edgesToTags allows providing additional tags for a given operation based on the
// eager-loaded schemas in the response schema.
func edgesToTags(cfg *Config, t *gen.Type) (tags []string) {
	for _, e := range t.Edges {
		ea := GetAnnotation(e)
		if !ea.GetSkip(cfg) && ea.GetEagerLoad(cfg) {
			tags = append(tags, Pluralize(e.Type.Name))
		}
	}
	return tags
}

// addSortableFields adds a schema entry for the provided type into the spec, returning
// the name of the schema entry.
func addSortableFields(spec *ogen.Spec, t *gen.Type, fields []string) (ref string) {
	ref = Singularize(t.Name) + "SortableFields"

	s := &ogen.Schema{
		Description: "All potential sortable fields for " + Singularize(t.Name) + " entities.",
		Type:        "string",
		Enum:        sliceToRawMessage(fields),
	}
	if t.ID != nil {
		s.Default = jsonschema.RawValue(`"id"`)
	}
	spec.Components.Schemas[ref] = s
	return ref
}

// addGlobalRequestHeaders adds the given headers to shared component parameters,
// then adds each of those parameters to each path root (rather than each request,
// to deduplicate references for those headers).
//
// NOTE: order of operations for this function is important. Ideally, it should be
// called after all requests have been added to the spec.
func addGlobalRequestHeaders(spec *ogen.Spec, headers map[string]*ogen.Header) {
	for k, v := range headers {
		spec.Components.Parameters[k] = v.InHeader().SetName(k)
	}

	for pathName := range spec.Paths {
		for k := range headers {
			spec.Paths[pathName].Parameters = append(
				spec.Paths[pathName].Parameters,
				&ogen.Parameter{Ref: "#/components/parameters/" + k},
			)
		}
	}
}

// addGlobalResponseHeaders adds the given headers to shared component headers,
// then adds each of those headers to every single response body.
//
// NOTE: order of operations for this function is important. Ideally, it should be
// called after all responses have been added to the spec (including error responses).
func addGlobalResponseHeaders(spec *ogen.Spec, headers map[string]*ogen.Header) {
	if spec.Components.Headers == nil {
		spec.Components.Headers = make(map[string]*ogen.Header)
	}

	for k, v := range headers {
		spec.Components.Headers[k] = v
	}

	for pathName, pathItem := range spec.Paths {
		spec.Paths[pathName] = PatchPathItem(pathItem, func(resp *ogen.Response) *ogen.Response {
			if resp.Ref != "" {
				return resp
			}

			if resp.Headers == nil {
				resp.Headers = make(map[string]*ogen.Header)
			}

			for k := range headers {
				resp.Headers[k] = &ogen.Header{Ref: "#/components/headers/" + k}
			}

			return resp
		})
	}

	for k := range headers {
		for r := range spec.Components.Responses {
			if spec.Components.Responses[r].Ref != "" {
				continue
			}

			if spec.Components.Responses[r].Headers == nil {
				spec.Components.Responses[r].Headers = make(map[string]*ogen.Header)
			}
			spec.Components.Responses[r].Headers[k] = &ogen.Header{Ref: "#/components/headers/" + k}
		}
	}
}

// addGlobalErrorResponses adds the given error responses to shared component
// responses, then adds each of those responses to all responses.
//
// NOTE: order of operations for this function is important. Ideally, it should be
// called before headers are added.
func addGlobalErrorResponses(cfg *Config, spec *ogen.Spec, responses map[int]*ogen.Schema) {
	// TODO: there is probably a more clean way of doing this, but this also covers
	// user-provided paths/operations passed in via config and hooks.

	if spec.Components.Responses == nil {
		spec.Components.Responses = map[string]*ogen.Response{}
	}

	for k, v := range responses {
		name := "Error" + PascalCase(http.StatusText(k))
		spec.Components.Schemas[name] = v
		spec.Components.Responses[name] = &ogen.Response{
			Description: fmt.Sprintf("%s (http status code %d)", http.StatusText(k), k),
			Content: map[string]ogen.Media{
				"application/json": {
					Schema: &ogen.Schema{Ref: "#/components/schemas/" + name},
				},
			},
		}
	}

	for pathName, pathItem := range spec.Paths {
		spec.Paths[pathName] = PatchOperations(pathItem, func(_ string, op *ogen.Operation) *ogen.Operation {
			if op == nil {
				return nil
			}

			if op.Responses == nil {
				op.Responses = map[string]*ogen.Response{}
			}

			for k := range responses {
				switch {
				case strings.HasPrefix(op.OperationID, "list") && k == http.StatusNotFound && !cfg.ListNotFound:
					continue
				case !strings.HasPrefix(op.OperationID, "create") && !strings.HasPrefix(op.OperationID, "update") && k == http.StatusConflict:
					continue
				}

				op.Responses[strconv.Itoa(k)] = &ogen.Response{Ref: "#/components/responses/Error" + PascalCase(http.StatusText(k))}
			}

			return op
		})
	}
}

// ErrorResponseObject returns a default error schema for the provided HTTP status code.
func ErrorResponseObject(code int) *ogen.Schema {
	return &ogen.Schema{
		Type: "object",
		Properties: []ogen.Property{
			{
				Name: "error",
				Schema: &ogen.Schema{
					Type:        "string",
					Description: "The underlying error, which may be masked when debugging is disabled.",
				},
			},
			{
				Name: "type",
				Schema: &ogen.Schema{
					Type:        "string",
					Description: "A summary of the error code based off the HTTP status code or application error code.",
					Example:     jsonschema.RawValue(fmt.Sprintf("%q", http.StatusText(code))),
				},
			},
			{
				Name: "code",
				Schema: &ogen.Schema{
					Type:        "integer",
					Description: "The HTTP status code or other internal application error code.",
					Example:     jsonschema.RawValue(strconv.Itoa(code)),
				},
			},
			{
				Name: "request_id",
				Schema: &ogen.Schema{
					Type:        "string",
					Description: "The unique request ID for this error.",
					Example:     jsonschema.RawValue(`"cb6f6f9c1783cdc9752cee2a4e95dd4c"`),
				},
			},
			{
				Name: "timestamp",
				Schema: &ogen.Schema{
					Type:        "string",
					Format:      "date-time",
					Description: "The timestamp of the error, in RFC3339 format.",
					Example:     jsonschema.RawValue(`"2024-04-26T12:19:01Z"`),
				},
			},
		},
		Required: []string{"error", "type", "code", "timestamp"},
	}
}

// GetOperationIDName returns the operation ID for the given operation, type, and optional
// edge, or the OperationID provided by the annotation if it exists.
func GetOperationIDName(op Operation, t *gen.Type, e *gen.Edge) string {
	if t == nil {
		panic("provided type is nil")
	}

	if e != nil {
		if id := GetAnnotation(e).GetOperationID(op); id != "" {
			return id
		}

		switch op {
		case OperationRead:
			return "get" + Singularize(t.Name) + Singularize(PascalCase(e.Name))
		case OperationList:
			return "list" + Singularize(t.Name) + Pluralize(PascalCase(e.Name))
		default:
			panic(fmt.Sprintf("unsupported operation %q", op))
		}
	}

	if id := GetAnnotation(t).GetOperationID(op); id != "" {
		return id
	}

	switch op {
	case OperationCreate:
		return "create" + Singularize(t.Name)
	case OperationUpdate:
		return "update" + Singularize(t.Name)
	case OperationRead:
		return "get" + Singularize(t.Name)
	case OperationList:
		return "list" + Pluralize(t.Name)
	case OperationDelete:
		return "delete" + Singularize(t.Name)
	default:
		panic(fmt.Sprintf("unsupported operation %q", op))
	}
}

// GetPathName returns the path name for the given operation, type, and optional edge,
// or the OperationID provided by the annotation if it exists. useUniqueID determines
// if the ID path parameter should be "{id}" or "{type|camel}ID".
func GetPathName(op Operation, t *gen.Type, e *gen.Edge, useUniqueID bool) string {
	id := "{id}"
	if useUniqueID {
		id = "{" + CamelCase(Singularize(t.Name)) + "ID}"
	}

	if e != nil {
		switch op {
		case OperationRead, OperationList:
			return "/" + Pluralize(KebabCase(t.Name)) + "/" + id + "/" + KebabCase(e.Name)
		default:
			panic(fmt.Sprintf("unsupported operation %q", op))
		}
	}

	switch op {
	case OperationRead, OperationUpdate, OperationDelete:
		return "/" + Pluralize(KebabCase(t.Name)) + "/" + id
	case OperationCreate, OperationList:
		return "/" + Pluralize(KebabCase(t.Name))
	default:
		panic(fmt.Sprintf("unsupported operation %q", op))
	}
}

// PatchOperations applies a callback to each operation in a path inside of the OpenAPI spec.
func PatchOperations(pathItem *ogen.PathItem, cb func(method string, op *ogen.Operation) *ogen.Operation) *ogen.PathItem {
	pathItem.Get = cb(http.MethodGet, pathItem.Get)
	pathItem.Post = cb(http.MethodPost, pathItem.Post)
	pathItem.Put = cb(http.MethodPut, pathItem.Put)
	pathItem.Patch = cb(http.MethodPatch, pathItem.Patch)
	pathItem.Delete = cb(http.MethodDelete, pathItem.Delete)
	pathItem.Options = cb(http.MethodOptions, pathItem.Options)
	pathItem.Head = cb(http.MethodHead, pathItem.Head)
	pathItem.Trace = cb(http.MethodTrace, pathItem.Trace)
	return pathItem
}

// PatchPathItem applies a callback to each response in a path inside of the OpenAPI spec.
func PatchPathItem(pathItem *ogen.PathItem, cb func(resp *ogen.Response) *ogen.Response) *ogen.PathItem {
	return PatchOperations(pathItem, func(_ string, op *ogen.Operation) *ogen.Operation {
		if op == nil {
			return nil
		}
		for k, v := range op.Responses {
			op.Responses[k] = cb(v)
		}
		return op
	})
}

func mergeOperation(overlap bool, orig, op *ogen.Operation) (*ogen.Operation, error) {
	if orig == nil {
		return op, nil
	}
	if op == nil {
		return orig, nil
	}

	orig.Tags = appendCompact(orig.Tags, op.Tags)
	orig.Summary = cmp.Or(op.Summary, orig.Summary)
	orig.Description = cmp.Or(op.Description, orig.Description)
	orig.OperationID = cmp.Or(op.OperationID, orig.OperationID)
	orig.Deprecated = orig.Deprecated || op.Deprecated

	// Merge parameters.
	orig.Parameters = appendCompactFunc(orig.Parameters, op.Parameters, func(oldParam, newParam *ogen.Parameter) bool {
		return oldParam.Name == newParam.Name
	})

	if orig.Responses == nil && op.Responses != nil {
		orig.Responses = map[string]*ogen.Response{}
	}
	if op.Responses != nil {
		err := mergeMap(overlap, orig.Responses, op.Responses)
		if err != nil {
			return nil, err
		}
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

		orig.Servers = appendCompactFunc(orig.Servers, spec.Servers, func(oldServer, newServer ogen.Server) bool {
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

			orig.Paths[k].Servers = appendCompactFunc(orig.Paths[k].Servers, v.Servers, func(oldServer, newServer ogen.Server) bool {
				return oldServer.URL == newServer.URL
			})

			orig.Paths[k].Parameters = appendCompactFunc(orig.Paths[k].Parameters, v.Parameters, func(oldParam, newParam *ogen.Parameter) bool {
				return oldParam.Name == newParam.Name
			})
		}

		if orig.Components == nil && spec.Components != nil {
			orig.Components = &ogen.Components{}
		}

		if spec.Components != nil {
			if spec.Components.Schemas != nil {
				if orig.Components.Schemas == nil {
					orig.Components.Schemas = map[string]*ogen.Schema{}
				}

				err = mergeMap(overlap, orig.Components.Schemas, spec.Components.Schemas)
				if err != nil {
					return err
				}
			}

			if spec.Components.Responses != nil {
				if orig.Components.Responses == nil {
					orig.Components.Responses = map[string]*ogen.Response{}
				}
				err = mergeMap(overlap, orig.Components.Responses, spec.Components.Responses)
				if err != nil {
					return err
				}
			}

			if spec.Components.Parameters != nil {
				if orig.Components.Parameters == nil {
					orig.Components.Parameters = map[string]*ogen.Parameter{}
				}
				err = mergeMap(overlap, orig.Components.Parameters, spec.Components.Parameters)
				if err != nil {
					return err
				}
			}

			if spec.Components.RequestBodies != nil {
				if orig.Components.RequestBodies == nil {
					orig.Components.RequestBodies = map[string]*ogen.RequestBody{}
				}
				err = mergeMap(overlap, orig.Components.RequestBodies, spec.Components.RequestBodies)
				if err != nil {
					return err
				}
			}

			if spec.Components.Headers != nil {
				if orig.Components.Headers == nil {
					orig.Components.Headers = map[string]*ogen.Header{}
				}
				err = mergeMap(overlap, orig.Components.Headers, spec.Components.Headers)
				if err != nil {
					return err
				}
			}

			if spec.Components.SecuritySchemes != nil {
				if orig.Components.SecuritySchemes == nil {
					orig.Components.SecuritySchemes = map[string]*ogen.SecurityScheme{}
				}
				err = mergeMap(overlap, orig.Components.SecuritySchemes, spec.Components.SecuritySchemes)
				if err != nil {
					return err
				}
			}

			if spec.Components.PathItems != nil {
				if orig.Components.PathItems == nil {
					orig.Components.PathItems = map[string]*ogen.PathItem{}
				}
				err = mergeMap(overlap, orig.Components.PathItems, spec.Components.PathItems)
				if err != nil {
					return err
				}
			}
		}

		orig.Tags = appendCompactFunc(orig.Tags, spec.Tags, func(oldTag, newTag ogen.Tag) bool {
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

// addOpenAPIEndpoint adds an endpoint to the OpenAPI spec that returns the OpenAPI
// spec itself, as JSON.
func addOpenAPIEndpoint(path string) *ogen.Spec {
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
						SetJSONContent(SchemaObjectAny),
				}),
		),
	)
}
