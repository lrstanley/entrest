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
)

type Follows struct {
	ent.Schema
}

func (Follows) Fields() []ent.Field {
	return []ent.Field{
		field.Time("followed_at").Default(time.Now),
		field.Int("user_id"),
		field.Int("pet_id"),
	}
}

func (Follows) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Unique().
			Required().
			Field("user_id"),
		edge.To("pet", Pet.Type).
			Unique().
			Required().
			Field("pet_id"),
	}
}

func (Follows) Annotations() []schema.Annotation {
	return []schema.Annotation{
		field.ID("user_id", "pet_id"),
	}
}
