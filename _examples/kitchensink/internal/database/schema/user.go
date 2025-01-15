// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package schema

import (
	"database/sql/driver"
	"fmt"
	"net/url"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/go-github/v66/github"
	"github.com/google/uuid"
	"github.com/lrstanley/entrest"
	"github.com/ogen-go/ogen"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.Nil).
			Default(uuid.New).
			Unique().
			Immutable().
			Annotations(
				entrest.WithSortable(true),
				entrest.WithFilter(entrest.FilterGroupEqual|entrest.FilterGroupArray),
			),
		field.String("name").
			Annotations(
				entrest.WithSortable(true),
				entrest.WithFilter(entrest.FilterGroupEqual|entrest.FilterGroupArray),
				entrest.WithFilterGroup("search"),
			).
			Comment("Name of the user."),
		field.Enum("type").
			NamedValues(
				"System", "SYSTEM",
				"User", "USER",
			).
			Default("USER").
			Annotations(
				entrest.WithExample("USER"),
				entrest.WithFilter(entrest.FilterGroupEqualExact|entrest.FilterGroupArray),
			).
			Comment("Type of object being defined (user or system which is for internal usecases)."),
		field.String("description").
			MinLen(1).
			MaxLen(1000).
			Optional().
			Nillable().
			Annotations(
				entrest.WithExample("Jon Smith"),
				entrest.WithFilter(entrest.FilterGroupContains|entrest.FilterGroupNil),
				entrest.WithFilterGroup("search"),
			).
			Comment("Full name if USER, otherwise null."),
		field.Bool("enabled").
			Annotations(entrest.WithFilter(entrest.FilterGroupEqualExact)).
			Default(true).Comment("If the user is still in the source system."),
		field.String("email").
			MinLen(1).
			MaxLen(320).
			Optional().
			Nillable().
			Annotations(
				entrest.WithSortable(true),
				entrest.WithExample("John.Smith@example.com"),
				entrest.WithFilter(entrest.FilterGroupEqual|entrest.FilterGroupArray),
				entrest.WithFilterGroup("search"),
			).
			Comment("Email associated with the user. Note that not all users have an associated email address."),
		field.Bytes("avatar").
			MinLen(1).
			MaxLen(1 * 1024 * 1024).
			Optional().
			Nillable().
			StructTag(`json:"-"`).
			Comment("Avatar data for the user. This should generally only apply to the USER user type."),
		field.String("password_hashed").
			Sensitive().
			NotEmpty().
			Annotations(
				// These should theoretically have no impact.
				entrest.WithFilter(entrest.FilterGroupEqual | entrest.FilterGroupArray),
			).
			Comment("Hashed password for the user, this shouldn't be readable in the spec anywhere."),
		// Make sure it imports the right github package version into the generated code.
		field.JSON("github_data", &github.User{}).
			Optional().
			Annotations(
				entrest.WithSchema(entrest.SchemaObjectAny),
			).
			Comment("The github user raw JSON data."),
		field.Other("profile_url", &ExampleValuer{}).
			Optional().
			SchemaType(map[string]string{
				dialect.Postgres: "varchar",
				dialect.SQLite:   "text",
			}).
			Default(DefaultExampleValuer()).
			Annotations(entrest.WithSchema(ogen.String())),
		field.Time("last_authenticated_at").
			Optional().
			Nillable().Annotations(
			entrest.WithFilter(entrest.FilterGroupEqual),
		),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("pets", Pet.Type).
			Comment("Pets owned by the user.").
			Annotations(
				entrest.WithEagerLoad(true),
				entrest.WithEagerLoadLimit(-1),
				entrest.WithFilter(entrest.FilterEdge),
				entsql.OnDelete(entsql.SetNull),
			),
		edge.To("followed_pets", Pet.Type).
			Through("following", Follows.Type).
			Comment("Pets that the user is following.").
			Annotations(
				entrest.WithFilter(entrest.FilterEdge),
				entsql.OnDelete(entsql.Cascade),
			),
		edge.To("friends", User.Type).
			Through("friendships", Friendship.Type).
			Comment("Friends of the user.").
			Annotations(
				entrest.WithFilter(entrest.FilterEdge),
				entsql.OnDelete(entsql.Cascade),
			),
	}
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entrest.WithDefaultSort("name"),
		entrest.WithDefaultOrder(entrest.OrderAsc),
		entrest.WithAllowClientIDs(true),
	}
}

type ExampleValuer struct {
	*url.URL
}

func DefaultExampleValuer() *ExampleValuer {
	u, _ := url.Parse("http://127.0.0.1/")
	return &ExampleValuer{URL: u}
}

func (l *ExampleValuer) Scan(value any) (err error) {
	switch v := value.(type) {
	case nil:
	case []byte:
		l.URL, err = url.Parse(string(v))
	case string:
		l.URL, err = url.Parse(v)
	default:
		err = fmt.Errorf("unexpected type %T", v)
	}
	return err
}

func (l ExampleValuer) Value() (driver.Value, error) {
	if l.URL == nil {
		return nil, nil //nolint:nilnil
	}
	return l.URL.String(), nil
}

func (l ExampleValuer) MarshalText() ([]byte, error) {
	if l.URL == nil {
		return nil, nil //nolint:nilnil
	}
	return []byte(l.URL.String()), nil
}

func (l *ExampleValuer) UnmarshalText(data []byte) (err error) {
	if len(data) == 0 {
		return nil
	}
	l.URL, err = url.Parse(string(data))
	return err
}
