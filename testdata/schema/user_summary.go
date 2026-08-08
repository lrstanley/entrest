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
)

// UserSummary is a read-only [ent.View] used to verify entrest only generates
// list/read operations for views.
type UserSummary struct {
	ent.View
}

func (UserSummary) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.View("SELECT id, name, user_id FROM users"),
	}
}

func (UserSummary) Fields() []ent.Field {
	return []ent.Field{
		field.String("id"),
		field.String("name"),
		field.String("user_id").
			Optional(),
	}
}

func (UserSummary) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Unique().
			Field("user_id"),
	}
}
