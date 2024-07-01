// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package entrest

import (
	"embed"
	"text/template"

	"entgo.io/ent/entc/gen"
)

var (
	funcMap = template.FuncMap{
		"zplural":   Pluralize,
		"zkebab":    KebabCase,
		"zsingular": Singularize,
		"zpascal":   PascalCase,
		"zcamel":    CamelCase,
		"zsnake":    SnakeCase,

		// Use this function when you want to invoke annotation functions (which are
		// often created if they depend on [Config]).
		"getAnnotation":       GetAnnotation,
		"getSortableFields":   GetSortableFields,
		"getFilterableFields": GetFilterableFields,
		"getOperationIDName":  GetOperationIDName,
		"getPathName":         GetPathName,
	}

	//go:embed templates
	templateDir embed.FS

	baseTemplates = gen.MustParse(
		gen.NewTemplate("rest").Funcs(funcMap).
			SkipIf(func(g *gen.Graph) bool { return GetConfig(g.Config).Handler == HandlerNone }).
			ParseFS(
				templateDir,
				"templates/*.tmpl",
				"templates/helper/*.tmpl",
				"templates/helper/**/*.tmpl",
			),
	)
	testingTemplates = gen.MustParse(
		gen.NewTemplate("resttest").Funcs(funcMap).
			SkipIf(func(g *gen.Graph) bool { return !GetConfig(g.Config).WithTesting }).
			ParseFS(
				templateDir,
				"templates/testing/*.tmpl",
			),
	)
)
