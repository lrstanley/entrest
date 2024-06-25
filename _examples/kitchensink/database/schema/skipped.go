// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/lrstanley/entrest"
)

type Skipped struct {
	ent.Schema
}

func (Skipped) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
	}
}

func (Skipped) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entrest.WithSkip(true),
	}
}
