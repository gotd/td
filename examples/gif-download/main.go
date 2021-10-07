// Binary gif-download implements example of gif backup.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/atomic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"
	"golang.org/x/xerrors"

	"github.com/gotd/contrib/middleware/ratelimit"

	"github.com/nnqq/td/telegram"
	"github.com/nnqq/td/telegram/auth"
	"github.com/nnqq/td/telegram/downloader"
	"github.com/nnqq/td/telegram/query/hasher"
	"github.com/nnqq/td/tg"
)

// terminalAuth implements auth.UserAuthenticator prompting the terminal for
// input.
type terminalAuth struct{}

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

func (terminalAuth) Phone(_ context.Context) (string, error) {
	fmt.Print("Enter phone: ")
	code, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(code), nil
}

func (terminalAuth) Password(_ context.Context) (string, error) {
	fmt.Print("Enter 2FA password: ")
	bytePwd, err := terminal.ReadPassword(0)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(bytePwd)), nil
}

func run(ctx context.Context) error {
	var (
		outputDir = flag.String("out", os.TempDir(), "output directory")
		inputDir  = flag.String("input", "", "input directory for uploads")
		jobs      = flag.Int("j", 3, "maximum concurrent download jobs")
		remove    = flag.Bool("rm", false, "remove downloaded gifs")
		rateLimit = flag.Duration("rate", time.Millisecond*100, "limit maximum rpc call rate")
		rateBurst = flag.Int("rate-burst", 3, "limit rpc call burst")
	)
	flag.Parse()

	log, _ := zap.NewDevelopment(zap.IncreaseLevel(zapcore.InfoLevel), zap.AddStacktrace(zapcore.FatalLevel))
	defer func() { _ = log.Sync() }()

	// Initializing client from environment.
	// Available environment variables:
	// 	APP_ID:         app_id of Telegram app.
	// 	APP_HASH:       app_hash of Telegram app.
	// 	SESSION_FILE:   path to session file
	// 	SESSION_DIR:    path to session directory, if SESSION_FILE is not set
	client, err := telegram.ClientFromEnvironment(telegram.Options{
		Logger: log,
		Middlewares: []telegram.Middleware{
			ratelimit.New(rate.Every(*rateLimit), *rateBurst),
		},
	})
	if err != nil {
		return err
	}

	// Setting up authentication flow.
	// Current flow will read phone, code and 2FA password from terminal.
	flow := auth.NewFlow(terminalAuth{}, auth.SendCodeOptions{})

	// Creating new RPC client.
	//
	// The tg.Client is generated from Telegram schema and implements
	// invocation of all defined Telegram MTProto methods on top of tg.Invoker.
	// E.g. api.MessagesSendMessage() is messages.sendMessage method.
	//
	// The tg.Invoker interface is implemented by client (telegram.Client) and
	// allows calling any MTProto method, like that:
	//	Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error
	api := client.API()

	// Connecting, performing authentication and downloading gifs.
	return client.Run(ctx, func(ctx context.Context) error {
		// Perform auth if no session is available.
		if err := client.Auth().IfNecessary(ctx, flow); err != nil {
			return xerrors.Errorf("auth: %w", err)
		}

		if *inputDir != "" {
			// Handling bulk upload.
			// Probably we can de-duplicate gifs by some criteria.
			if err := upload(ctx, log, api, *inputDir); err != nil {
				return xerrors.Errorf("upload: %w", err)
			}
		}

		// Processing gifs.
		gifs := make(chan *tg.Document, *jobs)
		g, ctx := errgroup.WithContext(ctx)
		g.Go(func() error {
			defer close(gifs)

			// Telegram allows up to 200 saved gifs, but only hides exceeding
			// ones.
			//
			// Hasher implements Telegram "pagination" hash calculation and
			// allows us to exhaust all gifs in "rm" mode.
			h := hasher.Hasher{}
			for {
				result, err := api.MessagesGetSavedGifs(ctx, int(h.Sum()))
				if err != nil {
					return xerrors.Errorf("get: %w", err)
				}

				h.Reset()
				switch result := result.(type) {
				case *tg.MessagesSavedGifsNotModified:
					// Done.
					return nil
				case *tg.MessagesSavedGifs:
					log.Info("Got gifs",
						zap.Int("count", len(result.Gifs)),
					)
					if len(result.Gifs) == 0 {
						// No results.
						return nil
					}

					// Processing batch.
					for _, doc := range result.Gifs {
						doc, ok := doc.AsNotEmpty()
						if !ok {
							continue
						}

						select {
						case gifs <- doc:
							h.Update64(uint64(doc.ID))
						case <-ctx.Done():
							return ctx.Err()
						}
					}
				}
			}
		})

		var (
			total      atomic.Int32
			downloaded atomic.Int32
		)
		for j := 0; j < *jobs; j++ {
			g.Go(func() error {
				// Process all discovered gifs.
				d := downloader.NewDownloader()
				for doc := range gifs {
					total.Inc()
					gifPath := filepath.Join(*outputDir, fmt.Sprintf("%d.mp4", doc.ID))
					log.Info("Got gif",
						zap.Int64("id", doc.ID),
						zap.Time("date", time.Unix(int64(doc.Date), 0)),
						zap.String("path", gifPath),
					)

					if _, err := os.Stat(gifPath); err == nil {
						// File exists, skipping.
						//
						// Note that we are not completely sure that existing
						// file is exactly same as this gif (e.g. partial
						// download), so not removing even with --rm flag.
						continue
					}

					// Downloading gif to gifPath.
					loc := doc.AsInputDocumentFileLocation()
					if _, err := d.Download(api, loc).ToPath(ctx, gifPath); err != nil {
						return xerrors.Errorf("download: %w", err)
					}
					downloaded.Inc()

					if *remove {
						log.Info("Removing gif after download",
							zap.Int64("id", doc.ID),
							zap.Time("date", time.Unix(int64(doc.Date), 0)),
						)
						if _, err := api.MessagesSaveGif(ctx, &tg.MessagesSaveGifRequest{
							ID:     doc.AsInput(),
							Unsave: true,
						}); err != nil {
							return xerrors.Errorf("remove: %w", err)
						}
					}
				}

				return nil
			})
		}

		if err := g.Wait(); err != nil {
			return err
		}
		log.Info("Finished OK",
			zap.Int32("downloaded", downloaded.Load()),
			zap.Int32("total", total.Load()),
		)

		return nil
	})
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := run(ctx); err != nil {
		panic(err)
	}
}
