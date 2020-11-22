package main

import (
	"flag"
	"os"

	"github.com/ernado/tl"

	"github.com/ernado/td/gen"
)

func main() {
	schemaPath := flag.String("schema", "", "Path to .tl file")
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

	if err := gen.Generate(os.Stdout, gen.Template(), schema); err != nil {
		panic(err)
	}
}
