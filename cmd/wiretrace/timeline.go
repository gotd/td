package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gotd/td/wiretrace"
)

// commonKeys are not repeated in the generic event details.
var commonKeys = map[string]bool{
	"seq": true, "t": true, "th": true, "ev": true, "dc": true, "shift": true, "sdc": true, "cause": true,
	"blob": true, "b64": true,
}

type causeKey struct {
	req   int64
	outer uint64
}

// effects are the events caused by one response.
type effects map[string]int

func (e effects) String() string {
	keys := make([]string, 0, len(e))
	for k := range e {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		if e[k] > 1 {
			parts = append(parts, fmt.Sprintf("%s×%d", k, e[k]))
		} else {
			parts = append(parts, k)
		}
	}
	return strings.Join(parts, ", ")
}

type timeline struct {
	names   *namer
	reqType map[int64]uint32
	// sent is the method of outgoing msg_ids, from msg summaries.
	sent   map[uint64]uint32
	caused map[causeKey]effects
}

func newTimeline(n *namer) *timeline {
	return &timeline{
		names:   n,
		reqType: map[int64]uint32{},
		sent:    map[uint64]uint32{},
		caused:  map[causeKey]effects{},
	}
}

// index is the first pass: request methods, outgoing methods and effects.
func (tl *timeline) index(ev *wiretrace.Event) {
	switch ev.Kind {
	case wiretrace.KindReqNew:
		tl.reqType[ev.Int("req")] = ev.Type("type")
	case wiretrace.KindMsg:
		if m, err := ev.Msg(); err == nil && m.Dir == wiretrace.DirOut {
			tl.sent[uint64(m.MsgID)] = innerID(m.Method, m.Type)
			for _, in := range m.Inner {
				tl.sent[uint64(in.MsgID)] = innerID(in.Method, in.Type)
			}
		}
	}
	if ev.Cause != nil {
		k := causeKey{req: ev.Cause.Req, outer: uint64(ev.Cause.Outer)}
		e := tl.caused[k]
		if e == nil {
			e = effects{}
			tl.caused[k] = e
		}
		e[tl.effectName(ev)]++
	}
}

func innerID(method, typ wiretrace.TypeID) uint32 {
	if method != 0 {
		return uint32(method)
	}
	return uint32(typ)
}

func (tl *timeline) effectName(ev *wiretrace.Event) string {
	switch ev.Kind {
	case wiretrace.KindReqNew:
		return "req.new " + tl.names.Name(ev.Type("type"))
	case wiretrace.KindStorageWrite, wiretrace.KindStorageRaw, wiretrace.KindStorageCommit,
		wiretrace.KindStorageRead, wiretrace.KindStorageRemove:
		return ev.Kind + " " + storageKind(ev.Str("file"))
	}
	return ev.Kind
}

var hexRunRe = regexp.MustCompile(`[0-9A-Fa-f]{8,}`)

// storageKind is a coarse label of a tdata file: its base name with long hex
// runs collapsed (e.g. "map", "key_data", "<hex>s").
func storageKind(file string) string {
	return hexRunRe.ReplaceAllString(path.Base(strings.ReplaceAll(file, "\\", "/")), "<hex>")
}

func dcLabel(ev *wiretrace.Event) string {
	if !ev.Has("sdc") {
		return ""
	}
	return "dc" + strconv.FormatInt(ev.Int("sdc"), 10)
}

// msgShape describes a msg event from its summary fields.
func (tl *timeline) msgShape(m *wiretrace.Msg) string {
	if len(m.Inner) > 0 {
		parts := make([]string, 0, len(m.Inner))
		for _, in := range m.Inner {
			parts = append(parts, tl.itemShape(m.Dir, in.Type, in.Method, nil, in.ReqMsgID, in.Result))
		}
		return "container[" + strings.Join(parts, ", ") + "]"
	}
	return tl.itemShape(m.Dir, m.Type, m.Method, m.Wrap, m.ReqMsgID, m.Result)
}

func (tl *timeline) itemShape(dir string, typ, method wiretrace.TypeID, wrap []wiretrace.TypeID,
	reqMsgID wiretrace.ID, result wiretrace.TypeID,
) string {
	n := tl.names.Name
	if dir == wiretrace.DirIn && uint32(typ) == 0xf35c6d01 {
		req := "?"
		if id, ok := tl.sent[uint64(reqMsgID)]; ok {
			req = n(id)
		}
		return fmt.Sprintf("rpc_result(%s → %s)", req, n(uint32(result)))
	}
	if method == 0 || method == typ {
		if len(wrap) == 0 {
			return n(uint32(typ))
		}
	}
	inner := n(innerID(method, typ))
	if len(wrap) == 0 && method != 0 && method != typ {
		// Container items: the summary has no wrap list.
		return n(uint32(typ)) + "(…" + inner + ")"
	}
	s := inner
	for i := len(wrap) - 1; i >= 0; i-- {
		s = n(uint32(wrap[i])) + "(" + s + ")"
	}
	return s
}

func (tl *timeline) causeSuffix(ev *wiretrace.Event) string {
	c := ev.Cause
	if c == nil {
		return ""
	}
	if c.Req == 0 {
		return fmt.Sprintf("  ⇐ update %s", tl.names.Name(uint32(c.Type)))
	}
	method := ""
	if t, ok := tl.reqType[c.Req]; ok {
		method = " " + tl.names.Name(t)
	}
	return fmt.Sprintf("  ⇐ req#%d%s → %s", c.Req, method, tl.names.Name(uint32(c.Type)))
}

func genericDetails(ev *wiretrace.Event) string {
	keys := make([]string, 0, len(ev.Fields))
	for k := range ev.Fields {
		if !commonKeys[k] {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		raw := ev.Fields[k]
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			s = string(raw)
		}
		parts = append(parts, k+"="+s)
	}
	return strings.Join(parts, " ")
}

func (tl *timeline) line(ev *wiretrace.Event) string {
	n := tl.names.Name
	var detail string
	switch ev.Kind {
	case wiretrace.KindHeader:
		return fmt.Sprintf("%12s %-5s hdr seg=%d app=%s layer=%d", "", "", ev.Int("seg"), ev.Str("app"), ev.Int("layer"))
	case wiretrace.KindDropped:
		detail = fmt.Sprintf("DROPPED %d events (writer queue overflow)", ev.ID("count"))
	case wiretrace.KindMsg:
		m, err := ev.Msg()
		if err != nil {
			detail = "msg: " + err.Error()
			break
		}
		arrow := "→"
		if m.Dir == wiretrace.DirIn {
			arrow = "←"
		}
		detail = fmt.Sprintf("%s %s %d seqno=%d len=%d %s", arrow, dcLabel(ev), uint64(m.MsgID), m.SeqNo, m.Len, tl.msgShape(m))
		if len(m.Redacted) > 0 {
			detail += " redacted=" + strings.Join(m.Redacted, ",")
		}
	case wiretrace.KindReqNew:
		detail = fmt.Sprintf("req#%d new %s %s size=%d", ev.Int("req"), n(ev.Type("type")), dcLabel(ev), ev.Int("size"))
		if ev.Bool("to_main") {
			detail += " to_main"
		}
		if v := ev.Int("can_wait"); v != 0 {
			detail += fmt.Sprintf(" can_wait=%dms", v)
		}
		if v := ev.Int("after"); v != 0 {
			detail += fmt.Sprintf(" after=req#%d", v)
		}
		if ev.Bool("needs_layer") {
			detail += " needs_layer"
		}
	case wiretrace.KindReqMsg:
		detail = fmt.Sprintf("req#%d → msg %d seqno=%d", ev.Int("req"), ev.ID("msg_id"), ev.Int("seqno"))
		if was := ev.ID("was"); was != 0 {
			detail += fmt.Sprintf(" (was %d)", was)
		}
	case wiretrace.KindReqCaller:
		detail = fmt.Sprintf("req#%d caller %s %s:%d", ev.Int("req"), ev.Str("subsystem"), ev.Str("file"), ev.Int("line"))
	case wiretrace.KindReqDone, wiretrace.KindReqFail:
		verb := "done"
		if ev.Kind == wiretrace.KindReqFail {
			verb = "FAIL"
		}
		detail = fmt.Sprintf("req#%d %s %s → %s lat=%.1fms", ev.Int("req"), verb, n(ev.Type("method")), n(ev.Type("type")),
			ev.Float("lat_ms"))
		if ev.Kind == wiretrace.KindReqFail {
			detail += fmt.Sprintf(" error=%d %s", ev.Int("code"), ev.Str("error"))
		}
		if e := tl.caused[causeKey{req: ev.Int("req"), outer: ev.ID("outer")}]; len(e) > 0 {
			detail += "  → caused " + e.String()
		}
	default:
		detail = ev.Kind
		if dc := dcLabel(ev); dc != "" {
			detail += " " + dc
		}
		if g := genericDetails(ev); g != "" {
			detail += " " + g
		}
	}
	return fmt.Sprintf("%12.6f %-5s %s%s", ev.T, ev.Thread, detail, tl.causeSuffix(ev))
}

func matchKind(kinds []string, kind string) bool {
	if len(kinds) == 0 {
		return true
	}
	for _, k := range kinds {
		if k == kind || (strings.HasSuffix(k, "*") && strings.HasPrefix(kind, strings.TrimSuffix(k, "*"))) {
			return true
		}
	}
	return false
}

func runTimeline(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("timeline", flag.ContinueOnError)
	schemaDir := fs.String("schema", "", "directory with api.tl and mtproto.tl, or comma-separated .tl files (optional, for names)")
	noStorage := fs.Bool("no-storage", false, "hide st.* events")
	kindsFlag := fs.String("kinds", "", "comma-separated event kinds to show, \"req.*\" style prefixes allowed")
	if err := parseFlags(fs, args, 1, "timeline [flags] RUN", stderr); err != nil {
		return err
	}
	run, err := openRun(fs.Arg(0))
	if err != nil {
		return err
	}
	defer func() { _ = run.Close() }()
	schema, err := loadSchema(stderr, *schemaDir, run, false)
	if err != nil {
		return err
	}
	var kinds []string
	if *kindsFlag != "" {
		kinds = strings.Split(*kindsFlag, ",")
	}
	tl := newTimeline(newNamer(schema))
	var bad badLines
	if err := run.Events(func(ev *wiretrace.Event) error {
		tl.index(ev)
		return nil
	}, bad.add); err != nil {
		return err
	}
	w := bufio.NewWriter(stdout)
	if err := run.Events(func(ev *wiretrace.Event) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if *noStorage && strings.HasPrefix(ev.Kind, "st.") {
			return nil
		}
		if !matchKind(kinds, ev.Kind) {
			return nil
		}
		_, err := fmt.Fprintln(w, tl.line(ev))
		return err
	}, nil); err != nil {
		return err
	}
	bad.report(stderr)
	return w.Flush()
}
