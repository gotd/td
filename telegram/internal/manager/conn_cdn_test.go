package manager

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/tdsync"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

type captureProto struct {
	invokeCalls int
	lastInput   bin.Encoder
	inputs      []bin.Encoder
}

func (p *captureProto) Invoke(_ context.Context, input bin.Encoder, _ bin.Decoder) error {
	p.invokeCalls++
	p.lastInput = input
	p.inputs = append(p.inputs, input)
	return nil
}

func (p *captureProto) Run(ctx context.Context, f func(ctx context.Context) error) error {
	return f(ctx)
}

func (*captureProto) Ping(context.Context) error {
	return nil
}

func newTestConn(mode ConnMode, proto protoConn) *Conn {
	c := &Conn{
		mode:        mode,
		dc:          203,
		appID:       42,
		device:      DeviceConfig{AppVersion: "test-app"},
		proto:       proto,
		clock:       clock.System,
		log:         zap.NewNop(),
		handler:     NoopHandler{},
		sessionInit: tdsync.NewReady(),
		gotConfig:   tdsync.NewReady(),
		dead:        tdsync.NewReady(),
	}
	if mode == ConnModeCDN {
		c.cdnNeedsInit.Store(true)
	}
	return c
}

func TestConnInitCDNNoHelpGetConfig(t *testing.T) {
	a := require.New(t)
	p := &captureProto{}
	c := newTestConn(ConnModeCDN, p)

	a.NoError(c.init(context.Background()))
	a.Equal(0, p.invokeCalls)

	select {
	case <-c.gotConfig.Ready():
	case <-time.After(time.Second):
		a.Fail("gotConfig should be signaled for CDN mode")
	}

	c.mux.Lock()
	cfg := c.cfg
	c.mux.Unlock()
	a.Equal(203, cfg.ThisDC)
}

func TestConnInvokeCDNWrappedUsesInitConnection(t *testing.T) {
	a := require.New(t)
	p := &captureProto{}
	c := newTestConn(ConnModeCDN, p)
	c.device = DeviceConfig{
		DeviceModel:    "private-model",
		SystemVersion:  "private-os",
		AppVersion:     "1.2.3",
		SystemLangCode: "ru",
		LangPack:       "ru-pack",
		LangCode:       "ru",
		Proxy: tg.InputClientProxy{
			Address: "127.0.0.1",
			Port:    1080,
		},
		Params: &tg.JSONObject{
			Value: []tg.JSONObjectValue{
				{
					Key:   "tz_offset",
					Value: &tg.JSONNumber{Value: 10800},
				},
			},
		},
	}
	err := c.invokeCDNWrapped(
		context.Background(),
		&tg.UploadGetCDNFileRequest{FileToken: []byte{1}, Offset: 0, Limit: 1024},
		&tg.UploadCDNFileBox{},
	)
	a.NoError(err)

	req, ok := p.lastInput.(*tg.InvokeWithLayerRequest)
	a.True(ok)
	a.Equal(tg.Layer, req.Layer)

	_, wrapped := req.Query.(*tg.InvokeWithoutUpdatesRequest)
	a.False(wrapped, "CDN query must not use invokeWithoutUpdates")

	initReq, ok := req.Query.(*tg.InitConnectionRequest)
	a.True(ok)
	a.Equal(42, initReq.APIID)
	a.Equal("n/a", initReq.DeviceModel)
	a.Equal("n/a", initReq.SystemVersion)
	a.Equal("1.2.3", initReq.AppVersion)
	a.Equal("ru", initReq.SystemLangCode)
	a.Equal("ru-pack", initReq.LangPack)
	a.Equal("ru", initReq.LangCode)
	a.Equal(tg.InputClientProxy{Address: "127.0.0.1", Port: 1080}, initReq.Proxy)

	params, ok := initReq.Params.(*tg.JSONObject)
	a.True(ok)
	a.Equal(
		[]tg.JSONObjectValue{{Key: "tz_offset", Value: &tg.JSONNumber{Value: 10800}}},
		params.Value,
	)

	query, ok := initReq.Query.(noopDecoder)
	a.True(ok)
	_, ok = query.Encoder.(*tg.UploadGetCDNFileRequest)
	a.True(ok)
}

type retryOnRawNotInitedProto struct {
	calls        []bin.Encoder
	rawErrBudget int
}

func (p *retryOnRawNotInitedProto) Invoke(_ context.Context, input bin.Encoder, _ bin.Decoder) error {
	p.calls = append(p.calls, input)
	if _, ok := input.(*tg.UploadGetCDNFileRequest); ok && p.rawErrBudget > 0 {
		p.rawErrBudget--
		return tgerr.New(400, "CONNECTION_NOT_INITED")
	}
	return nil
}

func (p *retryOnRawNotInitedProto) Run(ctx context.Context, f func(ctx context.Context) error) error {
	return f(ctx)
}

func (*retryOnRawNotInitedProto) Ping(context.Context) error {
	return nil
}

type rawMethodInvalidProto struct {
	calls        []bin.Encoder
	rawErrBudget int
}

func (p *rawMethodInvalidProto) Invoke(_ context.Context, input bin.Encoder, _ bin.Decoder) error {
	p.calls = append(p.calls, input)
	if _, ok := input.(*tg.UploadGetCDNFileRequest); ok && p.rawErrBudget > 0 {
		p.rawErrBudget--
		return tgerr.New(400, "METHOD_INVALID")
	}
	return nil
}

func (p *rawMethodInvalidProto) Run(ctx context.Context, f func(ctx context.Context) error) error {
	return f(ctx)
}

func (*rawMethodInvalidProto) Ping(context.Context) error {
	return nil
}

func TestConnInvokeCDNFirstCallWrapped(t *testing.T) {
	a := require.New(t)
	p := &captureProto{}
	c := newTestConn(ConnModeCDN, p)
	c.gotConfig.Signal()

	err := c.Invoke(
		context.Background(),
		&tg.UploadGetCDNFileRequest{FileToken: []byte{1}, Offset: 0, Limit: 1024},
		&tg.UploadCDNFileBox{},
	)
	a.NoError(err)
	a.Len(p.inputs, 1)

	_, wrapped := p.inputs[0].(*tg.InvokeWithLayerRequest)
	a.True(wrapped, "first CDN call must be wrapped with invokeWithLayer(initConnection)")
}

func TestConnInvokeCDNSecondCallRawAfterInit(t *testing.T) {
	a := require.New(t)
	p := &captureProto{}
	c := newTestConn(ConnModeCDN, p)
	c.gotConfig.Signal()

	req := &tg.UploadGetCDNFileRequest{FileToken: []byte{1}, Offset: 0, Limit: 1024}
	a.NoError(c.Invoke(context.Background(), req, &tg.UploadCDNFileBox{}))
	a.NoError(c.Invoke(context.Background(), req, &tg.UploadCDNFileBox{}))
	a.Len(p.inputs, 2)

	_, firstWrapped := p.inputs[0].(*tg.InvokeWithLayerRequest)
	a.True(firstWrapped, "first CDN call must initialize connection via wrapper")
	_, secondRaw := p.inputs[1].(*tg.UploadGetCDNFileRequest)
	a.True(secondRaw, "after successful init, next CDN call must be raw")
}

func TestConnInvokeCDNRawNotInitedRetryWrappedThenRaw(t *testing.T) {
	a := require.New(t)
	p := &retryOnRawNotInitedProto{rawErrBudget: 1}
	c := newTestConn(ConnModeCDN, p)
	c.gotConfig.Signal()

	req := &tg.UploadGetCDNFileRequest{FileToken: []byte{1}, Offset: 0, Limit: 1024}
	a.NoError(c.Invoke(context.Background(), req, &tg.UploadCDNFileBox{}))
	a.NoError(c.Invoke(context.Background(), req, &tg.UploadCDNFileBox{}))
	a.NoError(c.Invoke(context.Background(), req, &tg.UploadCDNFileBox{}))
	a.Len(p.calls, 4)

	_, ok := p.calls[0].(*tg.InvokeWithLayerRequest)
	a.True(ok, "cold start must be wrapped")
	_, ok = p.calls[1].(*tg.UploadGetCDNFileRequest)
	a.True(ok, "inited state must try raw")
	_, ok = p.calls[2].(*tg.InvokeWithLayerRequest)
	a.True(ok, "CONNECTION_NOT_INITED from raw must retry wrapped")
	_, ok = p.calls[3].(*tg.UploadGetCDNFileRequest)
	a.True(ok, "after successful retry, state must return to raw")
}

func TestConnInvokeCDNRawMethodInvalidNoWrappedFallback(t *testing.T) {
	a := require.New(t)
	p := &rawMethodInvalidProto{rawErrBudget: 1}
	c := newTestConn(ConnModeCDN, p)
	c.gotConfig.Signal()

	req := &tg.UploadGetCDNFileRequest{FileToken: []byte{1}, Offset: 0, Limit: 1024}
	a.NoError(c.Invoke(context.Background(), req, &tg.UploadCDNFileBox{}))

	err := c.Invoke(context.Background(), req, &tg.UploadCDNFileBox{})
	a.Error(err)
	a.True(tgerr.Is(err, "METHOD_INVALID"))
	a.Len(p.calls, 2)
	_, ok := p.calls[0].(*tg.InvokeWithLayerRequest)
	a.True(ok, "initial request must be wrapped")
	_, ok = p.calls[1].(*tg.UploadGetCDNFileRequest)
	a.True(ok, "raw METHOD_INVALID should be returned as-is")
}

func TestConnInvokeDataKeepsInvokeWithoutUpdates(t *testing.T) {
	a := require.New(t)
	p := &captureProto{}
	c := newTestConn(ConnModeData, p)
	c.gotConfig.Signal()

	a.NoError(c.Invoke(context.Background(), &tg.HelpGetConfigRequest{}, &tg.Config{}))

	outer, ok := p.lastInput.(*tg.InvokeWithoutUpdatesRequest)
	a.True(ok)

	withLayer, ok := outer.Query.(*tg.InvokeWithLayerRequest)
	a.True(ok)
	a.Equal(tg.Layer, withLayer.Layer)

	inner, ok := withLayer.Query.(*tg.InvokeWithoutUpdatesRequest)
	a.True(ok)

	query, ok := inner.Query.(noopDecoder)
	a.True(ok)
	_, ok = query.Encoder.(*tg.HelpGetConfigRequest)
	a.True(ok)
}
