// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

//nolint:goconst
package entrest

import (
	"cmp"
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"slices"
	"strings"

	"entgo.io/ent/entc/gen"
	"entgo.io/ent/schema/field"
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
		// TODO: sharing enum schemas between parameters and component schemas,
		// means that the default is used for both, even if the parameter version
		// shouldn't have a default.
		// var d json.RawMessage
		// if f.Default {
		// 	d, err = json.Marshal(f.DefaultValue().(string))
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// }
		schema = &ogen.Schema{Type: "string"}
		schema.Enum, err = ToEnum(f.EnumValues())
		if err != nil {
			return nil, err
		}
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

	if v := f.DefaultValue(); f.Default && v != nil && !f.IsEnum() {
		schema.Default, err = json.Marshal(f.DefaultValue())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal default value for field %s: %w", f.StructField(), err)
		}
	}

	schema.Description = cmp.Or(fa.Description, f.Comment())
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
			Description: cmp.Or(
				ta.GetOperationDescription(op),
				ta.Description,
				fmt.Sprintf("A single %s entity and the fields that can be created/updated.", entityName),
			),
			Type:       "object",
			Properties: ogen.Properties{},
			Required:   []string{},
		}

		var fieldSchema *ogen.Schema

		if op == OperationCreate && cfg.AllowClientUUIDs && t.ID != nil && t.ID.IsUUID() {
			fieldSchema, err = GetSchemaField(t.ID)
			if err != nil {
				panic(fmt.Sprintf("failed to generate schema for field %s: %v", t.ID.StructField(), err))
			}
			fieldSchema.Description = fmt.Sprintf("The ID of the %s entity.", entityName)
			schema.Properties = append(schema.Properties, *fieldSchema.ToProperty("id"))

			if !t.ID.Default {
				schema.Required = append(schema.Required, "id")
			}
		}

		for _, f := range t.Fields {
			fa := GetAnnotation(f)

			// Sensitive fields are allowed to be set in create/update by default.
			if fa.GetSkip(cfg) || fa.ReadOnly {
				continue
			}

			if op == OperationCreate || !f.Immutable {
				fieldSchema, err = GetSchemaField(f)
				if err != nil {
					panic(fmt.Sprintf("failed to generate schema for field %s: %v", f.StructField(), err))
				}

				// Hoist enums into components to reduce duplication where possible.
				if updated, ref, ok := hoistEnums(t, f, fieldSchema); ok {
					schemas[ref] = fieldSchema
					schema.Properties = append(schema.Properties, ogen.Property{
						Name:   f.Name,
						Schema: updated,
					})
				} else {
					schema.Properties = append(schema.Properties, *fieldSchema.ToProperty(f.Name))
				}

				if op == OperationCreate && !f.Optional && !f.Default {
					schema.Required = append(schema.Required, f.Name)
				}
			}
		}

		for _, e := range t.Edges {
			ea := GetAnnotation(e)

			if ea.GetSkip(cfg) || ea.ReadOnly {
				continue
			}
			if op == OperationUpdate && (e.Immutable || (e.Field() != nil && e.Field().Immutable)) {
				continue
			}

			if e.Type.ID == nil {
				// It's a through-edge, which means it's not directly settable, so skip.
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

				if !fa.GetSkip(cfg) {
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

				if ea.EdgeUpdateBulk {
					schema.Properties = append(schema.Properties, *fieldSchema.ToProperty(e.Name))
				}
			} else {
				schema.Properties = append(schema.Properties, *fieldSchema.ToProperty(e.Name))
			}

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
			Description: cmp.Or(ta.GetOperationDescription(op), ta.Description, fmt.Sprintf("A single %s entity.", entityName)),
			Type:        "object",
			Properties:  ogen.Properties{},
			Required:    []string{},
		}

		var fieldSchema *ogen.Schema

		if t.ID != nil {
			fieldSchema, err = GetSchemaField(t.ID)
			if err != nil {
				panic(fmt.Sprintf("failed to generate schema for field %s: %v", t.ID.StructField(), err))
			}
			fieldSchema.Description = fmt.Sprintf("The ID of the %s entity.", entityName)
			schema.Properties = append(schema.Properties, *fieldSchema.ToProperty("id"))

			if !t.ID.Optional {
				schema.Required = append(schema.Required, "id")
			}
		}

		for _, f := range t.Fields {
			fa := GetAnnotation(f)

			if fa.GetSkip(cfg) || f.Sensitive() {
				continue
			}

			if !f.Optional {
				schema.Required = append(schema.Required, f.Name)
			}

			fieldSchema, err = GetSchemaField(f)
			if err != nil {
				panic(fmt.Sprintf("failed to generate schema for field %s: %v", f.StructField(), err))
			}

			// Hoist enums into components to reduce duplication where possible.
			if updated, ref, ok := hoistEnums(t, f, fieldSchema); ok {
				schemas[ref] = fieldSchema
				schema.Properties = append(schema.Properties, ogen.Property{
					Name:   f.Name,
					Schema: updated,
				})
			} else {
				schema.Properties = append(schema.Properties, *fieldSchema.ToProperty(f.Name))
			}
		}

		edgeSchema := &ogen.Schema{
			Type:       "object",
			Properties: ogen.Properties{},
			Required:   []string{},
		}

		for _, e := range t.Edges {
			ea := GetAnnotation(e)

			if ea.GetSkip(cfg) || !ea.GetEagerLoad(cfg) {
				continue
			}

			prop := ogen.Property{
				Name:   e.Name,
				Schema: &ogen.Schema{Ref: "#/components/schemas/" + Singularize(e.Type.Name)},
			}

			if !e.Unique {
				prop.Schema = prop.Schema.AsArray()

				if limit := ea.GetEagerLoadLimit(cfg); limit > 0 {
					prop.Schema.MinItems = ptr(uint64(0))
					prop.Schema.MaxItems = ptr(uint64(limit))
					prop.Schema.Description = fmt.Sprintf("A list of %s entities. Limited to %d items. If there are more results than the limit, the results are capped and you must use the associated edge endpoint with pagination -- see also the 'EagerLoadLimit' config option.", Singularize(e.Type.Name), limit)
				}
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
		if edge != nil {
			// TODO: if /categories does not have pagination enabled, and then we have another
			// endpoint like /pets/{id}/categories, we could reuse the /categories schema,
			// because both are []Category. Would optimize the schema size, with the downside
			// that the description would be less accurate (can't describe the relationship).
			ea := GetAnnotation(edge)
			ra := GetAnnotation(edge.Type)

			if !ea.GetPagination(cfg, edge) || (!ra.GetPagination(cfg, edge) || cfg.DisableEagerLoadNonPagedOpt) {
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

// hoistEnums helps hoist field enums into components to reduce duplication where possible.
// If the existing schema is pointing to an enum, a new schema is returned which points to the
// provided component schema ref name.
func hoistEnums(t *gen.Type, f *gen.Field, existing *ogen.Schema) (updated *ogen.Schema, ref string, ok bool) {
	if existing == nil {
		return nil, "", false
	}

	name := Singularize(t.Name) + PascalCase(f.Name) + "Enum"

	if existing.Type == "array" {
		isEnum := existing.Items.Item != nil && existing.Items.Item.Enum != nil
		if !isEnum {
			for i := range existing.Items.Items {
				if len(existing.Items.Items[i].Enum) > 0 {
					isEnum = true
					break
				}
			}
		}
		if isEnum {
			updated = &ogen.Schema{Ref: "#/components/schemas/" + name}
			return updated.AsArray(), name, true
		}
	}
	if len(existing.Enum) > 0 {
		return &ogen.Schema{Ref: "#/components/schemas/" + name}, name, true
	}
	return existing, "", false
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
func GetSortableFields(t *gen.Type, edge *gen.Edge) (sortable []string) {
	cfg := GetConfig(t.Config)
	ta := GetAnnotation(t)
	fields := t.Fields

	if t.ID != nil && (edge == nil || edge.Field() == nil) {
		fields = append([]*gen.Field{t.ID}, fields...)
	}

	if edge == nil {
		sortable = append(sortable, "random")
	}

	for _, f := range fields {
		fa := GetAnnotation(f)
		if fa.GetSkip(cfg) || f.Sensitive() || (!fa.Sortable && f.Name != "id") {
			continue
		}
		if !f.IsString() && !f.IsTime() && !f.IsBool() && !f.IsInt() && !f.IsInt64() && !f.IsUUID() {
			continue
		}
		sortable = append(sortable, f.Name)
	}

	if edge == nil {
		for _, e := range t.Edges {
			ea := GetAnnotation(e)

			if ea.GetSkip(cfg) {
				continue
			}

			if !e.Unique {
				sortable = append(sortable, e.Name+".count")

				for _, f := range e.Type.Fields {
					fa := GetAnnotation(f)
					if fa.GetSkip(cfg) || f.Sensitive() || !fa.Sortable || (!f.IsInt() && !f.IsInt64()) {
						continue
					}
					sortable = append(sortable, e.Name+"."+f.Name+".sum")
				}

				continue
			}

			for _, f := range GetSortableFields(e.Type, e) {
				sortable = append(sortable, e.Name+"."+f)
			}
		}
	}

	if v := ta.GetDefaultSort(t.ID != nil && (edge == nil || edge.Field() == nil)); v != "" && !slices.Contains(sortable, v) {
		panic(fmt.Sprintf(
			"default sort field %q on schema %q does not exist (valid: %s) or does not have default sorting enabled",
			v,
			t.Name,
			strings.Join(sortable, ","),
		))
	}

	slices.Sort(sortable)
	return slices.Compact(sortable)
}

// FilterableFieldOp represents a filterable field, that filters based on a specific
// operation (e.g. eq, neq, gt, lt, etc).
type FilterableFieldOp struct {
	Type        *gen.Type
	Edge        *gen.Edge    // Edge may be nil.
	Field       *gen.Field   // Field may be nil (if so, assume we want a parameter to check for the edges existence).
	Operation   gen.Op       // The associated operation.
	fieldSchema *ogen.Schema // The base schema for the field, this may change based on the operation provided.
}

// ParameterName returns the raw query parameter name for the filterable field.
func (f *FilterableFieldOp) ParameterName() string {
	if f.Edge != nil {
		if f.Field == nil {
			return "has." + CamelCase(SnakeCase(Singularize(f.Edge.Name)))
		}
		return CamelCase(SnakeCase(Singularize(f.Edge.Name))) + "." + CamelCase(f.Field.Name) + "." + predicateFormat(f.Operation)
	}
	return CamelCase(f.Field.Name) + "." + predicateFormat(f.Operation)
}

// ComponentName returns the name/component alias for the parameter.
func (f *FilterableFieldOp) ComponentName() string {
	if f.Edge != nil {
		if f.Field == nil {
			return "EdgeHas" + PascalCase(Singularize(f.Edge.Name))
		}
		return "Edge" + PascalCase(Singularize(f.Edge.Name)) + PascalCase(f.Field.Name) + PascalCase(f.Operation.Name())
	}
	return PascalCase(f.Type.Name) + PascalCase(f.Field.Name) + PascalCase(f.Operation.Name())
}

// StructTag returns the struct tag for the filterable field, which uses json and
// github.com/go-playground/validator based tags by default.
func (f *FilterableFieldOp) StructTag() string {
	return fmt.Sprintf(
		`form:%q json:%q`,
		f.ParameterName()+",omitempty",
		SnakeCase(f.ComponentName())+",omitempty",
	)
}

func generatePredicateBuilder(
	t *gen.Type,
	f *gen.Field,
	e *gen.Edge,
	op gen.Op,
	structName, componentName string,
) string {
	if op.Niladic() {
		if e != nil {
			if f == nil {
				return fmt.Sprintf("%s.Has%s()", t.Package(), e.StructField())
			}

			pkg := t.Package()

			if e.Ref != nil {
				pkg = e.Ref.Type.Package()
			} else if e.Owner != nil {
				pkg = e.Owner.Package()
			}

			return fmt.Sprintf(
				"%s.Has%sWith(%s.%s%s())",
				pkg,
				e.StructField(),
				t.Package(),
				f.StructField(),
				op.Name(),
			)
		}
		return fmt.Sprintf("%s.%s%s()", t.Package(), f.StructField(), op.Name())
	}

	ftype := structName + "." + componentName

	if op.Variadic() {
		ftype += "..."
	} else {
		ftype = "*" + ftype
	}

	builder := fmt.Sprintf(
		"%s.%s%s(%s)",
		t.Package(),
		f.StructField(),
		op.Name(),
		ftype,
	)

	if e != nil {
		pkg := t.Package()

		if e.Ref != nil {
			pkg = e.Ref.Type.Package()
		} else if e.Owner != nil {
			pkg = e.Owner.Package()
		}

		return fmt.Sprintf("%s.Has%sWith(%s)", pkg, e.StructField(), builder)
	}
	return builder
}

func (f *FilterableFieldOp) PredicateBuilder(structName string) string {
	return generatePredicateBuilder(
		f.Type,
		f.Field,
		f.Edge,
		f.Operation,
		structName,
		f.ComponentName(),
	)
}

// TypeString returns the struct field type for the filterable field.
func (f *FilterableFieldOp) TypeString() string {
	if (f.Edge != nil && f.Field == nil) || f.Operation.Niladic() {
		return "*bool"
	}
	if f.Operation.Variadic() {
		return "[]" + f.Field.Type.String()
	}
	return "*" + f.Field.Type.String()
}

// Description returns a description for the filterable field.
func (f *FilterableFieldOp) Description() string {
	if f.Edge != nil && f.Field == nil {
		return fmt.Sprintf("If true, only return entities that have a %s edge.", Singularize(f.Edge.Name))
	}
	return predicateDescription(f.Field, f.Operation)
}

// Parameter returns the parameter for the filterable field.
func (f *FilterableFieldOp) Parameter() *ogen.Parameter {
	if f.Edge != nil && f.Field == nil {
		return &ogen.Parameter{
			Name:        f.ParameterName(),
			In:          "query",
			Description: f.Description(),
			Schema:      &ogen.Schema{Type: "boolean"},
		}
	}

	schema := &ogen.Schema{
		Ref:        f.fieldSchema.Ref,
		Type:       f.fieldSchema.Type,
		Items:      f.fieldSchema.Items,
		Minimum:    f.fieldSchema.Minimum,
		Maximum:    f.fieldSchema.Maximum,
		MinLength:  f.fieldSchema.MinLength,
		MaxLength:  f.fieldSchema.MaxLength,
		Enum:       f.fieldSchema.Enum,
		Deprecated: f.fieldSchema.Deprecated,
	}

	if f.Operation == gen.GT || f.Operation == gen.LT || f.Operation == gen.GTE || f.Operation == gen.LTE {
		if schema.Items != nil {
			schema.Items.Item.Type = "number"
		} else {
			schema.Type = "number"
		}
	}

	if f.Operation == gen.IsNil {
		if schema.Items != nil {
			schema.Items.Item.Type = "boolean"
		} else {
			schema.Type = "boolean"
		}
	}

	if f.Operation.Variadic() {
		schema = schema.AsArray() // If not already.
	}

	return &ogen.Parameter{
		Name:        f.ParameterName(),
		In:          "query",
		Description: f.Description(),
		Schema:      schema,
	}
}

// GetFilterableFields returns a list of filterable fields for the given type, where
// the key is the component name, and the value is the parameter which includes the
// name, description and schema for the parameter.
func GetFilterableFields(t *gen.Type, edge *gen.Edge) (filters []*FilterableFieldOp) {
	cfg := GetConfig(t.Config)
	ta := GetAnnotation(t)

	if ta.GetSkip(cfg) {
		return nil
	}

	fields := t.Fields

	if cfg.DefaultFilterID && t.ID != nil && (edge == nil || edge.Field() == nil) {
		ida := GetAnnotation(t.ID)
		if ida.Filter == 0 {
			ida.Filter = FilterGroupEqualExact | FilterGroupArray
		}
		t.ID.Annotations.Set(ida.Name(), ida)
		fields = append([]*gen.Field{t.ID}, fields...)
	}

	for _, f := range fields {
		fa := GetAnnotation(f)

		if fa.Filter == 0 || fa.GetSkip(cfg) || f.Sensitive() {
			continue
		}

		for _, op := range intersectSorted(f.Ops(), fa.Filter.Explode()) {
			if f.IsBool() && op == gen.NEQ {
				continue // Since you can use EQ=false instead.
			}

			fieldSchema, err := GetSchemaField(f)
			if err != nil {
				continue // Just skip things that can't be generated/easily mapped.
			}

			if updated, _, ok := hoistEnums(t, f, fieldSchema); ok {
				fieldSchema = updated
			}

			filters = append(filters, &FilterableFieldOp{
				Type:        t,
				Edge:        edge,
				Field:       f,
				Operation:   op,
				fieldSchema: fieldSchema,
			})
		}
	}

	if edge == nil {
		for _, e := range t.Edges {
			ea := GetAnnotation(e)

			if ea.GetSkip(cfg) || ea.Filter == 0 || ea.Filter != FilterEdge {
				continue
			}

			filters = append(filters, &FilterableFieldOp{
				Type:      t,
				Edge:      e,
				Operation: gen.IsNil,
				fieldSchema: &ogen.Schema{
					Type: "boolean",
				},
			})
			filters = append(filters, GetFilterableFields(e.Type, e)...)
		}
	}
	return filters
}

type FilterGroup struct {
	Name       string
	Type       *gen.Type
	FieldType  *field.TypeInfo
	Operations []gen.Op
	FieldPairs []*FilterGroupFieldPair
	Schema     *ogen.Schema
}

// ParameterName returns the raw query parameter name for the filter group.
func (g *FilterGroup) ParameterName(op gen.Op) string {
	return g.Name + "." + predicateFormat(op)
}

// Parameter returns the parameter for the filter group.
func (g *FilterGroup) Parameter(op gen.Op) *ogen.Parameter {
	return &ogen.Parameter{
		Name:        g.ParameterName(op),
		In:          "query",
		Description: g.Description(op),
		Schema:      g.Schema,
	}
}

func (g *FilterGroup) Description(op gen.Op) string {
	var fields []string
	for _, fp := range g.FieldPairs {
		if fp.Edge != nil {
			fields = append(fields, fp.Edge.Name+"."+fp.Field.Name)
			continue
		}
		fields = append(fields, fp.Field.Name)
	}
	return fmt.Sprintf(
		"Field %q filters across multiple fields (case insensitive): %s.",
		g.ParameterName(op),
		strings.Join(fields, ", "),
	)
}

// StructTag returns the struct tag for the filter group, which uses json and
// github.com/go-playground/validator based tags by default.
func (g *FilterGroup) StructTag(op gen.Op) string {
	return fmt.Sprintf(
		`form:%q json:%q`,
		g.ParameterName(op)+",omitempty",
		SnakeCase(g.ComponentName(op))+",omitempty",
	)
}

func (g *FilterGroup) PredicateBuilder(structName string, op gen.Op) string {
	component := g.ComponentName(op)
	var fields []string
	for _, fp := range g.FieldPairs {
		fields = append(fields, generatePredicateBuilder(
			fp.Type,
			fp.Field,
			fp.Edge,
			op,
			structName,
			component,
		))
	}
	return fmt.Sprintf("sql.OrPredicates(\n%s,\n)", strings.Join(fields, ",\n"))
}

// TypeString returns the struct field type for the filter group.
func (g *FilterGroup) TypeString(op gen.Op) string {
	if op.Niladic() {
		return "*bool"
	}
	if op.Variadic() {
		return "[]" + g.FieldType.String()
	}
	return "*" + g.FieldType.String()
}

// ComponentName returns the name/component alias for the parameter.
func (g *FilterGroup) ComponentName(op gen.Op) string {
	return PascalCase(g.Type.Name) + "FilterGroup" + PascalCase(g.Name) + PascalCase(op.Name())
}

type FilterGroupFieldPair struct {
	Type  *gen.Type
	Edge  *gen.Edge
	Field *gen.Field
}

func GetFilterGroups(t *gen.Type, edge *gen.Edge) []*FilterGroup {
	cfg := GetConfig(t.Config)
	ta := GetAnnotation(t)

	if ta.GetSkip(cfg) {
		return nil
	}

	groups := map[string]*FilterGroup{}

	for _, f := range t.Fields {
		fa := GetAnnotation(f)

		if fa.FilterGroup == "" || fa.GetSkip(cfg) || f.Sensitive() {
			continue
		}

		fieldSchema, err := GetSchemaField(f)
		if err != nil {
			panic(fmt.Sprintf(
				"failed to generate schema for field %s within filter group %q: %v",
				f.StructField(),
				fa.FilterGroup,
				err,
			))
		}

		ops := intersectSorted(f.Ops(), fa.Filter.Explode())

		if f.IsBool() {
			ops = slices.DeleteFunc(ops, func(op gen.Op) bool {
				return op == gen.NEQ
			})
		}

		if len(ops) == 0 {
			panic(fmt.Sprintf(
				"filter group %q on type %q and field %q has no filter predicates configured",
				fa.FilterGroup,
				t.Name,
				f.StructField(),
			))
		}

		if _, ok := groups[fa.FilterGroup]; !ok {
			groups[fa.FilterGroup] = &FilterGroup{
				Name:       fa.FilterGroup,
				Type:       t,
				FieldType:  f.Type,
				Operations: ops,
				Schema:     fieldSchema,
			}
		}

		if groups[fa.FilterGroup].FieldType.String() != f.Type.String() {
			panic(fmt.Sprintf(
				"filter group %q on type %q and field %q has a different type than another field in the group: %q vs %q",
				fa.FilterGroup,
				t.Name,
				f.StructField(),
				groups[fa.FilterGroup].FieldType.String(),
				f.Type.String(),
			))
		}

		if newOps := intersectSorted(groups[fa.FilterGroup].Operations, ops); len(groups[fa.FilterGroup].Operations) < len(newOps) {
			if len(newOps) == 0 {
				panic(fmt.Sprintf(
					"filter group %q on type %q and field %q has no intersecting predicates when compared to all other fields in the group",
					fa.FilterGroup,
					t.Name,
					f.StructField(),
				))
			}
			groups[fa.FilterGroup].Operations = newOps
		}

		groups[fa.FilterGroup].FieldPairs = append(groups[fa.FilterGroup].FieldPairs, &FilterGroupFieldPair{
			Type:  t,
			Field: f,
		})
	}

	if edge == nil {
		for _, e := range t.Edges {
			ea := GetAnnotation(e)

			// Only include the edge as part of the inner filter group if it has a filter group
			// that's a part of the main type.
			if _, ok := groups[ea.FilterGroup]; !ok || ea.GetSkip(cfg) {
				continue
			}

			edgeGroups := GetFilterGroups(e.Type, e)
			for _, eg := range edgeGroups {
				if _, ok := groups[eg.Name]; !ok {
					groups[eg.Name] = eg
					continue
				}

				groups[eg.Name].FieldPairs = append(groups[eg.Name].FieldPairs, eg.FieldPairs...)
			}
		}
	}

	return slices.SortedFunc(maps.Values(groups), func(a, b *FilterGroup) int {
		return strings.Compare(a.Name, b.Name)
	})
}
