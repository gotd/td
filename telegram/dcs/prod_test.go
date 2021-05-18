package dcs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProd(t *testing.T) {
	require.NotEmpty(t, Prod())

	// Check copying.
	a := Prod().Options
	a[0].IPAddress = "10"
	b := Prod().Options
	require.NotEqual(t, "10", b[0].IPAddress)
}
