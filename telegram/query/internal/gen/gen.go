package main

import (
	"bytes"
	"context"
	"embed"
	"flag"
	"fmt"
	"go/format"
	"io"
	"os"
	"os/signal"
	"text/template"

	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/gen"
)

//go:embed _template/*.tmpl
var templates embed.FS // nolint:gochecknoglobals

func generate(ctx context.Context, out io.Writer, c *collector) error {
	pkg, err := load(ctx, "github.com/gotd/td/tg")
	if err != nil {
		return xerrors.Errorf("load: %w", err)
	}

	config, err := c.Config(pkg)
	if err != nil {
		return xerrors.Errorf("collect: %w", err)
	}

	tmpl := template.New("templates").Funcs(gen.Funcs())
	tmpl = template.Must(tmpl.ParseFS(templates, "_template/*.tmpl"))
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "header", config); err != nil {
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

func run(ctx context.Context) error {
	var out io.Writer = os.Stdout

	set := flag.NewFlagSet("gen", flag.ExitOnError)
	output := set.String("out", "", "output file")
	cfg := collectorConfig{}
	cfg.fromFlags(set)
	if err := set.Parse(os.Args[1:]); err != nil {
		return xerrors.Errorf("parse flags: %w", err)
	}

	if *output != "" {
		f, err := os.Create(*output)
		if err != nil {
			return xerrors.Errorf("can't create file %q: %w", *output, err)
		}
		defer func() {
			_ = f.Close()
		}()
		out = f
	}

	c := newCollector(cfg)
	return generate(ctx, out, c)
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx); err != nil {
		fmt.Println(err)
		return
	}
}
