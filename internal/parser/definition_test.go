package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var bareString = Type{
	Name: "string",
	Bare: true,
}

var bareLong = Type{
	Name: "long",
	Bare: true,
}

func TestDefinition(t *testing.T) {
	for _, tt := range []struct {
		Case       string
		Input      string
		String     string
		Definition Definition
	}{
		{
			Case:  "inputPhoneCall",
			Input: "inputPhoneCall#1e36fded id:long access_hash:long = InputPhoneCall",
			Definition: Definition{
				ID:   0x1e36fded,
				Name: "inputPhoneCall",
				Params: []Parameter{
					{
						Name: "id",
						Type: bareLong,
					},
					{
						Name: "access_hash",
						Type: bareLong,
					},
				},
				Type: Type{Name: "InputPhoneCall"},
			},
		},
		{
			Case:   "userWithoutCRC",
			Input:  "user id:int first_name:string last_name:string = User;",
			String: "user#d23c81a3 id:int first_name:string last_name:string = User",
			Definition: Definition{
				ID:   0xd23c81a3,
				Name: "user",
				Type: Type{Name: "User"},
				Params: []Parameter{
					{Name: "id", Type: Type{Name: "int", Bare: true}},
					{Name: "first_name", Type: bareString},
					{Name: "last_name", Type: bareString},
				},
			},
		},
		{
			Case:   "OK",
			Input:  "ok = Ok;",
			String: "ok#d4edbe69 = Ok",
			Definition: Definition{
				ID:   0xd4edbe69,
				Name: "ok",
				Type: Type{Name: "Ok"},
			},
		},
		{
			Case:   "GroupWithoutFieldNames",
			String: "group#60fc45e0 int string string = Group",
			Input:  "group int string string = Group",
			Definition: Definition{
				ID:   0x60fc45e0,
				Name: "group",
				Type: Type{Name: "Group"},
				Params: []Parameter{
					{Type: Type{Name: "int", Bare: true}},
					{Type: bareString},
					{Type: bareString},
				},
			},
		},
		{
			Case:   "Zero",
			Input:  "0=0",
			String: "0#971b4490 = 0",
			Definition: Definition{
				Name: "0",
				ID:   0x971b4490,
				Type: Type{Name: "0", Bare: true},
			},
		},
		{
			Case:  "inputMediaUploadedPhoto",
			Input: "inputMediaUploadedPhoto#1e287d04 flags:# file:InputFile stickers:flags.0?Vector<InputDocument> ttl_seconds:flags.1?int = InputMedia",
			Definition: Definition{
				Name: "inputMediaUploadedPhoto",
				ID:   0x1e287d04,
				Type: Type{Name: "InputMedia"},
				Params: []Parameter{
					{Flags: true, Name: "flags"},
					{Name: "file", Type: Type{Name: "InputFile"}},
					{
						Name: "stickers",
						Flag: &Flag{
							Name:  "flags",
							Index: 0,
						},
						Type: Type{
							Name: "Vector",
							GenericArg: &Type{
								Name: "InputDocument",
							},
						},
					},
					{
						Name: "ttl_seconds",
						Flag: &Flag{
							Name:  "flags",
							Index: 1,
						},
						Type: Type{
							Name: "int",
							Bare: true,
						},
					},
				},
			},
		},
		{
			Case:  "invokeWithLayer",
			Input: "invokeWithLayer#da9b0d0d {X:Type} layer:int query:!X = X",
			Definition: Definition{
				Name:          "invokeWithLayer",
				ID:            0xda9b0d0d,
				Type:          Type{Name: "X"},
				GenericParams: []string{"X"},
				Params: []Parameter{
					{
						Name: "layer",
						Type: Type{Name: "int", Bare: true},
					},
					{
						Name: "query",
						Type: Type{Name: "X", GenericRef: true},
					},
				},
			},
		},
	} {
		var (
			input       = tt.Input
			expectedDef = tt.Definition
			expectedStr = tt.Input
		)
		if tt.String != "" {
			expectedStr = tt.String
		}
		t.Run(tt.Case, func(t *testing.T) {
			var d Definition
			if err := d.Parse(input); err != nil {
				t.Fatal(err)
			}
			require.Equal(t, expectedDef, d)
			require.Equal(t, expectedStr, d.String())
		})
	}
	t.Run("Error", func(t *testing.T) {
		for _, invalid := range []string{
			"=0",
			"0 :{.0?InputFi00=0",
		} {
			t.Run(invalid, func(t *testing.T) {
				var d Definition
				if err := d.Parse(invalid); err == nil {
					t.Error("should error")
				}
			})

		}
	})
}
