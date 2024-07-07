// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/lrstanley/entrest"
)

type Follows struct {
	ent.Schema
}

func (Follows) Fields() []ent.Field {
	return []ent.Field{
		field.Time("followed_at").
			Default(time.Now).
			Annotations(
				entrest.WithSortable(true),
				entrest.WithReadOnly(true),
			),
		field.Int("user_id"),
		field.Int("pet_id"),
	}
}

func (Follows) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Unique().
			Required().
			Field("user_id").
			Comment("The user that is following the pet.").
			Annotations(
				entrest.WithEagerLoad(true),
				entrest.WithFilter(entrest.FilterGroupEqualExact),
				entsql.OnDelete(entsql.Cascade),
			),
		edge.To("pet", Pet.Type).
			Unique().
			Required().
			Field("pet_id").
			Comment("The pet that is being followed by the user.").
			Annotations(
				entrest.WithEagerLoad(true),
				entrest.WithFilter(entrest.FilterGroupEqualExact),
				entsql.OnDelete(entsql.Cascade),
			),
	}
}

func (Follows) Annotations() []schema.Annotation {
	return []schema.Annotation{
		field.ID("user_id", "pet_id"),
	}
}
