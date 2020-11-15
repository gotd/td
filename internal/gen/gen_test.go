package gen

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
	"text/template"

	"github.com/ernado/tl"
)

func TestGen(t *testing.T) {
	tp, err := template.ParseGlob("_template/*.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Create("_testdata/output.go")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	data, err := ioutil.ReadFile("_testdata/Error.tl")
	if err != nil {
		t.Fatal(err)
	}
	schema, err := tl.Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if err := Generate(f, tp, schema); err != nil {
		t.Fatal(err)
	}
}
