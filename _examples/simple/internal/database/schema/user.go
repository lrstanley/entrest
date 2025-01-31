package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/lrstanley/entrest"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").
			Unique().
			Immutable().
			Annotations(
				entrest.WithSortable(true),
				entrest.WithFilter(entrest.FilterGroupEqual|entrest.FilterGroupArray),
				entrest.WithConditional(true),
			).
			Comment("Username of the user."),
		field.String("display_name").
			Annotations(
				entrest.WithSortable(true),
				entrest.WithFilter(entrest.FilterGroupEqual|entrest.FilterGroupArray),
			).
			Comment("full name/display name of the user."),
		field.String("email").
			MinLen(1).
			MaxLen(320).
			Annotations(
				entrest.WithSortable(true),
				entrest.WithExample("John.Smith@example.com"),
				entrest.WithFilter(entrest.FilterGroupEqual|entrest.FilterGroupArray),
			).
			Comment("Email associated with the user."),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("pets", Pet.Type).
			Comment("Pets owned by the user.").
			Annotations(
				entrest.WithEagerLoad(true),
				entrest.WithFilter(entrest.FilterEdge),
				entsql.OnDelete(entsql.SetNull),
			),
	}
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		MixinTime{},
	}
}
