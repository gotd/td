package telegram

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/gotd/log/logzap"

	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/session"
	"github.com/gotd/td/tdsync"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgtest"
	"github.com/gotd/td/tgtest/cluster"
	"github.com/gotd/td/transport"
)

// TestDCTransferSetupAfterKeyDrop checks that a new connection of a non-home
// DC pool, created after the DC key was dropped (ErrPFSDropKeysRequired),
// transfers authorization in its setup and becomes ready.
//
// Previously setup invoked auth.importAuthorization through the connection
// before it was ready, and the connection deadlocked on itself.
func TestDCTransferSetupAfterKeyDrop(t *testing.T) {
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
	c.Dispatch(homeDC, "home").HandleFunc(tg.AuthExportAuthorizationRequestTypeID,
		func(server *tgtest.Server, req *tgtest.Request) error {
			var r tg.AuthExportAuthorizationRequest
			if err := r.Decode(req.Buf); err != nil {
				return err
			}
			return server.SendResult(req, &tg.AuthExportedAuthorization{
				ID:    int64(r.DCID),
				Bytes: []byte{1},
			})
		},
	)

	// The first users.getUsers on otherDC blocks to keep the first pool
	// connection busy, so the pool has to create the second one.
	blocked := make(chan struct{})
	release := make(chan struct{})
	var getUsers atomic.Int64
	c.Dispatch(otherDC, "other").
		HandleFunc(tg.AuthImportAuthorizationRequestTypeID,
			func(server *tgtest.Server, req *tgtest.Request) error {
				imports.Add(1)
				return server.SendResult(req, &tg.AuthAuthorization{User: &tg.User{ID: 10}})
			},
		).
		HandleFunc(tg.UsersGetUsersRequestTypeID,
			func(server *tgtest.Server, req *tgtest.Request) error {
				if getUsers.Add(1) == 1 {
					close(blocked)
					select {
					case <-release:
					case <-req.RequestCtx.Done():
						return req.RequestCtx.Err()
					}
				}
				return server.SendVector(req, &tg.User{ID: 10})
			},
		)

	g := tdsync.NewCancellableGroup(ctx)
	g.Go(c.Up)
	g.Go(func(ctx context.Context) error {
		defer cancel()
		select {
		case <-c.Ready():
		case <-ctx.Done():
			return ctx.Err()
		}

		client := NewClient(1, "hash", Options{
			PublicKeys:     c.Keys(),
			Resolver:       c.Resolver(),
			DCList:         c.List(),
			Logger:         logzap.New(zapLog.Named("client")),
			SessionStorage: &session.StorageMemory{},
			NoUpdates:      true,
		})
		return client.Run(ctx, func(ctx context.Context) error {
			p, err := client.DC(ctx, otherDC, 2)
			if err != nil {
				return errors.Wrap(err, "DC")
			}
			defer func() { _ = p.Close() }()
			if n := imports.Load(); n != 1 {
				return errors.Errorf("expected 1 import after DC(), got %d", n)
			}

			// Server asked to drop keys: next pool connection generates a
			// new key and must transfer authorization in setup.
			client.handleDCConnDead(otherDC, mtproto.ErrPFSDropKeysRequired)

			api := tg.NewClient(p)
			first := make(chan error, 1)
			go func() {
				_, err := api.UsersGetUsers(ctx, []tg.InputUserClass{&tg.InputUserSelf{}})
				first <- err
			}()
			select {
			case <-blocked:
			case <-ctx.Done():
				return ctx.Err()
			}

			// Timeout turns a deadlock into a failure.
			reqCtx, reqCancel := context.WithTimeout(ctx, 10*time.Second)
			defer reqCancel()
			_, err = api.UsersGetUsers(reqCtx, []tg.InputUserClass{&tg.InputUserSelf{}})
			close(release)
			if err != nil {
				return errors.Wrap(err, "invoke via new connection")
			}
			if err := <-first; err != nil {
				return errors.Wrap(err, "invoke via first connection")
			}
			if n := imports.Load(); n != 2 {
				return errors.Errorf("expected 2 imports, got %d", n)
			}
			return nil
		})
	})

	err := g.Wait()
	if errors.Is(err, context.Canceled) {
		err = nil
	}
	require.NoError(t, err)
}
