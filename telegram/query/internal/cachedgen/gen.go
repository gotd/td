package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"

	"go.uber.org/multierr"
	"golang.org/x/xerrors"

	"github.com/nnqq/td/telegram/query/internal/genutil"
)

//go:embed _template/*.tmpl
var templates embed.FS

func generate(ctx context.Context, out io.Writer, pkgName string) error {
	pkg, err := genutil.Load(ctx, "github.com/nnqq/td/tg")
	if err != nil {
		return xerrors.Errorf("load: %w", err)
	}

	return genutil.WriteTemplate(templates, out, "header", Config{
		Queries: collect(pkg),
		Package: pkgName,
	})
}

func run(ctx context.Context) (err error) {
	var out io.Writer = os.Stdout

	set := flag.NewFlagSet("gen", flag.ExitOnError)
	output := set.String("out", "", "output file")
	pkgName := set.String("package", "cached", "name of package name to generate")
	if err := set.Parse(os.Args[1:]); err != nil {
		return xerrors.Errorf("parse flags: %w", err)
	}

	if *output != "" {
		f, err := os.Create(*output)
		if err != nil {
			return xerrors.Errorf("can't create file %q: %w", *output, err)
		}
		defer func() {
			multierr.AppendInto(&err, f.Close())
		}()
		out = f
	}

	return generate(ctx, out, *pkgName)
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx); err != nil {
		fmt.Println(err)
		return
	}
}
