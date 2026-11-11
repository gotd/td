package telegram_test

import (
	"context"
	"encoding/binary"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/mt"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
)

// Like Telegram Desktop, the client takes the server salt from
// new_session_created and bad_server_salt and never asks for future salts.
func TestClientNoFutureSalts(t *testing.T) {
	rec := &wireRecorder{}
	testCluster(transport.Intermediate, false, func(s clusterSetup) {
		s.Cluster.Common().Vector(tg.UsersGetUsersRequestTypeID, user)
		// Creates DC 2, the client's home DC.
		s.Cluster.Dispatch(2, "dc")
	}, func(ctx context.Context, c clientSetup) error {
		opts := c.Options
		opts.WireTracer = rec
		client := telegram.NewClient(1, "hash", opts)
		return client.Run(ctx, func(ctx context.Context) error {
			if _, err := client.Self(ctx); err != nil {
				return errors.Wrap(err, "self")
			}
			c.Complete()
			return nil
		})
	})(t)

	rec.mux.Lock()
	defer rec.mux.Unlock()
	out := 0
	for _, m := range rec.msgs {
		if m.m.Direction != mtproto.WireOut || len(m.body) < 4 {
			continue
		}
		out++
		require.NotEqual(t, uint32(mt.GetFutureSaltsRequestTypeID), binary.LittleEndian.Uint32(m.body),
			"get_future_salts sent")
	}
	require.Positive(t, out)
}
