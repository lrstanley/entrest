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

	"entgo.io/ent/entc/gen"
	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/jsonschema"
)

const eagerLoadDepthMessage = "If the entity has eager-loaded edges, the depth of when those will be loaded is limited to a depth of 1 (entity -> edge, not entity -> edge -> edge -> etc)."

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

	if !cfg.DisablePagination {
		spec.Components.Parameters["Page"] = &ogen.Parameter{
			Name:        "page",
			In:          "query",
			Description: "The page number to retrieve.",
			Schema: ogen.Int().
				SetMinimum(ptr(int64(1))).
				SetDefault(json.RawMessage(`1`)),
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

	return spec
}

func GenSpecType(t *gen.Type, op Operation) (*ogen.Spec, error) { // nolint:funlen,gocyclo,cyclop
	cfg := GetConfig(t.Config)
	ta := GetAnnotation(t)

	entityName := Singularize(t.Name)
	root := "/" + Pluralize(KebabCase(t.Name))

	spec := newBaseSpec(cfg)
	spec.Tags = append(spec.Tags, ogen.Tag{
		Name:        entityName,
		Description: ta.Description,
	})

	idSchema, err := GenSchemaField(t.ID)
	if err != nil {
		return nil, err
	}

	for k, v := range GenSchemaType(t, op) {
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
				{
					Name:        "id",
					In:          "path",
					Description: fmt.Sprintf("The ID of the %s to act upon.", entityName),
					Required:    true,
					Schema:      idSchema,
				},
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
				{
					Name:        "id",
					In:          "path",
					Description: fmt.Sprintf("The ID of the %s to act upon.", entityName),
					Required:    true,
					Schema:      idSchema,
				},
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
			Parameters: []*ogen.Parameter{
				{Ref: "#/components/parameters/Page"},
				{
					Name:        "itemsPerPage",
					In:          "query",
					Description: "The number of entities to retrieve per page.",
					Schema: ogen.Int().
						SetMinimum(ptr(int64(ta.GetMinItemsPerPage(cfg)))).
						SetMaximum(ptr(int64(ta.GetMaxItemsPerPage(cfg)))).
						SetDefault(json.RawMessage(strconv.Itoa(ta.GetItemsPerPage(cfg)))),
				},
			},
			Responses: ogen.Responses{
				strconv.Itoa(http.StatusOK): ogen.NewResponse().
					SetDescription(fmt.Sprintf("The requested %s.", entityName)).
					SetJSONContent(ogen.NewSchema().SetRef("#/components/schemas/" + entityName + "List")),
			},
		}

		// Greater than 1 because we want to sort by id by default.
		if sortable := GetSortableFields(t); len(sortable) > 1 {
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
			oper.Parameters = append(oper.Parameters, filters...)
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
				{
					Name:        "id",
					In:          "path",
					Description: fmt.Sprintf("The ID of the %s to act upon.", entityName),
					Required:    true,
					Schema:      idSchema,
				},
			},
		}
	default:
		panic(fmt.Sprintf("unsupported operation %q", op))
	}

	return spec, nil
}

func GenSpecEdge(t *gen.Type, e *gen.Edge, op Operation) (*ogen.Spec, error) {
	cfg := GetConfig(t.Config)
	ta := GetAnnotation(t)
	ea := GetAnnotation(e)
	ra := GetAnnotation(e.Type)

	if ea.Skip {
		return nil, errors.New("edge is skipped")
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

	idSchema, err := GenSchemaField(t.ID)
	if err != nil {
		return nil, err
	}

	for k, v := range GenSchemaType(t, op) {
		spec.Components.Schemas[k] = v
	}
	for k, v := range GenSchemaType(e.Type, op) {
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
				{
					Name:        "id",
					In:          "path",
					Description: fmt.Sprintf("The ID of the %s entity which this edge is attached to.", rootEntityName),
					Required:    true,
					Schema:      idSchema,
				},
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
			Parameters: []*ogen.Parameter{
				{Ref: "#/components/parameters/Page"},
				{
					Name:        "itemsPerPage",
					In:          "query",
					Description: "The number of entities to retrieve per page.",
					Schema: ogen.Int().
						SetMinimum(ptr(int64(ta.GetMinItemsPerPage(cfg)))).
						SetMaximum(ptr(int64(ta.GetMaxItemsPerPage(cfg)))).
						SetDefault(json.RawMessage(strconv.Itoa(ta.GetItemsPerPage(cfg)))),
				},
			},
			Responses: ogen.Responses{
				strconv.Itoa(http.StatusOK): ogen.NewResponse().
					SetDescription(fmt.Sprintf("The requested %s.", Pluralize(CamelCase(e.Name)))).
					SetJSONContent(ogen.NewSchema().SetRef("#/components/schemas/" + refEntityName + "Read")),
			},
		}

		// Greater than 1 because we want to sort by id by default.
		if sortable := GetSortableFields(e.Type); len(sortable) > 1 {
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
			oper.Parameters = append(oper.Parameters, filters...)
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
			},
		}
	default:
		panic(fmt.Sprintf("unsupported operation %q", op))
	}

	return spec, nil
}

func edgesToTags(cfg *Config, t *gen.Type) (tags []string) {
	for _, e := range t.Edges {
		ea := GetAnnotation(e)
		if !ea.Skip && ea.GetEagerLoad(cfg) {
			tags = append(tags, Singularize(e.Name))
		}
	}
	return tags
}
