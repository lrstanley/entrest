// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

//go:generate sh -c "mkdir -p ./database/schema"
//go:generate sh -c "cd database && go run -mod=readonly generate.go"

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"entgo.io/ent/dialect/sql/schema"
	"github.com/lrstanley/entrest/_examples/kitchensink/database/ent"
	"github.com/lrstanley/entrest/_examples/kitchensink/database/ent/rest"
	_ "github.com/lrstanley/entrest/_examples/kitchensink/database/ent/runtime" // Required by ent.
	"github.com/lrstanley/entrest/_examples/kitchensink/database/ent/user"
	"modernc.org/sqlite"
)

func main() {
	sql.Register("sqlite3", &sqlite.Driver{})
	db, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_pragma=foreign_keys(1)&_busy_timeout=15")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	ctx := context.Background()

	err = db.Schema.Create(
		ctx,
		schema.WithDropColumn(true),
		schema.WithDropIndex(true),
		schema.WithGlobalUniqueID(true),
		schema.WithForeignKeys(true),
	)
	if err != nil {
		panic(err)
	}

	john := db.User.Create().
		SetName("john").
		SetEmail("john@example.com").
		SetType(user.TypeUser).
		SetEnabled(true).
		SaveX(ctx)

	oreo := db.Pet.Create().
		SetName("Oreo").
		AddFollowedBy(john).
		SaveX(ctx)

	riley := db.Pet.Create().
		SetName("Riley").
		AddFriends(oreo).
		SetOwner(john).
		SaveX(ctx)

	printJSON := func(v any) {
		var b []byte
		b, err = json.MarshalIndent(v, "", "    ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(b)) //nolint:all
	}

	printJSON(john)
	printJSON(oreo)
	printJSON(riley)

	srv, err := rest.NewServer(db, nil)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", srv.Handler())
	http.ListenAndServe(":8080", mux) //nolint:all
}
