// Binary mtprint pretty-prints MTProto message from binary file.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/internal/tmap"
	"github.com/gotd/td/tg"
)

func main() {
	// TODO: Streaming mode.
	flag.Parse()
	name := flag.Arg(0)
	if name == "" {
		panic("no file provided")
	}

	buf, err := ioutil.ReadFile(name) // #nosec
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
