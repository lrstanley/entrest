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
		"getFilterGroups":     GetFilterGroups,
		"getOperationIDName":  GetOperationIDName,
		"getPathName":         GetPathName,
		"hasConditional":      hasConditional,
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

// FuncMaps export FuncMaps to use custom templates.
func FuncMaps() template.FuncMap {
	return funcMap
}

// hasConditional returns true if the graph has at least one schema with conditional functionality.
func hasConditional(g *gen.Graph, matchType *gen.Type) bool {
	cfg := GetConfig(g.Config)

	for _, t := range g.Nodes {
		if matchType != nil && t.Name != matchType.Name {
			continue
		}

		for _, f := range t.Fields {
			fa := GetAnnotation(f)

			if !fa.GetSkip(cfg) && fa.Conditional {
				return true
			}
		}
	}

	return false
}
