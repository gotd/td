package gen

import (
	"strings"
	"text/template"

	"github.com/gotd/td/internal/gen/internal"
)

// Template parses and returns vendored code generation templates.
func Template() *template.Template {
	tmpl := template.New("templates").Funcs(template.FuncMap{
		"trim":       strings.TrimSpace,
		"lower":      strings.ToLower,
		"trimPrefix": strings.TrimPrefix,
		"hasPrefix":  strings.HasPrefix,
	})
	for _, assetName := range internal.AssetNames() {
		tmpl = template.Must(tmpl.Parse(string(internal.MustAsset(assetName))))
	}

	return tmpl
}
