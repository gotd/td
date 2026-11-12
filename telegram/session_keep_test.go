package telegram

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/gotd/log/logzap"

	"github.com/gotd/td/session"
	"github.com/gotd/td/tdsync"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgtest"
	"github.com/gotd/td/tgtest/cluster"
	"github.com/gotd/td/transport"
)

// After a reconnect of the primary connection with the same key the client
// continues the server session: same session_id, no new_session_created, no
// update recovery needed.
func TestClientKeepsSessionAcrossReconnect(t *testing.T) {
	const dcID = 2
	zapLog := zaptest.NewLogger(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	c := cluster.NewCluster(cluster.Options{
		Logger:   logzap.New(zapLog.Named("cluster")),
		Protocol: transport.Intermediate,
	})
	var (
		mux      sync.Mutex
		sessions []int64
	)
	c.Dispatch(dcID, "dc").HandleFunc(tg.UsersGetUsersRequestTypeID,
		func(server *tgtest.Server, req *tgtest.Request) error {
			mux.Lock()
			sessions = append(sessions, req.Session.ID)
			mux.Unlock()
			return server.SendVector(req, &tg.User{ID: 10})
		},
	)

	var tooLong atomic.Int64
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
			UpdateHandler: UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
				if _, ok := u.(*tg.UpdatesTooLong); ok {
					tooLong.Add(1)
				}
				return nil
			}),
		})
		return client.Run(ctx, func(ctx context.Context) error {
			if _, err := client.API().UsersGetUsers(ctx, []tg.InputUserClass{&tg.InputUserSelf{}}); err != nil {
				return errors.Wrap(err, "first")
			}
			rctx, rcancel := context.WithTimeout(ctx, 10*time.Second)
			defer rcancel()
			if err := client.ensureRestart(rctx, restartTimeout); err != nil {
				return errors.Wrap(err, "restart")
			}
			if _, err := client.API().UsersGetUsers(ctx, []tg.InputUserClass{&tg.InputUserSelf{}}); err != nil {
				return errors.Wrap(err, "after reconnect")
			}
			return nil
		})
	})
	err := g.Wait()
	if errors.Is(err, context.Canceled) {
		err = nil
	}
	require.NoError(t, err)

	mux.Lock()
	defer mux.Unlock()
	require.GreaterOrEqual(t, len(sessions), 2)
	for _, id := range sessions[1:] {
		require.Equal(t, sessions[0], id, "session_id changed after reconnect")
	}
	require.Zero(t, tooLong.Load(), "no update recovery for a continued session")
}

// When the server starts a new session on the primary connection (it lost the
// previous one), updates sent to the old session may be lost: the update
// handler gets updatesTooLong once, but not for the first session.
func TestClientNewSessionRecoversUpdates(t *testing.T) {
	const dcID = 2
	zapLog := zaptest.NewLogger(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	c := cluster.NewCluster(cluster.Options{
		Logger:   logzap.New(zapLog.Named("cluster")),
		Protocol: transport.Intermediate,
	})
	var calls atomic.Int64
	c.Dispatch(dcID, "dc").HandleFunc(tg.UsersGetUsersRequestTypeID,
		func(server *tgtest.Server, req *tgtest.Request) error {
			if err := server.SendVector(req, &tg.User{ID: 10}); err != nil {
				return err
			}
			if calls.Add(1) == 1 {
				// The server forgets the session: the next message starts a
				// new one.
				server.ForgetSession(req.Session)
			}
			return nil
		},
	)

	var tooLong atomic.Int64
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
			NoUpdates:      false,
			UpdateHandler: UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
				if _, ok := u.(*tg.UpdatesTooLong); ok {
					tooLong.Add(1)
				}
				return nil
			}),
		})
		return client.Run(ctx, func(ctx context.Context) error {
			rctx, rcancel := context.WithTimeout(ctx, 20*time.Second)
			defer rcancel()
			for tooLong.Load() == 0 {
				if _, err := client.API().UsersGetUsers(rctx, []tg.InputUserClass{&tg.InputUserSelf{}}); err != nil {
					return errors.Wrap(err, "request")
				}
				select {
				case <-rctx.Done():
					return errors.Wrap(rctx.Err(), "no updatesTooLong")
				case <-time.After(50 * time.Millisecond):
				}
			}
			return nil
		})
	})
	err := g.Wait()
	if errors.Is(err, context.Canceled) {
		err = nil
	}
	require.NoError(t, err)
	require.Equal(t, int64(1), tooLong.Load())
}
