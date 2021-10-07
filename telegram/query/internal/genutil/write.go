package genutil

import (
	"bytes"
	"go/format"
	"io"
	"io/fs"
	"os"
	"text/template"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/internal/gen"
)

// WriteTemplate loads template from FS and executes it to given output writer.
func WriteTemplate(source fs.FS, out io.Writer, name string, data interface{}) error {
	tmpl := template.New("templates").Funcs(gen.Funcs())
	tmpl = template.Must(tmpl.ParseFS(source, "_template/*.tmpl"))
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return xerrors.Errorf("template: %w", err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		if _, cpyErr := io.Copy(os.Stdout, &buf); cpyErr != nil {
			return xerrors.Errorf("dump generated: %w, (original error: %s)", cpyErr, err.Error())
		}
		return xerrors.Errorf("format: %w", err)
	}

	_, err = out.Write(formatted)
	return err
}
