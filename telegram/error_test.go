package telegram

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
}

func TestAsFloodWait(t *testing.T) {
	err := func() error {
		return xerrors.Errorf("failed to perform operation: %w",
			mtproto.NewError(400, "FLOOD_WAIT_10"),
		)
	}()

	d, ok := AsFloodWait(err)
	assert.True(t, ok)
	assert.Equal(t, time.Second*10, d)
}
