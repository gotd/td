package parser

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"strings"
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
			t.Logf(" %s: %s", a.Name, a.Value)
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
			t.Logf(" %s: %s", a.Name, a.Value)
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
			t.Run("JSON", func(t *testing.T) {
				g := goldie.New(t,
					goldie.WithFixtureDir("_golden/parser/json"),
					goldie.WithDiffEngine(goldie.ColoredDiff),
					goldie.WithNameSuffix(".json"),
				)
				g.AssertJson(t, v, schema)
			})
			t.Run("WriteTo", func(t *testing.T) {
				b := new(bytes.Buffer)
				if _, err := schema.WriteTo(b); err != nil {
					t.Fatal(err)
				}
				g := goldie.New(t,
					goldie.WithFixtureDir("_golden/parser/tl"),
					goldie.WithDiffEngine(goldie.ColoredDiff),
					goldie.WithNameSuffix(".tl"),
				)
				g.Assert(t, v, b.Bytes())

				parsedSchema, err := Parse(b)
				if err != nil {
					t.Fatal(err)
				}
				require.Equal(t, schema, parsedSchema, "parsed schema should be equal to written")
			})
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
		{
			File: "categories.tl",
			Schema: &Schema{
				Definitions: []SchemaDefinition{
					{
						Category: CategoryType,
						Definition: Definition{
							ID:   1,
							Name: "first",
							Type: Type{Name: "Foo"},
						},
					},
					{
						Category: CategoryFunction,
						Definition: Definition{
							ID:   6,
							Name: "func",
							Type: Type{Name: "Call"},
							Params: []Parameter{
								{Name: "id", Type: Type{Name: "int", Bare: true}},
							},
						},
					},
					{
						Category: CategoryType,
						Definition: Definition{
							ID:   2,
							Name: "second",
							Type: Type{Name: "Foo"},
						},
					},
					{
						Category: CategoryFunction,
						Definition: Definition{
							ID:   4,
							Name: "secFunc",
							Type: Type{Name: "Foo"},
							Params: []Parameter{
								{Name: "id", Type: Type{Name: "int", Bare: true}},
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
			require.Equal(t, tt.Schema, schema)

			t.Run("JSON", func(t *testing.T) {
				g := goldie.New(t,
					goldie.WithFixtureDir("_golden/parser_strict/json"),
					goldie.WithDiffEngine(goldie.ColoredDiff),
					goldie.WithNameSuffix(".json"),
				)
				g.AssertJson(t, tt.File, schema)
			})
			t.Run("WriteTo", func(t *testing.T) {
				b := new(bytes.Buffer)
				if _, err := schema.WriteTo(b); err != nil {
					t.Fatal(err)
				}
				g := goldie.New(t,
					goldie.WithFixtureDir("_golden/parser_strict/tl"),
					goldie.WithDiffEngine(goldie.ColoredDiff),
					goldie.WithNameSuffix(".tl"),
				)
				g.Assert(t, tt.File, b.Bytes())

				parsedSchema, err := Parse(b)
				if err != nil {
					t.Fatal(err)
				}
				require.Equal(t, schema, parsedSchema, "parsed schema should be equal to written")
			})
		})
	}
}

func TestParserFailures(t *testing.T) {
	for _, tt := range []string{
		"//@0 @0\n0=0",
		"//@\xef\f\f\f\f/@class Stat" +
			"isticsGraph@descript" +
			"ion /@r a@n a@a t@n " +
			"h\ng=StatisticsGra00",
	} {
		if _, err := Parse(strings.NewReader(tt)); err == nil {
			t.Error("should error")
		}
	}
}

func TestParserConsistent(t *testing.T) {
	for _, tt := range []string{
		"//@0 ////\\n0=0",
	} {
		parsed, err := Parse(strings.NewReader(tt))
		if err != nil {
			t.Fatal(err)
		}
		first := new(bytes.Buffer)
		if _, err := parsed.WriteTo(first); err != nil {
			t.Fatal(err)
		}
		secondParsed, err := Parse(bytes.NewReader(first.Bytes()))
		if err != nil {
			t.Fatal(err)
		}
		second := new(bytes.Buffer)
		if _, err := secondParsed.WriteTo(second); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(first.Bytes(), second.Bytes()) {
			t.Logf("%s", first)
			t.Logf("%s", second)
			t.Error("mismatch")
		}
	}
}
