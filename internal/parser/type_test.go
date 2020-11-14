package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParameterType(t *testing.T) {
	for _, tt := range []struct {
		Type   Type
		String string
	}{
		{
			Type:   Type{Name: "Ok"},
			String: "Ok",
		},
		{
			Type:   Type{Name: "foo", Bare: true},
			String: "foo",
		},
		{
			Type:   Type{Name: "baz", Bare: true, Namespace: []string{"foo", "bar"}},
			String: "foo.bar.baz",
		},
		{
			Type: Type{
				Name:      "Vec",
				Namespace: []string{"basic"},
				GenericArg: &Type{
					Name:      "T",
					Namespace: []string{"generic"},
				},
			},
			String: "basic.Vec<generic.T>",
		},
		{
			Type: Type{
				Name:       "X",
				Namespace:  []string{"foo"},
				GenericRef: true,
			},
			String: "!foo.X",
		},
	} {
		t.Run(tt.String, func(t *testing.T) {
			t.Run("String", func(t *testing.T) {
				if v := tt.Type.String(); v != tt.String {
					t.Errorf("(%s).String = %s", tt.String, v)
				}
			})
			t.Run("Parse", func(t *testing.T) {
				var result Type
				if err := result.Parse(tt.String); err != nil {
					t.Fatal(err)
				}
				require.Equal(t, result, tt.Type)
			})
		})
	}
}
