package telegram

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"

	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/tg"
)

func TestError(t *testing.T) {
	// Ensure that code generation works for errors.
	err := func() error {
		rpcErr := &mtproto.Error{
			Type: tg.ErrAccessTokenExpired,
		}

		return xerrors.Errorf("failed to perform operation: %w", rpcErr)
	}()
	require.True(t, mtproto.IsErr(err, tg.ErrAccessTokenExpired))
	require.True(t, tg.IsAccessTokenExpired(err))
}
