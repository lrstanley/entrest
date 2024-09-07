// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package entrest

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"entgo.io/ent/entc/gen"
	"entgo.io/ent/schema/field"
	"github.com/ogen-go/ogen"
)

// Predicate represents a filtering predicate provided by ent.
type Predicate int

// Mirrored from entgo.io/ent/entc/gen with special groupings and support for bitwise operations.
const (
	// FilterEdge is a special filter which is applied to the edge itself, indicating
	// that all of the edges fields should also be included in filtering options.
	FilterEdge Predicate = 1 << iota

	FilterEQ           // =
	FilterNEQ          // <>
	FilterGT           // >
	FilterGTE          // >=
	FilterLT           // <
	FilterLTE          // <=
	FilterIsNil        // IS NULL / has
	FilterIn           // within
	FilterNotIn        // without
	FilterEqualFold    // equals case-insensitive
	FilterContains     // containing
	FilterContainsFold // containing case-insensitive
	FilterHasPrefix    // startingWith
	FilterHasSuffix    // endingWith
	// FilterNotNil // Since you can use IsNil=false (we have three states for passed parameters).

	// FilterGroupEqualExact includes: eq, neq, equal fold, is nil.
	FilterGroupEqualExact = FilterEQ | FilterNEQ | FilterEqualFold | FilterGroupNil
	// FilterGroupEqual includes: eq, neq, equal fold, contains, contains case, prefix, suffix, nil.
	FilterGroupEqual = FilterGroupEqualExact | FilterGroupContains | FilterHasPrefix | FilterHasSuffix
	// FilterGroupContains includes: contains, contains case, is nil.
	FilterGroupContains = FilterContains | FilterContainsFold | FilterGroupNil
	// FilterGroupNil includes: is nil.
	FilterGroupNil = FilterIsNil
	// FilterGroupLength includes: gt, lt (often gte/lte isn't really needed).
	FilterGroupLength = FilterGT | FilterLT
	// FilterGroupArray includes: in, not in.
	FilterGroupArray = FilterIn | FilterNotIn
)

// filterMap maps a predicate to the entgo.io/ent/entc/gen.Op (to get string representation).
var filterMap = map[Predicate]gen.Op{
	FilterEQ:           gen.EQ,
	FilterNEQ:          gen.NEQ,
	FilterGT:           gen.GT,
	FilterGTE:          gen.GTE,
	FilterLT:           gen.LT,
	FilterLTE:          gen.LTE,
	FilterIsNil:        gen.IsNil,
	FilterIn:           gen.In,
	FilterNotIn:        gen.NotIn,
	FilterEqualFold:    gen.EqualFold,
	FilterContains:     gen.Contains,
	FilterContainsFold: gen.ContainsFold,
	FilterHasPrefix:    gen.HasPrefix,
	FilterHasSuffix:    gen.HasSuffix,
}

// String returns the gen.Op string representation of a predicate.
func (p Predicate) String() string {
	if _, ok := filterMap[p]; ok {
		return filterMap[p].Name()
	}
	panic("predicate.String() called with grouped predicate, use Explode() first")
}

// Has returns if the predicate has the provided predicate.
func (p Predicate) Has(v Predicate) bool {
	return p&v != 0
}

// Add adds the provided predicate to the current predicate.
func (p Predicate) Add(v Predicate) Predicate {
	p |= v
	return p
}

// Remove removes the provided predicate from the current predicate.
func (p Predicate) Remove(v Predicate) Predicate {
	p &^= v
	return p
}

// Explode returns all individual predicates as []gen.Op.
func (p Predicate) Explode() (ops []gen.Op) {
	for pred, op := range filterMap {
		if p.Has(pred) {
			ops = append(ops, op)
		}
	}
	return ops
}

// predicateFormat returns the query string representation of a filter predicate.
func predicateFormat(op gen.Op) string {
	switch op {
	case gen.Contains:
		return "has"
	case gen.ContainsFold:
		return "ihas"
	case gen.EqualFold:
		return "ieq"
	case gen.HasPrefix:
		return "prefix"
	case gen.HasSuffix:
		return "suffix"
	case gen.IsNil:
		return "null"
	default:
		return CamelCase(SnakeCase(op.Name()))
	}
}

// predicateDescription returns the description of a filter predicate.
func predicateDescription(f *gen.Field, op gen.Op) string {
	switch op {
	case gen.EQ:
		return fmt.Sprintf("Filters field %q to be equal to the provided value.", f.Name)
	case gen.NEQ:
		return fmt.Sprintf("Filters field %q to be not equal to the provided value.", f.Name)
	case gen.GT:
		if f.IsString() {
			return fmt.Sprintf("Filters field %q to be longer than the provided value.", f.Name)
		}
		return fmt.Sprintf("Filters field %q to be greater than the provided value.", f.Name)
	case gen.GTE:
		if f.IsString() {
			return fmt.Sprintf("Filters field %q to be longer than or equal in length to the provided value.", f.Name)
		}
		return fmt.Sprintf("Filters field %q to be greater than or equal to the provided value.", f.Name)
	case gen.LT:
		if f.IsString() {
			return fmt.Sprintf("Filters field %q to be shorter than the provided value.", f.Name)
		}
		return fmt.Sprintf("Filters field %q to be less than the provided value.", f.Name)
	case gen.LTE:
		if f.IsString() {
			return fmt.Sprintf("Filters field %q to be shorter than or equal in length to the provided value.", f.Name)
		}
		return fmt.Sprintf("Filters field %q to be less than or equal to the provided value.", f.Name)
	case gen.IsNil:
		return fmt.Sprintf("Filters field %q to be null/nil.", f.Name)
	case gen.NotNil:
		return fmt.Sprintf("Filters field %q to be not null/nil.", f.Name)
	case gen.In:
		return fmt.Sprintf("Filters field %q to be within the provided values.", f.Name)
	case gen.NotIn:
		return fmt.Sprintf("Filters field %q to be not within the provided values.", f.Name)
	case gen.EqualFold:
		return fmt.Sprintf("Filters field %q to be equal to the provided value, case-insensitive.", f.Name)
	case gen.Contains:
		return fmt.Sprintf("Filters field %q to contain the provided value.", f.Name)
	case gen.ContainsFold:
		return fmt.Sprintf("Filters field %q to contain the provided value, case-insensitive.", f.Name)
	case gen.HasPrefix:
		return fmt.Sprintf("Filters field %q to start with the provided value.", f.Name)
	case gen.HasSuffix:
		return fmt.Sprintf("Filters field %q to end with the provided value.", f.Name)
	default:
		panic("unknown predicate")
	}
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
