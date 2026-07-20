package manager

import (
	"context"
	"testing"

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
	c.gotConfig.Signal()

	a.NoError(c.Invoke(context.Background(), &tg.HelpGetConfigRequest{}, &tg.Config{}))

	outer, ok := p.lastInput.(*tg.InvokeWithoutUpdatesRequest)
	a.True(ok)

	withLayer, ok := outer.Query.(*tg.InvokeWithLayerRequest)
	a.True(ok)
	a.Equal(42, withLayer.Layer)
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
