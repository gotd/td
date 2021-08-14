// Binary gotdgen generates go source code from TL schema.
package main

import (
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"

	"github.com/gotd/tl"

	"github.com/gotd/td/internal/gen"
)

type formattedSource struct {
	Format bool
	Root   string
}

func (t formattedSource) WriteFile(name string, content []byte) error {
	out := content
	if t.Format {
		buf, err := format.Source(content)
		if err != nil {
			return err
		}
		out = buf
	}
	return os.WriteFile(filepath.Join(t.Root, name), out, 0600)
}

func main() {
	schemaPath := flag.String("schema", "", "Path to .tl file")
	targetDir := flag.String("target", "td", "Path to target dir")
	packageName := flag.String("package", "td", "Target package name")
	performFormat := flag.Bool("format", true, "Perform code formatting")
	docBase := flag.String("doc", "", "Base documentation url")
	docLineLimit := flag.Int("line-limit", 0, "GoDoc comment line length limit")
	clean := flag.Bool("clean", false, "Clean generated files before generation")
	client := flag.Bool("client", true, "Generate client definition")
	registry := flag.Bool("registry", true, "Generate type ID registry")
	server := flag.Bool("server", false, "Generate server handlers")

	flag.Parse()
	if *schemaPath == "" {
		panic("no schema provided")
	}
	f, err := os.Open(*schemaPath)
	if err != nil {
		panic(err)
	}
	defer func() { _ = f.Close() }()

	schema, err := tl.Parse(f)
	if err != nil {
		panic(err)
	}
	files, err := os.ReadDir(*targetDir)
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	if os.IsNotExist(err) {
		if err := os.Mkdir(*targetDir, 0750); err != nil {
			panic(err)
		}
	}
	if *clean {
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			name := f.Name()
			if !strings.HasSuffix(name, "_gen.go") {
				continue
			}
			if !strings.HasPrefix(name, "tl_") {
				continue
			}
			if err := os.Remove(filepath.Join(*targetDir, name)); err != nil {
				panic(err)
			}
		}
	}

	fs := formattedSource{
		Root:   *targetDir,
		Format: *performFormat,
	}
	var opts []gen.Option
	if *client {
		opts = append(opts, gen.WithClient())
	}
	if *registry {
		opts = append(opts, gen.WithRegistry())
	}
	if *server {
		opts = append(opts, gen.WithServer())
	}
	if *docBase != "" {
		opts = append(opts, gen.WithDocumentation(*docBase))
	}
	if *docLineLimit != 0 {
		opts = append(opts, gen.WithDocLineLimit(*docLineLimit))
	}
	g, err := gen.NewGenerator(schema, opts...)
	if err != nil {
		panic(fmt.Sprintf("%+v", err))
	}

	if err := g.WriteSource(fs, *packageName, gen.Template()); err != nil {
		panic(fmt.Sprintf("%+v", err))
	}
}
