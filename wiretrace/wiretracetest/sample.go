package wiretracetest

import (
	"encoding/binary"
	"testing"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/mt"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/wiretrace"
)

// Message ids of Sample.
const (
	MsgInit      uint64 = 0x6000000000000004
	MsgAck       uint64 = 0x6000000000000008
	MsgSignIn    uint64 = 0x600000000000000c
	MsgOutBox    uint64 = 0x6000000000000010
	MsgDelete    uint64 = 0x6000000000000014
	MsgAfter     uint64 = 0x6000000000000018
	MsgSend      uint64 = 0x600000000000001c
	MsgInBox     uint64 = 0x6000000000000001
	MsgInInit    uint64 = 0x6000000000000005
	MsgInSignIn  uint64 = 0x6000000000000009
	MsgInSession uint64 = 0x600000000000000d
	MsgInDelete  uint64 = 0x6000000000000011
	MsgInAfter   uint64 = 0x6000000000000015
	MsgInSend    uint64 = 0x6000000000000019
	MsgInUnknown uint64 = 0x600000000000001d
)

// UnknownID is the unknown constructor id used by Sample.
const UnknownID uint32 = 0xdeadbeef

// Sample writes a synthetic session covering every decoder feature:
//
//   - lifecycle events (conn.start, conn.up, key.set, session.new_id,
//     state, salt.change, conn.reason, conn.restart, config.loaded)
//   - an outgoing container of invokeWithLayer(initConnection(
//     help.getNearestDc)), msgs_ack and a redacted auth.signIn
//   - an incoming container of rpc_results (one redacted) and
//     new_session_created
//   - photos.deletePhotos answered by a top-level bare Vector<long>
//   - invokeAfterMsg(help.getConfig) answered by a gzip_packed config
//   - messages.sendMessage failing with FLOOD_WAIT_17
//   - an update with an unknown constructor
//   - request lifecycle (req.new, req.msg, req.caller after the message,
//     req.done, req.fail), causes, storage events and a dropped record.
func Sample(t testing.TB) *Trace {
	t.Helper()
	tr := New(t, wiretrace.WriterOptions{InlineMax: 64})
	F := wiretrace.F
	const sdc = 2

	tr.Event("t1", wiretrace.KindConnStart, sdc, F("pdc", 2), F("tests", 2), F("proxy", 0), F("dc_type", 0))
	tr.Event("t1", wiretrace.KindConnUp, sdc, F("tag", "tcp-v1#1"))
	tr.Event("t1", wiretrace.KindKeySet, sdc, F("key", wiretrace.ID(1)), F("key_hex", "0100000000000000"),
		F("key_kind", "temp"), F("key_dc", 2), F("expires_at", 1700000000))
	tr.Event("t1", wiretrace.KindSessionNewID, sdc, F("old", wiretrace.SignedID(0)), F("new", wiretrace.SignedID(-42)))
	tr.Event("t1", wiretrace.KindState, sdc, F("from", -1), F("to", 1))

	// Requests 1 (getNearestDc) and 2 (signIn).
	tr.Event("main", wiretrace.KindReqNew, sdc, F("req", 1), F("type", wiretrace.TypeID(tg.HelpGetNearestDCRequestTypeID)),
		F("to_main", true), F("size", 4), F("needs_layer", true))
	tr.Event("main", wiretrace.KindReqNew, sdc, F("req", 2), F("type", wiretrace.TypeID(tg.AuthSignInRequestTypeID)),
		F("to_main", true), F("size", 40), F("needs_layer", false))
	tr.Event("t1", wiretrace.KindReqMsg, sdc, F("req", 1), F("was", wiretrace.ID(0)), F("msg_id", wiretrace.ID(MsgInit)), F("seqno", 1))
	tr.Event("t1", wiretrace.KindReqMsg, sdc, F("req", 2), F("was", wiretrace.ID(0)), F("msg_id", wiretrace.ID(MsgSignIn)), F("seqno", 5))

	init := &tg.InvokeWithLayerRequest{
		Layer: 232,
		Query: &tg.InitConnectionRequest{
			APIID:          1,
			DeviceModel:    "PC",
			SystemVersion:  "Linux",
			AppVersion:     "1.0",
			SystemLangCode: "en",
			LangPack:       "tdesktop",
			LangCode:       "en",
			Params: &tg.JSONObject{Value: []tg.JSONObjectValue{
				{Key: "tz_offset", Value: &tg.JSONNumber{Value: 0}},
			}},
			Query: &tg.HelpGetNearestDCRequest{},
		},
	}
	signIn := &tg.AuthSignInRequest{PhoneNumber: "+10000000000", PhoneCodeHash: "hash", PhoneCode: "12345"}
	signIn.SetFlags()
	tr.OutContainer(MsgOutBox, 6,
		Item{
			MsgID: MsgInit, SeqNo: 1, Body: Encode(t, init),
			Method: tg.HelpGetNearestDCRequestTypeID,
		},
		Item{MsgID: MsgAck, SeqNo: 2, Body: Encode(t, &mt.MsgsAck{MsgIDs: []int64{1, 2}})},
		Item{
			MsgID: MsgSignIn, SeqNo: 5, Body: Redact(Encode(t, signIn)),
			Method: tg.AuthSignInRequestTypeID, Redacted: []string{"auth.signIn"},
		},
	)
	tr.Event("main", wiretrace.KindReqCaller, -1, F("req", 1), F("file", "mtproto/mtp_instance.cpp"), F("line", 100),
		F("function", "void MTP::Instance::Private::start()"), F("subsystem", "mtproto"))
	tr.Event("main", wiretrace.KindReqCaller, -1, F("req", 2), F("file", "intro/intro_code.cpp"), F("line", 200),
		F("function", "void Intro::details::CodeWidget::submitCode()"), F("subsystem", "intro"))

	tr.Now += 0.1
	nearest := Encode(t, &tg.NearestDC{Country: "XX", ThisDC: 2, NearestDC: 2})
	auth := Encode(t, &tg.AuthAuthorization{User: &tg.UserEmpty{ID: 1}})
	tr.InContainer(MsgInBox, 7,
		Item{MsgID: MsgInInit, SeqNo: 2, Body: RPCResult(MsgInit, nearest)},
		Item{
			MsgID: MsgInSignIn, SeqNo: 4, Body: Redact12(RPCResult(MsgSignIn, auth)),
			Redacted: []string{"auth.signIn.result"},
		},
		Item{MsgID: MsgInSession, SeqNo: 6, Body: Encode(t, &mt.NewSessionCreated{FirstMsgID: int64(MsgInit), UniqueID: 1, ServerSalt: 8})},
	)
	tr.Event("t1", wiretrace.KindSaltChange, sdc, F("old", wiretrace.ID(7)), F("new", wiretrace.ID(8)), F("accepted", true))
	reqDone(tr, 1, tg.HelpGetNearestDCRequestTypeID, tg.NearestDCTypeID, MsgInBox, 100.5)
	reqDone(tr, 2, tg.AuthSignInRequestTypeID, tg.AuthAuthorizationTypeID, MsgInBox, 101)

	// Request 3: photos.deletePhotos → Vector<long>.
	tr.Event("main", wiretrace.KindReqNew, sdc, F("req", 3), F("type", wiretrace.TypeID(tg.PhotosDeletePhotosRequestTypeID)),
		F("to_main", true), F("size", 32), F("needs_layer", false), F("cause", wiretrace.Cause{Req: 1, Outer: wiretrace.ID(MsgInBox),
			Type: wiretrace.TypeID(tg.NearestDCTypeID)}))
	tr.Event("t1", wiretrace.KindReqMsg, sdc, F("req", 3), F("was", wiretrace.ID(0)), F("msg_id", wiretrace.ID(MsgDelete)), F("seqno", 7))
	tr.Out(Item{MsgID: MsgDelete, SeqNo: 7, Body: Encode(t, &tg.PhotosDeletePhotosRequest{
		ID: []tg.InputPhotoClass{&tg.InputPhoto{ID: 10, AccessHash: 11, FileReference: []byte{1, 2, 3}}},
	})})
	tr.Now += 0.05
	tr.In(Item{MsgID: MsgInDelete, SeqNo: 8, Body: RPCResult(MsgDelete, Encode(t, &tg.LongVector{Elems: []int64{10, 20}}))})
	reqDone(tr, 3, tg.PhotosDeletePhotosRequestTypeID, 0x1cb5c415, MsgInDelete, 50)

	// Request 4: invokeAfterMsg(help.getConfig) → gzip_packed config.
	tr.Event("main", wiretrace.KindReqNew, sdc, F("req", 4), F("type", wiretrace.TypeID(tg.HelpGetConfigRequestTypeID)),
		F("to_main", true), F("size", 4), F("after", 3), F("needs_layer", false))
	tr.Event("t1", wiretrace.KindReqMsg, sdc, F("req", 4), F("was", wiretrace.ID(0)), F("msg_id", wiretrace.ID(MsgAfter)), F("seqno", 9))
	tr.Out(Item{
		MsgID: MsgAfter, SeqNo: 9,
		Body:   Encode(t, &tg.InvokeAfterMsgRequest{MsgID: int64(MsgDelete), Query: &tg.HelpGetConfigRequest{}}),
		Method: tg.HelpGetConfigRequestTypeID,
		Wrap:   []uint32{tg.InvokeAfterMsgRequestTypeID},
	})
	tr.Now += 0.02
	cfg := &tg.Config{ThisDC: 2, DCOptions: []tg.DCOption{{ID: 2, IPAddress: "127.0.0.1", Port: 443}}}
	tr.In(Item{MsgID: MsgInAfter, SeqNo: 10, Body: RPCResult(MsgAfter, GZIP(t, Encode(t, cfg)))})
	cause := F("cause", wiretrace.Cause{Req: 4, Outer: wiretrace.ID(MsgInAfter), Type: wiretrace.TypeID(tg.ConfigTypeID)})
	reqDone(tr, 4, tg.HelpGetConfigRequestTypeID, tg.ConfigTypeID, MsgInAfter, 20)
	tr.Event("main", wiretrace.KindConfigLoaded, 0, F("dcs", 1), F("this_dc", 2), F("date", 0), cause)
	tr.Event("main", wiretrace.KindStorageWrite, -1, F("file", "D877F783D5D3EF8C/map"), F("len", 100),
		F("caller", "void Storage::Account::writeMap()"), cause)
	tr.Event("t2", wiretrace.KindStorageCommit, -1, F("file", "D877F783D5D3EF8C/maps"), F("len", 140))

	// Request 5: messages.sendMessage → FLOOD_WAIT_17.
	tr.Event("main", wiretrace.KindReqNew, sdc, F("req", 5), F("type", wiretrace.TypeID(tg.MessagesSendMessageRequestTypeID)),
		F("to_main", true), F("size", 64), F("needs_layer", false))
	tr.Event("t1", wiretrace.KindReqMsg, sdc, F("req", 5), F("was", wiretrace.ID(0)), F("msg_id", wiretrace.ID(MsgSend)), F("seqno", 11))
	tr.Out(Item{MsgID: MsgSend, SeqNo: 11, Body: Encode(t, &tg.MessagesSendMessageRequest{
		Peer: &tg.InputPeerSelf{}, Message: "hello", RandomID: 99,
	})})
	tr.Now += 0.01
	tr.In(Item{MsgID: MsgInSend, SeqNo: 12, Body: RPCResult(MsgSend, Encode(t, &mt.RPCError{ErrorCode: 420, ErrorMessage: "FLOOD_WAIT_17"}))})
	tr.Event("main", wiretrace.KindReqFail, -1, F("req", 5), F("method", wiretrace.TypeID(tg.MessagesSendMessageRequestTypeID)),
		F("lat_ms", 10.0), F("type", wiretrace.TypeID(0x2144ca19)), F("outer", wiretrace.ID(MsgInSend)), F("len", 24),
		F("code", 420), F("error", "FLOOD_WAIT_17"))

	// An update with an unknown constructor.
	tr.In(Item{MsgID: MsgInUnknown, SeqNo: 14, Body: unknownUpdate()})

	if _, err := tr.W.WriteDropped(3); err != nil {
		t.Fatal(err)
	}
	tr.Event("t1", wiretrace.KindConnReason, sdc, F("reason", "recv_timeout"))
	tr.Event("t1", wiretrace.KindConnRestart, sdc, F("reason", "recv_timeout"), F("retry_ms", 1))
	tr.Event("t1", wiretrace.KindConnDown, sdc, F("tag", "tcp-v1#1"))
	tr.Close()
	return tr
}

// Redact12 redacts the rpc_result result (after the 12-byte rpc_result
// header), as the tracer does for responses of sensitive requests.
func Redact12(body []byte) []byte {
	out := append([]byte(nil), body[:12]...)
	return append(out, Redact(body[12:])...)
}

func reqDone(tr *Trace, req int, method, typ uint32, outer uint64, lat float64) {
	tr.t.Helper()
	F := wiretrace.F
	tr.Event("main", wiretrace.KindReqDone, -1, F("req", req), F("method", wiretrace.TypeID(method)), F("lat_ms", lat),
		F("type", wiretrace.TypeID(typ)), F("outer", wiretrace.ID(outer)), F("len", 16))
}

// unknownUpdate is updateShort#78d4dec1 update:Update date:int with an
// unknown update constructor.
func unknownUpdate() []byte {
	var b bin.Buffer
	b.PutID(tg.UpdateShortTypeID)
	b.PutID(UnknownID)
	b.PutInt(1)
	b.PutInt(2)
	var date [4]byte
	binary.LittleEndian.PutUint32(date[:], 1700000000)
	b.Put(date[:])
	return b.Copy()
}
