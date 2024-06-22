// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Pet struct {
	ent.Schema
}

func (Pet) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id"),
		field.String("name"),
		field.JSON("nicknames", []string{}).Optional(),
		field.Int("age").Optional(),
	}
}

func (Pet) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("categories", Category.Type).Ref("pets"),
		edge.From("owner", User.Type).Ref("pets").Unique(),
		edge.To("friends", Pet.Type),
		edge.From("followed_by", User.Type).
			Ref("followed_pets").
			Through("following", Follows.Type),
	}
}
