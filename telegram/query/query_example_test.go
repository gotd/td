package query_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/nnqq/td/telegram"
	"github.com/nnqq/td/telegram/downloader"
	"github.com/nnqq/td/telegram/query"
	"github.com/nnqq/td/telegram/query/channels/participants"
	"github.com/nnqq/td/telegram/query/dialogs"
	"github.com/nnqq/td/telegram/query/messages"
	"github.com/nnqq/td/tg"
)

func ExampleQuery_iterAllMessages() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	client, err := telegram.ClientFromEnvironment(telegram.Options{})
	if err != nil {
		panic(err)
	}

	// This example iterates over all messages of all dialogs of user and prints them.
	if err := client.Run(ctx, func(ctx context.Context) error {
		raw := tg.NewClient(client)
		cb := func(ctx context.Context, dlg dialogs.Elem) error {
			// Skip deleted dialogs.
			if dlg.Deleted() {
				return nil
			}

			return dlg.Messages(raw).ForEach(ctx, func(ctx context.Context, elem messages.Elem) error {
				msg, ok := elem.Msg.(*tg.Message)
				if !ok {
					return nil
				}
				fmt.Println(msg.Message)

				return nil
			})
		}

		return query.GetDialogs(raw).ForEach(ctx, cb)
	}); err != nil {
		panic(err)
	}
}

func ExampleQuery_downloadSaved() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	client, err := telegram.ClientFromEnvironment(telegram.Options{})
	if err != nil {
		panic(err)
	}

	// This example downloads all attachments (photo, video, docs, etc.)
	// from SavedMessages dialog.
	if err := client.Run(ctx, func(ctx context.Context) error {
		raw := tg.NewClient(client)
		d := downloader.NewDownloader()
		return query.Messages(raw).GetHistory(&tg.InputPeerSelf{}).ForEach(ctx,
			func(ctx context.Context, elem messages.Elem) error {
				f, ok := elem.File()
				if !ok {
					return nil
				}

				_, err := d.Download(raw, f.Location).ToPath(ctx, f.Name)
				return err
			})
	}); err != nil {
		panic(err)
	}
}

func ExampleQuery_getAdmins() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	client, err := telegram.ClientFromEnvironment(telegram.Options{})
	if err != nil {
		panic(err)
	}

	// This example iterates over all channels and prints admins.
	if err := client.Run(ctx, func(ctx context.Context) error {
		raw := tg.NewClient(client)
		cb := func(ctx context.Context, dlg dialogs.Elem) error {
			// Skip deleted dialogs.
			if dlg.Deleted() {
				return nil
			}

			q, ok := dlg.Participants(raw)
			if !ok {
				return nil
			}

			return q.ForEach(ctx, func(ctx context.Context, elem participants.Elem) error {
				user, admin, ok := elem.Admin()
				if !ok {
					return nil
				}

				fmt.Println(user.Username, "admin")
				if admin.AdminRights.ChangeInfo {
					fmt.Println("\t+ ChangeInfo")
				}
				if admin.AdminRights.PostMessages {
					fmt.Println("\t+ PostMessages")
				}
				if admin.AdminRights.EditMessages {
					fmt.Println("\t+ EditMessages")
				}
				if admin.AdminRights.DeleteMessages {
					fmt.Println("\t+ DeleteMessages")
				}
				if admin.AdminRights.BanUsers {
					fmt.Println("\t+ BanUsers")
				}
				if admin.AdminRights.InviteUsers {
					fmt.Println("\t+ InviteUsers")
				}
				if admin.AdminRights.PinMessages {
					fmt.Println("\t+ PinMessages")
				}
				if admin.AdminRights.AddAdmins {
					fmt.Println("\t+ AddAdmins")
				}
				if admin.AdminRights.Anonymous {
					fmt.Println("\t+ Anonymous")
				}
				if admin.AdminRights.ManageCall {
					fmt.Println("\t+ ManageCall")
				}
				if admin.AdminRights.Other {
					fmt.Println("\t+ Other")
				}
				return nil
			})
		}

		return query.GetDialogs(raw).ForEach(ctx, cb)
	}); err != nil {
		panic(err)
	}
}
