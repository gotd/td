package gen

import (
	"text/template"

	"github.com/gotd/td/gen/internal"
)

func Template() *template.Template {
	tmpl := template.New("templates")
	tmpl = template.Must(tmpl.Parse(string(internal.MustAsset("_template/header.tmpl"))))
	tmpl = template.Must(tmpl.Parse(string(internal.MustAsset("_template/registry.tmpl"))))
	tmpl = template.Must(tmpl.Parse(string(internal.MustAsset("_template/main.tmpl"))))
	return tmpl
}
