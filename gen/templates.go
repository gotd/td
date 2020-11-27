package gen

import (
	"text/template"

	"github.com/ernado/td/gen/internal"
)

func Template() *template.Template {
	tmpl := template.New("templates")
	tmpl = template.Must(tmpl.Parse(string(internal.MustAsset("_template/main.tmpl"))))
	tmpl = template.Must(tmpl.Parse(string(internal.MustAsset("_template/header.tmpl"))))
	return tmpl
}
