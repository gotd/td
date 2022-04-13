package auth_test

import (
	"context"
	"fmt"

	"github.com/go-faster/errors"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
)

func ExampleClient_UpdatePassword() {
	ctx := context.Background()
	client := telegram.NewClient(telegram.TestAppID, telegram.TestAppHash, telegram.Options{})
	if err := client.Run(ctx, func(ctx context.Context) error {
		// Updating password.
		if err := client.Auth().UpdatePassword(ctx, []byte("new_password"), auth.UpdatePasswordOptions{
			// Hint sets new password hint.
			Hint: "new password hint",
			// Password will be called if old password is requested by Telegram.
			//
			// If password was requested and Password is nil, auth.ErrPasswordNotProvided error will be returned.
			Password: func(ctx context.Context) (string, error) {
				return "old_password", nil
			},
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		panic(err)
	}
}

func ExampleClient_ResetPassword() {
	ctx := context.Background()
	client := telegram.NewClient(telegram.TestAppID, telegram.TestAppHash, telegram.Options{})
	if err := client.Run(ctx, func(ctx context.Context) error {
		wait, err := client.Auth().ResetPassword(ctx)
		var waitErr *auth.ResetFailedWaitError
		switch {
		case errors.As(err, &waitErr):
			// Telegram requested wait until making new reset request.
			fmt.Printf("Wait until %s to reset password.\n", wait.String())
		case err != nil:
			return err
		}

		// If returned time is zero, password was successfully reset.
		if wait.IsZero() {
			fmt.Println("Password was reset.")
			return nil
		}

		fmt.Printf("Password will be reset on %s.\n", wait.String())
		return nil
	}); err != nil {
		panic(err)
	}
}
