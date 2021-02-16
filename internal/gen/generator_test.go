package gen

import (
	"bytes"
	"go/format"
	"os"
	"testing"

	"golang.org/x/xerrors"

	"github.com/gotd/tl"
)

type formattedSource struct{}

func (t formattedSource) WriteFile(name string, content []byte) error {
	if name == "" {
		return xerrors.New("name is blank")
	}
	_, err := format.Source(content)
	return err
}

func TestGenerator(t *testing.T) {
	data, err := os.ReadFile("_testdata/example.tl")
	if err != nil {
		t.Fatal(err)
	}
	schema, err := tl.Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	g, err := NewGenerator(schema, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := g.WriteSource(formattedSource{}, "pkg", Template()); err != nil {
		t.Fatal(err)
	}
}

func TestGeneratorTelegram(t *testing.T) {
	data, err := os.ReadFile("_testdata/telegram.tl")
	if err != nil {
		t.Fatal(err)
	}
	schema, err := tl.Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	g, err := NewGenerator(schema, "https://core.telegram.org/")
	if err != nil {
		t.Fatal(err)
	}
	if err := g.WriteSource(formattedSource{}, "pkg", Template()); err != nil {
		t.Fatal(err)
	}
}
