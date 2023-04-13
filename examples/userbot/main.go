package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gotd/contrib/middleware/ratelimit"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/time/rate"
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

func ensureDir(name string) error {
	if err := os.MkdirAll(name, 0700); err != nil {
		return err
	}

	return nil
}

// terminalAuth implements auth.UserAuthenticator prompting the terminal for
// input.
type terminalAuth struct {
	phone string
}

func (terminalAuth) SignUp(ctx context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, xerrors.New("not implemented")
}

func (terminalAuth) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	return &auth.SignUpRequired{TermsOfService: tos}
}

func (terminalAuth) Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	fmt.Print("Enter code: ")
	code, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(code), nil
}

func (a terminalAuth) Phone(_ context.Context) (string, error) {
	return a.phone, nil
}

func (terminalAuth) Password(_ context.Context) (string, error) {
	fmt.Print("Enter 2FA password: ")
	bytePwd, err := terminal.ReadPassword(syscall.Stdin)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(bytePwd)), nil
}

func run(ctx context.Context) error {
	log, _ := zap.NewDevelopment(zap.IncreaseLevel(zapcore.DebugLevel))
	defer func() { _ = log.Sync() }()

	// Setting up session storage.
	// This is needed to reuse session and not login every time.
	sessionDir := filepath.Join("session")
	if err := ensureDir(sessionDir); err != nil {
		return err
	}
	phone := os.Getenv("TG_PHONE")
	if phone == "" {
		return xerrors.New("no phone")
	}
	// So, we are storing session information in current directory, under subdirectory "session".
	// Session file name is based on phone number.
	storage := &telegram.FileSessionStorage{
		Path: filepath.Join(sessionDir, fmt.Sprintf("session.phone.%x.json", []byte(phone))),
	}

	log.Info("Session storage", zap.String("path", storage.Path))

	// APP_HASH, APP_ID is from https://my.telegram.org/.
	appID, err := strconv.Atoi(os.Getenv("APP_ID"))
	if err != nil {
		return xerrors.Errorf("failed to parse app id: %w", err)
	}

	var (
		// Dispatcher is used to register handlers for events.
		dispatcher = tg.NewUpdateDispatcher()
		client     = telegram.NewClient(appID, os.Getenv("APP_HASH"), telegram.Options{
			Logger:         log,
			SessionStorage: storage,    // Setting up session storage,
			UpdateHandler:  dispatcher, // Setting up update handler to event dispatcher.
			Middlewares: []telegram.Middleware{
				// Setting up ratelimit so we don't get flood wait errors.
				ratelimit.New(rate.Every(time.Millisecond*100), 5),
			},
		})
	)

	// Registering handler for new private messages.
	dispatcher.OnNewMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateNewMessage) error {
		switch m := u.Message.(type) {
		case *tg.Message:
			if m.Out {
				// Outgoing message.
				return nil
			}
			switch peer := m.PeerID.(type) {
			case *tg.PeerUser:
				l := log.With(zap.String("text", m.Message))
				if user, ok := e.Users[peer.UserID]; ok {
					l = l.With(
						zap.Int64("user_id", user.ID),
						zap.String("user_first_name", user.FirstName),
						zap.String("username", user.Username),
					)
				}
				l.Info("Got message")
			default:
				log.Warn("Unsupported peer", zap.Any("peer", peer))
				return nil
			}
		}
		return nil
	})

	if err := client.Run(ctx, func(ctx context.Context) error {
		if self, err := client.Self(ctx); err != nil || self.Bot {
			// Need to authenticate.
			if err := auth.NewFlow(terminalAuth{phone: phone}, auth.SendCodeOptions{}).Run(ctx, client.Auth()); err != nil {
				return xerrors.Errorf("failed to auth: %w", err)
			}
		} else {
			log.Info("Already authenticated")
		}

		// Getting info about current user.
		self, err := client.Self(ctx)
		if err != nil {
			return xerrors.Errorf("failed to call self: %w", err)
		}

		log.Info("Logged in",
			zap.String("first_name", self.FirstName),
			zap.String("last_name", self.LastName),
			zap.String("username", self.Username),
			zap.Int64("id", self.ID),
		)

		// Waiting until context is done.
		<-ctx.Done()
		return ctx.Err()
	}); err != nil {
		return xerrors.Errorf("run: %w", err)
	}

	return nil
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx); err != nil {
		panic(err)
	}
}
