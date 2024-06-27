// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/lrstanley/entrest"
)

type Pet struct {
	ent.Schema
}

func (Pet) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id"),
		field.String("name").
			Annotations(
				entrest.WithExample("Kuro"),
				entrest.WithSortable(true),
				entrest.WithFilter(entrest.FilterGroupEqual|entrest.FilterGroupArray),
			),
		field.JSON("nicknames", []string{}).
			Optional().
			Annotations(
				entrest.WithFilter(entrest.FilterGroupEqual | entrest.FilterGroupArray),
			),
		field.Int("age").
			Optional().
			Annotations(
				entrest.WithExample(2),
				entrest.WithSortable(true),
				entrest.WithFilter(entrest.FilterGroupEqualExact|entrest.FilterGroupArray|entrest.FilterGroupLength),
			),
	}
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		AuditableTimestamp{},
	}
}

func (Pet) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("categories", Category.Type).
			Ref("pets").
			Comment("Categories that the pet belongs to.").
			Annotations(
				entrest.WithEagerLoad(true),
				entrest.WithFilter(entrest.FilterEdge),
				entrest.WithEdgeUpdateBulk(true),
			),
		edge.From("owner", User.Type).
			Ref("pets").
			Unique().
			Comment("The user that owns the pet.").
			Annotations(
				entrest.WithEagerLoad(true),
				entrest.WithFilter(entrest.FilterEdge),
			),
		edge.To("friends", Pet.Type).
			Comment("Pets that this pet is friends with.").
			Annotations(
				entrest.WithFilter(entrest.FilterEdge),
			),
		edge.From("followed_by", User.Type).
			Ref("followed_pets").
			Through("following", Follows.Type).
			Comment("Users that this pet is followed by.").
			Annotations(
				entrest.WithFilter(entrest.FilterEdge),
				entsql.OnDelete(entsql.Cascade),
			),
	}
}

func (Pet) Annotations() []schema.Annotation {
	return []schema.Annotation{}
}
