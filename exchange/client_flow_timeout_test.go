package exchange

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/gotd/log/logzap"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/testutil"
	"github.com/gotd/td/transport"
)

// blackholeAfterNRecv wraps a transport.Conn so that Send behaves normally
// but every Recv beyond the first allowed calls blocks until ctx is done
// instead of returning data.
//
// This reproduces a server that answers the first key exchange steps and
// then goes silent: it "accepts our writes" (Send always succeeds, exactly
// like a real peer that received the client's data) yet never replies to
// the next request, leaving a bare conn.Recv with no way to unblock other
// than its context.
type blackholeAfterNRecv struct {
	transport.Conn
	allowed int32
	count   atomic.Int32
}

func (c *blackholeAfterNRecv) Recv(ctx context.Context, b *bin.Buffer) error {
	if c.count.Add(1) <= c.allowed {
		return c.Conn.Recv(ctx, b)
	}
	<-ctx.Done()
	return ctx.Err()
}

// TestClientExchangeRespectsTimeout pins ExchangeTimeout to the raw reads at
// key exchange steps 5 and 7.
//
// Exchanger.WithTimeout documents itself as setting the deadline of "every
// exchange request", and the server flow applies it to all three of its
// reads. The client flow applies it to every write and to the step 2 read,
// then reads steps 5 and 7 through a bare conn.Recv. transport.connection.Recv
// derives its socket deadline solely from ctx.Deadline(), so those two reads
// are unbounded whenever the caller's context carries no deadline -- which in
// PFS mode is always, because mtproto.Conn.connect scopes DialTimeout to the
// dial and hands the raw context to the exchange.
//
// The caller context here is deliberately context.Background(): the point is
// that a context with no deadline of its own must still be bounded by the
// configured ExchangeTimeout.
//
// Both sites are covered independently: allowing 1 reply through blackholes
// the step 5 read (ServerDHParams), allowing 2 blackholes the step 7 read
// (DhGen). A regression that reintroduces a bare conn.Recv at either site
// fails its corresponding case.
func TestClientExchangeRespectsTimeout(t *testing.T) {
	tests := []struct {
		name          string
		allowed       int32
		wantErrSubstr string
	}{
		{
			name:          "step5 ServerDHParams read",
			allowed:       1,
			wantErrSubstr: "read ServerDHParams message",
		},
		{
			name:          "step7 DhGen read",
			allowed:       2,
			wantErrSubstr: "read DhGen message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const dc = 2

			log := zaptest.NewLogger(t)
			privateKey := PrivateKey{RSA: testutil.RSAPrivateKey()}

			i := transport.Intermediate
			rawClient, server := i.Pipe()

			// Real server side flow drives the replies allowed through. Once
			// the client below stops draining the pipe, the server's own next
			// Send blocks until t.Cleanup closes the connections; done is
			// closed on return so cleanup can wait for the goroutine to
			// actually exit instead of just unblocking it.
			done := make(chan struct{})
			go func() {
				defer close(done)
				_, _ = NewExchanger(server, dc).
					WithLogger(logzap.New(log.Named("server"))).
					WithRand(testutil.Rand([]byte("exchange-timeout-test-server"))).
					Server(privateKey).
					Run(context.Background())
			}()

			t.Cleanup(func() {
				_ = rawClient.Close()
				_ = server.Close()
				select {
				case <-done:
				case <-time.After(10 * time.Second):
					t.Error("server goroutine did not exit after connections were closed")
				}
			})

			client := &blackholeAfterNRecv{Conn: rawClient, allowed: tt.allowed}

			// The timeout under test bounds the blackholed raw read; it must
			// also comfortably cover the real RSA/DH driven reads that precede
			// it, which are slow under -race on a throttled CI runner. The
			// outer guard must in turn dwarf the real key exchange crypto that
			// runs before the step 7 read (RSA pad/decrypt, PQ factorization
			// and two 2048-bit DH modexps): that CPU bound work, not the read,
			// is what a tight guard trips over under the race detector.
			const (
				exchangeTimeout = 1 * time.Second
				hangGuard       = 30 * time.Second
			)

			e := NewExchanger(client, dc).
				WithLogger(logzap.New(log.Named("client"))).
				WithRand(testutil.Rand([]byte("exchange-timeout-test-client"))).
				WithTimeout(exchangeTimeout).
				Client([]PublicKey{privateKey.Public()})

			result := make(chan error, 1)
			go func() {
				_, err := e.Run(context.Background())
				result <- err
			}()

			select {
			case err := <-result:
				require.ErrorIs(t, err, context.DeadlineExceeded)
				require.ErrorContains(t, err, tt.wantErrSubstr)
			case <-time.After(hangGuard):
				t.Fatal("exchange hung despite configured timeout")
			}
		})
	}
}
