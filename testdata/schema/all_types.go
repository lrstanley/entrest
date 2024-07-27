// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package schema

import (
	"database/sql/driver"
	"fmt"
	"net/http"
	"net/url"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/lrstanley/entrest"
	"github.com/ogen-go/ogen"
)

type AllTypes struct{ ent.Schema }

// Fields of the AllTypes.
func (AllTypes) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.Nil).Default(uuid.New),
		field.Int("int"),
		field.Int8("int8"),
		field.Int16("int16"),
		field.Int32("int32"),
		field.Int64("int64"),
		field.Uint("uint"),
		field.Uint8("uint8"),
		field.Uint16("uint16"),
		field.Uint32("uint32"),
		field.Uint64("uint64"),
		field.Float32("float32"),
		field.Float("float64"),
		field.String("string_type"),
		field.Bool("bool"),
		field.Time("time"),
		field.Text("text"),
		field.Enum("state").Values("on", "off"),
		field.Strings("strings"),
		field.Ints("ints"),
		field.Floats("floats"),
		field.Bytes("bytes"),
		field.JSON("nicknames", []string{}),
		field.JSON("json_slice", []http.Dir{}).
			Annotations(entrest.WithSchema(ogen.String().AsArray())),
		field.JSON("json_obj", url.URL{}).
			Annotations(entrest.WithSchema(ogen.String())),
		field.Text("nilable").
			Nillable(),
		field.Other("other", &ExampleValuer{}).
			SchemaType(map[string]string{
				dialect.Postgres: "varchar",
				dialect.SQLite:   "text",
			}).
			Default(DefaultExampleValuer()).
			Annotations(entrest.WithSchema(ogen.String())),
	}
}

type ExampleValuer struct {
	*url.URL
}

func DefaultExampleValuer() *ExampleValuer {
	u, _ := url.Parse("127.0.0.1")
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
