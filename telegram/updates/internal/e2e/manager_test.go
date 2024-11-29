//go:build linux

package e2e

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"golang.org/x/sync/errgroup"

	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
)

func TestE2E(t *testing.T) {
	testManager(t, func(s *server, storage updates.StateStorage) chan *tg.Updates {
		t.Helper()

		c := make(chan *tg.Updates, 10)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var (
			biba = s.peers.createUser("biba")
			boba = s.peers.createUser("boba")
			chat = s.peers.createChat("chat")
		)

		var channels []*tg.PeerChannel
		require.NoError(t, storage.ForEachChannels(ctx, 123, func(ctx context.Context, channelID int64, pts int) error {
			channels = append(channels, &tg.PeerChannel{
				ChannelID: channelID,
			})
			return nil
		}))

		var wg sync.WaitGroup
		wg.Add(2)

		// Biba.
		go func() {
			defer wg.Done()
			for i := 0; i < 3; i++ {
				c <- s.CreateEvent(func(ev *EventBuilder) {
					ev.SendMessage(biba, chat, fmt.Sprintf("biba-%d", i))

					for mi, c := range channels {
						ev.SendMessage(biba, c, fmt.Sprintf("biba-channel-%d-%d", i, mi))
					}
				})
			}
		}()

		// Boba.
		go func() {
			defer wg.Done()
			for i := 0; i < 3; i++ {
				c <- s.CreateEvent(func(ev *EventBuilder) {
					ev.SendMessage(boba, chat, fmt.Sprintf("boba-%d", i))

					for _, c := range channels {
						ev.SendMessage(boba, c, fmt.Sprintf("boba-channel-%d", i))
					}
				})
			}
		}()

		go func() {
			wg.Wait()
			close(c)
		}()
		return c
	})
}

func testManager(t *testing.T, f func(s *server, storage updates.StateStorage) chan *tg.Updates) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())

	var (
		log     = zaptest.NewLogger(t)
		s       = newServer()
		h       = newHandler()
		storage = newMemStorage()
		hasher  = newMemAccessHasher()
	)

	const uid = 123

	require.NoError(t, storage.SetState(ctx, uid, updates.State{
		Pts:  0,
		Qts:  0,
		Date: 0,
		Seq:  0,
	}))

	for i := 0; i < 2; i++ {
		c := s.peers.createChannel(fmt.Sprintf("channel-%d", i))
		require.NoError(t, storage.SetChannelPts(ctx, uid, c.ChannelID, 0))
		require.NoError(t, hasher.SetChannelAccessHash(ctx, uid, c.ChannelID, c.ChannelID*2))
	}

	e := updates.New(updates.Config{
		Handler:      h,
		Logger:       log.Named("gaps"),
		Storage:      storage,
		AccessHasher: hasher,
	})

	uchan := loss(f(s, storage))
	g, ctx := errgroup.WithContext(ctx)
	ready := make(chan struct{})
	opts := updates.AuthOptions{
		OnStart: func(ctx context.Context) {
			t.Log("OnStart")
			close(ready)
		},
	}
	g.Go(func() error {
		t.Log("Starting manager")
		defer t.Log("Manager stopped")
		return e.Run(ctx, s, uid, opts)
	})
	g.Go(func() error {
		t.Log("Starting updates generator")
		defer t.Log("Updates generator stopped")

		defer cancel()

		select {
		case <-ready:
			t.Log("Ready")
		case <-ctx.Done():
			return ctx.Err()
		}

		var g errgroup.Group
		for i := 0; i < 2; i++ {
			g.Go(func() error {
				for {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case u, ok := <-uchan:
						if !ok {
							return nil
						}

						if err := e.Handle(ctx, u); err != nil {
							return err
						}
					}
				}
			})
		}

		t.Log("Waiting")
		if err := g.Wait(); err != nil {
			return err
		}

		t.Log("Sending pts changed")

		ups := []tg.UpdateClass{&tg.UpdatePtsChanged{}}
		if err := storage.ForEachChannels(ctx, uid, func(ctx context.Context, channelID int64, pts int) error {
			ups = append(ups, &tg.UpdateChannelTooLong{ChannelID: channelID})
			return nil
		}); err != nil {
			return err
		}

		t.Log("Handle")

		return e.Handle(ctx, &tg.Updates{
			Updates: ups,
		})
	})

	t.Log("Waiting for shutdown")
	require.ErrorIs(t, g.Wait(), context.Canceled)

	t.Log("Checking")
	require.Equal(t, s.messages, h.messages)
	require.Equal(t, s.peers.channels, h.ents.Channels)
	require.Equal(t, s.peers.chats, h.ents.Chats)
	require.Equal(t, s.peers.users, h.ents.Users)
}

func loss(in chan *tg.Updates) chan *tg.Updates {
	out := make(chan *tg.Updates)

	go func() {
		defer close(out)

		for u := range in {
			if rand.Intn(5) == 1 {
				continue
			}

			out <- u
		}
	}()

	return out
}
