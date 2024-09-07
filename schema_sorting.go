// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package entrest

import (
	"fmt"
	"slices"
	"strings"

	"entgo.io/ent/entc/gen"
)

// SortOrder represents the sorting order.
type SortOrder string

const (
	// OrderAsc represents an ascending order.
	OrderAsc SortOrder = "asc"
	// OrderDesc represents a descending order.
	OrderDesc SortOrder = "desc"
)

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
