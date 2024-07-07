// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/lrstanley/entrest"
)

type AuditableTimestamp struct{ mixin.Schema }

func (AuditableTimestamp) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Annotations(
				entrest.WithReadOnly(true),
				entrest.WithSortable(true),
				entrest.WithFilter(entrest.FilterGroupLength),
			).
			Comment("Time in which the resource was initially created."),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Annotations(
				entrest.WithReadOnly(true),
				entrest.WithSortable(true),
				entrest.WithFilter(entrest.FilterGroupLength),
			).
			Comment("Time that the resource was last updated."),
	}
}
