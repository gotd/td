package main

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-bindata/go-bindata"
)

type config struct {
	Package string
	Output  string
	Input   []bindata.InputConfig
}

func parseFlags() (config, error) {
	c := config{}

	set := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	set.StringVar(&c.Package, "pkg", c.Package, "Package name to use in the generated code.")
	set.StringVar(&c.Output, "o", c.Output, "Optional name of the output file to be generated.")

	if err := set.Parse(os.Args[1:]); err != nil {
		return c, err
	}

	if set.NFlag() == 0 {
		return c, errors.New("missing <input dir>")
	}

	// Create input configurations.
	c.Input = make([]bindata.InputConfig, set.NArg())
	for i := range c.Input {
		c.Input[i] = parseInput(set.Arg(i))
	}

	return c, nil
}

func glob(d bindata.InputConfig) string {
	if d.Recursive {
		return d.Path + "/*"
	}

	return d.Path
}

// parseRecursive determines whether the given path has a recrusive indicator and
// returns a new path with the recursive indicator chopped off if it does.
//
//  ex:
//      /path/to/foo/...    -> (/path/to/foo, true)
//      /path/to/bar        -> (/path/to/bar, false)
func parseInput(path string) bindata.InputConfig {
	if strings.HasSuffix(path, "/...") {
		return bindata.InputConfig{
			Path:      filepath.Clean(path[:len(path)-4]),
			Recursive: true,
		}
	}

	return bindata.InputConfig{
		Path:      filepath.Clean(path),
		Recursive: false,
	}
}
