// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/lrstanley/entrest"
)

type Settings struct {
	ent.Schema
}

func (Settings) Fields() []ent.Field {
	return []ent.Field{
		field.String("global_banner").
			NotEmpty().
			MaxLen(1000).
			Optional().
			Nillable().
			Comment("Global banner text to apply to the frontend."),
	}
}

func (Settings) Mixin() []ent.Mixin {
	return []ent.Mixin{
		AuditableTimestamp{},
	}
}

func (Settings) Hooks() []ent.Hook {
	return []ent.Hook{}
}

func (Settings) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("admins", User.Type).
			Annotations(
				entrest.WithEagerLoad(true),
			).
			Comment("Administrators for the platform."),
	}
}

func (Settings) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entrest.WithExcludeOperations(entrest.OperationCreate, entrest.OperationDelete),
		entrest.WithDescription("Settings contains the global settings for the platform. Generally only one should ever be returned."),
	}
}
