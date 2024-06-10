// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

//nolint:goconst
package entrest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"entgo.io/ent/entc/gen"
	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/jsonschema"
)

const eagerLoadDepthMessage = "If the entity has eager-loaded edges, the depth of when those will be loaded is limited to a depth of 1 (entity -> edge, not entity -> edge -> edge -> etc)."

func addPagination(spec *ogen.Spec, cfg *Config) {
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
				},
			},
			{
				Name: "last_page",
				Schema: &ogen.Schema{
					Type:        "integer",
					Description: "The number of the last page of results.",
					Example:     jsonschema.RawValue(`3`),
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
		},
		Required: []string{"page", "last_page", "is_last_page"},
	}

	if !cfg.DisableTotalCount {
		pagedSchema.Properties = append(pagedSchema.Properties, ogen.Property{
			Name: "total_count",
			Schema: &ogen.Schema{
				Type:        "integer",
				Description: "Total number of results.",
				Example:     jsonschema.RawValue(`100`),
			},
		})
		pagedSchema.Required = append(pagedSchema.Required, "total_count")
	}

	spec.Components.Schemas["PagedResponse"] = pagedSchema
}

func newBaseSpec(cfg *Config) *ogen.Spec {
	spec := &ogen.Spec{
		Paths: ogen.Paths{},
		Components: &ogen.Components{
			Schemas: map[string]*ogen.Schema{},
			Parameters: map[string]*ogen.Parameter{
				"PrettyResponse": {
					Name:        "pretty",
					In:          "query",
					Description: "If set to true, any JSON response will be indented. Not recommended for best performance.",
					Schema:      ogen.Bool(),
				},
				"SortOrder": {
					Name:        "order",
					In:          "query",
					Description: "Order the results in ascending or descending order.",
					Schema: &ogen.Schema{
						Type:    "string",
						Enum:    sliceToRawMessage([]string{"asc", "desc"}),
						Default: jsonschema.RawValue(`"desc"`),
					},
				},
				"FilterOperation": {
					Name:        "filter_op",
					In:          "query",
					Description: "Filter operation to use.",
					Schema: &ogen.Schema{
						Type:    "string",
						Enum:    sliceToRawMessage([]string{"and", "or"}),
						Default: jsonschema.RawValue(`"and"`),
					},
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
	root := "/" + Pluralize(KebabCase(t.Name))

	spec := newBaseSpec(cfg)
	spec.Tags = append(spec.Tags, ogen.Tag{
		Name:        entityName,
		Description: ta.Description,
	})

	idSchema, err := GetSchemaField(t.ID)
	if err != nil {
		return nil, err
	}

	if op != OperationList {
		spec.Components.Parameters[entityName+"ID"] = &ogen.Parameter{
			Name:        "id",
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
			Tags: []string{entityName},
			Summary: withDefault(
				ta.GetOperationSummary(op),
				fmt.Sprintf("Create a new %s entity", entityName),
			),
			Description: withDefault(
				ta.GetOperationDescription(op),
				fmt.Sprintf("Create a new %s entity. %s", entityName, eagerLoadDepthMessage),
			),
			OperationID: withDefault(ta.GetOperationID(op), fmt.Sprintf("create%s", entityName)),
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

		spec.Paths[root] = &ogen.PathItem{
			Summary:     oper.Summary,
			Description: oper.Description,
			Post:        oper,
			Parameters: []*ogen.Parameter{
				{Ref: "#/components/parameters/PrettyResponse"},
			},
		}
	case OperationUpdate:
		oper := &ogen.Operation{
			Tags: []string{entityName},
			Summary: withDefault(
				ta.GetOperationSummary(op),
				fmt.Sprintf("Update an existing %s entity", entityName),
			),
			Description: withDefault(
				ta.GetOperationDescription(op),
				fmt.Sprintf("Update an existing %s entity. %s", entityName, eagerLoadDepthMessage),
			),
			OperationID: withDefault(ta.GetOperationID(op), fmt.Sprintf("update%sByID", entityName)),
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

		spec.Paths[root+"/{id}"] = &ogen.PathItem{
			Summary:     fmt.Sprintf("Operate on a single %s entity", entityName),
			Description: fmt.Sprintf("Operate on a single %s entity by its ID.", entityName),
			Patch:       oper,
			Parameters: []*ogen.Parameter{
				{Ref: "#/components/parameters/PrettyResponse"},
				{Ref: "#/components/parameters/" + entityName + "ID"},
			},
		}
	case OperationRead:
		oper := &ogen.Operation{
			Tags: []string{entityName},
			Summary: withDefault(
				ta.GetOperationSummary(op),
				fmt.Sprintf("Retrieve a single %s entity", entityName),
			),
			Description: withDefault(
				ta.GetOperationDescription(op),
				fmt.Sprintf("Retrieve a single %s entity by its ID. %s", entityName, eagerLoadDepthMessage),
			),
			OperationID: withDefault(ta.GetOperationID(op), fmt.Sprintf("get%sByID", entityName)),
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

		spec.Paths[root+"/{id}"] = &ogen.PathItem{
			Summary:     fmt.Sprintf("Operate on a single %s entity", entityName),
			Description: fmt.Sprintf("Operate on a single %s entity by its ID.", entityName),
			Get:         oper,
			Parameters: []*ogen.Parameter{
				{Ref: "#/components/parameters/PrettyResponse"},
				{Ref: "#/components/parameters/" + entityName + "ID"},
			},
		}
	case OperationList:
		oper := &ogen.Operation{
			Tags: []string{entityName},
			Summary: withDefault(
				ta.GetOperationSummary(op),
				fmt.Sprintf("Query all %s entities", entityName),
			),
			Description: withDefault(
				ta.GetOperationDescription(op),
				fmt.Sprintf("Query all %s entities (including pagination, filtering, sorting, etc). %s", entityName, eagerLoadDepthMessage),
			),
			OperationID: withDefault(ta.GetOperationID(op), fmt.Sprintf("list%s", Pluralize(t.Name))),
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
					Name:        "itemsPerPage",
					In:          "query",
					Description: "The number of entities to retrieve per page.",
					Schema: ogen.Int().
						SetMinimum(ptr(int64(ta.GetMinItemsPerPage(cfg)))).
						SetMaximum(ptr(int64(ta.GetMaxItemsPerPage(cfg)))).
						SetDefault(json.RawMessage(strconv.Itoa(ta.GetItemsPerPage(cfg)))),
				},
			)
		}

		// Greater than 1 because we want to sort by id by default.
		if sortable := GetSortableFields(t, false); len(sortable) > 1 {
			oper.Parameters = append(
				oper.Parameters,
				&ogen.Parameter{
					Name:        "sort",
					In:          "query",
					Description: "Sort entity results by the given field.",
					Schema: &ogen.Schema{
						Type:    "string",
						Enum:    sliceToRawMessage(sortable),
						Default: jsonschema.RawValue(`"id"`),
					},
				},
				&ogen.Parameter{Ref: "#/components/parameters/SortOrder"},
			)
		}

		if filters := GetFilterableFields(t, nil); len(filters) > 0 {
			oper.Parameters = append(oper.Parameters, &ogen.Parameter{Ref: "#/components/parameters/FilterOperation"})

			for _, k := range mapKeys(filters) {
				spec.Components.Parameters[k] = filters[k]
				oper.Parameters = append(oper.Parameters, &ogen.Parameter{Ref: "#/components/parameters/" + k})
			}
		}

		if cfg.AddEdgesToTags {
			oper.Tags = append(oper.Tags, edgesToTags(cfg, t)...)
		}

		spec.Paths[root] = &ogen.PathItem{
			Summary:     oper.Summary,
			Description: oper.Description,
			Get:         oper,
			Parameters: []*ogen.Parameter{
				{Ref: "#/components/parameters/PrettyResponse"},
			},
		}
	case OperationDelete:
		oper := &ogen.Operation{
			Tags: []string{entityName},
			Summary: withDefault(
				ta.GetOperationSummary(op),
				fmt.Sprintf("Delete a single %s entity", entityName),
			),
			Description: withDefault(
				ta.GetOperationDescription(op),
				fmt.Sprintf("Delete a single %s entity by its ID.", entityName),
			),
			OperationID: withDefault(ta.GetOperationID(op), fmt.Sprintf("delete%sByID", entityName)),
			Deprecated:  ta.Deprecated,
			Parameters:  []*ogen.Parameter{},
			Responses: ogen.Responses{
				strconv.Itoa(http.StatusNoContent): ogen.NewResponse().
					SetDescription(fmt.Sprintf("The requested %s entity.", entityName)),
			},
		}

		spec.Paths[root+"/{id}"] = &ogen.PathItem{
			Summary:     fmt.Sprintf("Operate on a single %s entity", entityName),
			Description: fmt.Sprintf("Operate on a single %s entity by its ID.", entityName),
			Delete:      oper,
			Parameters: []*ogen.Parameter{
				{Ref: "#/components/parameters/" + entityName + "ID"},
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

	if ea.Skip {
		return nil, errors.New("edge is skipped")
	}

	if !ea.GetEdgeEndpoint(cfg) {
		return nil, errors.New("edge has endpoint disabled or edge is eager-loaded with global config to disable endpoints for edges which are also eager-loaded")
	}

	rootEntityName := Singularize(t.Name)
	refEntityName := Singularize(e.Type.Name)
	entityName := Singularize(PascalCase(e.Name))
	root := "/" + Pluralize(KebabCase(t.Name)) + "/{id}/" + KebabCase(e.Name)

	spec := newBaseSpec(cfg)
	spec.Tags = append(
		spec.Tags,
		ogen.Tag{
			Name:        rootEntityName,
			Description: ta.Description,
		},
		ogen.Tag{
			Name:        refEntityName,
			Description: ra.Description,
		},
	)

	idSchema, err := GetSchemaField(t.ID)
	if err != nil {
		return nil, err
	}

	spec.Components.Parameters[rootEntityName+"ID"] = &ogen.Parameter{
		Name:        "id",
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
			Tags: []string{rootEntityName, refEntityName},
			Summary: withDefault(
				ea.GetOperationSummary(op),
				e.Comment(),
				fmt.Sprintf("Get a %s associated %s", Pluralize(CamelCase(t.Name)), CamelCase(e.Name)),
			),
			Description: withDefault(
				ea.GetOperationDescription(op),
				fmt.Sprintf(
					"Get a %s associated %s (%s entity type). %s",
					Pluralize(CamelCase(t.Name)),
					CamelCase(e.Name),
					refEntityName,
					eagerLoadDepthMessage,
				),
			),
			OperationID: withDefault(ea.GetOperationID(op), fmt.Sprintf("get%s%sByID", rootEntityName, entityName)),
			Deprecated:  ta.Deprecated || ea.Deprecated || ra.Deprecated,
			Parameters:  []*ogen.Parameter{},
			Responses: ogen.Responses{
				strconv.Itoa(http.StatusOK): ogen.NewResponse().
					SetDescription(fmt.Sprintf("The requested %s entity.", CamelCase(e.Name))).
					SetJSONContent(&ogen.Schema{Ref: "#/components/schemas/" + refEntityName + "Read"}),
			},
		}

		spec.Paths[root] = &ogen.PathItem{
			Summary:     oper.Summary,     // Will probably always be the same.
			Description: oper.Description, // Will probably always be the same.
			Get:         oper,
			Parameters: []*ogen.Parameter{
				{Ref: "#/components/parameters/PrettyResponse"},
				{Ref: "#/components/parameters/" + rootEntityName + "ID"},
			},
		}
	case OperationList: // Not unique.
		if e.Unique {
			return nil, errors.New("edge is unique")
		}

		oper := &ogen.Operation{
			Tags: []string{rootEntityName, refEntityName},
			Summary: withDefault(
				ea.GetOperationSummary(op),
				e.Comment(),
				fmt.Sprintf("List a %s associated %s", Pluralize(CamelCase(t.Name)), Pluralize(CamelCase(e.Name))),
			),
			Description: withDefault(
				ea.GetOperationDescription(op),
				fmt.Sprintf(
					"List a %s associated %s (%s entity type). %s",
					Pluralize(CamelCase(t.Name)),
					Pluralize(CamelCase(e.Name)),
					refEntityName,
					eagerLoadDepthMessage,
				),
			),
			OperationID: withDefault(ea.GetOperationID(op), fmt.Sprintf("list%s%s", rootEntityName, entityName)),
			Deprecated:  ta.Deprecated || ea.Deprecated || ra.Deprecated,
			Parameters:  []*ogen.Parameter{},
			Responses: ogen.Responses{
				strconv.Itoa(http.StatusOK): ogen.NewResponse().
					SetDescription(fmt.Sprintf("The requested %s.", Pluralize(CamelCase(e.Name)))).
					SetJSONContent(ogen.NewSchema().SetRef("#/components/schemas/" + refEntityName + "List")),
			},
		}

		if cfg.DisableEagerLoadNonPagedOpt || ea.GetPagination(cfg, e) || ra.GetPagination(cfg, e) {
			addPagination(spec, cfg)
			oper.Parameters = append(oper.Parameters,
				&ogen.Parameter{Ref: "#/components/parameters/Page"},
				&ogen.Parameter{
					Name:        "itemsPerPage",
					In:          "query",
					Description: "The number of entities to retrieve per page.",
					Schema: ogen.Int().
						SetMinimum(ptr(int64(ta.GetMinItemsPerPage(cfg)))).
						SetMaximum(ptr(int64(ta.GetMaxItemsPerPage(cfg)))).
						SetDefault(json.RawMessage(strconv.Itoa(ta.GetItemsPerPage(cfg)))),
				},
			)
		} else {
			code := strconv.Itoa(http.StatusOK)
			oper.Responses[code] = oper.Responses[code].SetJSONContent(&ogen.Schema{
				Ref: "#/components/schemas/" + rootEntityName + entityName + "List",
			})
		}

		// Greater than 1 because we want to sort by id by default.
		if sortable := GetSortableFields(e.Type, false); len(sortable) > 1 {
			oper.Parameters = append(
				oper.Parameters,
				&ogen.Parameter{
					Name:        "sort",
					In:          "query",
					Description: "Sort entity results by the given field.",
					Schema: &ogen.Schema{
						Type:    "string",
						Enum:    sliceToRawMessage(sortable),
						Default: jsonschema.RawValue(`"id"`),
					},
				},
				&ogen.Parameter{Ref: "#/components/parameters/SortOrder"},
			)
		}

		if filters := GetFilterableFields(e.Type, nil); len(filters) > 0 {
			oper.Parameters = append(oper.Parameters, &ogen.Parameter{Ref: "#/components/parameters/FilterOperation"})

			for _, k := range mapKeys(filters) {
				spec.Components.Parameters[k] = filters[k]
				oper.Parameters = append(oper.Parameters, &ogen.Parameter{Ref: "#/components/parameters/" + k})
			}
		}

		if cfg.AddEdgesToTags {
			oper.Tags = append(oper.Tags, edgesToTags(cfg, e.Type)...)
		}

		spec.Paths[root] = &ogen.PathItem{
			Summary:     oper.Summary,
			Description: oper.Description,
			Get:         oper,
			Parameters: []*ogen.Parameter{
				{Ref: "#/components/parameters/PrettyResponse"},
				{Ref: "#/components/parameters/" + rootEntityName + "ID"},
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
		if !ea.Skip && ea.GetEagerLoad(cfg) {
			tags = append(tags, Singularize(e.Name))
		}
	}
	return tags
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
func addGlobalErrorResponses(spec *ogen.Spec, responses map[int]*ogen.Schema) {
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
				case strings.HasPrefix(op.OperationID, "list") && k == http.StatusNotFound:
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

// defaultErrorResponse returns a default error schema for the provided HTTP
// status code.
func defaultErrorResponse(code int) *ogen.Schema {
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
