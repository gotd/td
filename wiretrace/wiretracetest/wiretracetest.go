// Package wiretracetest builds synthetic wire traces for tests: TL bodies
// are encoded with gotd and written with wiretrace.Writer, with the same
// summary fields the tdesktop tracer writes.
package wiretracetest

import (
	"encoding/binary"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/proto"
	"github.com/gotd/td/wiretrace"
)

// Well-known ids.
const (
	idRPCResult    = 0xf35c6d01
	idMsgContainer = 0x73f1f8dc
)

// SchemaFiles returns the gotd schema files (_schema/telegram.tl and
// _schema/mt.tl) matching the tg and mt packages, for decoding traces built
// by this package.
func SchemaFiles() []string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return nil
	}
	root := filepath.Join(filepath.Dir(file), "..", "..", "_schema")
	return []string{filepath.Join(root, "telegram.tl"), filepath.Join(root, "mt.tl")}
}

// Encode encodes a TL object.
func Encode(t testing.TB, o bin.Encoder) []byte {
	t.Helper()
	var b bin.Buffer
	if err := o.Encode(&b); err != nil {
		t.Fatal(err)
	}
	return b.Copy()
}

// RPCResult builds rpc_result#f35c6d01 req_msg_id:long result:Object.
func RPCResult(reqMsgID uint64, result []byte) []byte {
	var b bin.Buffer
	b.PutID(idRPCResult)
	b.PutLong(int64(reqMsgID)) // #nosec G115
	b.Put(result)
	return b.Copy()
}

// GZIP builds gzip_packed#3072cfa1 packed_data:bytes.
func GZIP(t testing.TB, data []byte) []byte {
	t.Helper()
	return Encode(t, &proto.GZIP{Data: data})
}

// Item is one message: a top-level message or a container item.
type Item struct {
	MsgID uint64
	SeqNo uint32
	Body  []byte
	// Method is the innermost method below wrappers (outgoing summary).
	Method uint32
	// Wrap are the wrapper ids (outgoing summary of top-level messages).
	Wrap []uint32
	// Redacted is the "redacted" summary list.
	Redacted []string
}

// Container builds msg_container#73f1f8dc of items.
func Container(items ...Item) []byte {
	var b bin.Buffer
	b.PutID(idMsgContainer)
	b.PutInt(len(items))
	for _, it := range items {
		b.PutLong(int64(it.MsgID)) // #nosec G115
		b.PutUint32(it.SeqNo)
		b.PutInt(len(it.Body))
		b.Put(it.Body)
	}
	return b.Copy()
}

// Redact redacts body like the tdesktop tracer: the bytes after the
// constructor id become [marker][id][zeros...], keeping the length.
func Redact(body []byte) []byte {
	out := make([]byte, len(body))
	copy(out, body[:4])
	if len(body) >= 8 {
		binary.LittleEndian.PutUint32(out[4:], wiretrace.RedactedMarker)
	}
	if len(body) >= 12 {
		copy(out[8:12], body[:4])
	}
	return out
}

// Trace writes a synthetic trace with deterministic time.
type Trace struct {
	t   testing.TB
	Dir string
	W   *wiretrace.Writer

	// Now is the time of the next event, in seconds.
	Now float64
	// Step advances Now after each event.
	Step float64
	// SDC and Session are used for "msg" events.
	SDC     int
	Session int64
}

// New creates a trace in a new temporary directory.
func New(t testing.TB, opt wiretrace.WriterOptions) *Trace {
	t.Helper()
	tr := &Trace{t: t, Dir: filepath.Join(t.TempDir(), "run"), Now: 1, Step: 0.001, SDC: 2, Session: -42}
	opt.Now = func() float64 {
		v := tr.Now
		tr.Now += tr.Step
		return v
	}
	if opt.Header == nil {
		opt.Header = []wiretrace.Field{
			wiretrace.F("run", "test"),
			wiretrace.F("app", "test"),
			wiretrace.F("layer", 232),
		}
	}
	w, err := wiretrace.NewWriter(tr.Dir, opt)
	if err != nil {
		t.Fatal(err)
	}
	tr.W = w
	t.Cleanup(func() { _ = w.Close() })
	return tr
}

// Event writes a generic event with dc fields for sdc (if not negative).
func (tr *Trace) Event(thread, kind string, sdc int, fields ...wiretrace.Field) uint64 {
	tr.t.Helper()
	var all []wiretrace.Field
	if sdc >= 0 {
		all = append(all,
			wiretrace.F("sdc", sdc),
			wiretrace.F("dc", sdc%10000),
			wiretrace.F("shift", sdc/10000),
		)
	}
	all = append(all, fields...)
	seq, err := tr.W.Write(thread, kind, all, nil)
	if err != nil {
		tr.t.Fatal(err)
	}
	return seq
}

func typeOf(body []byte) wiretrace.TypeID {
	if len(body) < 4 {
		return 0
	}
	return wiretrace.TypeID(binary.LittleEndian.Uint32(body))
}

func wrapIDs(ids []uint32) []wiretrace.TypeID {
	out := make([]wiretrace.TypeID, 0, len(ids))
	for _, id := range ids {
		out = append(out, wiretrace.TypeID(id))
	}
	return out
}

func resultSummary(body []byte) (wiretrace.ID, wiretrace.TypeID) {
	if len(body) < 16 || binary.LittleEndian.Uint32(body) != idRPCResult {
		return 0, 0
	}
	return wiretrace.ID(binary.LittleEndian.Uint64(body[4:])), wiretrace.TypeID(binary.LittleEndian.Uint32(body[12:]))
}

func (tr *Trace) msg(dir string, msgID uint64, seqNo uint32, body []byte, items []Item, top Item) uint64 {
	tr.t.Helper()
	m := &wiretrace.Msg{
		Dir:     dir,
		SDC:     tr.SDC,
		DC:      tr.SDC % 10000,
		Shift:   tr.SDC / 10000,
		PDC:     tr.SDC % 10000,
		Conn:    "tcp-v1#1",
		Key:     1,
		KeyHex:  "0100000000000000",
		KeyKind: "temp",
		Session: wiretrace.SignedID(tr.Session),
		Salt:    7,
		MsgID:   wiretrace.ID(msgID),
		SeqNo:   int64(seqNo),
		Len:     int64(len(body)),
		Type:    typeOf(body),
	}
	if dir == wiretrace.DirOut && top.Method != 0 && wiretrace.TypeID(top.Method) != m.Type {
		m.Method = wiretrace.TypeID(top.Method)
	}
	m.Wrap = wrapIDs(top.Wrap)
	m.Redacted = append(m.Redacted, top.Redacted...)
	if dir == wiretrace.DirIn {
		m.ReqMsgID, m.Result = resultSummary(body)
	}
	for _, it := range items {
		in := wiretrace.Inner{
			MsgID: wiretrace.ID(it.MsgID),
			SeqNo: int64(it.SeqNo),
			Len:   int64(len(it.Body)),
			Type:  typeOf(it.Body),
		}
		if dir == wiretrace.DirOut && it.Method != 0 && wiretrace.TypeID(it.Method) != in.Type {
			in.Method = wiretrace.TypeID(it.Method)
		}
		if dir == wiretrace.DirIn {
			in.ReqMsgID, in.Result = resultSummary(it.Body)
		}
		m.Redacted = append(m.Redacted, it.Redacted...)
		m.Inner = append(m.Inner, in)
	}
	seq, err := tr.W.WriteMsg("t1", m, body)
	if err != nil {
		tr.t.Fatal(err)
	}
	return seq
}

// Out writes an outgoing top-level message.
func (tr *Trace) Out(it Item) uint64 {
	tr.t.Helper()
	return tr.msg(wiretrace.DirOut, it.MsgID, it.SeqNo, it.Body, nil, it)
}

// In writes an incoming top-level message.
func (tr *Trace) In(it Item) uint64 {
	tr.t.Helper()
	return tr.msg(wiretrace.DirIn, it.MsgID, it.SeqNo, it.Body, nil, it)
}

// OutContainer writes an outgoing msg_container of items.
func (tr *Trace) OutContainer(msgID uint64, seqNo uint32, items ...Item) uint64 {
	tr.t.Helper()
	return tr.msg(wiretrace.DirOut, msgID, seqNo, Container(items...), items, Item{})
}

// InContainer writes an incoming msg_container of items.
func (tr *Trace) InContainer(msgID uint64, seqNo uint32, items ...Item) uint64 {
	tr.t.Helper()
	return tr.msg(wiretrace.DirIn, msgID, seqNo, Container(items...), items, Item{})
}

// Close flushes and closes the writer.
func (tr *Trace) Close() {
	tr.t.Helper()
	if err := tr.W.Close(); err != nil {
		tr.t.Fatal(err)
	}
}
