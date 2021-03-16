package dcs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStagingDCs(t *testing.T) {
	require.NotEmpty(t, ProdDCs())

	// Check copying.
	a := StagingDCs()
	a[0].IPAddress = "10"
	b := StagingDCs()
	require.NotEqual(t, "10", b[0].IPAddress)
}
