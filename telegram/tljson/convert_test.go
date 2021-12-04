package tljson

import (
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestDecodeEncodeDecode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		expect  tg.JSONValueClass
	}{
		{"Empty", "", true, nil},
		{"InvalidNull", "nul", true, nil},
		{"InvalidTrue", "tru", true, nil},
		{"InvalidFalse", "falsy", true, nil},
		{"InvalidInt", `[1a]"`, true, nil},
		{"InvalidFloat", "1.", true, nil},
		{"InvalidString", `"hello`, true, nil},
		{"InvalidArray", "[1, 2, 3.]", true, nil},
		{"InvalidObject", `{"abc":"def}`, true, nil},

		{"Null", "null", false, nil},
		{"True", "true", false, nil},
		{"False", "false", false, nil},
		{"Int", "10", false, nil},
		{"Float", "1.1", false, nil},
		{"String", `"hello"`, false, nil},
		{"EmptyArray", "[]", false, nil},
		{"Array", "[1, 2, 3]", false, nil},
		{"EmptyObject", `{}`, false, nil},
		{"Object", `{"abc":"def"}`, false, nil},
		{"Tree", `{"a":1,"b":true,"c":null,"sub":{"abc":"def"}}`, false, nil},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			a := require.New(t)

			// Decode.
			d := jx.DecodeStr(tt.input)
			obj, err := Decode(d)
			if tt.wantErr {
				a.Error(err)
				return
			}
			a.NoError(err)

			// Encode.
			e := jx.GetEncoder()
			Encode(obj, e)

			// Decode.
			d.ResetBytes(e.Bytes())
			obj2, err := Decode(d)
			a.NoError(err)
			a.Equal(obj, obj2)
		})
	}
}
