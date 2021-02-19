package gen

import (
	"embed"
	"strings"
	"text/template"
)

// Funcs returns functions which used in templates.
func Funcs() template.FuncMap {
	return template.FuncMap{
		"trim":       strings.TrimSpace,
		"lower":      strings.ToLower,
		"trimPrefix": strings.TrimPrefix,
		"hasPrefix":  strings.HasPrefix,
		"hasSuffix":  strings.HasSuffix,
		"hasField": func(fields []fieldDef, name, typ string) bool {
			for _, f := range fields {
				if f.Name == name && f.Type == typ {
					return true
				}
			}

			return false
		},
		"concat": func(args ...interface{}) []interface{} {
			return args
		},
		"add": func(x, y int) int {
			return x + y
		},
	}
}

//go:embed _template/*.tmpl
var templates embed.FS // nolint:gochecknoglobals

// Template parses and returns vendored code generation templates.
func Template() *template.Template {
	tmpl := template.New("templates").Funcs(Funcs())
	tmpl = template.Must(tmpl.ParseFS(templates, "_template/*.tmpl"))
	return tmpl
}
