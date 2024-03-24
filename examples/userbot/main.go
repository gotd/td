package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"time"

	pebbledb "github.com/cockroachdb/pebble"
	"github.com/go-faster/errors"
	boltstor "github.com/gotd/contrib/bbolt"
	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/contrib/middleware/ratelimit"
	"github.com/gotd/contrib/pebble"
	"github.com/gotd/contrib/storage"
	"github.com/joho/godotenv"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/time/rate"
	lj "gopkg.in/natefinch/lumberjack.v2"

	"github.com/gotd/td/examples"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
)

func sessionFolder(phone string) string {
	var out []rune
	for _, r := range phone {
		if r >= '0' && r <= '9' {
			out = append(out, r)
		}
	}
	return "phone-" + string(out)
}

func run(ctx context.Context) error {
	var arg struct {
		FillPeerStorage bool
	}
	flag.BoolVar(&arg.FillPeerStorage, "fill-peer-storage", false, "fill peer storage")
	flag.Parse()

	// Using ".env" file to load environment variables.
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "load env")
	}

	// TG_PHONE is phone number in international format.
	// Like +4123456789.
	phone := os.Getenv("TG_PHONE")
	if phone == "" {
		return errors.New("no phone")
	}
	// APP_HASH, APP_ID is from https://my.telegram.org/.
	appID, err := strconv.Atoi(os.Getenv("APP_ID"))
	if err != nil {
		return errors.Wrap(err, " parse app id")
	}
	appHash := os.Getenv("APP_HASH")
	if appHash == "" {
		return errors.New("no app hash")
	}

	// Setting up session storage.
	// This is needed to reuse session and not login every time.
	sessionDir := filepath.Join("session", sessionFolder(phone))
	if err := os.MkdirAll(sessionDir, 0700); err != nil {
		return err
	}
	logFilePath := filepath.Join(sessionDir, "log.jsonl")

	fmt.Printf("Storing session in %s, logs in %s\n", sessionDir, logFilePath)

	// Setting up logging to file with rotation.
	//
	// Log to file, so we don't interfere with prompts and messages to user.
	logWriter := zapcore.AddSync(&lj.Logger{
		Filename:   logFilePath,
		MaxBackups: 3,
		MaxSize:    1, // megabytes
		MaxAge:     7, // days
	})
	logCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		logWriter,
		zap.DebugLevel,
	)
	lg := zap.New(logCore)
	defer func() { _ = lg.Sync() }()

	// So, we are storing session information in current directory, under subdirectory "session/phone_hash"
	sessionStorage := &telegram.FileSessionStorage{
		Path: filepath.Join(sessionDir, "session.json"),
	}
	// Peer storage, for resolve caching and short updates handling.
	db, err := pebbledb.Open(filepath.Join(sessionDir, "peers.pebble.db"), &pebbledb.Options{})
	if err != nil {
		return errors.Wrap(err, "create pebble storage")
	}
	peerDB := pebble.NewPeerStorage(db)
	lg.Info("Storage", zap.String("path", sessionDir))

	// Setting up client.
	//
	// Dispatcher is used to register handlers for events.
	dispatcher := tg.NewUpdateDispatcher()
	// Setting up update handler that will fill peer storage before
	// calling dispatcher handlers.
	updateHandler := storage.UpdateHook(dispatcher, peerDB)

	// Setting up persistent storage for qts/pts to be able to
	// recover after restart.
	boltdb, err := bbolt.Open(filepath.Join(sessionDir, "updates.bolt.db"), 0666, nil)
	if err != nil {
		return errors.Wrap(err, "create bolt storage")
	}
	updatesRecovery := updates.New(updates.Config{
		Handler: updateHandler, // using previous handler with peerDB
		Logger:  lg.Named("updates.recovery"),
		Storage: boltstor.NewStateStorage(boltdb),
	})

	// Handler of FLOOD_WAIT that will automatically retry request.
	waiter := floodwait.NewWaiter().WithCallback(func(ctx context.Context, wait floodwait.FloodWait) {
		// Notifying about flood wait.
		lg.Warn("Flood wait", zap.Duration("wait", wait.Duration))
		fmt.Println("Got FLOOD_WAIT. Will retry after", wait.Duration)
	})

	// Filling client options.
	options := telegram.Options{
		Logger:         lg,              // Passing logger for observability.
		SessionStorage: sessionStorage,  // Setting up session sessionStorage to store auth data.
		UpdateHandler:  updatesRecovery, // Setting up handler for updates from server.
		Middlewares: []telegram.Middleware{
			// Setting up FLOOD_WAIT handler to automatically wait and retry request.
			waiter,
			// Setting up general rate limits to less likely get flood wait errors.
			ratelimit.New(rate.Every(time.Millisecond*100), 5),
		},
	}
	client := telegram.NewClient(appID, appHash, options)
	api := client.API()

	// Setting up resolver cache that will use peer storage.
	resolver := storage.NewResolverCache(peer.Plain(api), peerDB)
	// Usage:
	//   if _, err := resolver.ResolveDomain(ctx, "tdlibchat"); err != nil {
	//	   return errors.Wrap(err, "resolve")
	//   }
	_ = resolver

	// Registering handler for new private messages.
	dispatcher.OnNewMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateNewMessage) error {
		msg, ok := u.Message.(*tg.Message)
		if !ok {
			return nil
		}
		if msg.Out {
			// Outgoing message.
			return nil
		}

		// Use PeerID to find peer because *Short updates does not contain any entities, so it necessary to
		// store some entities.
		//
		// Storage can be filled using PeerCollector (i.e. fetching all dialogs first).
		p, err := storage.FindPeer(ctx, peerDB, msg.GetPeerID())
		if err != nil {
			return err
		}

		fmt.Printf("%s: %s\n", p, msg.Message)
		return nil
	})

	// Authentication flow handles authentication process, like prompting for code and 2FA password.
	flow := auth.NewFlow(examples.Terminal{PhoneNumber: phone}, auth.SendCodeOptions{})

	return waiter.Run(ctx, func(ctx context.Context) error {
		// Spawning main goroutine.
		if err := client.Run(ctx, func(ctx context.Context) error {
			// Perform auth if no session is available.
			if err := client.Auth().IfNecessary(ctx, flow); err != nil {
				return errors.Wrap(err, "auth")
			}

			// Getting info about current user.
			self, err := client.Self(ctx)
			if err != nil {
				return errors.Wrap(err, "call self")
			}

			name := self.FirstName
			if self.Username != "" {
				// Username is optional.
				name = fmt.Sprintf("%s (@%s)", name, self.Username)
			}
			fmt.Println("Current user:", name)

			lg.Info("Login",
				zap.String("first_name", self.FirstName),
				zap.String("last_name", self.LastName),
				zap.String("username", self.Username),
				zap.Int64("id", self.ID),
			)

			if arg.FillPeerStorage {
				fmt.Println("Filling peer storage from dialogs to cache entities")
				collector := storage.CollectPeers(peerDB)
				if err := collector.Dialogs(ctx, query.GetDialogs(api).Iter()); err != nil {
					return errors.Wrap(err, "collect peers")
				}
				fmt.Println("Filled")
			}

			// Waiting until context is done.
			fmt.Println("Listening for updates. Interrupt (Ctrl+C) to stop.")
			return updatesRecovery.Run(ctx, api, self.ID, updates.AuthOptions{
				IsBot: self.Bot,
				OnStart: func(ctx context.Context) {
					fmt.Println("Update recovery initialized and started, listening for events")
				},
			})
		}); err != nil {
			return errors.Wrap(err, "run")
		}
		return nil
	})
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx); err != nil {
		if errors.Is(err, context.Canceled) && ctx.Err() == context.Canceled {
			fmt.Println("\rClosed")
			os.Exit(0)
		}
		_, _ = fmt.Fprintf(os.Stderr, "Error: %+v\n", err)
		os.Exit(1)
	} else {
		fmt.Println("Done")
		os.Exit(0)
	}
}
