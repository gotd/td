package genutil

import (
	"bytes"
	"go/format"
	"io"
	"io/fs"
	"os"
	"text/template"

	"github.com/go-faster/errors"

	"github.com/gotd/td/gen"
)

// WriteTemplate loads template from FS and executes it to given output writer.
func WriteTemplate(source fs.FS, out io.Writer, name string, data interface{}) error {
	tmpl := template.New("templates").Funcs(gen.Funcs())
	tmpl = template.Must(tmpl.ParseFS(source, "_template/*.tmpl"))
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return errors.Wrap(err, "template")
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		if _, cpyErr := io.Copy(os.Stdout, &buf); cpyErr != nil {
			return errors.Wrapf(cpyErr, "dump generated (original error: %v)", err)
		}
		return errors.Wrap(err, "format")
	}

	_, err = out.Write(formatted)
	return err
}
