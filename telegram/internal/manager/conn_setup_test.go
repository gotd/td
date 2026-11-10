package manager

import (
	"context"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

// TestConnInitSetupInvokesBeforeReady checks that setup callback can invoke
// requests through the connection before it is ready (DC auth transfer).
//
// Previously setup got the Conn itself, whose Invoke waits for readiness that
// is signaled only after setup returns, so init deadlocked.
func TestConnInitSetupInvokesBeforeReady(t *testing.T) {
	a := require.New(t)
	p := &captureProto{}
	c := newTestConn(ConnModeData, p)
	c.connBackoff = func(ctx context.Context) backoff.BackOff { return &backoff.StopBackOff{} }
	c.setup = func(ctx context.Context, inv tg.Invoker) error {
		select {
		case <-c.Ready():
			t.Error("connection must not be ready during setup")
		default:
		}
		_, err := tg.NewClient(inv).AuthImportAuthorization(ctx, &tg.AuthImportAuthorizationRequest{
			ID: 1,
		})
		return err
	}

	// Timeout turns a deadlock into a failure.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	a.NoError(c.init(ctx))

	select {
	case <-c.Ready():
	default:
		a.Fail("connection must be ready after init")
	}

	// initConnection + auth.importAuthorization.
	a.Len(p.inputs, 2)
	outer, ok := p.inputs[1].(*tg.InvokeWithoutUpdatesRequest)
	a.True(ok, "%T", p.inputs[1])
	withLayer, ok := outer.Query.(*tg.InvokeWithLayerRequest)
	a.True(ok, "%T", outer.Query)
	a.Equal(c.layer, withLayer.Layer)
	inner, ok := withLayer.Query.(*tg.InvokeWithoutUpdatesRequest)
	a.True(ok, "%T", withLayer.Query)
	q, ok := inner.Query.(noopDecoder)
	a.True(ok, "%T", inner.Query)
	req, ok := q.Encoder.(*tg.AuthImportAuthorizationRequest)
	a.True(ok, "%T", q.Encoder)
	a.Equal(int64(1), req.ID)
}
