package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseAnnotation(t *testing.T) {
	for _, tt := range []struct {
		Case   string
		Input  string
		Result []Annotation
	}{
		{
			Input: "//@name The name of the option @value The new value of the option",
			Result: []Annotation{
				{
					Name:  "name",
					Value: "The name of the option",
				},
				{
					Name:  "value",
					Value: "The new value of the option",
				},
			},
		},
	} {
		t.Run(tt.Case, func(t *testing.T) {
			ann, err := parseAnnotation(tt.Input)
			if err != nil {
				t.Fatal(err)
			}
			require.Equal(t, tt.Result, ann, "result should be equal")
		})
	}
	t.Run("String", func(t *testing.T) {
		for _, input := range []string{
			"//@foo bar",
			"//@bar baz.",
		} {
			ann, err := parseAnnotation(input)
			if err != nil {
				t.Fatal(err)
			}
			if len(ann) != 1 {
				t.Fatal("bad len")
			}
			if s := ann[0].String(); s != input {
				t.Errorf("%q != %q", s, input)
			}
		}
	})
	t.Run("Error", func(t *testing.T) {
		for _, input := range []string{
			"//@{} test",
			"//",
			"1",
			"//@\xef\f\f\f\f/@class StatisticsGraph@description /@r a@n a@a t@n h",
		} {
			if _, err := parseAnnotation(input); err == nil {
				t.Errorf("expected error on %q", input)
			}
		}
	})
	t.Run("SingleLine", func(t *testing.T) {
		if str := singleLineAnnotations([]Annotation{
			{Name: "class", Value: "foo"},
			{Name: "desc", Value: "bar"},
		}); str != "//@class foo @desc bar" {
			t.Error(str)
		}
	})
}
