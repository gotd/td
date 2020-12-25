// Binary mtprint pretty-prints MTProto message from binary file.
package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/internal/tmap"
	"github.com/gotd/td/tg"
)

func main() {
	inputName := flag.String("f", "", "input file (blank for stdin)")
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

	// TODO: Streaming mode via intermediate protocol.
	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}

	b := &bin.Buffer{Buf: buf}
	id, err := b.PeekID()
	if err != nil {
		panic(err)
	}

	c := tmap.NewConstructor(
		tg.TypesConstructorMap(),
		mt.TypesConstructorMap(),
	)
	v := c.New(id)
	if v == nil {
		panic(fmt.Sprintf("failed to find type 0x%x", id))
	}

	if err := v.Decode(b); err != nil {
		panic(err)
	}

	fmt.Print(v)
}
