//nolint:all
package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"entgo.io/ent/dialect/sql/schema"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/lrstanley/entrest/_examples/simple/internal/database/ent"
	"github.com/lrstanley/entrest/_examples/simple/internal/database/ent/rest"
	_ "github.com/lrstanley/entrest/_examples/simple/internal/database/ent/runtime" // Required by ent.
	"modernc.org/sqlite"
)

func main() {
	sql.Register("sqlite3", &sqlite.Driver{})
	db, err := ent.Open("sqlite3", "file:local.db?cache=shared&_pragma=foreign_keys(1)&_busy_timeout=15")
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

	srv, err := rest.NewServer(db, &rest.ServerConfig{})
	if err != nil {
		panic(err)
	}

	fmt.Println("running http server")
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Mount("/", srv.Handler())
	http.ListenAndServe(":8080", r)
}
