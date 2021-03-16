package dcs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProd(t *testing.T) {
	require.NotEmpty(t, ProdDCs())

	// Check copying.
	a := ProdDCs()
	a[0].IPAddress = "10"
	b := ProdDCs()
	require.NotEqual(t, "10", b[0].IPAddress)
}
