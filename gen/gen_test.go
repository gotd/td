package gen

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/ernado/tl"
)

type TestFS struct {
	Root string
}

func (t TestFS) WriteFile(name string, content []byte) error {
	return ioutil.WriteFile(filepath.Join(t.Root, name), content, 0600)
}

func TestGen(t *testing.T) {
	tp, err := template.ParseGlob("_template/*.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	data, err := ioutil.ReadFile("_testdata/Error.tl")
	if err != nil {
		t.Fatal(err)
	}
	schema, err := tl.Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if err := Generate(TestFS{Root: "example"}, tp, schema); err != nil {
		t.Fatal(err)
	}
}
