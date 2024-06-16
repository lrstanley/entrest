// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/lrstanley/entrest"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
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
			Default("USER").
			Annotations(
				entrest.WithExample("USER"),
				entrest.WithFilter(entrest.FilterGroupEqualExact|entrest.FilterGroupArray),
			).
			Comment("Type of object being defined (user or system which is for internal usecases)."),
		field.String("description").
			MinLen(1).
			MaxLen(1000).
			Optional().
			Nillable().
			Annotations(
				entrest.WithExample("Jon Smith"),
				entrest.WithFilter(entrest.FilterGroupContains|entrest.FilterGroupNil),
			).
			Comment("Full name if USER, otherwise null."),
		field.Bool("enabled").
			Annotations(
				entrest.WithFilter(entrest.FilterGroupEqualExact),
			).
			Default(true).Comment("If the user is still in the source system."),
		field.String("email").
			MinLen(1).
			MaxLen(320).
			Optional().
			Nillable().
			Annotations(
				entrest.WithSortable(true),
				entrest.WithExample("John.Smith@example.com"),
				entrest.WithFilter(entrest.FilterGroupEqual|entrest.FilterGroupArray),
			).
			Comment("Email associated with the user. Note that not all users have an associated email address."),
		field.Bytes("avatar").
			MinLen(1).
			MaxLen(1 * 1024 * 1024).
			Optional().
			Nillable().
			StructTag(`json:"-"`).
			Comment("Avatar data for the user. This should generally only apply to the USER user type."),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			Annotations(entrest.WithSortable(true)).
			Comment("Time the user was created in the source system."),
		field.Time("updated_at").
			Immutable().
			Default(time.Now).
			Annotations(entrest.WithSortable(true)).
			Comment("Last time the user was updated in the source system."),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("pets", Pet.Type).Annotations(
			entrest.WithEagerLoad(true),
			entrest.WithFilter(entrest.FilterEdge),
		),
	}
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{}
}
