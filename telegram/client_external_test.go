package telegram_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/tdsync"
	"github.com/gotd/td/internal/testutil"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/telegram/internal/e2etest"
	"github.com/gotd/td/telegram/internal/tgtest"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
)

func testTransportExternal(p dcs.Protocol, storage session.Storage) func(t *testing.T) {
	return func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		log := zaptest.NewLogger(t)
		defer func() { _ = log.Sync() }()

		err := telegram.TestClient(ctx, telegram.Options{
			Logger:         log.Named("client"),
			SessionStorage: storage,
			Resolver:       dcs.PlainResolver(dcs.PlainOptions{Protocol: p}),
		}, func(ctx context.Context, client *telegram.Client) error {
			if _, err := client.Self(ctx); err != nil {
				return xerrors.Errorf("self: %w", err)
			}

			return nil
		})

		require.NoError(t, err)
	}
}

func TestExternalE2EConnect(t *testing.T) {
	testutil.SkipExternal(t)

	// To re-use session.
	storage := &session.StorageMemory{}
	t.Run("Abridged", testTransportExternal(transport.Abridged, storage))
	t.Run("Intermediate", testTransportExternal(transport.Intermediate, storage))
	t.Run("PaddedIntermediate", testTransportExternal(transport.PaddedIntermediate, storage))
	t.Run("Full", testTransportExternal(transport.Full, storage))
}

const dialog = `— Да?
— Алё!
— Да да?
— Ну как там с деньгами?
— А?
— Как с деньгами-то там?
— Чё с деньгами?
— Чё?
— Куда ты звонишь?
— Тебе звоню.
— Кому?
— Ну тебе.`

func TestExternalE2EUsersDialog(t *testing.T) {
	testutil.SkipExternal(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	log := zaptest.NewLogger(t).WithOptions(zap.IncreaseLevel(zapcore.InfoLevel))

	cfg := e2etest.TestConfig{
		AppID:   telegram.TestAppID,
		AppHash: telegram.TestAppHash,
		DcID:    2,
	}
	suite := e2etest.NewSuite(tgtest.NewSuite(ctx, t, log), cfg, rand.Reader)

	auth := make(chan *tg.User, 1)
	g := tdsync.NewLogGroup(ctx, log.Named("group"))

	g.Go("echobot", func(ctx context.Context) error {
		return e2etest.NewEchoBot(suite, auth).Run(ctx)
	})

	user, ok := <-auth
	if ok {
		g.Go("terentyev", func(ctx context.Context) error {
			defer g.Cancel()
			return e2etest.NewUser(suite, strings.Split(dialog, "\n"), user.Username).Run(ctx)
		})
	}

	require.NoError(t, g.Wait())
}

func TestExternalE2EPartialUpload(t *testing.T) {
	testutil.SkipExternal(t)

	// Testing partial uploads.
	//
	// Partial uploads can ease some invariants:
	// 	* Know total chunk count only on last chunk
	//	* Upload first chunk (part=0) after last chunk
	//
	// Note that overwriting (re-uploading part with different content) is
	// unreliable and works only on very small files, so if you want to
	// finalize first chunk later, don't upload it until finalization.

	log := zaptest.NewLogger(t).WithOptions(zap.IncreaseLevel(zapcore.InfoLevel))
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	cfg := e2etest.TestConfig{
		AppID:   telegram.TestAppID,
		AppHash: telegram.TestAppHash,
		DcID:    2,
	}
	suite := e2etest.NewSuite(tgtest.NewSuite(ctx, t, log), cfg, rand.Reader)

	client := suite.Client(log, nil)
	require.NoError(t, client.Run(ctx, func(ctx context.Context) error {
		if err := suite.RetryAuthenticate(ctx, client); err != nil {
			return err
		}

		raw := tg.NewClient(client)
		fileID, err := crypto.RandInt64(rand.Reader)
		if err != nil {
			return err
		}

		const (
			totalParts    = 10
			allocParts    = totalParts + 1
			lastPart      = totalParts - 1
			chunkSize     = 1024 * 512 / 4
			lastChunkSize = chunkSize / 2
		)

		// Uploading chunks from 1 to totalParts.
		var totalSize int
		for i := 1; i < totalParts; i++ {
			req := &tg.UploadSaveBigFilePartRequest{
				FileID:         fileID,
				FilePart:       i,
				FileTotalParts: allocParts,
				Bytes:          bytes.Repeat([]byte{byte(i + 1)}, chunkSize),
			}

			// Last chunk should finalize part size, changing it from allocated parts
			// to real part count.
			if req.FilePart == lastPart {
				req.Bytes = req.Bytes[:lastChunkSize]
				req.FileTotalParts = totalParts
			}
			totalSize += len(req.Bytes)
			if _, err := raw.UploadSaveBigFilePart(ctx, req); err != nil {
				return xerrors.Errorf("part: %w", err)
			}
			t.Logf("Uploaded part %d of %d (len: %d)", req.FilePart+1, req.FileTotalParts, len(req.Bytes))
		}

		// Uploading first chunk.
		req := &tg.UploadSaveBigFilePartRequest{
			FileID:         fileID,
			FilePart:       0,
			FileTotalParts: totalParts,
			Bytes:          bytes.Repeat([]byte{50}, chunkSize),
		}
		totalSize += len(req.Bytes)
		if _, err := raw.UploadSaveBigFilePart(ctx, req); err != nil {
			return xerrors.Errorf("part: %w", err)
		}

		f := &tg.InputFileBig{
			ID:    fileID,
			Parts: totalParts,
			Name:  "data.bin",
		}

		mc, err := message.NewSender(raw).Self().UploadMedia(ctx, message.File(f))
		if err != nil {
			return xerrors.Errorf("file: %w", err)
		}

		doc, ok := mc.(*tg.MessageMediaDocument).Document.AsNotEmpty()
		if !ok {
			return xerrors.New("bad doc")
		}

		buf := new(bytes.Buffer)
		if _, err := downloader.NewDownloader().Download(raw, doc.AsInputDocumentFileLocation()).Stream(ctx, buf); err != nil {
			return err
		}

		assert.Equal(t, totalSize, buf.Len())
		assert.Equal(t, req.Bytes, buf.Bytes()[:chunkSize])

		t.Logf("File uploaded (id=%d, parts=%d, name=%s, size %d kb)", f.ID, f.Parts, f.Name, totalSize/1024)

		return nil
	}))
}
