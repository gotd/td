package telegram

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDownloadClientCDNCloserKeepsSharedPoolCached(t *testing.T) {
	c := newCDNPoolTestClient()
	defer c.cancel()

	d := downloadClient{client: c}
	_, closer, err := d.CDN(context.Background(), 203, 1)
	require.NoError(t, err)
	require.NotNil(t, closer)

	c.cdnPools.mux.Lock()
	require.EqualValues(t, 1, len(c.cdnPools.refs))
	require.EqualValues(t, 1, len(c.cdnPools.conns[203]))
	c.cdnPools.mux.Unlock()

	require.NoError(t, closer.Close())

	c.cdnPools.mux.Lock()
	require.EqualValues(t, 1, len(c.cdnPools.refs))
	require.EqualValues(t, 1, len(c.cdnPools.conns[203]))
	c.cdnPools.mux.Unlock()

	_, closer2, err := d.CDN(context.Background(), 203, 1)
	require.NoError(t, err)
	require.NoError(t, closer2.Close())
}
