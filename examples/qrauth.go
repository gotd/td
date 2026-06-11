package examples

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-faster/errors"
	"github.com/mdp/qrterminal/v3"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/auth/qrlogin"
	"github.com/gotd/td/tgerr"
)

// QRAuth makes sure the client is authorized, running an interactive QR login
// (with 2FA fallback) when it is not. loggedIn is the channel returned by
// qrlogin.OnLoginToken for the dispatcher passed to the client.
//
// This is example code: copy and adapt it to your needs.
func QRAuth(ctx context.Context, client *telegram.Client, loggedIn qrlogin.LoggedIn) error {
	status, err := client.Auth().Status(ctx)
	if err != nil {
		return errors.Wrap(err, "auth status")
	}
	if status.Authorized {
		return nil
	}

	fmt.Fprintln(os.Stderr,
		"\nScan this QR code with Telegram (Settings → Devices → Link Desktop Device):")
	show := func(ctx context.Context, token qrlogin.Token) error {
		qrterminal.Generate(token.URL(), qrterminal.L, os.Stderr)
		fmt.Fprintf(os.Stderr, "Or open: %s\n\nWaiting for scan...\n", token.URL())
		return nil
	}

	if _, err := client.QR().Auth(ctx, loggedIn, show); err != nil {
		if !tgerr.Is(err, "SESSION_PASSWORD_NEEDED") {
			return errors.Wrap(err, "QR auth")
		}
		if err := handle2FA(ctx, client.Auth()); err != nil {
			return err
		}
	}
	return nil
}

// handle2FA prompts for the cloud password and submits it, retrying on a wrong
// password until accepted or EOF.
func handle2FA(ctx context.Context, a *auth.Client) error {
	for {
		fmt.Fprint(os.Stderr, "Enter 2FA password: ")
		pwd, err := readLine()
		if err != nil {
			return errors.Wrap(err, "read 2FA password")
		}
		if _, err := a.Password(ctx, pwd); err != nil {
			if errors.Is(err, auth.ErrPasswordInvalid) {
				fmt.Fprintln(os.Stderr, "Wrong password, try again.")
				continue
			}
			return errors.Wrap(err, "2FA password")
		}
		return nil
	}
}

func readLine() (string, error) {
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
