package tljson_test

import (
	"encoding/binary"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/tl"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/mt"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/wiretrace/tljson"
	"github.com/gotd/td/wiretrace/wiretracetest"
)

func schema(t *testing.T) *tljson.Schema {
	t.Helper()
	s, err := tljson.LoadFiles(wiretracetest.SchemaFiles()...)
	require.NoError(t, err)
	return s
}

func toJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return string(b)
}

func decode(t *testing.T, d *tljson.Decoder, o bin.Encoder, function bool) string {
	t.Helper()
	v, err := d.Decode(wiretracetest.Encode(t, o), function)
	require.NoError(t, err)
	return toJSON(t, v)
}

func TestSchemaParse(t *testing.T) {
	const src = `
int ? = Int;
vector {t:Type} # [ t ] = Vector t;
int128 4*[ int ] = Int128;
//rpc_result#f35c6d01 req_msg_id:long result:Object = RpcResult; // parsed manually
boolFalse#bc799737 = Bool;
peerUser#59511722 user_id:long = Peer; // trailing comment
---functions---
help.getNearestDc#1fb33026 = NearestDc;
---types---
nearestDc#8e1a1775 country:string this_dc:int nearest_dc:int = NearestDc;
// LAYER 233
`
	s := tljson.NewSchema()
	require.NoError(t, s.Parse(strings.NewReader(src)))
	require.Equal(t, 233, s.Layer)
	require.Equal(t, 2, s.Skipped)
	require.Equal(t, "help.getNearestDc", s.Name(0x1fb33026))
	require.Equal(t, "nearestDc", s.Name(0x8e1a1775))
	require.Equal(t, "peerUser", s.Name(0x59511722))
	require.Equal(t, "#00000001", s.Name(1))
	_, ok := s.Function(0x1fb33026)
	require.True(t, ok)
	_, ok = s.Constructor(0x1fb33026)
	require.False(t, ok)
}

func TestFlags(t *testing.T) {
	d := &tljson.Decoder{Schema: schema(t)}
	u := &tg.User{
		Bot:            true,
		ID:             42,
		AccessHash:     7,
		Username:       "bot",
		StoriesHidden:  true, // flags2.3
		BotActiveUsers: 5,    // flags2.12
	}
	u.SetFlags()
	got := decode(t, d, u, false)
	require.Contains(t, got, `"_":"user"`)
	require.Contains(t, got, `"bot":true`)
	require.Contains(t, got, `"stories_hidden":true`)
	require.Contains(t, got, `"id":"42"`)
	require.Contains(t, got, `"access_hash":"7"`)
	require.Contains(t, got, `"username":"bot"`)
	require.Contains(t, got, `"bot_active_users":5`)
	require.NotContains(t, got, `"first_name"`)
	require.NotContains(t, got, `"self"`)
}

func TestVectorsAndPrimitives(t *testing.T) {
	d := &tljson.Decoder{Schema: schema(t)}
	require.Equal(t, `{"_":"msgs_ack","msg_ids":["1","-2"]}`, decode(t, d, &mt.MsgsAck{MsgIDs: []int64{1, -2}}, false))

	// Bare vector of bare constructors.
	got := decode(t, d, &mt.FutureSalts{ReqMsgID: 1, Now: 2, Salts: []mt.FutureSalt{{ValidSince: 3, ValidUntil: 4, Salt: 5}}}, false)
	require.Equal(t, `{"_":"future_salts","req_msg_id":"1","now":2,`+
		`"salts":[{"_":"future_salt","valid_since":3,"valid_until":4,"salt":"5"}]}`, got)

	// int128.
	var nonce bin.Int128
	nonce[0] = 0xab
	got = decode(t, d, &mt.ReqPqMultiRequest{Nonce: nonce}, true)
	require.Equal(t, `{"_":"req_pq_multi","nonce":"ab000000000000000000000000000000"}`, got)

	// int256 and bytes.
	var newNonce bin.Int256
	newNonce[31] = 1
	got = decode(t, d, &mt.PQInnerData{Pq: []byte{1}, P: []byte{2}, Q: []byte{3}, NewNonce: newNonce}, false)
	require.Contains(t, got, `"pq":"AQ=="`)
	require.Contains(t, got, `"new_nonce":"`+strings.Repeat("0", 62)+`01"`)

	// double and Bool.
	got = decode(t, d, &tg.GeoPoint{Long: 1.5, Lat: -2.25, AccessHash: 3}, false)
	require.Contains(t, got, `"long":1.5,"lat":-2.25`)
	got = decode(t, d, &tg.InputPeerNotifySettings{Silent: true}, false)
	require.Contains(t, got, `"silent":true`)

	// Boxed vector of boxed objects.
	got = decode(t, d, &tg.PhotosDeletePhotosRequest{ID: []tg.InputPhotoClass{&tg.InputPhotoEmpty{}}}, true)
	require.Equal(t, `{"_":"photos.deletePhotos","id":[{"_":"inputPhotoEmpty"}]}`, got)
}

func TestMaxBytes(t *testing.T) {
	d := &tljson.Decoder{Schema: schema(t), MaxBytes: 4}
	got := decode(t, d, &tg.AuthExportedAuthorization{ID: 1, Bytes: []byte("0123456789")}, false)
	require.Contains(t, got, `"bytes":{"_":"bytes","len":10,"sha256":"`)

	// Invalid UTF-8 strings are kept as bytes.
	got = decode(t, d, &tg.NearestDC{Country: "\xff\xfe"}, false)
	require.Contains(t, got, `"country":{"_":"bytes","b64":"//4="}`)
}

func TestGenericWrappers(t *testing.T) {
	d := &tljson.Decoder{Schema: schema(t)}
	req := &tg.InvokeWithLayerRequest{
		Layer: 232,
		Query: &tg.InitConnectionRequest{
			APIID:       1,
			DeviceModel: "PC",
			LangPack:    "tdesktop",
			Query: &tg.InvokeAfterMsgRequest{
				MsgID: 9,
				Query: &tg.InvokeWithoutUpdatesRequest{Query: &tg.HelpGetConfigRequest{}},
			},
		},
	}
	got := decode(t, d, req, true)
	require.Equal(t, `{"_":"invokeWithLayer","layer":232,"query":{"_":"initConnection","flags":0,"api_id":1,`+
		`"device_model":"PC","system_version":"","app_version":"","system_lang_code":"","lang_pack":"tdesktop",`+
		`"lang_code":"","query":{"_":"invokeAfterMsg","msg_id":"9","query":{"_":"invokeWithoutUpdates",`+
		`"query":{"_":"help.getConfig"}}}}}`, got)
}

func vectorLong(values ...int64) []byte {
	var b bin.Buffer
	b.PutID(tljson.IDVector)
	b.PutInt(len(values))
	for _, v := range values {
		b.PutLong(v)
	}
	return b.Copy()
}

func TestRPCResultBareVector(t *testing.T) {
	s := schema(t)
	def, ok := s.FunctionByName("photos.deletePhotos")
	require.True(t, ok)
	body := wiretracetest.RPCResult(77, vectorLong(10, 20))

	// Resolved through the request method return type, Vector<long>.
	d := &tljson.Decoder{Schema: s, ResultType: func(req int64) (tl.Type, bool) {
		require.Equal(t, int64(77), req)
		return def.Type, true
	}}
	v, err := d.Decode(body, false)
	require.NoError(t, err)
	require.Equal(t, `{"_":"rpc_result","req_msg_id":"77","result":["10","20"]}`, toJSON(t, v))

	// Vector<int> must not be confused with Vector<long>.
	var b bin.Buffer
	b.PutID(tljson.IDVector)
	b.PutInt(2)
	b.PutInt(1)
	b.PutInt(2)
	d.ResultType = func(int64) (tl.Type, bool) {
		return tl.Type{Name: "Vector", GenericArg: &tl.Type{Name: "int", Bare: true}}, true
	}
	v, err = d.Decode(wiretracetest.RPCResult(77, b.Copy()), false)
	require.NoError(t, err)
	require.Equal(t, `{"_":"rpc_result","req_msg_id":"77","result":[1,2]}`, toJSON(t, v))

	// Unknown request: the element type is guessed.
	d.ResultType = nil
	v, err = d.Decode(body, false)
	require.NoError(t, err)
	require.Equal(t, `{"_":"rpc_result","req_msg_id":"77","result":["10","20"]}`, toJSON(t, v))

	v, err = d.Decode(wiretracetest.RPCResult(77, wiretracetest.Encode(t, &tg.UserClassVector{Elems: []tg.UserClass{
		&tg.UserEmpty{ID: 1},
	}})), false)
	require.NoError(t, err)
	require.Equal(t, `{"_":"rpc_result","req_msg_id":"77","result":[{"_":"userEmpty","id":"1"}]}`, toJSON(t, v))

	// Top-level bare vector decoded as a type.
	v, err = d.DecodeAs(vectorLong(5), def.Type)
	require.NoError(t, err)
	require.Equal(t, `["5"]`, toJSON(t, v))
}

func TestRPCError(t *testing.T) {
	s := schema(t)
	def, _ := s.FunctionByName("photos.deletePhotos")
	d := &tljson.Decoder{Schema: s, ResultType: func(int64) (tl.Type, bool) { return def.Type, true }}
	body := wiretracetest.RPCResult(1, wiretracetest.Encode(t, &mt.RPCError{ErrorCode: 420, ErrorMessage: "FLOOD_WAIT_3"}))
	v, err := d.Decode(body, false)
	require.NoError(t, err)
	require.Equal(t, `{"_":"rpc_result","req_msg_id":"1","result":{"_":"rpc_error","error_code":420,"error_message":"FLOOD_WAIT_3"}}`,
		toJSON(t, v))
}

func TestGzip(t *testing.T) {
	s := schema(t)
	def, _ := s.FunctionByName("photos.deletePhotos")
	d := &tljson.Decoder{Schema: s, ResultType: func(int64) (tl.Type, bool) { return def.Type, true }}
	packed := wiretracetest.GZIP(t, vectorLong(1, 2, 3))
	v, err := d.Decode(wiretracetest.RPCResult(5, packed), false)
	require.NoError(t, err)
	got := toJSON(t, v)
	require.Contains(t, got, `"result":{"_":"gzip_packed","gzip":{"len":`)
	require.Contains(t, got, `"unpacked":32},"inner":["1","2","3"]}`)

	// Broken gzip data.
	var b bin.Buffer
	b.PutID(tljson.IDGzipPacked)
	b.PutBytes([]byte("not gzip"))
	v, err = d.Decode(b.Copy(), false)
	require.Error(t, err)
	require.Contains(t, toJSON(t, v), `"partial":true`)
}

func TestContainer(t *testing.T) {
	d := &tljson.Decoder{Schema: schema(t)}
	body := wiretracetest.Container(
		wiretracetest.Item{MsgID: 4, SeqNo: 1, Body: wiretracetest.Encode(t, &mt.PingRequest{PingID: 1})},
		wiretracetest.Item{MsgID: 8, SeqNo: 3, Body: []byte{0xef, 0xbe, 0xad, 0xde, 1, 2, 3, 4}},
		wiretracetest.Item{MsgID: 12, SeqNo: 5, Body: wiretracetest.Encode(t, &mt.MsgsAck{MsgIDs: []int64{1}})},
	)
	v, err := d.Decode(body, true)
	require.NoError(t, err, "containers recover from broken items")
	got := toJSON(t, v)
	require.Contains(t, got, `{"_":"message","msg_id":"4","seqno":1,"bytes":12,"body":{"_":"ping","ping_id":"1"}}`)
	require.Contains(t, got, `"body":{"_":"#deadbeef","hex":"01020304"},"partial":true}`)
	require.Contains(t, got, `"body":{"_":"msgs_ack","msg_ids":["1"]}}`)
}

func TestRedacted(t *testing.T) {
	d := &tljson.Decoder{Schema: schema(t)}
	signIn := &tg.AuthSignInRequest{PhoneNumber: "+1", PhoneCodeHash: "h", PhoneCode: "12345"}
	signIn.SetFlags()
	raw := wiretracetest.Encode(t, &tg.InvokeWithLayerRequest{Layer: 1, Query: signIn})
	// Redact after the auth.signIn id (offset 8), as the tracer does.
	red := append(append([]byte(nil), raw[:8]...), wiretracetest.Redact(raw[8:])...)
	require.NotContains(t, string(red), "12345")
	v, err := d.Decode(red, true)
	require.NoError(t, err)
	require.Equal(t, `{"_":"invokeWithLayer","layer":1,"query":{"_":"redacted","orig":"auth.signIn","len":`+
		itoa(len(raw)-8)+`}}`, toJSON(t, v))

	// Redacted rpc_result body.
	res := wiretracetest.Redact12(wiretracetest.RPCResult(1, wiretracetest.Encode(t, &tg.AuthExportedAuthorization{
		ID: 1, Bytes: []byte("secret"),
	})))
	v, err = d.Decode(res, false)
	require.NoError(t, err)
	require.Equal(t, `{"_":"rpc_result","req_msg_id":"1","result":{"_":"redacted","orig":"auth.exportedAuthorization","len":20}}`,
		toJSON(t, v))

	// Marker-only redaction of an 8-byte body.
	var b bin.Buffer
	b.PutID(tg.AuthExportAuthorizationRequestTypeID)
	b.PutUint32(tljson.RedactedMarker)
	v, err = d.Decode(b.Copy(), true)
	require.NoError(t, err)
	require.Equal(t, `{"_":"redacted","orig":"auth.exportAuthorization","len":8}`, toJSON(t, v))
}

func itoa(v int) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func TestUnknownConstructor(t *testing.T) {
	d := &tljson.Decoder{Schema: schema(t)}
	var b bin.Buffer
	b.PutID(tg.UpdateShortTypeID)
	b.PutID(0xdeadbeef)
	b.PutInt(7)
	v, err := d.Decode(b.Copy(), false)
	var e *tljson.Error
	require.ErrorAs(t, err, &e)
	require.Equal(t, uint32(0xdeadbeef), e.Unknown)
	require.Equal(t, `{"_":"updateShort","update":{"_":"#deadbeef","hex":"07000000"},"partial":true}`, toJSON(t, v))

	// Unknown top-level.
	v, err = d.Decode([]byte{1, 0, 0, 0, 2}, false)
	require.Error(t, err)
	require.Equal(t, `{"_":"#00000001","hex":"02"}`, toJSON(t, v))
}

func TestTruncated(t *testing.T) {
	d := &tljson.Decoder{Schema: schema(t)}
	raw := wiretracetest.Encode(t, &tg.NearestDC{Country: "XX", ThisDC: 2, NearestDC: 3})
	v, err := d.Decode(raw[:len(raw)-2], false)
	require.Error(t, err)
	got := toJSON(t, v)
	require.Contains(t, got, `"this_dc":2`)
	require.Contains(t, got, `"error":"at `)
	require.Contains(t, got, `"partial":true`)

	// Trailing bytes.
	_, err = d.Decode(append(raw, 0, 0, 0, 0), false)
	require.ErrorContains(t, err, "trailing")

	// Huge vector count.
	var b bin.Buffer
	b.PutID(mt.MsgsAckTypeID)
	b.PutID(tljson.IDVector)
	b.PutInt(1 << 30)
	_, err = d.Decode(b.Copy(), false)
	require.ErrorContains(t, err, "overflows")
}

func TestDepthLimit(t *testing.T) {
	d := &tljson.Decoder{Schema: schema(t), MaxDepth: 3}
	var q bin.Object = &tg.HelpGetConfigRequest{}
	for i := 0; i < 5; i++ {
		q = &tg.InvokeWithoutUpdatesRequest{Query: q}
	}
	_, err := d.Decode(wiretracetest.Encode(t, q), true)
	require.ErrorContains(t, err, "max depth")
}

func TestNaN(t *testing.T) {
	d := &tljson.Decoder{Schema: schema(t)}
	var b bin.Buffer
	b.PutID(tg.JSONNumberTypeID)
	var nan [8]byte
	binary.LittleEndian.PutUint64(nan[:], 0x7ff8000000000001)
	b.Put(nan[:])
	v, err := d.Decode(b.Copy(), false)
	require.NoError(t, err)
	require.Equal(t, `{"_":"jsonNumber","value":"NaN"}`, toJSON(t, v))
}
