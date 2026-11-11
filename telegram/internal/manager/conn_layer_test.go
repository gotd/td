package manager

import (
	"context"
	"testing"

	"github.com/cenkalti/backoff/v4"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/clock"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/tg"
)

func TestConnOptionsLayerDefault(t *testing.T) {
	a := require.New(t)

	opts := ConnOptions{}
	opts.setDefaults(clock.System)
	a.Equal(tg.Layer, opts.Layer)

	opts = ConnOptions{Layer: 42}
	opts.setDefaults(clock.System)
	a.Equal(42, opts.Layer)
}

func TestConnInvokeUsesConfiguredLayer(t *testing.T) {
	a := require.New(t)
	p := &captureProto{}
	c := newTestConn(ConnModeData, p)
	c.layer = 42
	c.connBackoff = func(ctx context.Context) backoff.BackOff { return &backoff.StopBackOff{} }

	// init sends the layer with initConnection.
	a.NoError(c.init(context.Background()))
	outer, ok := p.inputs[0].(*tg.InvokeWithoutUpdatesRequest)
	a.True(ok)
	withLayer, ok := outer.Query.(*tg.InvokeWithLayerRequest)
	a.True(ok)
	a.Equal(42, withLayer.Layer)

	// After init, queries go without invokeWithLayer, like Telegram Desktop
	// and TDLib.
	a.NoError(c.Invoke(context.Background(), &tg.HelpGetNearestDCRequest{}, &tg.NearestDC{}))
	outer, ok = p.lastInput.(*tg.InvokeWithoutUpdatesRequest)
	a.True(ok)
	q, ok := outer.Query.(noopDecoder)
	a.True(ok, "%T", outer.Query)
	_, ok = q.Encoder.(*tg.HelpGetNearestDCRequest)
	a.True(ok)

	// The layer is sent again only after the server loses the state.
	c.needsInit.Store(true)
	a.NoError(c.Invoke(context.Background(), &tg.HelpGetNearestDCRequest{}, &tg.NearestDC{}))
	outer, ok = p.lastInput.(*tg.InvokeWithoutUpdatesRequest)
	a.True(ok)
	withLayer, ok = outer.Query.(*tg.InvokeWithLayerRequest)
	a.True(ok)
	a.Equal(42, withLayer.Layer)
	a.False(c.needsInit.Load())
}

func TestCreateConnPropagatesLayer(t *testing.T) {
	a := require.New(t)

	conn := CreateConn(nil, ConnModeData, 42, mtproto.Options{Clock: clock.System}, ConnOptions{
		Layer: 42,
	})
	a.Equal(42, conn.layer)

	conn = CreateConn(nil, ConnModeData, 42, mtproto.Options{Clock: clock.System}, ConnOptions{})
	a.Equal(tg.Layer, conn.layer)
}
