package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/lrstanley/entrest"
)

type Pet struct {
	ent.Schema
}

func (Pet) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id"),
		field.String("name").
			Annotations(
				entrest.WithExample("Kuro"),
				entrest.WithSortable(true),
				entrest.WithFilter(entrest.FilterGroupEqual|entrest.FilterGroupArray),
			),
		field.Int("age").
			Min(0).
			Max(50).
			Annotations(
				entrest.WithExample(2),
				entrest.WithSortable(true),
				entrest.WithFilter(entrest.FilterGroupEqualExact|entrest.FilterGroupArray|entrest.FilterGroupLength),
			),
		field.Enum("type").
			NamedValues(
				"Dog", "DOG",
				"Cat", "CAT",
				"Bird", "BIRD",
				"Fish", "FISH",
				"Amphibian", "AMPHIBIAN",
				"Reptile", "REPTILE",
				"Other", "OTHER",
			).
			Annotations(
				entrest.WithExample("DOG"),
				entrest.WithSortable(true),
				entrest.WithFilter(entrest.FilterGroupEqualExact|entrest.FilterGroupArray),
			),
	}
}

func (Pet) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("pets").
			Unique().
			Comment("The user that owns the pet.").
			Annotations(
				entrest.WithEagerLoad(true),
				entrest.WithFilter(entrest.FilterEdge),
			),
		edge.To("friends", Pet.Type).
			Comment("Pets that this pet is friends with.").
			Annotations(
				entrest.WithFilter(entrest.FilterEdge),
			),
	}
}

func (Pet) Mixin() []ent.Mixin {
	return []ent.Mixin{
		MixinTime{},
	}
}
