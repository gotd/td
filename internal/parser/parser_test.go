package parser

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

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
		"telegram_api_header.tl",
	} {
		t.Run(v, func(t *testing.T) {
			data, err := ioutil.ReadFile(filepath.Join("_testdata", v))
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

func TestParserStrict(t *testing.T) {
	for _, tt := range []struct {
		File   string
		Schema *Schema
	}{
		{
			File: "fields.tl",
			Schema: &Schema{
				Definitions: []SchemaDefinition{
					{
						Category: CategoryType,
						Definition: Definition{
							Name: "inputMediaPhoto",
							ID:   0xb3ba0635,
							Type: Type{
								Name: "InputMedia",
							},
							Params: []Parameter{
								{Flags: true, Name: "flags"},
								{Name: "id", Type: Type{Name: "InputPhoto"}},
								{Name: "ttl_seconds", Type: Type{Name: "int", Bare: true}, Flag: &Flag{Name: "flags", Index: 0}},
							},
						},
					},
				},
			},
		},
	} {
		t.Run(tt.File, func(t *testing.T) {
			data, err := ioutil.ReadFile(filepath.Join("_testdata", tt.File))
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
			g.AssertJson(t, tt.File, schema)
			require.Equal(t, tt.Schema, schema)
		})
	}
}
