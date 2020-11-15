package parser

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/sebdah/goldie/v2"
)

func TestParserBase(t *testing.T) {
	data, err := ioutil.ReadFile("_testdata/base.tl")
	if err != nil {
		t.Fatal(err)
	}
	schema, err := Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range schema.Classes {
		t.Logf("Class %s: %s", c.Name, c.Description)
	}
	for _, d := range schema.Definitions {
		t.Logf("%s = %s (0x%x)", d.Definition.Name, d.Definition.Type, d.Definition.ID)
		for _, a := range d.Annotations {
			t.Logf(" %s: %s", a.Key, a.Value)
		}
	}
}

func TestParserError(t *testing.T) {
	data, err := ioutil.ReadFile("_testdata/Error.tl")
	if err != nil {
		t.Fatal(err)
	}
	schema, err := Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range schema.Classes {
		t.Logf("Class %s: %s", c.Name, c.Description)
	}
	for _, d := range schema.Definitions {
		t.Logf("%s = %s", d.Definition.Name, d.Definition.Type)
		for _, a := range d.Annotations {
			t.Logf(" %s: %s", a.Key, a.Value)
		}
	}
}

func TestParser(t *testing.T) {
	for _, v := range []string{
		"td_api.tl",
		"telegram_api.tl",
	} {
		t.Run(v, func(t *testing.T) {
			data, err := ioutil.ReadFile("_testdata/td_api.tl")
			if err != nil {
				t.Fatal(err)
			}
			schema, err := Parse(bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			g := goldie.New(t,
				goldie.WithFixtureDir("_golden"),
				goldie.WithDiffEngine(goldie.ColoredDiff),
				goldie.WithNameSuffix(".golden.json"),
			)
			g.AssertJson(t, v, schema)
		})
	}
}
