package gen

import (
	"embed"
	"strings"
	"text/template"
)

// Funcs returns functions which used in templates.
func Funcs() template.FuncMap {
	return template.FuncMap{
		"trim":                 strings.TrimSpace,
		"lower":                strings.ToLower,
		"trimPrefix":           strings.TrimPrefix,
		"trimSuffix":           strings.TrimSuffix,
		"hasPrefix":            strings.HasPrefix,
		"hasSuffix":            strings.HasSuffix,
		"hasField":             hasField,
		"mapCollectableFields": mapCollectableFields,
		"sortableFields":       sortableFields,
		"generateSliceHelper":  generateSliceHelper,
		"concat": func(args ...interface{}) []interface{} {
			return args
		},
		"add": func(x, y int) int {
			return x + y
		},
		"notEmpty": func(s string) bool{
			return strings.TrimSpace(s) != ""
		},
	}
}

//go:embed _template/*.tmpl
var templates embed.FS

// Template parses and returns vendored code generation templates.
func Template() *template.Template {
	tmpl := template.New("templates").Funcs(Funcs())
	tmpl = template.Must(tmpl.ParseFS(templates, "_template/*.tmpl"))
	return tmpl
}
