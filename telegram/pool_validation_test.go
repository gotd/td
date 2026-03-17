package telegram

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPoolMethodValidation(t *testing.T) {
	c := &Client{}

	_, err := c.Pool(-1)
	require.ErrorContains(t, err, "invalid max value -1")

	_, err = c.DC(context.Background(), 2, -1)
	require.ErrorContains(t, err, "invalid max value -1")

	_, err = c.MediaOnly(context.Background(), 2, -1)
	require.ErrorContains(t, err, "invalid max value -1")

	_, err = c.CDN(context.Background(), 203, -1)
	require.ErrorContains(t, err, "invalid max value -1")
}
