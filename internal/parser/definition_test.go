package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseDefinition(t *testing.T) {
	for _, tt := range []struct {
		Case       string
		Input      string
		Definition Definition
	}{
		{
			Case:  "inputPhoneCall",
			Input: "inputPhoneCall#1e36fded id:long access_hash:long = InputPhoneCall;",
			Definition: Definition{
				ID:        0x1e36fded,
				Name:      "inputPhoneCall",
				Interface: "InputPhoneCall",
				Fields: []Field{
					{
						Name: "id",
						Type: "long",
					},
					{
						Name: "access_hash",
						Type: "long",
					},
				},
			},
		},
		{
			Case:  "userWithoutCRC",
			Input: "user id:int first_name:string last_name:string = User;",
			Definition: Definition{
				ID:        0xd23c81a3,
				Name:      "user",
				Interface: "User",
				Fields: []Field{
					{
						Name: "id",
						Type: "int",
					},
					{
						Name: "first_name",
						Type: "string",
					},
					{
						Name: "last_name",
						Type: "string",
					},
				},
			},
		},
	} {
		var (
			input       = tt.Input
			expectedDef = tt.Definition
		)
		t.Run(tt.Case, func(t *testing.T) {
			d, err := parseDefinition(input)
			if err != nil {
				t.Fatal(err)
			}
			require.Equal(t, expectedDef, d)
		})
	}
}
