package tdp_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tdp"
	"github.com/nnqq/td/tdp/internal/schema"
)

func TestFormat(t *testing.T) {
	for _, tt := range []struct {
		Input   tdp.Object
		Output  string
		Options []tdp.Option
	}{
		{
			Output: "<nil>",
		},
		{
			Input: &schema.DCOption{
				ID:        10,
				IPAddress: "127.0.0.1",
				Port:      1010,
			},
			Options: []tdp.Option{tdp.WithTypeID},
			Output:  "dcOption#18b7a10d",
		},

		{
			Input: &schema.Config{
				DCOptions: []schema.DCOption{
					{
						ID:        1,
						IPAddress: "127.0.0.1",
						Port:      1010,
					},
				},
			},
			Options: []tdp.Option{tdp.WithTypeID},
		},
	} {
		t.Skip("TODO: Use golden files")
		t.Run(tt.Output, func(t *testing.T) {
			require.Equal(t, tt.Output, tdp.Format(tt.Input, tt.Options...))
		})
	}
}
