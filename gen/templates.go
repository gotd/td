package gen

import (
	"text/template"

	"github.com/ernado/td/gen/internal"
)

//go:generate go run github.com/go-bindata/go-bindata/go-bindata -pkg=internal -o=internal/bindata.go -mode=420 -modtime=1 ./_template/...

func Template() *template.Template {
	tmpl := template.New("templates")
	tmpl = template.Must(tmpl.Parse(string(internal.MustAsset("_template/main.tmpl"))))
	tmpl = template.Must(tmpl.Parse(string(internal.MustAsset("_template/header.tmpl"))))
	return tmpl
}
