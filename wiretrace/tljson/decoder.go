package tljson

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"unicode/utf8"

	"github.com/go-faster/errors"

	"github.com/gotd/tl"
)

// RedactedMarker follows the constructor id of an object whose body was
// redacted by the tracer: [id][RedactedMarker][id][zeros...].
const RedactedMarker uint32 = 0x7EDAC7ED

// Built-in type names.
const (
	typeInt    = "int"
	typeLong   = "long"
	typeTrue   = "true"
	typeVector = "vector"
)

// Limits of the decoder.
const (
	DefaultMaxDepth = 64
	// MaxUnpacked limits the size of inflated gzip_packed data.
	MaxUnpacked = 32 << 20
	maxVector   = 1 << 22
)

// Decoder decodes TL binary data into Objects using a Schema.
type Decoder struct {
	Schema *Schema
	// MaxBytes replaces TL bytes values longer than this with
	// {"_":"bytes","len":N,"sha256":"<first 8 bytes>"}. Zero means no limit.
	MaxBytes int
	// MaxDepth limits nesting, DefaultMaxDepth if zero.
	MaxDepth int
	// ResultType resolves the result type of rpc_result by req_msg_id (the
	// return type of the request method). If it returns false, the result
	// is decoded as a boxed Object.
	ResultType func(reqMsgID int64) (tl.Type, bool)
}

// Error is returned when decoding stopped before the end of the data. The
// returned value is still usable and is marked "partial":true.
type Error struct {
	// Unknown is the unknown constructor id that stopped decoding, if any.
	Unknown uint32
	Msg     string
}

func (e *Error) Error() string { return e.Msg }

// Decode decodes a boxed object (function or constructor) from data.
//
// If function is true, function ids take precedence over constructor ids
// (outgoing messages).
func (d *Decoder) Decode(data []byte, function bool) (any, error) {
	s := d.state(data)
	v, err := s.boxed(nil, function)
	if err != nil {
		return v, err
	}
	return v, s.trailing()
}

// DecodeAs decodes a value of type t (boxed or bare) from data.
func (d *Decoder) DecodeAs(data []byte, t tl.Type) (any, error) {
	s := d.state(data)
	v, err := s.value(t)
	if err != nil {
		return v, err
	}
	return v, s.trailing()
}

func (d *Decoder) state(data []byte) *state {
	depth := d.MaxDepth
	if depth == 0 {
		depth = DefaultMaxDepth
	}
	return &state{d: d, buf: data, maxDepth: depth}
}

type state struct {
	d        *Decoder
	buf      []byte
	pos      int
	depth    int
	maxDepth int
	// redacted is set when a redacted object consumed the rest of buf.
	redacted bool
}

func (s *state) fail(format string, args ...any) error {
	return &Error{Msg: fmt.Sprintf("at %d: ", s.pos) + fmt.Sprintf(format, args...)}
}

func (s *state) trailing() error {
	if s.redacted || s.pos >= len(s.buf) {
		return nil
	}
	return s.fail("%d trailing bytes", len(s.buf)-s.pos)
}

func (s *state) remaining() int { return len(s.buf) - s.pos }

func (s *state) u32() (uint32, error) {
	if s.remaining() < 4 {
		return 0, s.fail("unexpected end of data (want 4 bytes, have %d)", s.remaining())
	}
	v := binary.LittleEndian.Uint32(s.buf[s.pos:])
	s.pos += 4
	return v, nil
}

func (s *state) u64() (uint64, error) {
	if s.remaining() < 8 {
		return 0, s.fail("unexpected end of data (want 8 bytes, have %d)", s.remaining())
	}
	v := binary.LittleEndian.Uint64(s.buf[s.pos:])
	s.pos += 8
	return v, nil
}

func (s *state) raw(n int) ([]byte, error) {
	if n < 0 || s.remaining() < n {
		return nil, s.fail("unexpected end of data (want %d bytes, have %d)", n, s.remaining())
	}
	v := s.buf[s.pos : s.pos+n]
	s.pos += n
	return v, nil
}

// tlBytes reads TL string/bytes with padding.
func (s *state) tlBytes() ([]byte, error) {
	if s.remaining() < 1 {
		return nil, s.fail("unexpected end of data (bytes)")
	}
	n := int(s.buf[s.pos])
	header := 1
	switch n {
	case 254:
		if s.remaining() < 4 {
			return nil, s.fail("unexpected end of data (bytes header)")
		}
		n = int(s.buf[s.pos+1]) | int(s.buf[s.pos+2])<<8 | int(s.buf[s.pos+3])<<16
		header = 4
	case 255:
		return nil, s.fail("bad bytes length prefix 255")
	}
	full := (header + n + 3) &^ 3
	if s.remaining() < full {
		return nil, s.fail("bytes of %d overflow data (have %d)", n, s.remaining())
	}
	v := s.buf[s.pos+header : s.pos+header+n]
	s.pos += full
	return v, nil
}

func (s *state) bytesValue(v []byte) any {
	if s.d.MaxBytes > 0 && len(v) > s.d.MaxBytes {
		sum := sha256.Sum256(v)
		o := NewObject("bytes")
		o.Set("len", len(v))
		o.Set("sha256", hex.EncodeToString(sum[:8]))
		return o
	}
	return Bytes(append([]byte(nil), v...))
}

func (s *state) stringValue(v []byte) any {
	if !utf8.Valid(v) {
		o := NewObject("bytes")
		o.Set("b64", Bytes(append([]byte(nil), v...)))
		return o
	}
	return string(v)
}

// isRedacted checks for [RedactedMarker][id] right after id at s.pos.
func (s *state) isRedacted(id uint32) bool {
	if s.remaining() < 4 || binary.LittleEndian.Uint32(s.buf[s.pos:]) != RedactedMarker {
		return false
	}
	if s.remaining() >= 8 {
		return binary.LittleEndian.Uint32(s.buf[s.pos+4:]) == id
	}
	return true
}

func (s *state) enter() error {
	s.depth++
	if s.depth > s.maxDepth {
		return s.fail("max depth %d exceeded", s.maxDepth)
	}
	return nil
}

func (s *state) leave() { s.depth-- }

// boxed decodes a boxed value. expect is the expected type, if known.
func (s *state) boxed(expect *tl.Type, function bool) (any, error) {
	if err := s.enter(); err != nil {
		return nil, err
	}
	defer s.leave()

	start := s.pos
	id, err := s.u32()
	if err != nil {
		return nil, err
	}
	if s.isRedacted(id) {
		o := NewObject("redacted")
		o.Set("orig", s.d.Schema.Name(id))
		o.Set("len", len(s.buf)-start)
		s.pos = len(s.buf)
		s.redacted = true
		return o, nil
	}
	switch id {
	case IDVector:
		if expect != nil && TypeName(*expect) == "Vector" && expect.GenericArg != nil {
			return s.vector(*expect.GenericArg)
		}
		return s.vectorGuess()
	case IDGzipPacked:
		return s.gzip(expect, function)
	case IDRPCResult:
		return s.rpcResult()
	case IDMsgContainer:
		return s.container(function)
	case IDBoolTrue:
		return true, nil
	case IDBoolFalse:
		return false, nil
	}
	def, ok := s.d.Schema.Lookup(id, function)
	if !ok {
		o := NewObject(UnknownName(id))
		o.Set("hex", hex.EncodeToString(s.buf[s.pos:]))
		s.pos = len(s.buf)
		return o, &Error{Unknown: id, Msg: fmt.Sprintf("at %d: unknown constructor %s", start, UnknownName(id))}
	}
	o := NewObject(def.Name)
	return o, s.params(def, o)
}

func markPartial(o *Object, err error) {
	var e *Error
	if errors.As(err, &e) && e.Unknown == 0 && o.Get("error") == nil && !hasErrorBelow(o) {
		o.Set("error", e.Msg)
	}
	o.Set("partial", true)
}

// hasErrorBelow reports whether a nested object already carries the error.
func hasErrorBelow(o *Object) bool {
	for _, f := range o.Fields {
		switch v := f.Value.(type) {
		case *Object:
			if v.Get("error") != nil || hasErrorBelow(v) {
				return true
			}
		case []any:
			for _, e := range v {
				if c, ok := e.(*Object); ok && (c.Get("error") != nil || hasErrorBelow(c)) {
					return true
				}
			}
		}
	}
	return false
}

func (s *state) params(def *Def, o *Object) error {
	flags := map[string]uint32{}
	for _, p := range def.Params {
		if s.redacted {
			return nil
		}
		if p.Flags {
			v, err := s.u32()
			if err != nil {
				markPartial(o, err)
				return err
			}
			flags[p.Name] = v
			o.Set(p.Name, v)
			continue
		}
		if p.Flag != nil {
			if flags[p.Flag.Name]&(1<<uint(p.Flag.Index)) == 0 {
				continue
			}
			if TypeName(p.Type) == typeTrue {
				o.Set(p.Name, true)
				continue
			}
		}
		v, err := s.value(p.Type)
		if v != nil {
			o.Set(p.Name, v)
		}
		if err != nil {
			markPartial(o, err)
			return err
		}
	}
	return nil
}

// value decodes a parameter of type t.
func (s *state) value(t tl.Type) (any, error) {
	if t.GenericRef {
		return s.boxed(nil, true)
	}
	switch TypeName(t) {
	case typeInt:
		v, err := s.u32()
		return int32(v), err // #nosec G115
	case "#":
		return s.u32()
	case typeLong:
		v, err := s.u64()
		return Long(v), err // #nosec G115
	case "double":
		v, err := s.u64()
		return math.Float64frombits(v), err
	case "string":
		v, err := s.tlBytes()
		if err != nil {
			return nil, err
		}
		return s.stringValue(v), nil
	case "bytes":
		v, err := s.tlBytes()
		if err != nil {
			return nil, err
		}
		return s.bytesValue(v), nil
	case "int128":
		v, err := s.raw(16)
		return hex.EncodeToString(v), err
	case "int256":
		v, err := s.raw(32)
		return hex.EncodeToString(v), err
	case typeTrue:
		return true, nil
	case typeVector:
		if t.GenericArg == nil {
			return nil, s.fail("bare vector without element type")
		}
		return s.vector(*t.GenericArg)
	case "Object":
		return s.boxed(nil, false)
	}
	if t.Bare {
		def, ok := s.d.Schema.bareConstructor(t)
		if !ok {
			return nil, s.fail("unknown bare type %s", t)
		}
		if err := s.enter(); err != nil {
			return nil, err
		}
		defer s.leave()
		o := NewObject(def.Name)
		return o, s.params(def, o)
	}
	return s.boxed(&t, false)
}

func minSize(t tl.Type) int {
	switch TypeName(t) {
	case "long", "double":
		return 8
	case "int128":
		return 16
	case "int256":
		return 32
	case "true":
		return 0
	}
	if t.Bare && TypeName(t) != "int" && TypeName(t) != "string" && TypeName(t) != "bytes" {
		return 0
	}
	return 4
}

func (s *state) count(elemSize int) (int, error) {
	n, err := s.u32()
	if err != nil {
		return 0, err
	}
	if n > maxVector || (elemSize > 0 && int(n) > s.remaining()/elemSize) {
		return 0, s.fail("vector of %d elements overflows data (%d bytes)", n, s.remaining())
	}
	return int(n), nil
}

func (s *state) vector(elem tl.Type) (any, error) {
	n, err := s.count(minSize(elem))
	if err != nil {
		return nil, err
	}
	out := make([]any, 0, n)
	for i := 0; i < n; i++ {
		v, err := s.value(elem)
		if v != nil {
			out = append(out, v)
		}
		if err != nil {
			return out, err
		}
		if s.redacted {
			break
		}
	}
	return out, nil
}

// vectorGuess decodes a vector with an unknown element type: boxed objects
// if the first element is a known constructor, else long or int by size.
func (s *state) vectorGuess() (any, error) {
	n, err := s.count(4)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return []any{}, nil
	}
	var elem tl.Type
	first := binary.LittleEndian.Uint32(s.buf[s.pos:])
	_, known := s.d.Schema.Lookup(first, false)
	switch {
	case known || first == IDVector || first == IDBoolTrue || first == IDBoolFalse:
		elem = tl.Type{Name: "Object"}
	case s.remaining() == n*8:
		elem = tl.Type{Name: "long", Bare: true}
	case s.remaining() == n*4:
		elem = tl.Type{Name: "int", Bare: true}
	default:
		o := NewObject("vector")
		o.Set("count", n)
		o.Set("hex", hex.EncodeToString(s.buf[s.pos:]))
		s.pos = len(s.buf)
		return o, s.fail("vector with unknown element type")
	}
	s.pos -= 4 // re-read count
	return s.vector(elem)
}

func (s *state) gzip(expect *tl.Type, function bool) (any, error) {
	packed, err := s.tlBytes()
	if err != nil {
		return nil, err
	}
	o := NewObject("gzip_packed")
	info := &Object{}
	info.Set("len", len(packed))
	o.Set("gzip", info)
	r, err := gzip.NewReader(bytes.NewReader(packed))
	if err != nil {
		markPartial(o, s.fail("gzip: %v", err))
		return o, s.fail("gzip: %v", err)
	}
	data, err := io.ReadAll(io.LimitReader(r, MaxUnpacked+1))
	if err != nil {
		markPartial(o, s.fail("gzip: %v", err))
		return o, s.fail("gzip: %v", err)
	}
	if len(data) > MaxUnpacked {
		return o, s.fail("gzip: unpacked data exceeds %d bytes", MaxUnpacked)
	}
	info.Set("unpacked", len(data))
	sub := &state{d: s.d, buf: data, depth: s.depth, maxDepth: s.maxDepth}
	v, err := sub.boxed(expect, function)
	if err == nil {
		err = sub.trailing()
	}
	o.Set("inner", v)
	if err != nil {
		o.Set("partial", true)
		return o, err
	}
	return o, nil
}

func (s *state) rpcResult() (any, error) {
	o := NewObject("rpc_result")
	req, err := s.u64()
	if err != nil {
		markPartial(o, err)
		return o, err
	}
	o.Set("req_msg_id", Long(req)) // #nosec G115
	var expect *tl.Type
	if s.d.ResultType != nil {
		if t, ok := s.d.ResultType(int64(req)); ok { // #nosec G115
			expect = &t
		}
	}
	// rpc_error is not of the method type.
	if s.remaining() >= 4 && binary.LittleEndian.Uint32(s.buf[s.pos:]) == IDRPCError {
		expect = nil
	}
	v, err := s.boxed(expect, false)
	o.Set("result", v)
	if err != nil {
		o.Set("partial", true)
		return o, err
	}
	return o, nil
}

func (s *state) container(function bool) (any, error) {
	o := NewObject("msg_container")
	n, err := s.count(16)
	if err != nil {
		markPartial(o, err)
		return o, err
	}
	var items []any
	for i := 0; i < n; i++ {
		item := NewObject("message")
		items = append(items, item)
		msgID, err := s.u64()
		if err != nil {
			markPartial(o, err)
			o.Set("messages", items)
			return o, err
		}
		seqNo, err := s.u32()
		if err != nil {
			markPartial(o, err)
			o.Set("messages", items)
			return o, err
		}
		length, err := s.u32()
		if err != nil {
			markPartial(o, err)
			o.Set("messages", items)
			return o, err
		}
		body, err := s.raw(int(length))
		if err != nil {
			markPartial(o, err)
			o.Set("messages", items)
			return o, err
		}
		item.Set("msg_id", Long(msgID))  // #nosec G115
		item.Set("seqno", int32(seqNo))  // #nosec G115
		item.Set("bytes", int32(length)) // #nosec G115
		sub := &state{d: s.d, buf: body, depth: s.depth, maxDepth: s.maxDepth}
		v, err := sub.boxed(nil, function)
		if err == nil {
			err = sub.trailing()
		}
		item.Set("body", v)
		if err != nil {
			// Items have explicit lengths, the container recovers.
			item.Set("partial", true)
			o.Set("partial", true)
		}
	}
	o.Set("messages", items)
	return o, nil
}
