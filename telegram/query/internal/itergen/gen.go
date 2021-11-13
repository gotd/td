package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/go-faster/errors"
	"go.uber.org/multierr"

	"github.com/gotd/td/telegram/query/internal/genutil"
)

//go:embed _template/*.tmpl
var templates embed.FS

func generate(ctx context.Context, out io.Writer, cfg collectorConfig) error {
	pkg, err := genutil.Load(ctx, "github.com/gotd/td/tg")
	if err != nil {
		return errors.Wrap(err, "load")
	}

	c := newCollector(pkg, cfg)
	config, err := c.Config()
	if err != nil {
		return errors.Wrap(err, "collect")
	}

	return genutil.WriteTemplate(templates, out, "header", config)
}

func run(ctx context.Context) (err error) {
	var out io.Writer = os.Stdout

	set := flag.NewFlagSet("gen", flag.ExitOnError)
	output := set.String("out", "", "output file")
	cfg := collectorConfig{}
	cfg.fromFlags(set)
	if err := set.Parse(os.Args[1:]); err != nil {
		return errors.Wrap(err, "parse flags")
	}

	if *output != "" {
		f, err := os.Create(*output)
		if err != nil {
			return errors.Wrapf(err, "can't create file %q", *output)
		}
		defer func() {
			multierr.AppendInto(&err, f.Close())
		}()
		out = f
	}

	return generate(ctx, out, cfg)
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx); err != nil {
		fmt.Println(err)
		return
	}
}
