//go:build ignore

package main

import (
	"log"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"github.com/lrstanley/entrest"
)

func main() {
	ex, err := entrest.NewExtension(&entrest.Config{
		Handler:      entrest.HandlerStdlib,
		WithTesting:  true,
		StrictMutate: true,
	})
	if err != nil {
		log.Fatalf("creating entrest extension: %v", err)
	}

	err = entc.Generate(
		"./database/schema",
		&gen.Config{
			Target:  "./database/ent",
			Schema:  "github.com/lrstanley/entrest/_examples/simple/internal/database/schema",
			Package: "github.com/lrstanley/entrest/_examples/simple/internal/database/ent",
		},
		entc.Extensions(ex),
	)
	if err != nil {
		log.Fatalf("failed to run ent codegen: %v", err)
	}
}
