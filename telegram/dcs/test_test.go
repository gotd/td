package dcs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTestDCs(t *testing.T) {
	require.NotEmpty(t, Prod())

	// Check copying.
	a := Test().Options
	a[0].IPAddress = "10"
	b := Test().Options
	require.NotEqual(t, "10", b[0].IPAddress)
}
