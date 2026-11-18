package telegram

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/pool"
	"github.com/gotd/td/telegram/internal/manager"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// Connections of a non-home DC pool must carry the test DC flag like the
// primary connection does, otherwise the key exchange with a test server
// fails with a wrong DC id.
func TestClientPoolPassesTestDC(t *testing.T) {
	ctx := context.Background()

	for _, test := range []bool{false, true} {
		name := "Production"
		if test {
			name = "Test"
		}
		t.Run(name, func(t *testing.T) {
			client := newMigrationClient(t, func(id int64, dc int, body bin.Encoder) (bin.Encoder, error) {
				switch body.(type) {
				case *tg.AuthExportAuthorizationRequest:
					return &tg.AuthExportedAuthorization{ID: 1, Bytes: []byte{1}}, nil
				case *tg.AuthImportAuthorizationRequest:
					return &tg.AuthAuthorization{User: &tg.User{ID: 1}}, nil
				case *tg.HelpGetNearestDCRequest:
					return &tg.NearestDC{ThisDC: dc}, nil
				default:
					return nil, tgerr.New(400, "UNEXPECTED")
				}
			})
			client.testDC = test

			var (
				mux sync.Mutex
				got []bool
			)
			create := client.create
			client.create = func(
				dialer mtproto.Dialer,
				mode manager.ConnMode,
				appID int,
				opts mtproto.Options,
				connOpts manager.ConnOptions,
			) pool.Conn {
				mux.Lock()
				got = append(got, connOpts.Test)
				mux.Unlock()
				return create(dialer, mode, appID, opts, connOpts)
			}

			var raw crypto.Key
			raw[0] = 42
			client.session.Store(pool.Session{DC: 2, AuthKey: raw.WithID()})
			client.onTransfer = noopOnTransfer

			for _, open := range []func() (CloseInvoker, error){
				func() (CloseInvoker, error) { return client.DC(ctx, 10, 1) },
				func() (CloseInvoker, error) { return client.MediaOnly(ctx, 10, 1) },
			} {
				inv, err := open()
				require.NoError(t, err)

				_, err = tg.NewClient(inv).HelpGetNearestDC(ctx)
				require.NoError(t, err)
				require.NoError(t, inv.Close())
			}

			mux.Lock()
			defer mux.Unlock()
			require.NotEmpty(t, got)
			for _, v := range got {
				require.Equal(t, test, v, "pool connection test DC flag")
			}
		})
	}
}
