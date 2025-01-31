package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/lrstanley/entrest"
)

var _ ent.Mixin = (*MixinTime)(nil)

type MixinTime struct {
	mixin.Schema
}

func (MixinTime) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Annotations(
				entrest.WithSortable(true),
				entrest.WithReadOnly(true),
				entrest.WithFilter(entrest.FilterGroupEqualExact|entrest.FilterGroupLength),
			).
			Comment("Time the entity was created."),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Annotations(
				entrest.WithSortable(true),
				entrest.WithReadOnly(true),
				entrest.WithFilter(entrest.FilterGroupEqualExact|entrest.FilterGroupLength),
				entrest.WithConditional(true),
			).
			Comment("Time the entity was last updated."),
	}
}
