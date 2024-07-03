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

type Category struct {
	ent.Schema
}

func (Category) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.String("readonly").
			Annotations(entrest.WithReadOnly(true)),
		field.String("skip_in_spec").
			Optional().
			Annotations(entrest.WithSkip(true)),
	}
}

func (Category) Mixin() []ent.Mixin {
	return []ent.Mixin{
		AuditableTimestamp{},
	}
}

func (Category) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("pets", Pet.Type),
	}
}

func (Category) Annotations() []schema.Annotation {
	return []schema.Annotation{}
}
