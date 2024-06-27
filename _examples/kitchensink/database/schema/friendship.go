// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Friendship struct {
	ent.Schema
}

func (Friendship) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Default(time.Now),
		field.Int("user_id"),
		field.Int("friend_id"),
	}
}

func (Friendship) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Required().
			Unique().
			Field("user_id").
			Annotations(
				entsql.OnDelete(entsql.Cascade),
			),
		edge.To("friend", User.Type).
			Required().
			Unique().
			Field("friend_id").
			Annotations(
				entsql.OnDelete(entsql.Cascade),
			),
	}
}
