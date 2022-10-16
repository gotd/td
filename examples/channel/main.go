// Channel, create & make it public
package main

import (
	"context"

	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

func main() {
	// Environment variables:
	// 	APP_ID:        app_id of Telegram app.
	// 	APP_HASH:      app_hash of Telegram app.
	// 	SESSION_FILE:  path to session file
	// 	SESSION_DIR:   path to session directory, if SESSION_FILE is not set
	client, err := telegram.ClientFromEnvironment(telegram.Options{})
	if err != nil {
		panic(err)
	}
	if err := client.Run(context.Background(), func(ctx context.Context) error {

		api := client.API()

		// Create a new channel
		channelsCreateChannel, err := api.ChannelsCreateChannel(ctx, &tg.ChannelsCreateChannelRequest{
			Broadcast: true,
			Megagroup: false,
			ForImport: false,
			Title:     "My example channel",
			About:     "This is my example channel",
		})
		if err != nil {
			return err
		}

		// Unpack the response
		channel := channelsCreateChannel.(*tg.Updates).Chats[0].(*tg.Channel)

		ok, err := api.ChannelsCheckUsername(ctx, &tg.ChannelsCheckUsernameRequest{
			Channel:  channel.AsInput(),
			Username: "myusernameexample",
		})
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("username is not available")
		}

		ok, err = api.ChannelsUpdateUsername(ctx, &tg.ChannelsUpdateUsernameRequest{
			Channel:  channel.AsInput(),
			Username: "myusernameexample",
		})

		if err != nil {
			return err
		}
		if !ok {
			return errors.New("unable to update username")
		}

		return nil
	}); err != nil {
		panic(err)
	}
}
