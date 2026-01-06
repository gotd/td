package takeout_test

import (
	"context"
	"fmt"

	"github.com/gotd/td/telegram/takeout"
	"github.com/gotd/td/tg"
)

func ExampleRun() {
	// This example demonstrates how to use the takeout API wrapper.
	// In a real application, you would use a proper tg.Invoker.

	ctx := context.Background()
	var invoker tg.Invoker // obtained from telegram.Client

	// Configure what data to export
	cfg := takeout.Config{
		Contacts:          true,
		MessageUsers:      true,
		MessageChats:      true,
		MessageMegagroups: true,
		MessageChannels:   true,
		Files:             true,
		FileMaxSize:       512 * 1024 * 1024, // 512 MB
	}

	err := takeout.Run(ctx, invoker, cfg, func(ctx context.Context, client *takeout.Client) error {
		// All API calls made with client are wrapped with takeout session.
		// Use tg.NewClient(client) to get a full API client.
		api := tg.NewClient(client)

		// For example, get dialogs:
		dialogs, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
			Limit: 100,
		})
		if err != nil {
			return err
		}

		fmt.Printf("Got dialogs: %T\n", dialogs)
		return nil
	})
	if err != nil {
		// Handle error
		_ = err
	}
}
