// Binary dlkey extracts public keys from remote repo.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	"path"

	"go.uber.org/multierr"
	"golang.org/x/xerrors"
)

func run(ctx context.Context) (rErr error) {
	var (
		name   = flag.String("f", "mtproto_dc_options.cpp", "file name to download")
		base   = flag.String("base", "https://raw.githubusercontent.com/telegramdesktop/tdesktop", "base url")
		branch = flag.String("branch", "dev", "branch to use")
		dir    = flag.String("dir", "Telegram/SourceFiles/mtproto/", "directory")
		out    = flag.String("o", "", "output file name (blank to stdout)")
	)
	flag.Parse()

	u, err := url.Parse(*base)
	if err != nil {
		return xerrors.Errorf("parse base: %w", err)
	}
	u.Path = path.Join(u.Path, *branch, *dir, *name)

	keys, err := extractKeys(ctx, u)
	if err != nil {
		return xerrors.Errorf("extract keys: %w", err)
	}

	available, err := getAvailable(ctx, keys)
	if err != nil {
		return xerrors.Errorf("get fingerprints: %w", err)
	}

	var w io.Writer = os.Stdout
	if p := *out; p != "" {
		f, err := os.Create(p)
		if err != nil {
			return xerrors.Errorf("create: %w", err)
		}
		defer multierr.AppendInvoke(&rErr, multierr.Close(f))
	}

	return available.Print(w)
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}
}
