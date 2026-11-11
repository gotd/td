package decode_test

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/mt"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/wiretrace"
	"github.com/gotd/td/wiretrace/decode"
	"github.com/gotd/td/wiretrace/tljson"
	"github.com/gotd/td/wiretrace/wiretracetest"
)

func decodeRun(t *testing.T, dir string) []decode.Message {
	t.Helper()
	schema, err := tljson.LoadFiles(wiretracetest.SchemaFiles()...)
	require.NoError(t, err)
	run, err := wiretrace.Open(dir)
	require.NoError(t, err)
	defer func() { require.NoError(t, run.Close()) }()
	d := decode.New(schema, decode.Options{})
	events, err := run.All()
	require.NoError(t, err)
	for _, ev := range events {
		d.Index(ev)
	}
	var out []decode.Message
	for _, ev := range events {
		if ev.Kind != wiretrace.KindMsg {
			continue
		}
		body, err := run.Blob(ev)
		require.NoError(t, err)
		msgs, err := d.Messages(ev, body)
		require.NoError(t, err)
		out = append(out, msgs...)
	}
	return out
}

func byMsgID(t *testing.T, msgs []decode.Message, id uint64) decode.Message {
	t.Helper()
	for _, m := range msgs {
		if m.MsgIDValue == id {
			return m
		}
	}
	t.Fatalf("no message %d", id)
	return decode.Message{}
}

func bodyJSON(t *testing.T, m decode.Message) string {
	t.Helper()
	b, err := json.Marshal(m.Body)
	require.NoError(t, err)
	return string(b)
}

func TestSample(t *testing.T) {
	a := require.New(t)
	tr := wiretracetest.Sample(t)
	msgs := decodeRun(t, tr.Dir)
	a.Len(msgs, 13)

	box := strconv.FormatUint(wiretracetest.MsgOutBox, 10)
	init := byMsgID(t, msgs, wiretracetest.MsgInit)
	a.Equal("out", init.Dir)
	a.Equal(box, init.Container)
	a.Equal([]string{"invokeWithLayer:232", "initConnection"}, init.Wrap)
	a.Equal("invokeWithLayer", init.Type)
	a.Equal("help.getNearestDc", init.Method)
	a.Equal(int64(1), init.Req)
	a.Equal("mtproto", init.Caller.Subsystem, "req.caller written after the message is joined")
	a.Empty(init.Redacted)
	a.Equal("-42", init.Session)
	a.False(init.Partial)

	ack := byMsgID(t, msgs, wiretracetest.MsgAck)
	a.Equal("msgs_ack", ack.Type)
	a.Zero(ack.Req)
	a.Equal(`{"_":"msgs_ack","msg_ids":["1","2"]}`, bodyJSON(t, ack))

	signIn := byMsgID(t, msgs, wiretracetest.MsgSignIn)
	a.Equal("auth.signIn", signIn.Type)
	a.Equal("auth.signIn", signIn.Method)
	a.Equal([]string{"auth.signIn"}, signIn.Redacted)
	a.Equal(`{"_":"redacted","orig":"auth.signIn","len":40}`, bodyJSON(t, signIn))
	a.Equal("intro", signIn.Caller.Subsystem)

	nearest := byMsgID(t, msgs, wiretracetest.MsgInInit)
	a.Equal("in", nearest.Dir)
	a.Equal("rpc_result", nearest.Type)
	a.Equal("help.getNearestDc", nearest.Method)
	a.Equal("nearestDc", nearest.Result)
	a.Equal(strconv.FormatUint(wiretracetest.MsgInit, 10), nearest.ReqMsgID)
	a.NotNil(nearest.LatencyMS)
	a.InDelta(103, *nearest.LatencyMS, 0.01)
	a.Equal(int64(1), nearest.Req)
	a.Empty(nearest.Redacted)

	authRes := byMsgID(t, msgs, wiretracetest.MsgInSignIn)
	a.Equal("auth.authorization", authRes.Result)
	a.Equal([]string{"auth.signIn.result"}, authRes.Redacted)

	a.Equal("new_session_created", byMsgID(t, msgs, wiretracetest.MsgInSession).Type)
	a.Empty(byMsgID(t, msgs, wiretracetest.MsgInSession).Redacted)

	// Top-level bare Vector<long> result.
	del := byMsgID(t, msgs, wiretracetest.MsgInDelete)
	a.Equal("photos.deletePhotos", del.Method)
	a.Equal("vector", del.Result)
	a.Equal(`{"_":"rpc_result","req_msg_id":"`+strconv.FormatUint(wiretracetest.MsgDelete, 10)+`","result":["10","20"]}`,
		bodyJSON(t, del))
	a.Empty(del.Container)

	after := byMsgID(t, msgs, wiretracetest.MsgAfter)
	a.Equal([]string{"invokeAfterMsg"}, after.Wrap)
	a.Equal(strconv.FormatUint(wiretracetest.MsgDelete, 10), after.After)
	a.Equal("photos.deletePhotos", after.AfterMethod)
	a.Equal("help.getConfig", after.Method)

	cfg := byMsgID(t, msgs, wiretracetest.MsgInAfter)
	a.True(cfg.Gzip)
	a.Equal("config", cfg.Result)
	a.Equal("help.getConfig", cfg.Method)
	a.Contains(bodyJSON(t, cfg), `"ip_address":"127.0.0.1"`)

	flood := byMsgID(t, msgs, wiretracetest.MsgInSend)
	a.Equal("rpc_error", flood.Result)
	a.Equal(&decode.RPCError{Code: 420, Message: "FLOOD_WAIT_17"}, flood.Error)
	a.Equal(int64(5), flood.Req)

	unknown := byMsgID(t, msgs, wiretracetest.MsgInUnknown)
	a.True(unknown.Partial)
	a.Equal("updateShort", unknown.Type)
	a.Contains(unknown.DecodeError, "unknown constructor #deadbeef")
	a.Contains(bodyJSON(t, unknown), `"update":{"_":"#deadbeef"`)
}

func TestContainerFailure(t *testing.T) {
	a := require.New(t)
	tr := wiretracetest.New(t, wiretrace.WriterOptions{})
	// The item claims 100 bytes, the container has 4.
	var b bin.Buffer
	b.PutID(tljson.IDMsgContainer)
	b.PutInt(1)
	b.PutLong(1)
	b.PutInt(1)
	b.PutInt(100)
	b.PutInt(0)
	tr.In(wiretracetest.Item{MsgID: 9, Body: b.Copy()})
	tr.Close()
	msgs := decodeRun(t, tr.Dir)
	a.Len(msgs, 1)
	a.True(msgs[0].Partial)
	a.Equal("msg_container", msgs[0].Type)
	a.Contains(msgs[0].DecodeError, "overflows")
}

func TestGzipContainer(t *testing.T) {
	a := require.New(t)
	tr := wiretracetest.New(t, wiretrace.WriterOptions{})
	inner := wiretracetest.Container(
		wiretracetest.Item{MsgID: 5, SeqNo: 2, Body: wiretracetest.Encode(t, &mt.Pong{MsgID: 4, PingID: 1})},
	)
	tr.In(wiretracetest.Item{MsgID: 9, Body: wiretracetest.GZIP(t, inner)})
	tr.Close()
	msgs := decodeRun(t, tr.Dir)
	a.Len(msgs, 1)
	a.Equal("pong", msgs[0].Type)
	a.True(msgs[0].Gzip)
	a.Equal("9", msgs[0].Container)
}

func TestRedactionAttribution(t *testing.T) {
	a := require.New(t)
	tr := wiretracetest.New(t, wiretrace.WriterOptions{})
	signIn := wiretracetest.Encode(t, &tg.AuthSignInRequest{PhoneNumber: "+1", PhoneCodeHash: "h"})
	export := wiretracetest.Encode(t, &tg.AuthExportAuthorizationRequest{DCID: 2})
	tr.OutContainer(100, 6,
		wiretracetest.Item{MsgID: 104, Body: wiretracetest.Redact(signIn), Redacted: []string{"auth.signIn"}},
		wiretracetest.Item{MsgID: 108, Body: export},
	)
	auth := wiretracetest.Encode(t, &tg.AuthAuthorization{User: &tg.UserEmpty{ID: 1}})
	exported := wiretracetest.Encode(t, &tg.AuthExportedAuthorization{ID: 1, Bytes: []byte("secret")})
	tr.InContainer(101, 7,
		wiretracetest.Item{MsgID: 105, Body: wiretracetest.Redact12(wiretracetest.RPCResult(104, auth)),
			Redacted: []string{"auth.signIn.result"}},
		wiretracetest.Item{MsgID: 109, Body: wiretracetest.Redact12(wiretracetest.RPCResult(108, exported)),
			Redacted: []string{"auth.exportAuthorization.result"}},
	)
	tr.Close()
	msgs := decodeRun(t, tr.Dir)
	a.Len(msgs, 4)
	a.Equal([]string{"auth.signIn"}, byMsgID(t, msgs, 104).Redacted)
	a.Empty(byMsgID(t, msgs, 108).Redacted)
	a.Equal([]string{"auth.signIn.result"}, byMsgID(t, msgs, 105).Redacted)
	a.Equal([]string{"auth.exportAuthorization.result"}, byMsgID(t, msgs, 109).Redacted)
	a.Equal("auth.exportedAuthorization", byMsgID(t, msgs, 109).Result)
}
