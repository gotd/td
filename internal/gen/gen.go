package gen

import (
	"io"
	"text/template"

	"github.com/ernado/tl"
)

func Generate(w io.Writer, t *template.Template, s *tl.Schema) error {
	cfg := Config{
		Package: "td",
	}
	renderCtx := Context{
		Config: cfg,
	}
	return t.ExecuteTemplate(w, "simple", renderCtx)
}
