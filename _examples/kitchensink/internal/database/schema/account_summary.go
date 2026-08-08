// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/lrstanley/entrest"
)

// AccountSummary is a read-only [ent.View] over a subset of user columns. entrest
// automatically exposes only list/read operations for views.
//
// Named AccountSummary (not UserSummary) so it does not sort last among schema
// types; ent's client template currently emits invalid hooks code when the last
// node is a view.
type AccountSummary struct {
	ent.View
}

func (AccountSummary) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.View("SELECT id, name, type, email, enabled FROM users"),
		entrest.WithDescription("Read-only summary of users without sensitive fields."),
		entrest.WithDefaultSort("name"),
		entrest.WithDefaultOrder(entrest.OrderAsc),
	}
}

func (AccountSummary) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.Nil).
			Annotations(
				entrest.WithSortable(true),
				entrest.WithFilter(entrest.FilterGroupEqual|entrest.FilterGroupArray),
			),
		field.String("name").
			Annotations(
				entrest.WithSortable(true),
				entrest.WithFilter(entrest.FilterGroupEqual|entrest.FilterGroupArray),
			).
			Comment("Name of the user."),
		field.Enum("type").
			NamedValues(
				"System", "SYSTEM",
				"User", "USER",
			).
			Annotations(
				entrest.WithFilter(entrest.FilterGroupEqualExact|entrest.FilterGroupArray),
			).
			Comment("Type of object being defined (user or system which is for internal usecases)."),
		field.String("email").
			Optional().
			Nillable().
			Annotations(
				entrest.WithSortable(true),
				entrest.WithFilter(entrest.FilterGroupEqual|entrest.FilterGroupArray),
			).
			Comment("Email associated with the user."),
		field.Bool("enabled").
			Annotations(entrest.WithFilter(entrest.FilterGroupEqualExact)).
			Comment("If the user is still in the source system."),
	}
}
