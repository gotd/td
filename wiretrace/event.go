package wiretrace

import (
	"encoding/json"
	"strconv"

	"github.com/go-faster/errors"
)

// Event is one record of an events-NNNN.jsonl segment.
//
// Common fields are parsed, every field (including the common ones) is
// available as raw JSON in Fields.
type Event struct {
	Seq    uint64
	T      float64
	Thread string
	Kind   string
	// Blob references the body of the event, if any.
	Blob *BlobRef
	// Inline is the decoded "b64" field, if the writer inlined the body.
	Inline []byte
	// Cause is the causal context, if any.
	Cause *Cause
	// UI is the UI action stamp, if any. It is never set for "hdr" records
	// (their "ui" is HeaderUI).
	UI *UIStamp
	// Fields holds all fields of the record.
	Fields map[string]json.RawMessage

	// Segment is the index of the events segment the record was read from.
	Segment int
	// Line is the 1-based line number in that segment.
	Line int
	// Raw is the (repaired, see RepairLegacyNumbers) JSON line.
	Raw []byte
}

type eventCommon struct {
	Seq    uint64   `json:"seq"`
	T      float64  `json:"t"`
	Thread string   `json:"th"`
	Kind   string   `json:"ev"`
	Blob   *BlobRef `json:"blob"`
	B64    []byte   `json:"b64"`
	Cause  *Cause   `json:"cause"`
	UI     *UIStamp `json:"ui"`
}

// ParseEvent parses one JSON line.
//
// Lines written by older tdesktop builds under a locale with a comma decimal
// separator (`"t":12,345678`) are repaired, see RepairLegacyNumbers.
func ParseEvent(line []byte) (*Event, error) {
	ev, err := parseEvent(line)
	if err == nil {
		return ev, nil
	}
	repaired, ok := RepairLegacyNumbers(line)
	if !ok {
		return nil, err
	}
	ev, rErr := parseEvent(repaired)
	if rErr != nil {
		return nil, err
	}
	return ev, nil
}

func parseEvent(line []byte) (*Event, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(line, &fields); err != nil {
		return nil, errors.Wrap(err, "parse event")
	}
	var c eventCommon
	if err := json.Unmarshal(line, &c); err != nil {
		// A malformed "ui" object must not hide the rest of the event.
		rest := make(map[string]json.RawMessage, len(fields))
		for k, v := range fields {
			if k != "ui" {
				rest[k] = v
			}
		}
		c = eventCommon{}
		raw, mErr := json.Marshal(rest)
		if mErr != nil || json.Unmarshal(raw, &c) != nil {
			return nil, errors.Wrap(err, "parse event fields")
		}
	}
	if c.Kind == KindHeader || (c.UI != nil && c.UI.A == 0 && c.UI.Rel == "") {
		c.UI = nil
	}
	return &Event{
		Seq:    c.Seq,
		T:      c.T,
		Thread: c.Thread,
		Kind:   c.Kind,
		Blob:   c.Blob,
		Inline: c.B64,
		Cause:  c.Cause,
		UI:     c.UI,
		Fields: fields,
		Raw:    line,
	}, nil
}

// RepairLegacyNumbers rewrites numbers like `:12,345678` (comma decimal
// separator, written by tdesktop trace builds before the locale fix) into
// `:12.345678` when they are followed by `,`, `}` or `]`.
//
// Valid JSON never has `:<digits>,<digits>` followed by one of those
// characters outside of strings, so the repair is applied only to lines that
// failed to parse. Returns false if nothing was changed.
func RepairLegacyNumbers(line []byte) ([]byte, bool) {
	out := make([]byte, 0, len(line))
	changed := false
	inString := false
	for i := 0; i < len(line); i++ {
		c := line[i]
		if inString {
			out = append(out, c)
			switch c {
			case '\\':
				if i+1 < len(line) {
					i++
					out = append(out, line[i])
				}
			case '"':
				inString = false
			}
			continue
		}
		if c == '"' {
			inString = true
			out = append(out, c)
			continue
		}
		if c != ':' {
			out = append(out, c)
			continue
		}
		out = append(out, c)
		// :[-]digits,digits[,}\]]
		j := i + 1
		if j < len(line) && line[j] == '-' {
			j++
		}
		intStart := j
		for j < len(line) && isDigit(line[j]) {
			j++
		}
		if j == intStart || j >= len(line) || line[j] != ',' {
			continue
		}
		comma := j
		j++
		fracStart := j
		for j < len(line) && isDigit(line[j]) {
			j++
		}
		if j == fracStart || j >= len(line) {
			continue
		}
		switch line[j] {
		case ',', '}', ']':
		default:
			continue
		}
		out = append(out, line[i+1:comma]...)
		out = append(out, '.')
		out = append(out, line[fracStart:j]...)
		i = j - 1
		changed = true
	}
	if !changed {
		return line, false
	}
	return out, true
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

// Has reports whether the field is present.
func (e *Event) Has(key string) bool {
	_, ok := e.Fields[key]
	return ok
}

// Decode decodes the field into v. Returns false if the field is missing or
// can't be decoded.
func (e *Event) Decode(key string, v any) bool {
	raw, ok := e.Fields[key]
	if !ok {
		return false
	}
	return json.Unmarshal(raw, v) == nil
}

// Str returns a string field, or "" if missing.
func (e *Event) Str(key string) string {
	var s string
	if !e.Decode(key, &s) {
		return ""
	}
	return s
}

// Int returns an integer field (JSON number or decimal string), or 0.
func (e *Event) Int(key string) int64 {
	raw, ok := e.Fields[key]
	if !ok {
		return 0
	}
	v, err := strconv.ParseInt(unquote(raw), 10, 64)
	if err != nil {
		f, fErr := strconv.ParseFloat(unquote(raw), 64)
		if fErr != nil {
			return 0
		}
		return int64(f)
	}
	return v
}

// ID returns an unsigned 64-bit id field (decimal string or number), or 0.
func (e *Event) ID(key string) uint64 {
	raw, ok := e.Fields[key]
	if !ok {
		return 0
	}
	v, err := parseUint(raw)
	if err != nil {
		return 0
	}
	return v
}

// Float returns a number field, or 0.
func (e *Event) Float(key string) float64 {
	raw, ok := e.Fields[key]
	if !ok {
		return 0
	}
	v, err := strconv.ParseFloat(unquote(raw), 64)
	if err != nil {
		return 0
	}
	return v
}

// Bool returns a boolean field, or false.
func (e *Event) Bool(key string) bool {
	var v bool
	if !e.Decode(key, &v) {
		return false
	}
	return v
}

// Type returns a constructor id field ("0x%08x"), or 0.
func (e *Event) Type(key string) uint32 {
	var v TypeID
	if !e.Decode(key, &v) {
		return 0
	}
	return uint32(v)
}

// Msg is the parsed form of a "msg" event.
type Msg struct {
	Dir       string   `json:"dir"`
	SDC       int      `json:"sdc"`
	DC        int      `json:"dc"`
	Shift     int      `json:"shift"`
	PDC       int      `json:"pdc"`
	Conn      string   `json:"conn,omitempty"`
	Key       ID       `json:"key"`
	KeyHex    string   `json:"key_hex,omitempty"`
	KeyKind   string   `json:"key_kind,omitempty"`
	Session   SignedID `json:"session"`
	Salt      ID       `json:"salt"`
	MsgID     ID       `json:"msg_id"`
	SeqNo     int64    `json:"seqno"`
	Len       int64    `json:"len"`
	Truncated bool     `json:"truncated,omitempty"`
	Type      TypeID   `json:"type"`
	Method    TypeID   `json:"method,omitempty"`
	Wrap      []TypeID `json:"wrap,omitempty"`
	ReqMsgID  ID       `json:"req_msg_id,omitempty"`
	Result    TypeID   `json:"result,omitempty"`
	Inner     []Inner  `json:"inner,omitempty"`
	Redacted  []string `json:"redacted,omitempty"`
	PII       bool     `json:"pii,omitempty"`
	Partial   bool     `json:"partial,omitempty"`
}

// Msg parses a "msg" event.
func (e *Event) Msg() (*Msg, error) {
	if e.Kind != KindMsg {
		return nil, errors.Errorf("event %d is %q, not %q", e.Seq, e.Kind, KindMsg)
	}
	var m Msg
	if err := json.Unmarshal(e.Raw, &m); err != nil {
		return nil, errors.Wrap(err, "parse msg")
	}
	return &m, nil
}

// Header parses a "hdr" event.
func (e *Event) Header() (*Header, error) {
	if e.Kind != KindHeader {
		return nil, errors.Errorf("event is %q, not %q", e.Kind, KindHeader)
	}
	var h Header
	if err := json.Unmarshal(e.Raw, &h); err != nil {
		return nil, errors.Wrap(err, "parse hdr")
	}
	return &h, nil
}
