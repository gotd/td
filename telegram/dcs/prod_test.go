package dcs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProd(t *testing.T) {
	require.NotEmpty(t, ProdDCs())

	// Check copying.
	a := ProdDCs().Options
	a[0].IPAddress = "10"
	b := ProdDCs().Options
	require.NotEqual(t, "10", b[0].IPAddress)
}
