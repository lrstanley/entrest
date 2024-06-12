// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

//nolint:goconst
package entrest

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strings"

	"entgo.io/ent/entc/gen"
	"github.com/ogen-go/ogen"
)

// mapTypeToSchema returns an ogen.Schema for the given gen.Field, if it exists.
// returns nil if the type is not supported.
func mapTypeToSchema(baseType string) *ogen.Schema {
	switch baseType {
	case "bool":
		return ogen.Bool()
	case "time.Time":
		return ogen.DateTime()
	case "string":
		return ogen.String()
	case "[]byte":
		return ogen.Bytes()
	case "uuid.UUID":
		return ogen.UUID()
	case "int":
		return ogen.Int()
	case "int8":
		return ogen.Int32().SetMinimum(ptr(int64(math.MinInt8))).SetMaximum(ptr(int64(math.MaxInt8)))
	case "int16":
		return ogen.Int32().SetMinimum(ptr(int64(math.MinInt16))).SetMaximum(ptr(int64(math.MaxInt16)))
	case "int32":
		return ogen.Int32().SetMinimum(ptr(int64(math.MinInt32))).SetMaximum(ptr(int64(math.MaxInt32)))
	case "int64":
		return ogen.Int64().SetMinimum(ptr(int64(math.MinInt64))).SetMaximum(ptr(int64(math.MaxInt64)))
	case "uint":
		return ogen.Int64().SetMinimum(ptr(int64(0))).SetMaximum(ptr(int64(math.MaxUint32)))
	case "uint8":
		return ogen.Int32().SetMinimum(ptr(int64(0))).SetMaximum(ptr(int64(math.MaxUint8)))
	case "uint16":
		return ogen.Int32().SetMinimum(ptr(int64(0))).SetMaximum(ptr(int64(math.MaxUint16)))
	case "uint32":
		return ogen.Int64().SetMinimum(ptr(int64(0))).SetMaximum(ptr(int64(math.MaxUint32)))
	case "uint64":
		return ogen.Int64().SetMinimum(ptr(int64(0)))
	case "float32":
		return ogen.Float()
	case "float64":
		return ogen.Double()
	default:
		return nil
	}
}

// GetSchemaField generates a schema for the given field, if its supported. If the
// field you have provided is not supported, use the [WithSchema] annotation on the
// field to provide a custom schema (primarily beneficial for JSON fields).
func GetSchemaField(f *gen.Field) (*ogen.Schema, error) {
	fa := GetAnnotation(f)

	var err error

	if fa.Schema != nil {
		return fa.Schema, nil
	}

	var schema *ogen.Schema
	baseType := f.Type.String()

	if f.IsEnum() {
		var d json.RawMessage
		if f.Default {
			d, err = json.Marshal(f.DefaultValue().(string))
			if err != nil {
				return nil, err
			}
		}
		values := make([]json.RawMessage, len(f.EnumValues()))
		for i, e := range f.EnumValues() {
			values[i], err = json.Marshal(e)
			if err != nil {
				return nil, err
			}
		}
		schema = ogen.String().AsEnum(d, values...)
	}

	if schema == nil {
		if strings.HasPrefix(baseType, "[]") {
			schema = mapTypeToSchema(baseType[2:])
			if schema != nil {
				schema = schema.AsArray()
			}
		}

		if schema == nil {
			schema = mapTypeToSchema(baseType)
		}
	}

	if schema == nil {
		return nil, fmt.Errorf("no openapi type exists for type %q of field %s", baseType, f.StructField())
	}

	if f.Nillable {
		schema.Nullable = true
		// TODO: when we switch to 3.1, we can use:
		// schema = ogen.NewSchema().SetOneOf([]*ogen.Schema{schema, {Type: "null"}})
	}

	if v := f.DefaultValue(); f.Default && v != nil {
		schema.Default, err = json.Marshal(f.DefaultValue())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal default value for field %s: %w", f.StructField(), err)
		}
	}

	schema.Description = withDefault(fa.Description, f.Comment())
	schema.Deprecated = fa.Deprecated

	if fa.Example != nil {
		schema.Example, err = json.Marshal(fa.Example)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal example for field %s: %w", f.StructField(), err)
		}
	}

	return schema, nil
}

// GetSchemaType returns a map of ogen.Schemas for the given gen.Type. Multiple may be
// returned if the type has multiple schemas (e.g. a list of entities, or an entity which
// has edges). Note that depending on the operation, this schema may be for the request or
// response, or both. Edge should be provided only if the type is from an edge schema.
func GetSchemaType(t *gen.Type, op Operation, edge *gen.Edge) map[string]*ogen.Schema { // nolint:funlen,gocyclo,cyclop
	cfg := GetConfig(t.Config)
	ta := GetAnnotation(t)

	schemas := map[string]*ogen.Schema{}
	var dependencies []Operation

	var err error

	entityName := Singularize(t.Name)

	switch op {
	case OperationCreate, OperationUpdate:
		schema := &ogen.Schema{
			Description: withDefault(
				ta.GetOperationDescription(op),
				ta.Description,
				fmt.Sprintf("A single %s entity and the fields that can be created/updated.", entityName),
			),
			Type:       "object",
			Properties: ogen.Properties{},
			Required:   []string{},
		}

		var fieldSchema *ogen.Schema

		if op == OperationCreate && cfg.AllowClientUUIDs && t.ID.IsUUID() {
			fieldSchema, err = GetSchemaField(t.ID)
			if err != nil {
				panic(fmt.Sprintf("failed to generate schema for field %s: %v", t.ID.StructField(), err))
			}
			fieldSchema.Description = fmt.Sprintf("The ID of the %s entity.", entityName)
			schema.Properties = append(schema.Properties, *fieldSchema.ToProperty("id"))

			if t.ID.Default {
				schema.Required = append(schema.Required, "id")
			}
		}

		for _, f := range t.Fields {
			fa := GetAnnotation(f)

			if fa.Skip || fa.ReadOnly {
				continue
			}

			if op == OperationCreate || !f.Immutable {
				fieldSchema, err = GetSchemaField(f)
				if err != nil {
					panic(fmt.Sprintf("failed to generate schema for field %s: %v", f.StructField(), err))
				}
				schema.Properties = append(schema.Properties, *fieldSchema.ToProperty(f.Name))

				if op == OperationCreate && !f.Optional && !f.Default {
					schema.Required = append(schema.Required, f.Name)
				}
			}
		}

		for _, e := range t.Edges {
			ea := GetAnnotation(e)

			if ea.Skip || ea.ReadOnly {
				continue
			}
			if op == OperationUpdate && (e.Immutable || (e.Field() != nil && e.Field().Immutable)) {
				continue
			}

			fieldSchema, err = GetSchemaField(e.Type.ID)
			if err != nil {
				panic(fmt.Sprintf("failed to generate schema for field %s: %v", e.Type.ID.StructField(), err))
			}

			if !e.Unique {
				fieldSchema = fieldSchema.AsArray()
			}

			if e.Field() != nil {
				fa := GetAnnotation(e.Field())

				if fa.ReadOnly {
					continue
				}

				if !fa.Skip {
					// If the edge has a field, and the field isn't skipped, then there is no
					// point in having two fields that can be used during create (especially
					// if both are required).
					continue
				}

				if (op == OperationCreate && !e.Optional && !e.Field().Default) || (op == OperationUpdate && !e.Field().UpdateDefault) {
					schema.Required = append(schema.Required, e.Name)
				}
			}

			if op == OperationUpdate && !e.Unique {
				// Handle adding the "add_<field>" and "remove_<field>" object properties for update operations.
				schema.Properties = append(
					schema.Properties,
					*fieldSchema.ToProperty("add_" + e.Name),
					*fieldSchema.ToProperty("remove_" + e.Name),
				)
			}

			schema.Properties = append(schema.Properties, *fieldSchema.ToProperty(e.Name))

			if !slices.Contains(schema.Required, e.Name) && op == OperationCreate && !e.Optional {
				schema.Required = append(schema.Required, e.Name)
			}
		}

		switch op {
		case OperationCreate:
			schemas[entityName+"Create"] = schema
		case OperationUpdate:
			schemas[entityName+"Update"] = schema
		default:
			panic("unreachable")
		}

		dependencies = append(dependencies, OperationRead)
	case OperationRead:
		schema := &ogen.Schema{
			Description: withDefault(ta.GetOperationDescription(op), ta.Description, fmt.Sprintf("A single %s entity.", entityName)),
			Type:        "object",
			Properties:  ogen.Properties{},
			Required:    []string{},
		}

		var fieldSchema *ogen.Schema

		fieldSchema, err = GetSchemaField(t.ID)
		if err != nil {
			panic(fmt.Sprintf("failed to generate schema for field %s: %v", t.ID.StructField(), err))
		}
		fieldSchema.Description = fmt.Sprintf("The ID of the %s entity.", entityName)
		schema.Properties = append(schema.Properties, *fieldSchema.ToProperty("id"))

		for _, f := range t.Fields {
			fa := GetAnnotation(f)

			if fa.Skip {
				continue
			}

			if !f.Optional {
				schema.Required = append(schema.Required, f.Name)
			}

			fieldSchema, err = GetSchemaField(f)
			if err != nil {
				panic(fmt.Sprintf("failed to generate schema for field %s: %v", f.StructField(), err))
			}

			schema.Properties = append(schema.Properties, *fieldSchema.ToProperty(f.Name))
		}

		edgeSchema := &ogen.Schema{
			Type:       "object",
			Properties: ogen.Properties{},
			Required:   []string{},
		}

		for _, e := range t.Edges {
			ea := GetAnnotation(e)

			if ea.Skip || !ea.GetEagerLoad(cfg) {
				continue
			}

			prop := ogen.Property{
				Name:   e.Name,
				Schema: &ogen.Schema{Ref: "#/components/schemas/" + Singularize(e.Type.Name)},
			}

			if !e.Unique {
				prop.Schema = prop.Schema.AsArray()
			}

			if !e.Optional {
				// TODO: nullable?
				// prop.Schema.Nullable = true
				edgeSchema.Required = append(edgeSchema.Required, e.Name)
			}

			edgeSchema.Properties = append(edgeSchema.Properties, prop)

			if edge == nil {
				for k, v := range GetSchemaType(e.Type, OperationRead, e) {
					schemas[k] = v
				}
			}
		}

		// Apply main schema.
		schemas[entityName] = schema

		if len(edgeSchema.Properties) > 0 {
			schemas[entityName+"Read"] = &ogen.Schema{
				Description: schema.Description,
				AllOf: []*ogen.Schema{
					{Ref: "#/components/schemas/" + entityName},
					{
						Type: "object",
						Properties: ogen.Properties{
							{Name: "edges", Schema: &ogen.Schema{Ref: "#/components/schemas/" + entityName + "Edges"}},
						},
						Required: []string{"edges"},
					},
				},
			}
			schemas[entityName+"Edges"] = edgeSchema
		} else {
			// No-op these references/shortcut them to the main schema.
			schemas[entityName+"Read"] = &ogen.Schema{Ref: "#/components/schemas/" + entityName}
		}
	case OperationList:
		if !cfg.DisableEagerLoadNonPagedOpt && edge != nil {
			// TODO: if /categories does not have pagination enabled, and then we have another
			// endpoint like /pets/{id}/categories, we could reuse the /categories schema,
			// because both are []Category. Would optimize the schema size, with the downside
			// that the description would be less accurate (can't describe the relationship).
			ea := GetAnnotation(edge)
			ra := GetAnnotation(edge.Type)

			if !ea.GetPagination(cfg, edge) || !ra.GetPagination(cfg, edge) {
				// This should allow setting the normal list operation as well, so don't return.
				schema := ogen.NewSchema().SetRef("#/components/schemas/" + Singularize(edge.Type.Name) + "Read").AsArray()
				schema.Description = fmt.Sprintf(
					"List of %s associated with %s (%s entity type).",
					Pluralize(CamelCase(edge.Name)),
					Pluralize(CamelCase(t.Name)),
					Singularize(CamelCase(edge.Type.Name)),
				)

				// If edge pagination is enabled, but edge type isn't paginated, we cannot re-use
				// the paginated schema from the edge type.
				if !ra.GetPagination(cfg, edge) && ea.GetPagination(cfg, edge) {
					schema = toPagedSchema(schema)
				}

				// We're setting a specific schema for the edge response because we cannot re-use
				// the paginated schema from the edge type (as it's paginated and the edge isn't).
				schemas[entityName+Singularize(PascalCase(edge.Name))+"List"] = schema
			}
		}

		if !ta.GetPagination(cfg, nil) {
			schema := ogen.NewSchema().SetRef("#/components/schemas/" + entityName + "Read").AsArray()
			schema.Description = fmt.Sprintf("A list of %s entities. Includes eager-loaded edges (if any) for each entity.", entityName)
			schemas[entityName+"List"] = schema
			return schemas
		}

		schemas[entityName+"List"] = toPagedSchema(
			ogen.NewSchema().
				SetRef("#/components/schemas/" + entityName + "Read").
				SetDescription(fmt.Sprintf("A paginated result set of %s entities. Includes eager-loaded edges (if any) for each entity.", entityName)),
		)

		dependencies = append(dependencies, OperationRead)
	case OperationDelete:
	default:
		panic(fmt.Sprintf("unsupported operation %q", op))
	}

	// If one operation depends on schemas from another, then we should recurse
	// and generates the schemas for that operation as well.
	for _, oper := range dependencies {
		if oper == op {
			continue
		}
		for k, v := range GetSchemaType(t, oper, nil) {
			schemas[k] = v
		}
	}

	return schemas
}

// toPagedSchema converts a response schema to a paged response schema, hoisting the
// description from the response schema to the paged response schema.
func toPagedSchema(schema *ogen.Schema) *ogen.Schema {
	desc := schema.Description
	schema.Description = ""

	if schema.Items == nil {
		schema = schema.AsArray()
	}

	return &ogen.Schema{
		Description: desc,
		AllOf: []*ogen.Schema{
			{Ref: "#/components/schemas/PagedResponse"},
			{
				Type: "object",
				Properties: ogen.Properties{{
					Name:   "content",
					Schema: schema,
				}},
				Required: []string{"content"},
			},
		},
	}
}

// GetSortableFields returnsd a list of sortable fields for the given type. It
// recurses through edges to find sortable fields as well.
func GetSortableFields(t *gen.Type, isEdge bool) (sortable []string) {
	fields := t.Fields

	if t.ID != nil {
		fields = append([]*gen.Field{t.ID}, fields...)
	}

	for _, f := range fields {
		fa := GetAnnotation(f)

		if fa.Skip || f.Sensitive() || (!fa.Sortable && f.Name != "id") {
			continue
		}

		if !f.IsString() && !f.IsTime() && !f.IsBool() && !f.IsInt() && !f.IsInt64() && !f.IsUUID() {
			continue
		}

		sortable = append(sortable, f.Name)
	}

	if !isEdge {
		for _, e := range t.Edges {
			ea := GetAnnotation(e)

			if !e.Unique || ea.Skip {
				continue
			}

			for _, f := range GetSortableFields(e.Type, true) {
				sortable = append(sortable, fmt.Sprintf("%s.%s", e.Name, f))
			}
		}
	}

	return sortable
}

// GetFilterableFields returns a list of filterable fields for the given type, where
// the key is the component name, and the value is the parameter which includes the
// name, description and schema for the parameter.
func GetFilterableFields(t *gen.Type, edge *gen.Edge) (filters map[string]*ogen.Parameter) {
	ta := GetAnnotation(t.ID)

	if ta.Skip {
		return nil
	}

	filters = map[string]*ogen.Parameter{}

	for _, f := range t.Fields {
		fa := GetAnnotation(f)

		if fa.Filter == 0 {
			continue
		}

		requestedOps := fa.Filter.Explode()

		for _, op := range f.Ops() {
			if !slices.Contains(requestedOps, op) {
				continue
			}

			if op == gen.NotNil {
				continue // Since you can use IsNil=false (we have three states for passed parameters).
			}

			if f.IsBool() && op == gen.NEQ {
				continue // Since you can use EQ=false instead.
			}

			fieldSchema, err := GetSchemaField(f)
			if err != nil {
				continue // Just skip things that can't be generated/easily mapped.
			}

			schema := &ogen.Schema{
				Type:       fieldSchema.Type,
				Items:      fieldSchema.Items,
				Minimum:    fieldSchema.Minimum,
				Maximum:    fieldSchema.Maximum,
				MinLength:  fieldSchema.MinLength,
				MaxLength:  fieldSchema.MaxLength,
				Enum:       fieldSchema.Enum,
				Deprecated: fieldSchema.Deprecated,
			}

			if op == gen.GT || op == gen.LT || op == gen.GTE || op == gen.LTE {
				if schema.Items != nil {
					schema.Items.Item.Type = "number"
				} else {
					schema.Type = "number"
				}
			}

			if op == gen.IsNil {
				if schema.Items != nil {
					schema.Items.Item.Type = "boolean"
				} else {
					schema.Type = "boolean"
				}
			}

			if op.Variadic() {
				schema = schema.AsArray() // If not already.
			}

			param := &ogen.Parameter{
				In:          "query",
				Description: predicateDescription(f, op),
				Schema:      schema,
			}

			if edge != nil {
				param.Name = CamelCase(SnakeCase(Singularize(edge.Name))) + "." + CamelCase(f.Name) + "." + predicateFormat(op)
				filters["Edge"+PascalCase(Singularize(edge.Name))+PascalCase(f.Name)+PascalCase(op.Name())] = param
			} else {
				param.Name = CamelCase(f.Name) + "." + predicateFormat(op)
				filters[PascalCase(t.Name)+PascalCase(f.Name)+PascalCase(op.Name())] = param
			}
		}
	}

	if edge == nil {
		for _, e := range t.Edges {
			ea := GetAnnotation(e)

			if ea.Skip || ea.Filter == 0 || ea.Filter != FilterEdge {
				continue
			}

			entityName := Singularize(e.Name)

			filters["EdgeHas"+PascalCase(Singularize(e.Name))] = &ogen.Parameter{
				Name:        "has." + CamelCase(SnakeCase(entityName)),
				In:          "query",
				Description: fmt.Sprintf("If true, only return entities that have a %s edge.", entityName),
				Schema:      &ogen.Schema{Type: "boolean"},
			}

			for k, v := range GetFilterableFields(e.Type, e) {
				filters[k] = v
			}
		}
	}

	return filters
}
