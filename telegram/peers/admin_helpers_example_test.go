package peers_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/telegram/peers/members"
	"github.com/gotd/td/tg"
)

func administrate(ctx context.Context) error {
	client, err := telegram.ClientFromEnvironment(telegram.Options{})
	if err != nil {
		return err
	}

	return client.Run(ctx, func(ctx context.Context) error {
		m := peers.Options{}.Build(client.API())

		// Resolve a supergroup to administrate.
		p, err := m.Resolve(ctx, "gotd_test")
		if err != nil {
			return err
		}
		ch, ok := p.(peers.Channel)
		if !ok {
			return fmt.Errorf("%q is not a channel", "gotd_test")
		}
		sg, ok := ch.ToSupergroup()
		if !ok {
			return fmt.Errorf("%q is not a supergroup", "gotd_test")
		}

		// Set a public username.
		if err := sg.SetUsername(ctx, "gotd_example"); err != nil {
			return err
		}

		// Require admin approval for new members and hide message
		// history from them.
		if err := sg.ToggleJoinRequest(ctx, true); err != nil {
			return err
		}
		if err := sg.TogglePreHistoryHidden(ctx, true); err != nil {
			return err
		}

		// Promote a user to an admin allowed to ban and pin.
		admin, err := m.Resolve(ctx, "gotd_admin")
		if err != nil {
			return err
		}
		user, ok := admin.(peers.User)
		if !ok {
			return fmt.Errorf("%q is not a user", "gotd_admin")
		}
		if err := members.Channel(sg.Channel).Promote(ctx, user.InputUser(), members.AdminRights{
			Rank:        "moderator",
			BanUsers:    true,
			PinMessages: true,
		}); err != nil {
			return err
		}

		// Walk the recent admin actions log.
		return sg.AdminLog().ForEach(ctx, func(event tg.ChannelAdminLogEvent) error {
			fmt.Println("admin action", event.ID, "by", event.UserID)
			return nil
		})
	})
}

func ExampleChannel_AdminLog() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := administrate(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}
}
