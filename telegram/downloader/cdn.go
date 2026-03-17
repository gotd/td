package downloader

import (
	"io"
	"sync"

	"golang.org/x/sync/singleflight"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

// cdn is a download schema that starts on the master DC and switches to CDN on
// upload.fileCdnRedirect without losing the original request.
type cdn struct {
	provider CDNProvider
	client   Client
	pool     *bin.Pool
	// retryHandler observes retried transient downloader errors.
	retryHandler RetryHandler

	// master preserves regular path that may return redirect errors when
	// allowCDN=true.
	master master
	// max is forwarded to provider pool creation, usually mapped from number of
	// download threads.
	max int64
	// verify enables inline verification for decrypted CDN chunks.
	verify bool

	// stateMux guards mode/redirect/client pointer/revision.
	stateMux sync.RWMutex
	// refreshMux serializes redirect refreshes so only one goroutine asks master
	// for new token when CDN reports token invalid.
	refreshMux sync.Mutex
	// clientMux serializes CDN client (re)creation per schema instance.
	clientMux sync.Mutex
	// hashesMux guards in-memory cache of CDN hashes by offset.
	hashesMux sync.RWMutex
	// windowsMux guards bounded cache of verified CDN hash windows used to
	// handle custom part sizes that split hash windows.
	windowsMux sync.Mutex
	// windowsLoad deduplicates concurrent fetches of the same full hash window.
	windowsLoad singleflight.Group

	mode        cdnMode
	redirect    *tg.UploadFileCDNRedirect
	cdn         CDN
	closer      io.Closer
	clientDC    int
	rev         uint64
	hashes      map[int64]tg.FileHash
	hashOffsets []int64
	windows     map[int64][]byte
	windowsFIFO []int64
}

var _ schema = (*cdn)(nil)

type cdnMode uint8

const (
	// modeMaster means request master first and switch only on redirect.
	modeMaster cdnMode = iota
	// modeCDN means active redirect exists and chunks should be fetched from CDN.
	modeCDN
)

// maxVerifiedWindowCache bounds memory used by split-window verification.
//
// Split windows are needed only when downloader part size does not align with
// Telegram CDN hash window size (typically 128KB). In that case we may fetch a
// full hash window once, verify it, and reuse verified bytes for neighboring
// chunks. A small bounded cache is enough because sequential/parallel readers
// usually work on nearby offsets.
const maxVerifiedWindowCache = 16

func newCDNSchema(
	masterSchema master,
	provider CDNProvider,
	pool *bin.Pool,
	max int64,
	verifyCDNInline bool,
	retryHandler RetryHandler,
) *cdn {
	if max < 1 {
		max = 1
	}

	return &cdn{
		provider:     provider,
		client:       masterSchema.client,
		pool:         pool,
		retryHandler: retryHandler,
		master:       masterSchema,
		max:          max,
		verify:       verifyCDNInline,
		mode:         modeMaster,
	}
}

func (c *cdn) reportRetry(operation string, attempt int, err error) {
	if attempt < 1 || err == nil || c.retryHandler == nil {
		return
	}
	c.retryHandler(RetryEvent{
		Operation: operation,
		Attempt:   attempt,
		Err:       err,
	})
}

func (c *cdn) Close() error {
	// Close is called by Builder defer path and should release only schema-local
	// CDN resources. Shared client-level pools are managed in telegram package.
	c.stateMux.Lock()
	closer := c.closer
	c.cdn = nil
	c.closer = nil
	c.clientDC = 0
	c.stateMux.Unlock()

	if closer != nil {
		return closer.Close()
	}
	return nil
}

func (c *cdn) closeClient() {
	// Internal best-effort close used on fingerprint/token recovery loops.
	c.stateMux.Lock()
	closer := c.closer
	c.cdn = nil
	c.closer = nil
	c.clientDC = 0
	c.stateMux.Unlock()

	if closer != nil {
		_ = closer.Close()
	}
}

func (c *cdn) snapshot() (mode cdnMode, redirect *tg.UploadFileCDNRedirect, rev uint64) {
	c.stateMux.RLock()
	defer c.stateMux.RUnlock()

	return c.mode, c.redirect, c.rev
}

func (c *cdn) setRedirect(redirect *tg.UploadFileCDNRedirect) {
	c.stateMux.Lock()
	c.mode = modeCDN
	c.redirect = redirect
	c.rev++
	c.stateMux.Unlock()

	// Redirect update invalidates hash cache scope (file token / offset range
	// may change). Seed with hashes returned in redirect when available.
	c.resetHashes()
	c.resetWindows()
	if redirect != nil {
		c.cacheHashes(redirect.FileHashes)
	}
}

func (c *cdn) setMaster() {
	c.stateMux.Lock()
	c.mode = modeMaster
	c.redirect = nil
	c.rev++
	c.stateMux.Unlock()

	// Leaving CDN mode invalidates CDN hash cache.
	c.resetHashes()
	c.resetWindows()
}
