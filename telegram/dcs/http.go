package dcs

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/proto/codec"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
)

const (
	// defaultHTTPMaxWait is the default http_wait max_wait (ms): the server holds
	// the long-poll response open up to this long. Matches the MTProto default.
	defaultHTTPMaxWait = 25000
	// httpClientTimeoutMargin is added on top of max_wait for the HTTP client
	// timeout so a full-length long-poll response is not cut off.
	httpClientTimeoutMargin = 5 * time.Second
	// httpWaitRetryInterval throttles the poll loop when a wait frame cannot be
	// built (e.g. during key exchange) or a POST fails, avoiding a busy loop.
	httpWaitRetryInterval = 500 * time.Millisecond
	// httpInboxSize buffers response frames until Recv consumes them.
	httpInboxSize = 32
	// httpMaxConnsPerHost bounds concurrent connections of the default client to
	// a single DC, so a burst of sends cannot exhaust file descriptors.
	httpMaxConnsPerHost = 16
)

// recvResult is a decoded response delivered to Recv: either a frame or a
// transport-level protocol error (e.g. auth_key_not_found).
type recvResult struct {
	data []byte
	err  error
}

// httpConn implements transport.Conn over the MTProto HTTP transport.
//
// See https://core.telegram.org/mtproto/transports#http-transport.
//
// MTProto over HTTP is request/response: the server can only deliver messages
// as the body of a response to a client POST. Send posts one raw MTProto frame
// (framing is done by HTTP Content-Length, there is no codec tag) and the
// response body — the messages the server had queued for the session — is
// delivered to Recv through inbox. Send is non-blocking: the POST and its
// (possibly long-polling) response are handled on a separate goroutine so a 25s
// http_wait long-poll never stalls the mtproto write path.
type httpConn struct {
	client *http.Client
	// urls holds one /api endpoint per candidate DC address. Requests use the
	// current index; a failed POST rotates to the next candidate.
	urls   []string
	urlIdx atomic.Uint32

	// inbox buffers responses (frames or protocol errors) until Recv consumes
	// them.
	inbox chan recvResult

	// http_wait parameters in milliseconds, reported via HTTPWaitParams and used
	// by the mtproto layer to build http_wait messages.
	maxDelay  int
	waitAfter int
	maxWait   int

	ctx       context.Context
	cancel    context.CancelFunc
	startOnce sync.Once
	closeOnce sync.Once
}

func newHTTPConn(client *http.Client, urls []string, maxDelay, waitAfter, maxWait int) *httpConn {
	ctx, cancel := context.WithCancel(context.Background())
	return &httpConn{
		client:    client,
		urls:      urls,
		inbox:     make(chan recvResult, httpInboxSize),
		maxDelay:  maxDelay,
		waitAfter: waitAfter,
		maxWait:   maxWait,
		ctx:       ctx,
		cancel:    cancel,
	}
}

var _ transport.Conn = (*httpConn)(nil)

// Send posts a single MTProto frame. It is non-blocking: the POST and response
// are handled on a separate goroutine, so a long-poll response never stalls the
// caller. Delivery reliability is provided by the mtproto rpc engine
// (acks/retransmits) and liveness by the ping loop; a failed POST is therefore
// dropped here and recovered upstream.
func (c *httpConn) Send(ctx context.Context, b *bin.Buffer) error {
	select {
	case <-c.ctx.Done():
		return errors.Wrap(net.ErrClosed, "send")
	default:
	}

	// b.Buf is reused by the caller once Send returns; copy before handing it to
	// the POST goroutine.
	frame := make([]byte, len(b.Buf))
	copy(frame, b.Buf)
	// One goroutine per send: bounded in practice by the mtproto rpc engine's
	// in-flight window and by the client's MaxConnsPerHost (which the goroutines
	// block on). All are cancelled on Close via c.ctx.
	go func() { _, _ = c.roundtrip(frame) }()
	return nil
}

// Recv blocks until an incoming frame or a transport-level protocol error is
// available.
func (c *httpConn) Recv(ctx context.Context, b *bin.Buffer) error {
	select {
	case r := <-c.inbox:
		if r.err != nil {
			return r.err
		}
		b.Buf = append(b.Buf[:0], r.data...)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-c.ctx.Done():
		return errors.Wrap(net.ErrClosed, "recv")
	}
}

// Close releases the connection and cancels in-flight requests.
func (c *httpConn) Close() error {
	c.closeOnce.Do(c.cancel)
	return nil
}

// roundtrip posts one frame and delivers the response to inbox. It reports
// whether anything was delivered (a frame or a protocol error) and returns an
// error only for transport-level failures (POST error, non-200), which callers
// use to back off.
func (c *httpConn) roundtrip(frame []byte) (delivered bool, _ error) {
	url := c.urls[c.urlIndex()]
	req, err := http.NewRequestWithContext(c.ctx, http.MethodPost, url, bytes.NewReader(frame))
	if err != nil {
		return false, errors.Wrap(err, "build request")
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := c.client.Do(req)
	if err != nil {
		c.rotateURL() // fail over to the next candidate address
		return false, errors.Wrap(err, "do request")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, resp.Body)
		return false, errors.Errorf("unexpected status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, errors.Wrap(err, "read body")
	}
	switch {
	case len(body) == 0:
		// Empty body is a valid "dummy" long-poll response (max_wait elapsed with
		// no messages); nothing to deliver.
		return false, nil
	case len(body) == bin.Word:
		// A 4-byte body is a transport-level protocol error (e.g. -404 auth key
		// not found), not an encrypted message. The TCP codecs surface it via
		// checkProtocolError; do the same so the read loop can react (e.g.
		// regenerate the auth key) instead of feeding garbage to the decrypter.
		code := int32(binary.LittleEndian.Uint32(body))
		c.deliver(recvResult{err: &codec.ProtocolErr{Code: -code}})
		return true, nil
	default:
		c.deliver(recvResult{data: body})
		return true, nil
	}
}

// urlIndex returns the current candidate index. The modulo is taken on the
// unsigned counter so the index stays in range even if the counter wraps.
func (c *httpConn) urlIndex() int {
	return int(c.urlIdx.Load() % uint32(len(c.urls)))
}

func (c *httpConn) deliver(r recvResult) {
	select {
	case c.inbox <- r:
	case <-c.ctx.Done():
	}
}

// rotateURL advances to the next candidate address after a failed POST. Failover
// is best-effort: a burst of concurrent failures may advance the index by more
// than one, but the poll loop's round-robin still reaches every candidate.
func (c *httpConn) rotateURL() {
	if len(c.urls) > 1 {
		c.urlIdx.Add(1)
	}
}

// HTTPWaitParams reports the http_wait fields (milliseconds) the transport wants
// the mtproto layer to use. It is part of the mtproto HTTP long-poll capability.
func (c *httpConn) HTTPWaitParams() (maxDelay, waitAfter, maxWait int) {
	return c.maxDelay, c.waitAfter, c.maxWait
}

// StartHTTPWait starts the long-poll loop. frame yields a freshly-encrypted
// http_wait service message — only the mtproto layer can encrypt it. The loop
// keeps exactly one long-poll POST outstanding, re-issuing it as soon as the
// previous response arrives, to minimize update latency. It is safe to call at
// most once; further calls are no-ops.
//
// The loop runs in a goroutine whose lifetime is bound solely to Close: it is
// not part of the mtproto Run errgroup, so the connection must be closed to stop
// it (Conn.handleClose does this on teardown).
func (c *httpConn) StartHTTPWait(frame func(ctx context.Context) (*bin.Buffer, error)) {
	c.startOnce.Do(func() {
		go c.pollLoop(frame)
	})
}

func (c *httpConn) pollLoop(frame func(ctx context.Context) (*bin.Buffer, error)) {
	// A conforming server holds an empty response until max_wait; a response that
	// comes back much sooner than this signals a middlebox not honoring the
	// long-poll, and must be throttled to avoid a request storm.
	fastEmpty := time.Duration(c.maxWait) * time.Millisecond / 2

	for {
		if c.ctx.Err() != nil {
			return
		}
		b, err := frame(c.ctx)
		if err != nil {
			// Wait frame cannot be built (e.g. key exchange in progress) or the
			// connection is closing. Back off to avoid a hot loop.
			if !c.sleep(httpWaitRetryInterval) {
				return
			}
			continue
		}
		// Synchronous: blocks up to max_wait, then immediately loops to re-issue
		// the next http_wait (low update latency, one poll outstanding).
		start := time.Now()
		delivered, err := c.roundtrip(b.Buf)
		switch {
		case err != nil:
			// POST failed; back off instead of hammering a dead connection. The
			// ping loop ultimately detects and tears down a dead connection.
			if !c.sleep(httpWaitRetryInterval) {
				return
			}
		case !delivered && time.Since(start) < fastEmpty:
			// Instant empty 200 from a non-conforming middlebox: throttle.
			if !c.sleep(httpWaitRetryInterval) {
				return
			}
		}
	}
}

// sleep waits for d, returning false if the connection was closed while waiting.
func (c *httpConn) sleep(d time.Duration) bool {
	select {
	case <-c.ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}

var _ Resolver = httpResolver{}

type httpResolver struct {
	client     *http.Client
	scheme     string
	port       int
	preferIPv6 bool
	maxDelay   int
	waitAfter  int
	maxWait    int
	// startIdx rotates the preferred candidate across successive Primary calls
	// (i.e. across reconnects), so a dead first address does not trap the client.
	startIdx *atomic.Uint32
}

func (r httpResolver) urls(candidates []tg.DCOption) []string {
	// Rotate the starting candidate per call for cross-reconnect failover. Add(1)-1
	// makes the first call start at index 0 (the highest-priority DC); the modulo
	// is taken on the unsigned counter to stay in range on wrap.
	off := int((r.startIdx.Add(1) - 1) % uint32(len(candidates)))
	urls := make([]string, len(candidates))
	for i := range candidates {
		dc := candidates[(off+i)%len(candidates)]
		addr := net.JoinHostPort(dc.IPAddress, strconv.Itoa(r.port))
		urls[i] = r.scheme + "://" + addr + "/api"
	}
	return urls
}

func (r httpResolver) Primary(ctx context.Context, dc int, list List) (transport.Conn, error) {
	candidates := FindPrimaryDCs(list.Options, dc, r.preferIPv6)
	// The HTTP transport cannot speak to TCP-obfuscated-only DCs.
	n := 0
	for _, x := range candidates {
		if !x.TCPObfuscatedOnly {
			candidates[n] = x
			n++
		}
	}
	candidates = candidates[:n]
	if len(candidates) == 0 {
		return nil, errors.Errorf("no addresses for DC %d", dc)
	}
	return newHTTPConn(r.client, r.urls(candidates), r.maxDelay, r.waitAfter, r.maxWait), nil
}

func (r httpResolver) MediaOnly(ctx context.Context, dc int, list List) (transport.Conn, error) {
	return nil, errors.Errorf("can't resolve %d: MediaOnly is unsupported", dc)
}

func (r httpResolver) CDN(ctx context.Context, dc int, list List) (transport.Conn, error) {
	return nil, errors.Errorf("can't resolve %d: CDN is unsupported", dc)
}

// HTTPOptions is HTTP resolver creation options.
type HTTPOptions struct {
	// Client is the HTTP client used for POST requests. If nil, a client with a
	// bounded transport and a timeout derived from MaxWait is used. A custom
	// client's timeout must exceed MaxWait, otherwise long-poll responses are cut
	// off. For Scheme "https" a custom client with an appropriate tls.Config
	// (SNI/certificates for the DC) is required.
	Client *http.Client
	// Scheme is "http" (default) or "https".
	//
	// Note: https support is experimental. Connecting over https to a bare DC IP
	// needs a custom Client whose tls.Config matches the DC certificate; the
	// default client will fail TLS verification.
	Scheme string
	// Port overrides the transport port. Defaults to 80 for http and 443 for
	// https.
	//
	// The HTTP transport uses this fixed port for every DC and intentionally
	// ignores tg.DCOption.Port (which is the TCP MTProto port), per the MTProto
	// HTTP transport specification.
	Port int
	// PreferIPv6 gives IPv6 DCs higher precedence.
	// Default is to prefer IPv4 DCs over IPv6.
	PreferIPv6 bool
	// MaxDelay, WaitAfter and MaxWait are the http_wait fields in milliseconds.
	// Defaults: 0, 0, 25000.
	//
	// See https://core.telegram.org/mtproto/service_messages.
	MaxDelay  int
	WaitAfter int
	MaxWait   int
}

func (o *HTTPOptions) setDefaults() {
	if o.Scheme == "" {
		o.Scheme = "http"
	}
	if o.Port == 0 {
		if o.Scheme == "https" {
			o.Port = 443
		} else {
			o.Port = 80
		}
	}
	if o.MaxWait == 0 {
		o.MaxWait = defaultHTTPMaxWait
	}
	if o.Client == nil {
		o.Client = &http.Client{
			Timeout: time.Duration(o.MaxWait)*time.Millisecond + httpClientTimeoutMargin,
			Transport: &http.Transport{
				Proxy:               http.ProxyFromEnvironment,
				MaxConnsPerHost:     httpMaxConnsPerHost,
				MaxIdleConnsPerHost: httpMaxConnsPerHost,
			},
		}
	}
}

// HTTP creates an MTProto-over-HTTP DC resolver with http_wait long-polling.
//
// It suits environments where a raw persistent MTProto TCP socket cannot be
// held but HTTP POST to the DC is possible. Updates are delivered via http_wait
// long-polling, so their latency is bounded by the round-trip rather than being
// instantaneous. MediaOnly and CDN DCs are not supported.
//
// See https://core.telegram.org/mtproto/transports#http-transport.
func HTTP(opts HTTPOptions) Resolver {
	opts.setDefaults()
	return httpResolver{
		client:     opts.Client,
		scheme:     opts.Scheme,
		port:       opts.Port,
		preferIPv6: opts.PreferIPv6,
		maxDelay:   opts.MaxDelay,
		waitAfter:  opts.WaitAfter,
		maxWait:    opts.MaxWait,
		startIdx:   new(atomic.Uint32),
	}
}
