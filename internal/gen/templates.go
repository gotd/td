package gen

import (
	"strings"
	"text/template"

	"github.com/gotd/td/internal/gen/internal"
)

// Funcs returns functions which used in templates.
func Funcs() template.FuncMap {
	return template.FuncMap{
		"trim":       strings.TrimSpace,
		"lower":      strings.ToLower,
		"trimPrefix": strings.TrimPrefix,
		"hasPrefix":  strings.HasPrefix,
		"add": func(x, y int) int {
			return x + y
		},
	}
}

// Template parses and returns vendored code generation templates.
func Template() *template.Template {
	tmpl := template.New("templates").Funcs(Funcs())
	for _, assetName := range internal.AssetNames() {
		tmpl = template.Must(tmpl.Parse(string(internal.MustAsset(assetName))))
	}

	return tmpl
}
