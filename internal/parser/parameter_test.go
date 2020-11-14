package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParameter(t *testing.T) {
	for _, tt := range []struct {
		String string
		Value  Parameter
	}{
		{
			String: "flags:#",
			Value:  Parameter{Name: "flags", Flags: true},
		},
		{
			String: "code:int32",
			Value:  Parameter{Name: "code", Type: Type{Name: "int32", Bare: true}},
		},
		{
			String: "status_code:f.1?Code",
			Value: Parameter{
				Name: "status_code",
				Flag: Flag{Name: "f", Index: 1},
				Type: Type{Name: "Code"},
			},
		},
		{
			String: "int",
			Value: Parameter{
				Name: "",
				Type: Type{Name: "int", Bare: true},
			},
		},
	} {
		t.Run(tt.String, func(t *testing.T) {
			t.Run("String", func(t *testing.T) {
				if v := tt.Value.String(); v != tt.String {
					t.Errorf("(%s).String = %s", tt.String, v)
				}
			})
			t.Run("Parse", func(t *testing.T) {
				var result Parameter
				if err := result.Parse(tt.String); err != nil {
					t.Fatal(err)
				}
				require.Equal(t, result, tt.Value)
			})
		})
	}
	t.Run("Error", func(t *testing.T) {
		for _, s := range []string{
			".1",
			"",
			"{a:b}",
			"{c",
		} {
			var value Parameter
			if err := value.Parse(s); err == nil {
				t.Errorf("Expected error on %q", s)
			}
		}
	})
	t.Run("Conditional", func(t *testing.T) {
		for _, conditional := range []Parameter{
			{
				Name: "Foo",
				Flag: Flag{Name: "flags", Index: 1},
			},
		} {
			t.Run(conditional.Name, func(t *testing.T) {
				if !conditional.Conditional() {
					t.Error("expected conditional")
				}
			})
		}
		for _, nonConditional := range []Parameter{
			{
				Name: "Bar",
				Flag: Flag{},
			},
		} {
			t.Run(nonConditional.Name, func(t *testing.T) {
				if nonConditional.Conditional() {
					t.Error("expected non-conditional")
				}
			})
		}
	})
}
