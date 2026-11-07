package telegram

import (
	"context"
	"sync"
	"sync/atomic"
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

func TestClientDCHomeChanged(t *testing.T) {
	ctx := context.Background()
	a := require.New(t)

	var (
		block   = make(chan struct{})
		started = make(chan struct{})
		first   atomic.Bool
	)
	client := newMigrationClient(t, func(id int64, dc int, body bin.Encoder) (bin.Encoder, error) {
		switch body.(type) {
		case *tg.HelpGetNearestDCRequest:
			// First request holds its connection, so the next one
			// needs a new connection.
			if first.CompareAndSwap(false, true) {
				close(started)
				<-block
			}
			return &tg.NearestDC{ThisDC: dc}, nil
		default:
			return &tg.AuthExportedAuthorization{ID: 1, Bytes: []byte{1}}, nil
		}
	})
	client.onTransfer = noopOnTransfer

	var (
		mux   sync.Mutex
		conns []mtproto.Options
		dcs   []int
	)
	create := client.create
	client.create = func(
		dialer mtproto.Dialer, mode manager.ConnMode, appID int,
		opts mtproto.Options, connOpts manager.ConnOptions,
	) pool.Conn {
		mux.Lock()
		conns = append(conns, opts)
		dcs = append(dcs, connOpts.DC)
		mux.Unlock()
		return create(dialer, mode, appID, opts, connOpts)
	}

	var raw2, raw4 crypto.Key
	raw2[0] = 2
	raw4[0] = 4
	key2, key4 := raw2.WithID(), raw4.WithID()
	client.session.Store(pool.Session{DC: 2, AuthKey: key2})

	inv, err := client.DC(ctx, 2, 2)
	a.NoError(err)
	defer func() { _ = inv.Close() }()

	errs := make(chan error, 1)
	go func() {
		_, err := tg.NewClient(inv).HelpGetNearestDC(ctx)
		errs <- err
	}()
	<-started

	// Home DC changes while the pool is alive.
	client.session.Store(pool.Session{DC: 4, AuthKey: key4})

	r, err := tg.NewClient(inv).HelpGetNearestDC(ctx)
	a.NoError(err)
	a.Equal(2, r.ThisDC)
	close(block)
	a.NoError(<-errs)

	mux.Lock()
	defer mux.Unlock()
	a.Len(conns, 2)
	a.Equal([]int{2, 2}, dcs, "pool connections must dial DC 2")
	a.Equal(key2.ID, conns[0].Key.ID, "first connection uses primary key")
	a.NotEqual(key4.ID, conns[1].Key.ID, "DC 2 connection must not use DC 4 key")
}
