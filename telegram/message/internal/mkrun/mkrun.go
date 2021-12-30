// Package mkrun contains some helpers for generation scripts.
package mkrun

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io"
	"os"
	"strings"
	"text/template"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/go-faster/errors"
	"go.uber.org/multierr"
)

// Config is generation config.
type Config struct {
	PackageName string
	Data        interface{}
}

func generate(w io.Writer, pkgName string, g Generator) error {
	start := time.Now()
	data, err := g.Data()
	if err != nil {
		return err
	}
	collectInfoTime := time.Since(start)

	start = time.Now()
	buf := bytes.Buffer{}
	t, err := template.New("gen").Funcs(template.FuncMap{
		"trimPrefix": strings.TrimPrefix,
		"trimSuffix": strings.TrimSuffix,
		"lowerFirst": func(s string) string {
			r, size := utf8.DecodeRuneInString(s)
			if r == utf8.RuneError || unicode.IsLower(r) {
				return s
			}
			return string(unicode.ToLower(r)) + s[size:]
		},
	}).Parse(g.Template())
	if err != nil {
		return errors.Errorf("parse: %w", err)
	}

	if err := t.Execute(&buf, Config{
		PackageName: pkgName,
		Data:        data,
	}); err != nil {
		return errors.Errorf("execute: %w", err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		_, _ = os.Stderr.Write(buf.Bytes())
		return errors.Errorf("format: %w", err)
	}

	if _, err := w.Write(formatted); err != nil {
		return errors.Errorf("write: %w", err)
	}
	writeTime := time.Since(start)

	fmt.Printf("Generation %s complete, collect time: %s, write time: %s\n",
		g.Name(),
		collectInfoTime,
		writeTime,
	)

	return nil
}

func run(g Generator) (rErr error) {
	set := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	var (
		o       = set.String("output", "", "output file")
		pkgName = set.String("package", os.Getenv("GOPACKAGE"), "package name")
	)
	g.Flags(set)
	if err := set.Parse(os.Args[1:]); err != nil {
		return errors.Wrap(err, "parse")
	}

	if *pkgName == "" {
		if *o != "" {
			return errors.New("package name is empty")
		}
		*pkgName = "pkg"
	}

	var w io.Writer = os.Stdout
	if path := *o; path != "" {
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer multierr.AppendInvoke(&rErr, multierr.Close(f))
		w = f
	}

	return generate(w, *pkgName, g)
}

// Main is generation main function.
func Main(g Generator) {
	if err := run(g); err != nil {
		panic(err)
	}
}
