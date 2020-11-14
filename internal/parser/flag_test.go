package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlag(t *testing.T) {
	for _, tt := range []struct {
		String string
		Flag   Flag
	}{
		{
			String: "flags.1",
			Flag:   Flag{Index: 1, Name: "flags"},
		},
	} {
		t.Run(tt.String, func(t *testing.T) {
			t.Run("String", func(t *testing.T) {
				if v := tt.Flag.String(); v != tt.String {
					t.Errorf("(%s).String = %s", tt.String, v)
				}
			})
			t.Run("Parse", func(t *testing.T) {
				var result Flag
				if err := result.Parse(tt.String); err != nil {
					t.Fatal(err)
				}
				require.Equal(t, result, tt.Flag)
			})
		})
	}
	t.Run("Error", func(t *testing.T) {
		for _, s := range []string{
			".1",
			"flag",
			"",
			"foo.bar",
		} {
			var f Flag
			if err := f.Parse(s); err == nil {
				t.Errorf("Expected error on %q", s)
			}
		}
	})
}
