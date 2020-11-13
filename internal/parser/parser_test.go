package parser

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestParser(t *testing.T) {
	data, err := ioutil.ReadFile("_testdata/td_api.tl")
	if err != nil {
		t.Fatal(err)
	}
	schema, err := Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range schema.Types {
		t.Logf("%s = %s", d.Definition.Name, d.Definition.Interface)
		for _, a := range d.Annotations {
			t.Logf(" %s: %s", a.Key, a.Value)
		}
	}
}
