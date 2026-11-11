package manager

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// notInitedProto answers requests not wrapped in initConnection with errType
// while budget lasts, and requests wrapped in it with wrappedErr.
type notInitedProto struct {
	calls      []bin.Encoder
	errType    string
	budget     int
	wrappedErr error
}

func hasInitConnection(e bin.Encoder) bool {
	for {
		switch v := e.(type) {
		case *tg.InvokeWithoutUpdatesRequest:
			e = v.Query
		case *tg.InvokeWithLayerRequest:
			e = v.Query
		case *tg.InitConnectionRequest:
			return true
		default:
			return false
		}
	}
}

func (p *notInitedProto) Invoke(_ context.Context, input bin.Encoder, _ bin.Decoder) error {
	p.calls = append(p.calls, input)
	if hasInitConnection(input) {
		return p.wrappedErr
	}
	if p.budget > 0 {
		p.budget--
		return tgerr.New(400, p.errType)
	}
	return nil
}

func (p *notInitedProto) Run(ctx context.Context, f func(ctx context.Context) error) error {
	return f(ctx)
}

func (*notInitedProto) Ping(context.Context) error { return nil }

func TestConnInvokeReinitAfterNotInited(t *testing.T) {
	for _, errType := range []string{"CONNECTION_NOT_INITED", "CONNECTION_LAYER_INVALID"} {
		for _, mode := range []ConnMode{ConnModeUpdates, ConnModeData} {
			t.Run(errType+"/"+mode.String(), func(t *testing.T) {
				a := require.New(t)
				p := &notInitedProto{errType: errType, budget: 1}
				c := newTestConn(mode, p)
				c.device.DeviceModel = "model"
				c.gotConfig.Signal()

				req := &tg.HelpGetNearestDCRequest{}
				a.NoError(c.Invoke(context.Background(), req, &tg.NearestDC{}))
				a.Len(p.calls, 2)
				a.False(hasInitConnection(p.calls[0]))
				a.True(hasInitConnection(p.calls[1]), "not-inited error must be retried with initConnection")
				a.False(c.needsInit.Load())

				// The re-sent request keeps the regular wrapping.
				outer := p.calls[1]
				if mode == ConnModeData {
					w, ok := outer.(*tg.InvokeWithoutUpdatesRequest)
					a.True(ok)
					outer = w.Query
				}
				withLayer, ok := outer.(*tg.InvokeWithLayerRequest)
				a.True(ok)
				a.Equal(tg.Layer, withLayer.Layer)
				q := withLayer.Query
				if mode == ConnModeData {
					w, ok := q.(*tg.InvokeWithoutUpdatesRequest)
					a.True(ok)
					q = w.Query
				}
				initReq, ok := q.(*tg.InitConnectionRequest)
				a.True(ok)
				a.Equal(42, initReq.APIID)
				a.Equal("model", initReq.DeviceModel)

				// Re-initialized: the next request is not wrapped.
				a.NoError(c.Invoke(context.Background(), req, &tg.NearestDC{}))
				a.Len(p.calls, 3)
				a.False(hasInitConnection(p.calls[2]))
			})
		}
	}
}

func TestConnInvokeReinitKeepsWrappingUntilSuccess(t *testing.T) {
	a := require.New(t)
	p := &notInitedProto{errType: "CONNECTION_NOT_INITED", budget: 1, wrappedErr: tgerr.New(500, "RPC_CALL_FAIL")}
	c := newTestConn(ConnModeUpdates, p)
	c.gotConfig.Signal()

	err := c.Invoke(context.Background(), &tg.HelpGetNearestDCRequest{}, &tg.NearestDC{})
	a.True(tgerr.Is(err, "RPC_CALL_FAIL"))
	// An RPC error from the wrapped request means initConnection was handled.
	a.False(c.needsInit.Load())

	p.budget = 1
	p.wrappedErr = context.DeadlineExceeded // transport-level: state unknown
	err = c.Invoke(context.Background(), &tg.HelpGetNearestDCRequest{}, &tg.NearestDC{})
	a.ErrorIs(err, context.DeadlineExceeded)
	a.True(c.needsInit.Load(), "must stay not-inited until a wrapped request succeeds")

	p.wrappedErr = nil
	a.NoError(c.Invoke(context.Background(), &tg.HelpGetNearestDCRequest{}, &tg.NearestDC{}))
	a.True(hasInitConnection(p.calls[len(p.calls)-1]), "not-inited connection must send initConnection first")
	a.False(c.needsInit.Load())
}

func TestConnInvokeOtherErrorNoReinit(t *testing.T) {
	a := require.New(t)
	p := &notInitedProto{errType: "METHOD_INVALID", budget: 1}
	c := newTestConn(ConnModeUpdates, p)
	c.gotConfig.Signal()

	err := c.Invoke(context.Background(), &tg.HelpGetNearestDCRequest{}, &tg.NearestDC{})
	a.True(tgerr.Is(err, "METHOD_INVALID"))
	a.Len(p.calls, 1)
	a.False(c.needsInit.Load())
}
