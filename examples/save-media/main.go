// Binary save-media is a userbot that saves the media of a message you reply
// to with the text "save".
//
// It addresses the request from https://github.com/gotd/td/issues/166:
// "reply to a picture / video message with key phrase like 'save' and it'll
// save the referred picture / video".
//
// Reply "save" to any photo, video or document in a private chat or basic
// group and the file is downloaded to --out. It demonstrates the update
// dispatcher, fetching the replied-to message and the downloader helper,
// reusing query/messages Elem.File to locate the attachment.
package main

import (
	"context"
	"flag"
	"path/filepath"
	"strings"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/gotd/log/logzap"

	"github.com/gotd/td/examples"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/query/messages"
	"github.com/gotd/td/tg"
)

func main() {
	outDir := flag.String("out", ".", "directory to save media to")
	flag.Parse()

	examples.Run(func(ctx context.Context, log *zap.Logger) error {
		dispatcher := tg.NewUpdateDispatcher()
		client, err := telegram.ClientFromEnvironment(telegram.Options{
			Logger:        logzap.New(log),
			UpdateHandler: dispatcher,
		})
		if err != nil {
			return err
		}

		api := client.API()
		sender := message.NewSender(api)
		d := downloader.NewDownloader()

		// Handle messages from private chats and basic groups. Channel and
		// supergroup messages arrive via OnNewChannelMessage instead and would
		// use channels.getMessages to fetch the replied-to message.
		dispatcher.OnNewMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateNewMessage) error {
			msg, ok := u.Message.(*tg.Message)
			if !ok || msg.Out {
				return nil
			}
			if strings.ToLower(strings.TrimSpace(msg.Message)) != "save" {
				return nil
			}

			// The "save" message must be a reply to the media message.
			reply, ok := msg.ReplyTo.(*tg.MessageReplyHeader)
			if !ok || reply.ReplyToMsgID == 0 {
				_, err := sender.Reply(e, u).Text(ctx, "Reply 'save' to a media message.")
				return err
			}

			// Fetch the replied-to message by ID.
			res, err := api.MessagesGetMessages(ctx, []tg.InputMessageClass{
				&tg.InputMessageID{ID: reply.ReplyToMsgID},
			})
			if err != nil {
				return errors.Wrap(err, "get replied message")
			}
			var raw []tg.MessageClass
			switch m := res.(type) {
			case *tg.MessagesMessages:
				raw = m.Messages
			case *tg.MessagesMessagesSlice:
				raw = m.Messages
			default:
				return nil
			}
			if len(raw) == 0 {
				return nil
			}
			replied, ok := raw[0].(*tg.Message)
			if !ok {
				return nil
			}

			// Reuse query/messages Elem.File to locate the downloadable file.
			file, ok := messages.Elem{Msg: replied}.File()
			if !ok {
				_, err := sender.Reply(e, u).Text(ctx, "Replied message has no media.")
				return err
			}

			path := filepath.Join(*outDir, file.Name)
			log.Info("Saving media", zap.String("path", path), zap.String("mime", file.MIMEType))
			if _, err := d.Download(api, file.Location).ToPath(ctx, path); err != nil {
				return errors.Wrap(err, "download")
			}

			_, err = sender.Reply(e, u).Text(ctx, "Saved "+file.Name)
			return err
		})

		// Reading phone, code and 2FA password from terminal when no session.
		flow := auth.NewFlow(examples.Terminal{}, auth.SendCodeOptions{})
		return client.Run(ctx, func(ctx context.Context) error {
			if err := client.Auth().IfNecessary(ctx, flow); err != nil {
				return errors.Wrap(err, "auth")
			}
			log.Info("Userbot started, reply 'save' to media messages (Ctrl-C to stop)")
			<-ctx.Done()
			return ctx.Err()
		})
	})
}
