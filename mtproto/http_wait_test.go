package mtproto

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/mt"
)

// fakeHTTPWaiter is a transport.Conn that advertises the http_wait capability.
type fakeHTTPWaiter struct {
	maxDelay, waitAfter, maxWait int

	paramsCalled bool
	frame        func(ctx context.Context) (*bin.Buffer, error)
}

func (f *fakeHTTPWaiter) Send(context.Context, *bin.Buffer) error { return nil }
func (f *fakeHTTPWaiter) Recv(context.Context, *bin.Buffer) error { return nil }
func (f *fakeHTTPWaiter) Close() error                            { return nil }

func (f *fakeHTTPWaiter) HTTPWaitParams() (maxDelay, waitAfter, maxWait int) {
	f.paramsCalled = true
	return f.maxDelay, f.waitAfter, f.maxWait
}

func (f *fakeHTTPWaiter) StartHTTPWait(frame func(ctx context.Context) (*bin.Buffer, error)) {
	f.frame = frame
}

// plainConn is a transport.Conn without the http_wait capability.
type plainConn struct{}

func (plainConn) Send(context.Context, *bin.Buffer) error { return nil }
func (plainConn) Recv(context.Context, *bin.Buffer) error { return nil }
func (plainConn) Close() error                            { return nil }

func TestConn_startHTTPWait(t *testing.T) {
	a := require.New(t)
	c := newTestClient(func(int64, int32, bin.Encoder) (bin.Encoder, error) { return nil, nil })

	fake := &fakeHTTPWaiter{maxDelay: 1, waitAfter: 2, maxWait: 25000}
	c.conn = fake

	c.startHTTPWait()

	a.True(fake.paramsCalled, "HTTPWaitParams must be consulted")
	a.NotNil(fake.frame, "StartHTTPWait must receive a frame factory")

	// The factory must produce a valid encrypted http_wait carrying the params.
	buf, err := fake.frame(context.Background())
	a.NoError(err)
	a.NotEmpty(buf.Buf)

	// Decrypt server-side and decode the inner http_wait to verify the fields.
	msg, err := crypto.NewServerCipher(c.rand).DecryptFromBuffer(c.authKey, buf)
	a.NoError(err)
	var wait mt.HTTPWaitRequest
	a.NoError(wait.Decode(&bin.Buffer{Buf: msg.Data()}))
	a.Equal(1, wait.MaxDelay)
	a.Equal(2, wait.WaitAfter)
	a.Equal(25000, wait.MaxWait)
}

func TestConn_startHTTPWait_NonHTTPNoop(t *testing.T) {
	c := newTestClient(func(int64, int32, bin.Encoder) (bin.Encoder, error) { return nil, nil })
	c.conn = plainConn{}
	// A non-HTTP transport must not be driven; this must not panic.
	c.startHTTPWait()
}
