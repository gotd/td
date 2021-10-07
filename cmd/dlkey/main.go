// Binary dlkey extracts public keys from remote repo.
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/nnqq/td/internal/keyparser"
)

func main() {
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
		panic(err)
	}

	u.Path = path.Join(u.Path, *branch, *dir, *name)

	res, err := http.Get(u.String())
	if err != nil {
		panic(err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode/100 != 2 {
		panic(fmt.Sprintf("status code %d", res.StatusCode))
	}

	var outWriter io.Writer = os.Stdout
	if *out != "" {
		w, err := os.Create(*out)
		if err != nil {
			panic(err)
		}
		defer func() {
			if err := w.Close(); err != nil {
				panic(err)
			}
		}()
		outWriter = w
	}

	if err := keyparser.Extract(res.Body, outWriter); err != nil {
		panic(err)
	}
}
