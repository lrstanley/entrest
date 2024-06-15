// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package schema

import (
	"net/http"
	"net/url"

	"entgo.io/ent"
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
	}
}
