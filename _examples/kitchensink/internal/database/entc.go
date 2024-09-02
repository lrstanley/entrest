//go:build ignore

// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package main

import (
	"log"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"github.com/lrstanley/entrest"
)

func main() {
	ex, err := entrest.NewExtension(&entrest.Config{
		SpecFromPath:          "../base-openapi.json", // Using a base spec to start with, not required.
		Handler:               entrest.HandlerStdlib,
		WithTesting:           true,
		StrictMutate:          true,
		ListNotFound:          true,
		GlobalRequestHeaders:  entrest.RequestIDHeader,
		GlobalResponseHeaders: entrest.RateLimitHeaders,
	})
	if err != nil {
		log.Fatalf("creating entrest extension: %v", err)
	}

	err = entc.Generate(
		"./database/schema",
		&gen.Config{
			Target:  "./database/ent",
			Schema:  "github.com/lrstanley/entrest/_examples/kitchensink/internal/database/schema",
			Package: "github.com/lrstanley/entrest/_examples/kitchensink/internal/database/ent",
			Features: []gen.Feature{
				gen.FeaturePrivacy,
			},
		},
		entc.Extensions(ex),
	)
	if err != nil {
		log.Fatalf("failed to run ent codegen: %v", err)
	}
}
