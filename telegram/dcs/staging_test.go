package dcs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStagingDCs(t *testing.T) {
	require.NotEmpty(t, Prod())

	// Check copying.
	a := Staging().Options
	a[0].IPAddress = "10"
	b := Staging().Options
	require.NotEqual(t, "10", b[0].IPAddress)
}
