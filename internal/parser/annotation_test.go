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
		{},
		{
			Input: "//@name The name of the option @value The new value of the option",
			Result: []Annotation{
				{
					Key:   "name",
					Value: "The name of the option",
				},
				{
					Key:   "value",
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
}
