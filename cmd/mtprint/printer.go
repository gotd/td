package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/k0kubun/pp/v3"
	"golang.org/x/xerrors"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/mt"
	"github.com/nnqq/td/internal/proto/codec"
	"github.com/nnqq/td/internal/tmap"
	"github.com/nnqq/td/tdp"
	"github.com/nnqq/td/tg"
	"github.com/nnqq/td/transport"
)

// Object is abstraction for TL schema object.
type Object interface {
	bin.Object
	tdp.Object
}

// Formatter formats given bin.Object and prints it to io.Writer.
type Formatter interface {
	Format(w io.Writer, i Object) error
}

// FormatterFunc is functional adapter for Formatter.
type FormatterFunc func(w io.Writer, i Object) error

// Format implements Formatter.
func (f FormatterFunc) Format(w io.Writer, i Object) error {
	return f(w, i)
}

func formats(name string) Formatter {
	switch name {
	case "json":
		return FormatterFunc(func(w io.Writer, i Object) error {
			e := json.NewEncoder(w)
			e.SetIndent("", "\t")
			return e.Encode(i)
		})
	case "pp":
		return FormatterFunc(func(w io.Writer, i Object) error {
			_, err := pp.Fprintln(w, i)
			return err
		})
	case "tdp":
		return FormatterFunc(func(w io.Writer, i Object) error {
			_, err := io.WriteString(w, tdp.Format(i))
			return err
		})
	default: // "go" format
		return FormatterFunc(func(w io.Writer, i Object) error {
			_, err := fmt.Fprintln(w, i)
			return err
		})
	}
}

// Printer decodes messages from given reader and prints is using Formatter.
type Printer struct {
	src    io.Reader
	codec  transport.Codec
	format Formatter
}

// NewPrinter creates new Printer.
// If format is nil, "go" format will be used.
// If c is nil, codec.Intermediate will be use.
func NewPrinter(src io.Reader, format Formatter, c transport.Codec) Printer {
	if c == nil {
		c = codec.Intermediate{}
	}
	if format == nil {
		format = formats("go")
	}
	return Printer{
		src:    src,
		codec:  c,
		format: format,
	}
}

// Print prints decoded messages to output.
func (p Printer) Print(output io.Writer) error {
	b := &bin.Buffer{}

	m := tmap.NewConstructor(
		tg.TypesConstructorMap(),
		mt.TypesConstructorMap(),
	)
	for {
		b.Reset()
		if err := p.codec.Read(p.src, b); err != nil {
			if xerrors.Is(err, io.EOF) {
				return nil
			}

			return err
		}

		id, err := b.PeekID()
		if err != nil {
			return err
		}

		obj := m.New(id)
		if obj == nil {
			return xerrors.Errorf("find type %#x", id)
		}

		v, ok := obj.(Object)
		if !ok {
			return xerrors.Errorf("unexpected type %T", obj)
		}

		if err := v.Decode(b); err != nil {
			return err
		}

		if err := p.format.Format(output, v); err != nil {
			return err
		}
	}
}
