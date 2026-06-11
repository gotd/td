// Binary groupcall joins a Telegram group voice chat and streams an MP3 file
// into it as audio.
//
// It logs in via QR code (with 2FA fallback), resolves a supergroup/channel by
// username, joins its active voice chat and streams the MP3 (transcoded to Opus
// with ffmpeg) until the file ends.
//
// Usage:
//
//	APP_ID=... APP_HASH=... SESSION_FILE=session.json \
//	    go run ./examples/groupcall -chat @mygroup -audio song.mp3
//
// The group must already have a voice chat started. Pass -test to run against
// Telegram's test servers instead of production. Requirements: ffmpeg on PATH
// (with libopus).
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

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
	var (
		chat  = flag.String("chat", "", "username of the supergroup/channel whose voice chat to join")
		audio = flag.String("audio", "", "path to an MP3 file to stream")
		test  = flag.Bool("test", false, "use Telegram test servers instead of production")
	)
	flag.Parse()

	if *chat == "" || *audio == "" {
		fmt.Fprintln(os.Stderr, "usage: groupcall -chat @group -audio file.mp3")
		os.Exit(2)
	}

	examples.Run(func(ctx context.Context, log *zap.Logger) error {
		dispatcher := tg.NewUpdateDispatcher()
		loggedIn := qrlogin.OnLoginToken(&dispatcher)

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

		gc := calls.NewGroupCall(client.API(), calls.Options{Logger: log})
		gc.Register(dispatcher)
		gc.OnParticipants(func(p []tg.GroupCallParticipant) {
			log.Info("Participants updated", zap.Int("count", len(p)))
		})

		return client.Run(ctx, func(ctx context.Context) error {
			if err := examples.QRAuth(ctx, client, loggedIn); err != nil {
				return err
			}
			self, err := client.Self(ctx)
			if err != nil {
				return errors.Wrap(err, "self")
			}

			call, err := resolveGroupCall(ctx, client.API(), *chat)
			if err != nil {
				return err
			}

			joinAs := &tg.InputPeerUser{UserID: self.ID, AccessHash: self.AccessHash}
			log.Info("Joining voice chat", zap.Int64("call", call.ID))
			if err := gc.Join(ctx, call, joinAs); err != nil {
				return errors.Wrap(err, "join group call")
			}
			defer func() {
				if err := gc.Leave(ctx); err != nil {
					log.Warn("Leave", zap.Error(err))
				}
			}()
			log.Info("Joined; streaming audio", zap.String("file", *audio))

			if err := examples.StreamMP3(ctx, gc.WriteAudio, *audio); err != nil {
				return errors.Wrap(err, "stream audio")
			}
			log.Info("Playback finished")
			return nil
		})
	})
}

// resolveGroupCall resolves a @username to the active group call of its
// supergroup/channel.
func resolveGroupCall(ctx context.Context, api *tg.Client, username string) (*tg.InputGroupCall, error) {
	resolved, err := api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
		Username: strings.TrimPrefix(username, "@"),
	})
	if err != nil {
		return nil, errors.Wrap(err, "resolve username")
	}

	var channel *tg.Channel
	for _, c := range resolved.Chats {
		if ch, ok := c.(*tg.Channel); ok {
			channel = ch
			break
		}
	}
	if channel == nil {
		return nil, errors.Errorf("%q is not a supergroup or channel", username)
	}

	full, err := api.ChannelsGetFullChannel(ctx, &tg.InputChannel{
		ChannelID:  channel.ID,
		AccessHash: channel.AccessHash,
	})
	if err != nil {
		return nil, errors.Wrap(err, "get full channel")
	}
	channelFull, ok := full.FullChat.(*tg.ChannelFull)
	if !ok {
		return nil, errors.Errorf("unexpected full chat %T", full.FullChat)
	}
	call, ok := channelFull.Call.(*tg.InputGroupCall)
	if !ok {
		return nil, errors.New("the group has no active voice chat")
	}
	return call, nil
}
