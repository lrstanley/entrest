// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Comment("Name of the identity."),
		field.Enum("type").
			NamedValues(
				"System", "SYSTEM",
				"User", "USER",
			).
			Default("USER").
			Comment("Type of object being defined (user and system which is for internal usecases)."),
		field.String("description").
			MinLen(1).
			MaxLen(1000).
			Optional().
			Nillable().
			Comment("Full name if USER or SYSTEM, otherwise null."),
		field.String("email").
			MinLen(1).
			MaxLen(320).
			Optional().
			Nillable().
			Comment("Email associated with the identity. Note that not all identities have an associated email address."),
		field.Bytes("avatar").
			MinLen(1).
			MaxLen(1 * 1024 * 1024). // 1MB (ish).
			Optional().
			Nillable().
			StructTag(`json:"-"`).
			Comment("Avatar data for the identity. This should generally only apply to the USER identity type."),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			Comment("Time the identity was created in the source system."),
		field.Time("updated_at").
			Immutable().
			Default(time.Now).
			Comment("Last time the identity was updated in the source system."),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("pets", Pet.Type),
		edge.To("followed_pets", Pet.Type).
			Through("following", Follows.Type),
		edge.To("friends", User.Type).
			Through("friendships", Friendship.Type),
	}
}
