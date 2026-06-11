// Binary call answers any incoming message with a Telegram 1:1 voice call back
// to the sender, streaming an MP3 file as the outgoing audio.
//
// It logs in via QR code (with 2FA fallback), then waits for an incoming
// message. For each message it places a call to the sender and, once connected,
// transcodes the MP3 to Opus with ffmpeg and streams it into the call.
//
// Usage:
//
//	APP_ID=... APP_HASH=... SESSION_FILE=session.json \
//	    go run ./examples/call -audio song.mp3
//
// Then message the logged-in account from another account to get a call back.
// Pass -test to run against Telegram's test servers instead of production.
//
// Requirements: ffmpeg on PATH (with libopus).
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/gotd/td/examples"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth/qrlogin"
	"github.com/gotd/td/telegram/calls"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tg"
)

func main() {
	audio := flag.String("audio", "", "path to an MP3 file to stream as voice")
	test := flag.Bool("test", false, "use Telegram test servers instead of production")
	flag.Parse()

	if *audio == "" {
		fmt.Fprintln(os.Stderr, "usage: call -audio file.mp3")
		os.Exit(2)
	}

	examples.Run(func(ctx context.Context, log *zap.Logger) error {
		dispatcher := tg.NewUpdateDispatcher()
		loggedIn := qrlogin.OnLoginToken(&dispatcher)

		// Caches sender access hashes seen in updates so we can build an
		// InputUser for the call back.
		users := newUserCache()

		// User IDs of incoming-message senders to call back. Buffered and
		// deduplicated so messages that arrive while we are busy don't pile up.
		incoming := make(chan int64, 1)
		dispatcher.OnNewMessage(func(_ context.Context, e tg.Entities, u *tg.UpdateNewMessage) error {
			m, ok := u.Message.(*tg.Message)
			if !ok || m.Out {
				return nil // Not a regular message, or one we sent ourselves.
			}
			peer, ok := m.PeerID.(*tg.PeerUser)
			if !ok {
				return nil // Only call back private chats.
			}
			// Full and difference updates carry the sender (with access hash) in
			// entities; "short" updates do not, so we resolve those via dialogs.
			if user, ok := e.Users[peer.UserID]; ok {
				users.put(user)
			}
			select {
			case incoming <- peer.UserID:
			default: // A call back is already queued or in progress.
			}
			return nil
		})

		opts := telegram.Options{
			Logger:        log,
			UpdateHandler: dispatcher,
		}
		if *test {
			// Telegram test servers live on DC 2.
			opts.DC = 2
			opts.DCList = dcs.Test()
		}
		client, err := telegram.ClientFromEnvironment(opts)
		if err != nil {
			return err
		}
		api := client.API()

		callClient := calls.NewClient(api, calls.Options{Logger: log})
		callClient.Register(dispatcher)

		return client.Run(ctx, func(ctx context.Context) error {
			if err := examples.QRAuth(ctx, client, loggedIn); err != nil {
				return err
			}
			self, err := client.Self(ctx)
			if err != nil {
				return errors.Wrap(err, "self")
			}
			log.Info("Logged in; message this account to get a call back",
				zap.String("user", self.FirstName), zap.Int64("id", self.ID))

			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case id := <-incoming:
					user, err := users.resolve(ctx, api, id)
					if err != nil {
						log.Warn("Resolve sender", zap.Int64("id", id), zap.Error(err))
						continue
					}
					if err := placeCall(ctx, log, callClient, user, *audio); err != nil {
						log.Warn("Call back failed", zap.Error(err))
					}
				}
			}
		})
	})
}

// userCache remembers users (and their access hashes) seen in updates and can
// resolve an unknown user ID through the dialog list.
type userCache struct {
	mu sync.Mutex
	m  map[int64]*tg.InputUser
}

func newUserCache() *userCache { return &userCache{m: make(map[int64]*tg.InputUser)} }

func (c *userCache) put(u *tg.User) {
	c.mu.Lock()
	c.m[u.ID] = &tg.InputUser{UserID: u.ID, AccessHash: u.AccessHash}
	c.mu.Unlock()
}

func (c *userCache) get(id int64) (*tg.InputUser, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	in, ok := c.m[id]
	return in, ok
}

// resolve returns an InputUser for id, falling back to the dialog list (a fresh
// DM bumps its sender to the top) when the user has not been seen in an update.
func (c *userCache) resolve(ctx context.Context, api *tg.Client, id int64) (*tg.InputUser, error) {
	if in, ok := c.get(id); ok {
		return in, nil
	}

	dialogs, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      100,
	})
	if err != nil {
		return nil, errors.Wrap(err, "get dialogs")
	}
	var users []tg.UserClass
	switch d := dialogs.(type) {
	case *tg.MessagesDialogs:
		users = d.Users
	case *tg.MessagesDialogsSlice:
		users = d.Users
	}
	for _, uc := range users {
		if u, ok := uc.(*tg.User); ok {
			c.put(u)
		}
	}
	if in, ok := c.get(id); ok {
		return in, nil
	}
	return nil, errors.Errorf("user %d not found in dialogs", id)
}

// placeCall requests a call, waits for it to connect, streams the audio and
// then hangs up.
func placeCall(
	ctx context.Context,
	log *zap.Logger,
	callClient *calls.Client,
	user tg.InputUserClass,
	audio string,
) error {
	log.Info("Requesting call")
	conn, err := callClient.Request(ctx, user)
	if err != nil {
		return errors.Wrap(err, "request call")
	}
	defer func() {
		if err := callClient.Discard(ctx, calls.DiscardHangup); err != nil {
			log.Warn("Discard", zap.Error(err))
		}
	}()

	connected := make(chan struct{})
	var once sync.Once
	conn.OnConnected(func() {
		once.Do(func() {
			close(connected)
			log.Info("Call connected")
		})
	})
	conn.OnDisconnected(func() { log.Info("Call disconnected") })

	select {
	case <-connected:
	case <-time.After(60 * time.Second):
		return errors.New("timed out waiting for the call to connect")
	case <-ctx.Done():
		return ctx.Err()
	}

	log.Info("Streaming audio", zap.String("file", audio))
	if err := examples.StreamMP3(ctx, conn.AudioTrack().WriteRTP, audio); err != nil {
		return errors.Wrap(err, "stream audio")
	}
	log.Info("Playback finished")
	return nil
}
