package rpc

import (
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
)

func TestNotifyErrorBefore(t *testing.T) {
	e := New(NopSend, Options{})
	got := map[int64]error{}
	for _, id := range []int64{10, 20, 30} {
		e.rpc[id] = func(_ *bin.Buffer, err error) error {
			got[id] = err
			return nil
		}
	}
	reset := errors.New("reset")
	e.NotifyErrorBefore(25, reset)
	require.Equal(t, map[int64]error{10: reset, 20: reset}, got)
}
