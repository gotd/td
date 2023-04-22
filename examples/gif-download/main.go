// Binary gif-download implements example of gif backup.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/go-faster/errors"
	"github.com/gotd/contrib/middleware/ratelimit"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"

	"github.com/gotd/td/examples"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/telegram/query/hasher"
	"github.com/gotd/td/tg"
)

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
	flow := auth.NewFlow(examples.Terminal{}, auth.SendCodeOptions{})

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
			return errors.Wrap(err, "auth")
		}

		if *inputDir != "" {
			// Handling bulk upload.
			// Probably we can de-duplicate gifs by some criteria.
			if err := upload(ctx, log, api, *inputDir); err != nil {
				return errors.Wrap(err, "upload")
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
				result, err := api.MessagesGetSavedGifs(ctx, int64(int(h.Sum())))
				if err != nil {
					return errors.Wrap(err, "get")
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
						return errors.Wrap(err, "download")
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
							return errors.Wrap(err, "remove")
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
