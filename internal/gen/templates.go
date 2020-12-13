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
	tmpl = template.Must(tmpl.Parse(string(internal.MustAsset("_template/utils.tmpl"))))
	tmpl = template.Must(tmpl.Parse(string(internal.MustAsset("_template/header.tmpl"))))
	tmpl = template.Must(tmpl.Parse(string(internal.MustAsset("_template/registry.tmpl"))))
	tmpl = template.Must(tmpl.Parse(string(internal.MustAsset("_template/client.tmpl"))))
	tmpl = template.Must(tmpl.Parse(string(internal.MustAsset("_template/main.tmpl"))))
	tmpl = template.Must(tmpl.Parse(string(internal.MustAsset("_template/handlers.tmpl"))))
	return tmpl
}
