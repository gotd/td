// Package decode turns "msg" events of a wire trace into logical messages:
// containers are split, wrappers (invokeWithLayer, initConnection,
// invokeAfterMsg, ...) are unwrapped, rpc_result is joined with its request
// (method, latency, request id and caller).
package decode

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/go-faster/errors"

	"github.com/gotd/tl"

	"github.com/gotd/td/wiretrace"
	"github.com/gotd/td/wiretrace/tljson"
)

// Caller is the call site of a request (req.caller).
type Caller struct {
	File      string `json:"file,omitempty"`
	Line      int64  `json:"line,omitempty"`
	Function  string `json:"function,omitempty"`
	Subsystem string `json:"subsystem,omitempty"`
}

// RPCError is a decoded rpc_error.
type RPCError struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

// Message is one logical MTProto message.
type Message struct {
	Seq       uint64   `json:"seq"`
	T         float64  `json:"t"`
	Dir       string   `json:"dir"`
	SDC       int      `json:"sdc"`
	Session   string   `json:"session"`
	MsgID     string   `json:"msg_id"`
	SeqNo     int64    `json:"seqno"`
	Container string   `json:"container,omitempty"`
	Gzip      bool     `json:"gzip,omitempty"`
	Wrap      []string `json:"wrap,omitempty"`
	// After is the target msg_id of invokeAfterMsg(s) (the last one).
	After       string `json:"after,omitempty"`
	AfterMethod string `json:"after_method,omitempty"`
	// Type is the constructor (or function) of the message body.
	Type string `json:"type"`
	// Method is the innermost method of an outgoing query, or the method of
	// the request an rpc_result answers.
	Method string  `json:"method,omitempty"`
	Req    int64   `json:"req,omitempty"`
	Caller *Caller `json:"caller,omitempty"`
	// ReqMsgID, Result, LatencyMS and Error are set for rpc_result.
	ReqMsgID  string    `json:"req_msg_id,omitempty"`
	Result    string    `json:"result,omitempty"`
	LatencyMS *float64  `json:"latency_ms,omitempty"`
	Error     *RPCError `json:"error,omitempty"`

	Redacted    []string `json:"redacted,omitempty"`
	Partial     bool     `json:"partial,omitempty"`
	DecodeError string   `json:"decode_error,omitempty"`
	Body        any      `json:"body"`

	// MsgIDValue is the numeric MsgID.
	MsgIDValue uint64 `json:"-"`
	// TypeID is the constructor id of the body (after gzip).
	TypeID uint32 `json:"-"`
	// MethodID is the id of Method, if known.
	MethodID uint32 `json:"-"`
}

type sent struct {
	t      float64
	method uint32
}

// Options configures Decoder.
type Options struct {
	// MaxBytes is passed to tljson.Decoder.
	MaxBytes int
}

// Decoder decodes "msg" events into logical messages.
//
// Feed every event to Index first (pass 1), then call Messages for "msg"
// events (pass 2). Index collects facts that may be written after the
// message they describe (req.caller is emitted from another thread).
type Decoder struct {
	schema *tljson.Schema
	tl     *tljson.Decoder

	sent      map[uint64]sent
	msgReq    map[uint64]int64
	reqCaller map[int64]*Caller
}

// New creates a Decoder.
func New(schema *tljson.Schema, opt Options) *Decoder {
	d := &Decoder{
		schema:    schema,
		sent:      map[uint64]sent{},
		msgReq:    map[uint64]int64{},
		reqCaller: map[int64]*Caller{},
	}
	d.tl = &tljson.Decoder{
		Schema:     schema,
		MaxBytes:   opt.MaxBytes,
		ResultType: d.resultType,
	}
	return d
}

// Schema returns the schema of the decoder.
func (d *Decoder) Schema() *tljson.Schema { return d.schema }

func (d *Decoder) resultType(reqMsgID int64) (tl.Type, bool) {
	s, ok := d.sent[uint64(reqMsgID)] // #nosec G115
	if !ok || s.method == 0 {
		return tl.Type{}, false
	}
	def, ok := d.schema.Function(s.method)
	if !ok || def.Type.GenericRef {
		return tl.Type{}, false
	}
	return def.Type, true
}

// Index records request and message facts from an event (pass 1).
func (d *Decoder) Index(ev *wiretrace.Event) {
	switch ev.Kind {
	case wiretrace.KindReqMsg:
		if id := ev.ID("msg_id"); id != 0 {
			d.msgReq[id] = ev.Int("req")
		}
	case wiretrace.KindReqCaller:
		d.reqCaller[ev.Int("req")] = &Caller{
			File:      ev.Str("file"),
			Line:      ev.Int("line"),
			Function:  ev.Str("function"),
			Subsystem: ev.Str("subsystem"),
		}
	case wiretrace.KindMsg:
		if ev.Str("dir") != wiretrace.DirOut {
			return
		}
		m, err := ev.Msg()
		if err != nil {
			return
		}
		if len(m.Inner) > 0 && uint32(m.Type) == tljson.IDMsgContainer {
			for _, in := range m.Inner {
				d.remember(uint64(in.MsgID), ev.T, innerMethod(in.Method, in.Type))
			}
			return
		}
		d.remember(uint64(m.MsgID), ev.T, innerMethod(m.Method, m.Type))
	}
}

func innerMethod(method, typ wiretrace.TypeID) uint32 {
	if method != 0 {
		return uint32(method)
	}
	return uint32(typ)
}

func (d *Decoder) remember(msgID uint64, t float64, method uint32) {
	if msgID == 0 {
		return
	}
	prev, ok := d.sent[msgID]
	if ok && prev.method != 0 && method == 0 {
		method = prev.method
	}
	if ok && prev.t != 0 && prev.t < t {
		t = prev.t
	}
	d.sent[msgID] = sent{t: t, method: method}
}

// Messages decodes a "msg" event with its body into logical messages.
func (d *Decoder) Messages(ev *wiretrace.Event, body []byte) ([]Message, error) {
	m, err := ev.Msg()
	if err != nil {
		return nil, err
	}
	base := Message{
		Seq:      ev.Seq,
		T:        ev.T,
		Dir:      m.Dir,
		SDC:      m.SDC,
		Session:  strconv.FormatInt(int64(m.Session), 10),
		Redacted: m.Redacted,
		Partial:  m.Partial,
	}
	out := m.Dir == wiretrace.DirOut

	data, gz, gzErr := unpack(body)
	if gzErr == nil && len(data) >= 4 && binary.LittleEndian.Uint32(data) == tljson.IDMsgContainer {
		return d.container(base, uint64(m.MsgID), data, gz, out), nil
	}
	msg := base
	msg.MsgIDValue = uint64(m.MsgID)
	msg.MsgID = strconv.FormatUint(msg.MsgIDValue, 10)
	msg.SeqNo = m.SeqNo
	d.fill(&msg, body, out)
	return []Message{msg}, nil
}

// unpack inflates a top-level gzip_packed body.
func unpack(body []byte) (data []byte, packed bool, err error) {
	if len(body) < 8 || binary.LittleEndian.Uint32(body) != tljson.IDGzipPacked {
		return body, false, nil
	}
	n := int(body[4])
	header := 1
	if n == 254 {
		n = int(body[5]) | int(body[6])<<8 | int(body[7])<<16
		header = 4
	}
	if 4+header+n > len(body) {
		return nil, true, errors.New("short gzip_packed")
	}
	r, err := gzip.NewReader(bytes.NewReader(body[4+header : 4+header+n]))
	if err != nil {
		return nil, true, errors.Wrap(err, "gzip")
	}
	data, err = io.ReadAll(io.LimitReader(r, tljson.MaxUnpacked))
	if err != nil {
		return nil, true, errors.Wrap(err, "gzip")
	}
	return data, true, nil
}

func (d *Decoder) container(base Message, parent uint64, data []byte, gz, out bool) []Message {
	var msgs []Message
	fail := func(reason string) []Message {
		msg := base
		msg.MsgIDValue = parent
		msg.MsgID = strconv.FormatUint(parent, 10)
		msg.Type = "msg_container"
		msg.Partial = true
		msg.DecodeError = reason
		return append(msgs, msg)
	}
	if len(data) < 8 {
		return fail("short container")
	}
	n := binary.LittleEndian.Uint32(data[4:])
	pos := 8
	for i := uint32(0); i < n; i++ {
		if len(data)-pos < 16 {
			return fail("short container item header")
		}
		msgID := binary.LittleEndian.Uint64(data[pos:])
		seqNo := binary.LittleEndian.Uint32(data[pos+8:])
		length := int(binary.LittleEndian.Uint32(data[pos+12:]))
		pos += 16
		if length < 0 || len(data)-pos < length {
			return fail("container item overflows body")
		}
		msg := base
		msg.MsgIDValue = msgID
		msg.MsgID = strconv.FormatUint(msgID, 10)
		msg.SeqNo = int64(seqNo)
		msg.Container = strconv.FormatUint(parent, 10)
		msg.Gzip = gz
		msg.Redacted = nil
		d.fill(&msg, data[pos:pos+length], out)
		msgs = append(msgs, msg)
		pos += length
	}
	attributeRedactions(msgs, base.Redacted)
	return msgs
}

// attributeRedactions assigns the event-level "redacted" entries to the
// container items they describe: entries naming the item method, one of its
// wrappers or its rpc_result, or items whose body has a redacted object.
// Entries that match no item are kept on every item.
func attributeRedactions(msgs []Message, entries []string) {
	for _, e := range entries {
		matched := false
		for i := range msgs {
			if redactionMatches(&msgs[i], e) {
				msgs[i].Redacted = append(msgs[i].Redacted, e)
				matched = true
			}
		}
		if matched {
			continue
		}
		for i := range msgs {
			msgs[i].Redacted = append(msgs[i].Redacted, e)
		}
	}
}

func redactionMatches(m *Message, entry string) bool {
	hasPrefix := func(name string) bool {
		return name != "" && (entry == name || strings.HasPrefix(entry, name+"."))
	}
	if m.Dir == wiretrace.DirOut && hasPrefix(m.Method) {
		return true
	}
	for _, w := range m.Wrap {
		if i := strings.IndexByte(w, ':'); i >= 0 {
			w = w[:i]
		}
		if hasPrefix(w) {
			return true
		}
	}
	if m.Type == "rpc_result" && (entry == "result" || strings.HasSuffix(entry, ".result")) {
		if entry != "result" && m.Method != "" {
			return entry == m.Method+".result"
		}
		return hasRedacted(m.Body)
	}
	return false
}

func hasRedacted(v any) bool {
	switch v := v.(type) {
	case *tljson.Object:
		if v.Name() == "redacted" {
			return true
		}
		for _, f := range v.Fields {
			if hasRedacted(f.Value) {
				return true
			}
		}
	case []any:
		for _, e := range v {
			if hasRedacted(e) {
				return true
			}
		}
	}
	return false
}

func (d *Decoder) fill(msg *Message, body []byte, out bool) {
	v, err := d.tl.Decode(body, out)
	msg.Body = v
	if err != nil {
		msg.Partial = true
		msg.DecodeError = err.Error()
	}
	obj, _ := v.(*tljson.Object)
	if len(body) >= 4 {
		msg.TypeID = binary.LittleEndian.Uint32(body)
	}
	if g := obj.Name(); g == "gzip_packed" {
		msg.Gzip = true
		if inner, ok := obj.Get("inner").(*tljson.Object); ok {
			obj = inner
		}
	}
	msg.Type = redactedName(obj)
	if msg.Type == "" {
		msg.Type = d.schema.Name(msg.TypeID)
	}
	if out {
		d.fillOut(msg, obj)
		return
	}
	if msg.Type == "rpc_result" {
		d.fillResult(msg, obj)
	}
}

func (d *Decoder) fillOut(msg *Message, obj *tljson.Object) {
	cur := obj
	for cur != nil {
		def, ok := d.schema.FunctionByName(cur.Name())
		if !ok {
			break
		}
		param, ok := def.GenericParam()
		if !ok {
			break
		}
		w := def.Name
		switch def.Name {
		case "invokeWithLayer":
			if l, ok := cur.Get("layer").(int32); ok {
				w += ":" + strconv.Itoa(int(l))
			}
		case "invokeAfterMsg":
			if id, ok := cur.Get("msg_id").(tljson.Long); ok {
				msg.After = strconv.FormatUint(uint64(id), 10) // #nosec G115
			}
		case "invokeAfterMsgs":
			if ids, ok := cur.Get("msg_ids").([]any); ok && len(ids) > 0 {
				if id, ok := ids[len(ids)-1].(tljson.Long); ok {
					msg.After = strconv.FormatUint(uint64(id), 10) // #nosec G115
				}
			}
		}
		msg.Wrap = append(msg.Wrap, w)
		next, _ := cur.Get(param).(*tljson.Object)
		cur = next
	}
	if cur != nil {
		msg.Method = redactedName(cur)
		if def, ok := d.schema.FunctionByName(msg.Method); ok {
			msg.MethodID = def.ID
		}
	}
	if msg.After != "" {
		if id, err := strconv.ParseUint(msg.After, 10, 64); err == nil {
			if s, ok := d.sent[id]; ok && s.method != 0 {
				msg.AfterMethod = d.schema.Name(s.method)
			}
		}
	}
	if msg.MethodID != 0 {
		d.remember(msg.MsgIDValue, msg.T, msg.MethodID)
	}
	if req, ok := d.msgReq[msg.MsgIDValue]; ok {
		msg.Req = req
		msg.Caller = d.reqCaller[req]
	}
}

func (d *Decoder) fillResult(msg *Message, obj *tljson.Object) {
	reqMsgID, _ := obj.Get("req_msg_id").(tljson.Long)
	id := uint64(reqMsgID) // #nosec G115
	msg.ReqMsgID = strconv.FormatUint(id, 10)
	if s, ok := d.sent[id]; ok {
		if s.method != 0 {
			msg.MethodID = s.method
			msg.Method = d.schema.Name(s.method)
		}
		lat := (msg.T - s.t) * 1000
		lat = math.Round(lat*1000) / 1000
		msg.LatencyMS = &lat
	}
	if req, ok := d.msgReq[id]; ok {
		msg.Req = req
		msg.Caller = d.reqCaller[req]
	}
	val := obj.Get("result")
	if g, ok := val.(*tljson.Object); ok && g.Name() == "gzip_packed" {
		msg.Gzip = true
		val = g.Get("inner")
	}
	res, _ := val.(*tljson.Object)
	switch v := val.(type) {
	case *tljson.Object:
		msg.Result = redactedName(v)
	case []any:
		msg.Result = "vector"
	case bool:
		if v {
			msg.Result = "boolTrue"
		} else {
			msg.Result = "boolFalse"
		}
	}
	if msg.Result == "rpc_error" {
		code, _ := res.Get("error_code").(int32)
		text, _ := res.Get("error_message").(string)
		msg.Error = &RPCError{Code: code, Message: text}
	}
}

// redactedName returns the object name, or the original name of a redacted
// object.
func redactedName(o *tljson.Object) string {
	if o.Name() == "redacted" {
		if orig, ok := o.Get("orig").(string); ok {
			return orig
		}
	}
	return o.Name()
}
