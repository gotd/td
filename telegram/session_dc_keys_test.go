package telegram

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/gotd/log"
	"github.com/gotd/log/logzap"

	"github.com/gotd/td/crypto"
	"github.com/gotd/td/pool"
	"github.com/gotd/td/session"
	"github.com/gotd/td/tdsync"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgtest"
	"github.com/gotd/td/tgtest/cluster"
	"github.com/gotd/td/transport"
)

// Keys of the other DCs are persisted with their authorization, so after a
// restart a pool to such DC skips key exchange and auth.importAuthorization.
// The self ID is persisted too.
func TestClientRestoresAuthorizedDCKeys(t *testing.T) {
	const (
		homeDC  = 2
		otherDC = 1
	)
	zapLog := zaptest.NewLogger(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	c := cluster.NewCluster(cluster.Options{
		Logger:   logzap.New(zapLog.Named("cluster")),
		Protocol: transport.Intermediate,
	})
	var imports atomic.Int64
	c.Dispatch(homeDC, "home").
		HandleFunc(tg.AuthExportAuthorizationRequestTypeID,
			func(server *tgtest.Server, req *tgtest.Request) error {
				var r tg.AuthExportAuthorizationRequest
				if err := r.Decode(req.Buf); err != nil {
					return err
				}
				return server.SendResult(req, &tg.AuthExportedAuthorization{ID: int64(r.DCID), Bytes: []byte{1}})
			},
		).
		HandleFunc(tg.UsersGetUsersRequestTypeID,
			func(server *tgtest.Server, req *tgtest.Request) error {
				return server.SendVector(req, &tg.User{ID: 10, Bot: true})
			},
		)
	c.Dispatch(otherDC, "other").
		HandleFunc(tg.AuthImportAuthorizationRequestTypeID,
			func(server *tgtest.Server, req *tgtest.Request) error {
				imports.Add(1)
				return server.SendResult(req, &tg.AuthAuthorization{User: &tg.User{ID: 10}})
			},
		).
		HandleFunc(tg.UsersGetUsersRequestTypeID,
			func(server *tgtest.Server, req *tgtest.Request) error {
				return server.SendVector(req, &tg.User{ID: 10})
			},
		)

	storage := &session.StorageMemory{}
	var firstKey [8]byte
	run := func(ctx context.Context, name string, check func(client *Client) error) error {
		client := NewClient(1, "hash", Options{
			PublicKeys:     c.Keys(),
			Resolver:       c.Resolver(),
			DCList:         c.List(),
			Logger:         logzap.New(zapLog.Named(name)),
			SessionStorage: storage,
			NoUpdates:      true,
		})
		return client.Run(ctx, func(ctx context.Context) error {
			if err := check(client); err != nil {
				return err
			}
			p, err := client.DC(ctx, otherDC, 1)
			if err != nil {
				return errors.Wrap(err, "DC")
			}
			defer func() { _ = p.Close() }()
			if _, err := tg.NewClient(p).UsersGetUsers(ctx, []tg.InputUserClass{&tg.InputUserSelf{}}); err != nil {
				return errors.Wrap(err, "invoke")
			}
			client.sessionsMux.Lock()
			key := client.sessions[otherDC].Load().AuthKey.ID
			client.sessionsMux.Unlock()
			if name == "first" {
				firstKey = key
			} else if key != firstKey {
				return errors.New("other DC key was not reused")
			}
			return nil
		})
	}

	g := tdsync.NewCancellableGroup(ctx)
	g.Go(c.Up)
	g.Go(func(ctx context.Context) error {
		defer cancel()
		select {
		case <-c.Ready():
		case <-ctx.Done():
			return ctx.Err()
		}
		if err := run(ctx, "first", func(client *Client) error {
			// Run doesn't call Self in NoUpdates mode.
			_, err := client.Self(ctx)
			return err
		}); err != nil {
			return errors.Wrap(err, "first")
		}
		if n := imports.Load(); n != 1 {
			return errors.Errorf("expected 1 import, got %d", n)
		}
		data, err := (&session.Loader{Storage: storage}).Load(ctx)
		if err != nil {
			return errors.Wrap(err, "load")
		}
		if k, ok := data.DCs[otherDC]; !ok || !k.Authorized {
			return errors.Errorf("other DC key not stored: %+v", data.DCs)
		}
		if data.UserID != 10 || !data.Bot {
			return errors.Errorf("self not stored: %d %v", data.UserID, data.Bot)
		}

		if err := run(ctx, "second", func(client *Client) error {
			if id, bot, ok := client.SelfID(); !ok || id != 10 || !bot {
				return errors.Errorf("SelfID not restored: %d %v %v", id, bot, ok)
			}
			return nil
		}); err != nil {
			return errors.Wrap(err, "second")
		}
		if n := imports.Load(); n != 1 {
			return errors.Errorf("restored key was transferred again: %d imports", n)
		}
		return nil
	})

	err := g.Wait()
	if errors.Is(err, context.Canceled) {
		err = nil
	}
	require.NoError(t, err)
}

// A DC key is persisted only while it is marked authorized; a 401 on a cached
// pool clears the mark, so the next pool transfers the authorization again.
func TestClientDCAuthorizedMark(t *testing.T) {
	a := require.New(t)
	client := &Client{log: log.For(log.Nop)}
	client.init()
	var key crypto.AuthKey
	key.Value[0] = 1
	key = key.Value.WithID()
	client.sessions[4] = pool.NewSyncSession(pool.Session{DC: 4, AuthKey: key, Salt: 3})

	a.False(client.dcKeyAuthorized(4))
	client.setDCAuthorized(context.Background(), 4, true)
	a.True(client.dcKeyAuthorized(4))
	data := &session.Data{DC: 2}
	client.fillDCKeys(data)
	a.Equal(map[int]session.DCKey{4: {
		AuthKey: key.Value[:], AuthKeyID: key.ID[:], Authorized: true, Salt: 3,
	}}, data.DCs)

	client.setDCAuthorized(context.Background(), 4, false)
	a.False(client.dcKeyAuthorized(4))
	client.fillDCKeys(data)
	a.Nil(data.DCs)

	// A new key of the DC is not authorized by the old mark.
	client.setDCAuthorized(context.Background(), 4, true)
	var other crypto.AuthKey
	other.Value[0] = 2
	client.sessions[4].Store(pool.Session{DC: 4, AuthKey: other.Value.WithID()})
	a.False(client.dcKeyAuthorized(4))
}
