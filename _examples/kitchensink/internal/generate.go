package internal

//go:generate go run -mod=readonly database/entc.go

import (
	// Import tools that are used in "go:build ignore" based files, which won't
	// automatically be tracked in go.mod.
	_ "entgo.io/ent/entc/gen"
	_ "github.com/lrstanley/entrest"
	_ "github.com/ogen-go/ogen"
)
