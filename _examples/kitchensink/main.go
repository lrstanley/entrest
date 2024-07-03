// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

//go:generate sh -c "mkdir -p ./database/schema"
//go:generate sh -c "cd database && go run -mod=readonly generate.go"

package main

import (
	"context"
	"database/sql"
	"net/http"

	"entgo.io/ent/dialect/sql/schema"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

	for range 100 {
		db.User.Create().
			SetName(gofakeit.FirstName()).
			SetEmail(gofakeit.Email()).
			SetType(user.TypeUser).
			SetEnabled(true).
			ExecX(ctx)
	}

	oreo := db.Pet.Create().
		SetName("Riley").
		AddFollowedBy(john).
		SaveX(ctx)

	for range 100 {
		db.Pet.Create().
			SetName(gofakeit.PetName()).
			SetOwner(john).
			AddFriends(oreo).
			AddFollowedBy(john).
			ExecX(ctx)
	}

	srv, err := rest.NewServer(db, &rest.ServerConfig{})
	if err != nil {
		panic(err)
	}

	// Example of using net/http serve-mux:
	//	mux := http.NewServeMux()
	//	mux.Handle("/", srv.Handler())
	//	http.ListenAndServe(":8080", mux)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Mount("/", srv.Handler())
	http.ListenAndServe(":8080", r) //nolint:all
}
