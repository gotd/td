// Binary mtprint pretty-prints MTProto message from binary file.
package main

import (
	"flag"
	"io"
	"os"

	"github.com/nnqq/td/internal/proto/codec"
)

func main() {
	inputName := flag.String("f", "", "input file (blank for stdin)")
	format := flag.String("format", "go", "print format")
	flag.Parse()

	var reader io.Reader = os.Stdin
	if *inputName != "" {
		f, err := os.Open(*inputName)
		if err != nil {
			panic(err)
		}
		defer func() { _ = f.Close() }()
		reader = f
	}

	p := NewPrinter(reader, formats(*format), codec.Intermediate{})
	if err := p.Print(os.Stdout); err != nil {
		panic(err)
	}
}
