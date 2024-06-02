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

func GenSchemaField(f *gen.Field) (*ogen.Schema, error) {
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

// GenSchematype returns a map of ogen.Schemas for the given gen.Type. Multiple may be
// returned if the type has multiple schemas (e.g. a list of entities, or an entity which
// has edges). Note that depending on the operation, this schema may be for the request or
// response, or both. Recursive should always be true for the first call.
func GenSchemaType(t *gen.Type, op Operation) map[string]*ogen.Schema { // nolint:funlen,gocyclo,cyclop
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
			fieldSchema, err = GenSchemaField(t.ID)
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
				fieldSchema, err = GenSchemaField(f)
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

			fieldSchema, err = GenSchemaField(e.Type.ID)
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

		fieldSchema, err = GenSchemaField(t.ID)
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

			fieldSchema, err = GenSchemaField(f)
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

			if t.IsEdgeSchema() {
				for k, v := range GenSchemaType(e.Type, OperationRead) {
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
		schema := ogen.NewSchema().SetRef("#/components/schemas/" + entityName + "Read").AsArray()

		if !ta.GetPagination(cfg) {
			schema.Description = fmt.Sprintf("A list of %s entities. Includes eager-loaded edges (if any) for each entity.", entityName)
			schemas[entityName+"List"] = schema
			return schemas
		}

		schemas[entityName+"List"] = &ogen.Schema{
			Description: fmt.Sprintf("A paginated result set of %s entities. Includes eager-loaded edges (if any) for each entity.", entityName),
			AllOf: []*ogen.Schema{
				{Ref: "#/components/schemas/PagedResponse"},
				{
					Type:       "object",
					Properties: ogen.Properties{{Name: "content", Schema: schema}},
					Required:   []string{"content"},
				},
			},
		}

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
		for k, v := range GenSchemaType(t, oper) {
			schemas[k] = v
		}
	}

	return schemas
}

// GetSortableFields returnsd a list of sortable fields for the given type. It
// recurses through edges to find sortable fields as well.
func GetSortableFields(t *gen.Type) (sortable []string) {
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

	if !t.IsEdgeSchema() {
		for _, e := range t.Edges {
			ea := GetAnnotation(e)

			if !e.Unique || ea.Skip {
				continue
			}

			for _, f := range GetSortableFields(e.Type) {
				sortable = append(sortable, fmt.Sprintf("%s.%s", e.Name, f))
			}
		}
	}

	return sortable
}

func GetFilterableFields(t *gen.Type, edge *gen.Edge) (filters []*ogen.Parameter) {
	ta := GetAnnotation(t.ID)

	if ta.Skip {
		return nil
	}

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

			fieldSchema, err := GenSchemaField(f)
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
			} else {
				param.Name = CamelCase(f.Name) + "." + predicateFormat(op)
			}

			filters = append(filters, param)
		}
	}

	if edge == nil {
		for _, e := range t.Edges {
			ea := GetAnnotation(e)

			if ea.Skip || ea.Filter == 0 || ea.Filter != FilterEdge {
				continue
			}

			entityName := Singularize(e.Name)

			filters = append(filters, &ogen.Parameter{
				Name:        "has." + CamelCase(SnakeCase(entityName)),
				In:          "query",
				Description: fmt.Sprintf("If true, only return entities that have an %s edge.", entityName),
				Schema:      &ogen.Schema{Type: "boolean"},
			})

			filters = append(filters, GetFilterableFields(e.Type, e)...)
		}
	}

	slices.SortFunc(filters, func(a, b *ogen.Parameter) int {
		if a.Name < b.Name {
			return -1
		} else if a.Name > b.Name {
			return 1
		}
		return 0
	})

	return filters
}
