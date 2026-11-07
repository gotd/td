package telegram

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/pool"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

func TestClientDCHomeUsesPrimarySession(t *testing.T) {
	ctx := context.Background()
	a := require.New(t)

	var (
		mux      sync.Mutex
		requests []string
	)
	client := newMigrationClient(t, func(id int64, dc int, body bin.Encoder) (bin.Encoder, error) {
		mux.Lock()
		defer mux.Unlock()
		switch body.(type) {
		case *tg.AuthExportAuthorizationRequest:
			requests = append(requests, "export")
			// Telegram rejects export to the current DC.
			return nil, tgerr.New(400, "DC_ID_INVALID")
		case *tg.AuthImportAuthorizationRequest:
			requests = append(requests, "import")
			return &tg.AuthAuthorization{User: &tg.User{ID: 1}}, nil
		case *tg.HelpGetNearestDCRequest:
			requests = append(requests, "nearest")
			return &tg.NearestDC{ThisDC: dc}, nil
		default:
			return nil, tgerr.New(400, "UNEXPECTED")
		}
	})

	var raw crypto.Key
	raw[0] = 42
	key := raw.WithID()
	client.session.Store(pool.Session{DC: 2, AuthKey: key})
	client.onTransfer = noopOnTransfer

	for _, tt := range []struct {
		name string
		open func() (CloseInvoker, error)
	}{
		{"DC", func() (CloseInvoker, error) { return client.DC(ctx, 2, 1) }},
		{"MediaOnly", func() (CloseInvoker, error) { return client.MediaOnly(ctx, 2, 1) }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			inv, err := tt.open()
			require.NoError(t, err)
			defer func() { _ = inv.Close() }()

			r, err := tg.NewClient(inv).HelpGetNearestDC(ctx)
			require.NoError(t, err)
			require.Equal(t, 2, r.ThisDC)
		})
	}

	mux.Lock()
	a.Equal([]string{"nearest", "nearest"}, requests, "no authorization transfer to home DC")
	mux.Unlock()

	// Pool connections report their session back to the client, so
	// a separate zero-key session would replace the primary key.
	a.Equal(key, client.session.Load().AuthKey, "pool must use primary auth key")
}
