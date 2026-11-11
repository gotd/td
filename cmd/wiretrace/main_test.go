package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/mt"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/wiretrace"
	"github.com/gotd/td/wiretrace/wiretracetest"
)

func runCmd(t *testing.T, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	var out, errOut bytes.Buffer
	code = run(context.Background(), args, &out, &errOut)
	return out.String(), errOut.String(), code
}

func schemaFlag() string {
	return "--schema=" + strings.Join(wiretracetest.SchemaFiles(), ",")
}

func TestDecode(t *testing.T) {
	a := require.New(t)
	tr := wiretracetest.Sample(t)
	out, errOut, code := runCmd(t, "decode", schemaFlag(), tr.Dir)
	a.Equal(0, code, errOut)
	a.Contains(errOut, "dropped 3 events")

	lines := strings.Split(strings.TrimSpace(out), "\n")
	a.Len(lines, 13)
	var msgs []map[string]any
	for _, l := range lines {
		var m map[string]any
		a.NoError(json.Unmarshal([]byte(l), &m), l)
		msgs = append(msgs, m)
	}
	first := msgs[0]
	a.Equal("out", first["dir"])
	a.Equal("help.getNearestDc", first["method"])
	a.Equal([]any{"invokeWithLayer:232", "initConnection"}, first["wrap"])
	a.Equal("mtproto", first["caller"].(map[string]any)["subsystem"])
	a.Contains(lines[0], `"body":{"_":"invokeWithLayer","layer":232,"query":{"_":"initConnection"`)
	a.Contains(lines[7], `"result":["10","20"]`, "bare Vector<long> result")
	a.Contains(lines[9], `"gzip":true`)
	a.Contains(lines[11], `"error":{"code":420,"message":"FLOOD_WAIT_17"}`)
	a.Contains(lines[12], `"partial":true`)

	// Filters.
	out, _, code = runCmd(t, "decode", schemaFlag(), "--dir=out", "--method=help.", "--no-body", tr.Dir)
	a.Equal(0, code)
	lines = strings.Split(strings.TrimSpace(out), "\n")
	a.Len(lines, 2)
	for _, l := range lines {
		a.Contains(l, `"dir":"out"`)
		a.Contains(l, `"body":null`)
	}
}

func TestDecodeSchemaDiscovery(t *testing.T) {
	a := require.New(t)
	tr := wiretracetest.Sample(t)
	t.Setenv(schemaEnv, "")

	// No schema next to the trace.
	_, errOut, code := runCmd(t, "decode", tr.Dir)
	a.Equal(1, code)
	a.Contains(errOut, "pass --schema")

	// api.tl and mtproto.tl next to the run directory.
	files := wiretracetest.SchemaFiles()
	root := filepath.Dir(tr.Dir)
	for i, name := range []string{"api.tl", "mtproto.tl"} {
		data, err := os.ReadFile(files[i])
		a.NoError(err)
		a.NoError(os.WriteFile(filepath.Join(root, name), data, 0o600))
	}
	out, errOut, code := runCmd(t, "decode", "--no-body", tr.Dir)
	a.Equal(0, code, errOut)
	a.Len(strings.Split(strings.TrimSpace(out), "\n"), 13)

	// Schema directory flag.
	_, errOut, code = runCmd(t, "decode", "--schema", root, "--no-body", tr.Dir)
	a.Equal(0, code, errOut)
}

func TestTimeline(t *testing.T) {
	a := require.New(t)
	tr := wiretracetest.Sample(t)
	out, errOut, code := runCmd(t, "timeline", schemaFlag(), tr.Dir)
	a.Equal(0, code, errOut)
	for _, want := range []string{
		"conn.start dc2",
		"req#1 new help.getNearestDc dc2 size=4 to_main needs_layer",
		"req#1 → msg 6917529027641081860 seqno=1",
		"→ dc2 6917529027641081872 seqno=6 len=224 container[invokeWithLayer(…help.getNearestDc), msgs_ack, auth.signIn]" +
			" redacted=auth.signIn",
		"req#2 caller intro intro/intro_code.cpp:200",
		"← dc2 6917529027641081857 seqno=7 len=144 container[rpc_result(help.getNearestDc → nearestDc), " +
			"rpc_result(auth.signIn → auth.authorization), new_session_created]",
		"req#1 done help.getNearestDc → nearestDc lat=100.5ms  → caused req.new photos.deletePhotos",
		"req#3 new photos.deletePhotos dc2 size=32 to_main  ⇐ req#1 help.getNearestDc → nearestDc",
		"→ dc2 6917529027641081880 seqno=9 len=16 invokeAfterMsg(help.getConfig)",
		"req#4 done help.getConfig → config lat=20.0ms  → caused config.loaded, st.write map",
		"st.write caller=void Storage::Account::writeMap() file=D877F783D5D3EF8C/map len=100  ⇐ req#4 help.getConfig",
		"req#5 FAIL messages.sendMessage → rpc_error lat=10.0ms error=420 FLOOD_WAIT_17",
		"DROPPED 3 events",
		"conn.restart dc2 reason=recv_timeout retry_ms=1",
	} {
		a.Contains(out, want)
	}

	out, _, code = runCmd(t, "timeline", "--no-storage", "--kinds=req.*,st.write", tr.Dir)
	a.Equal(0, code)
	a.NotContains(out, "st.write caller=")
	a.NotContains(out, "conn.")
	a.Contains(out, "req#5 FAIL")
}

func TestStats(t *testing.T) {
	a := require.New(t)
	tr := wiretracetest.Sample(t)
	out, errOut, code := runCmd(t, "stats", schemaFlag(), tr.Dir)
	a.Equal(0, code, errOut)
	for _, want := range []string{
		"events: 39",
		"dropped: 3",
		"420 FLOOD_WAIT_X  messages.sendMessage",
		"1 waits       17s  messages.sendMessage",
		"recv_timeout",
		"rpc_result:nearestDc",
		"st.write map",
	} {
		a.Contains(out, want)
	}

	out, errOut, code = runCmd(t, "stats", "--json", tr.Dir)
	a.Equal(0, code, errOut)
	var s Stats
	a.NoError(json.Unmarshal([]byte(out), &s))
	a.Equal(39, s.Events)
	a.Equal(uint64(3), s.Dropped)
	a.Equal(1, s.Restarts["recv_timeout"])
	a.Len(s.DCs, 1)
	a.Equal(1, s.DCs[0].Sessions)
	a.Equal(4, s.DCs[0].MsgsOut)
	a.Equal(5, s.DCs[0].MsgsIn)
	a.Equal([]*FloodStats{{Method: "messages.sendMessage", Count: 1, Seconds: 17}}, s.Floods)
	methods := map[string]*MethodStats{}
	for _, m := range s.Methods {
		methods[m.Method] = m
	}
	a.Equal(100.5, methods["help.getNearestDc"].P50)
	a.Equal(1, methods["messages.sendMessage"].Fail)
	a.Equal(1, s.Subsystems["intro"])
}

// otherClient writes a trace of a client that wraps differently: no
// msgs_ack in the first container, ping_delay_disconnect, and getState.
func otherClient(t *testing.T) string {
	tr := wiretracetest.New(t, wiretrace.WriterOptions{})
	init := &tg.InvokeWithLayerRequest{
		Layer: 232,
		Query: &tg.InitConnectionRequest{
			APIID:       2,
			DeviceModel: "gotd",
			Query:       &tg.HelpGetNearestDCRequest{},
		},
	}
	tr.OutContainer(100, 2,
		wiretracetest.Item{MsgID: 104, SeqNo: 1, Body: wiretracetest.Encode(t, init), Method: tg.HelpGetNearestDCRequestTypeID},
		wiretracetest.Item{MsgID: 108, SeqNo: 2, Body: wiretracetest.Encode(t, &mt.PingDelayDisconnectRequest{PingID: 1, DisconnectDelay: 75})},
	)
	tr.In(wiretracetest.Item{MsgID: 101, Body: wiretracetest.RPCResult(104, wiretracetest.Encode(t, &tg.NearestDC{Country: "XX"}))})
	tr.Out(wiretracetest.Item{MsgID: 112, SeqNo: 3, Body: wiretracetest.Encode(t, &tg.UpdatesGetStateRequest{})})
	tr.In(wiretracetest.Item{MsgID: 105, Body: wiretracetest.RPCResult(112, wiretracetest.Encode(t, &tg.UpdatesState{Pts: 1}))})
	tr.Out(wiretracetest.Item{MsgID: 116, SeqNo: 5, Body: wiretracetest.Encode(t, &tg.MessagesSendMessageRequest{
		Peer: &tg.InputPeerSelf{}, Message: "x", RandomID: 1,
	})})
	tr.Close()
	return tr.Dir
}

func TestDiff(t *testing.T) {
	a := require.New(t)
	tr := wiretracetest.Sample(t)

	// A trace against itself: everything aligns.
	out, errOut, code := runCmd(t, "diff", schemaFlag(), tr.Dir, tr.Dir)
	a.Equal(0, code, errOut)
	a.Contains(out, "== phase auth: A 1 msgs, B 1 msgs")
	a.Contains(out, "  = container[invokeWithLayer(initConnection(help.getNearestDc)), msgs_ack(n), auth.signIn]")
	a.NotContains(out, "\n  - ")
	a.NotContains(out, "\n  + ")
	a.NotContains(out, "only in")

	other := otherClient(t)
	out, errOut, code = runCmd(t, "diff", schemaFlag(), "--timing", tr.Dir, other)
	a.Equal(0, code, errOut)
	for _, want := range []string{
		"== phase init: A 0 msgs, B 1 msgs",
		"  + container[invokeWithLayer(initConnection(help.getNearestDc)), ping_delay_disconnect(75)]   B+0.000s",
		"  - container[invokeWithLayer(initConnection(help.getNearestDc)), msgs_ack(n), auth.signIn]   A+0.000s",
		"== phase steady: A 0 msgs, B 1 msgs",
		"methods only in A:\n  auth.signIn×1\n  help.getConfig×1\n  msgs_ack×1\n  photos.deletePhotos×1",
		"methods only in B:\n  ping_delay_disconnect×1\n  updates.getState×1",
		"ping_delay_disconnect interval (median): A 0.0s (0 sent), B 0.0s (1 sent)",
	} {
		a.Contains(out, want)
	}
}

func TestDiffFields(t *testing.T) {
	a := require.New(t)
	tr := wiretracetest.Sample(t)
	b := wiretracetest.New(t, wiretrace.WriterOptions{})
	signIn := &tg.AuthSignInRequest{PhoneNumber: "+1", PhoneCodeHash: "h", PhoneCode: "1"}
	signIn.SetFlags()
	init := &tg.InvokeWithLayerRequest{
		Layer: 232,
		Query: &tg.InitConnectionRequest{
			APIID: 1, DeviceModel: "Other", SystemVersion: "Linux", AppVersion: "1.0", SystemLangCode: "en",
			LangPack: "tdesktop", LangCode: "en", Query: &tg.HelpGetNearestDCRequest{},
		},
	}
	b.OutContainer(100, 6,
		wiretracetest.Item{MsgID: 104, SeqNo: 1, Body: wiretracetest.Encode(t, init), Method: tg.HelpGetNearestDCRequestTypeID},
		wiretracetest.Item{MsgID: 108, SeqNo: 2, Body: wiretracetest.Encode(t, &mt.MsgsAck{MsgIDs: []int64{5}})},
		wiretracetest.Item{
			MsgID: 112, SeqNo: 5, Body: wiretracetest.Redact(wiretracetest.Encode(t, signIn)),
			Method: tg.AuthSignInRequestTypeID,
		},
	)
	b.Close()
	out, errOut, code := runCmd(t, "diff", schemaFlag(), tr.Dir, b.Dir)
	a.Equal(0, code, errOut)
	a.Contains(out, "  = container[invokeWithLayer(initConnection(help.getNearestDc)), msgs_ack(n), auth.signIn]")
	a.Contains(out, "~initConnection.device_model PC → Other")
	a.Contains(out, "-initConnection.params._=jsonObject")
}

func TestUsage(t *testing.T) {
	a := require.New(t)
	_, errOut, code := runCmd(t)
	a.Equal(2, code)
	a.Contains(errOut, "Usage: wiretrace")

	out, _, code := runCmd(t, "help")
	a.Equal(0, code)
	a.Contains(out, "decode")

	_, errOut, code = runCmd(t, "nope")
	a.Equal(2, code)
	a.Contains(errOut, `unknown command "nope"`)

	_, errOut, code = runCmd(t, "decode")
	a.Equal(2, code)
	a.Contains(errOut, "Usage: wiretrace decode")

	_, errOut, code = runCmd(t, "stats", t.TempDir())
	a.Equal(1, code)
	a.Contains(errOut, "no events-NNNN.jsonl")
}

func TestStorageKind(t *testing.T) {
	require.Equal(t, "map", storageKind("D877F783D5D3EF8C/map"))
	require.Equal(t, "<hex>s", storageKind(`tdata\D877F783D5D3EF8Cs`))
	require.Equal(t, "key_datas", storageKind("key_datas"))
}

func TestNormalizeError(t *testing.T) {
	require.Equal(t, "FLOOD_WAIT_X", normalizeError("FLOOD_WAIT_17"))
	require.Equal(t, "PHONE_MIGRATE_X", normalizeError("PHONE_MIGRATE_4"))
}
