package messages_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/gotd/log/logzap"
	"github.com/gotd/td/session"
	"github.com/gotd/td/tdsync"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgtest"
	"github.com/gotd/td/tgtest/cluster"
	"github.com/gotd/td/tgtest/services"
	"github.com/gotd/td/tgtest/services/config"
	"github.com/gotd/td/tgtest/services/messages"
)

func newClient(c *cluster.Cluster, opts telegram.Options) *telegram.Client {
	opts.PublicKeys = c.Keys()
	opts.DC = 2
	opts.DCList = c.List()
	opts.Resolver = c.Resolver()
	opts.SessionStorage = &session.StorageMemory{}
	opts.RetryInterval = 100 * time.Millisecond
	return telegram.NewClient(telegram.TestAppID, telegram.TestAppHash, opts)
}

// TestSendAndHistory checks that a sent message is echoed back in the response
// and is readable via messages.getHistory.
func TestSendAndHistory(t *testing.T) {
	a := require.New(t)
	log := zaptest.NewLogger(t)
	defer func() { _ = log.Sync() }()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	g := tdsync.NewCancellableGroup(ctx)
	c := cluster.NewCluster(cluster.Options{Logger: logzap.New(log.Named("cluster"))})

	svc := messages.NewService(messages.WithSelfResolver(func(tgtest.Session) *tg.User {
		return &tg.User{ID: 100, AccessHash: 100, Self: true}
	}))
	d := c.Dispatch(2, "server").Fallback(services.NotImplemented)
	config.NewService(&tg.Config{}, &tg.CDNConfig{}).Register(d)
	svc.Register(d)

	g.Go(c.Up)
	g.Go(func(ctx context.Context) error {
		select {
		case <-c.Ready():
		case <-ctx.Done():
			return ctx.Err()
		}
		defer g.Cancel()

		client := newClient(c, telegram.Options{NoUpdates: true, Logger: logzap.New(log.Named("client"))})
		return client.Run(ctx, func(ctx context.Context) error {
			api := client.API()

			peer := &tg.InputPeerUser{UserID: 200, AccessHash: 200}
			upd, err := api.MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
				Peer:     peer,
				Message:  "hello",
				RandomID: 42,
			})
			if err != nil {
				return errors.Wrap(err, "send")
			}

			u, ok := upd.(*tg.Updates)
			a.Truef(ok, "unexpected updates type %T", upd)
			a.Len(u.Updates, 2)
			mid, ok := u.Updates[0].(*tg.UpdateMessageID)
			a.True(ok)
			a.Equal(int64(42), mid.RandomID)
			nm, ok := u.Updates[1].(*tg.UpdateNewMessage)
			a.True(ok)
			sent := nm.Message.(*tg.Message)
			a.Equal("hello", sent.Message)
			a.True(sent.Out)
			a.Equal(mid.ID, sent.ID)

			hist, err := api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{Peer: peer})
			if err != nil {
				return errors.Wrap(err, "history")
			}
			m, ok := hist.(*tg.MessagesMessages)
			a.Truef(ok, "unexpected messages type %T", hist)
			a.Len(m.Messages, 1)
			a.Equal("hello", m.Messages[0].(*tg.Message).Message)

			return nil
		})
	})

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		a.NoError(err)
	}
}

// TestDeliver checks that a message sent to a connected recipient is delivered
// as an UpdateNewMessage.
func TestDeliver(t *testing.T) {
	a := require.New(t)
	log := zaptest.NewLogger(t)
	defer func() { _ = log.Sync() }()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	const (
		aliceID int64 = 100
		bobID   int64 = 200
	)
	users := map[[8]byte]int64{}
	var resolved []int64

	g := tdsync.NewCancellableGroup(ctx)
	c := cluster.NewCluster(cluster.Options{Logger: logzap.New(log.Named("cluster"))})

	svc := messages.NewService(messages.WithSelfResolver(func(s tgtest.Session) *tg.User {
		// Assign identities in connection order: Bob connects (binds) first,
		// then Alice.
		id, ok := users[s.AuthKey.ID]
		if !ok {
			next := []int64{bobID, aliceID}[len(resolved)]
			users[s.AuthKey.ID] = next
			resolved = append(resolved, next)
			id = next
		}
		return &tg.User{ID: id, AccessHash: id, Self: true}
	}))
	d := c.Dispatch(2, "server").Fallback(services.NotImplemented)
	config.NewService(&tg.Config{}, &tg.CDNConfig{}).Register(d)
	svc.Register(d)

	g.Go(c.Up)

	bobReady := make(chan struct{})
	bobGot := make(chan string, 1)

	// Bob: connects first, binds itself, then waits for an incoming message.
	g.Go(func(ctx context.Context) error {
		select {
		case <-c.Ready():
		case <-ctx.Done():
			return ctx.Err()
		}

		dispatcher := tg.NewUpdateDispatcher()
		dispatcher.OnNewMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateNewMessage) error {
			select {
			case bobGot <- u.Message.(*tg.Message).Message:
			default:
			}
			return nil
		})
		client := newClient(c, telegram.Options{
			UpdateHandler: dispatcher,
			Logger:        logzap.New(log.Named("bob")),
		})
		return client.Run(ctx, func(ctx context.Context) error {
			// Trigger session binding on the server.
			if _, err := client.API().MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
				Peer: &tg.InputPeerSelf{},
			}); err != nil {
				return errors.Wrap(err, "bob bind")
			}
			close(bobReady)

			select {
			case msg := <-bobGot:
				a.Equal("hi bob", msg)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	})

	// Alice: connects after Bob is ready and sends a message to Bob.
	g.Go(func(ctx context.Context) error {
		select {
		case <-bobReady:
		case <-ctx.Done():
			return ctx.Err()
		}
		defer g.Cancel()

		client := newClient(c, telegram.Options{NoUpdates: true, Logger: logzap.New(log.Named("alice"))})
		return client.Run(ctx, func(ctx context.Context) error {
			if _, err := client.API().MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
				Peer:     &tg.InputPeerUser{UserID: bobID, AccessHash: bobID},
				Message:  "hi bob",
				RandomID: 1,
			}); err != nil {
				return errors.Wrap(err, "alice send")
			}
			return nil
		})
	})

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		a.NoError(err)
	}
}
