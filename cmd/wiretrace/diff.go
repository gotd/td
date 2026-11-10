package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/gotd/td/wiretrace"
	"github.com/gotd/td/wiretrace/decode"
	"github.com/gotd/td/wiretrace/tljson"
)

// Phases of a trace, in order.
var phaseNames = []string{"connect", "bind", "init", "auth", "bootstrap", "steady"}

const (
	phaseBind = iota + 1
	phaseInit
	phaseAuth
	phaseBootstrap
	phaseSteady
)

// maxLCS limits the number of messages per phase compared with LCS.
const maxLCS = 3000

// outMsg is one outgoing top-level message (a container counts as one).
type outMsg struct {
	t       float64
	token   string
	methods []string
	// fields are the flattened non-volatile fields per item token.
	fields map[string]map[string]string
}

type phase struct {
	start float64
	msgs  []outMsg
}

type traceShape struct {
	phases    []*phase
	methods   map[string]int
	pings     []float64
	restarts  []int64
	initMsgs  map[string]bool
	authMsgs  map[string]bool
	cur       int
	authAwait int
}

func newTraceShape() *traceShape {
	s := &traceShape{
		methods:  map[string]int{},
		initMsgs: map[string]bool{},
		authMsgs: map[string]bool{},
	}
	for range phaseNames {
		s.phases = append(s.phases, &phase{start: -1})
	}
	return s
}

func (s *traceShape) advance(p int) {
	if p > s.cur {
		s.cur = p
	}
}

// volatileKeys are masked in field diffs.
var volatileKeys = map[string]bool{
	"msg_id": true, "msg_ids": true, "random_id": true, "nonce": true, "hash": true, "date": true,
	"ping_id": true, "access_hash": true, "file_reference": true, "query_id": true, "pts": true,
	"qts": true, "seq": true, "offset_id": true, "offset_date": true, "max_id": true, "min_id": true,
	"id": true, "user_id": true, "chat_id": true, "channel_id": true, "peer": true, "offset_peer": true,
	"bytes": true, "token": true, "except_ids": true, "phone_code_hash": true, "file_token": true,
	"offset": true, "limit": true, "location": true, "message": true, "phone_number": true,
}

func messageToken(m *decode.Message) string {
	name := m.Method
	if name == "" {
		name = m.Type
	}
	switch name {
	case "msgs_ack":
		name = "msgs_ack(n)"
	case "ping_delay_disconnect":
		if obj, ok := innermost(m).(*tljson.Object); ok {
			if v, ok := obj.Get("disconnect_delay").(int32); ok {
				name = fmt.Sprintf("ping_delay_disconnect(%d)", v)
			}
		}
	}
	for i := len(m.Wrap) - 1; i >= 0; i-- {
		w := m.Wrap[i]
		if j := strings.IndexByte(w, ':'); j >= 0 {
			w = w[:j]
		}
		name = w + "(" + name + ")"
	}
	return name
}

// innermost returns the innermost query of an outgoing message body.
func innermost(m *decode.Message) any {
	v := m.Body
	for range m.Wrap {
		obj, ok := v.(*tljson.Object)
		if !ok {
			return v
		}
		v = obj.Get("query")
	}
	return v
}

// wrapperObject returns the body object of the named wrapper.
func wrapperObject(m *decode.Message, name string) *tljson.Object {
	v := m.Body
	for {
		obj, ok := v.(*tljson.Object)
		if !ok {
			return nil
		}
		if obj.Name() == name {
			return obj
		}
		v = obj.Get("query")
	}
}

// flatten collects scalar leaves of v by dotted path, skipping volatile keys
// and wrapped queries.
func flatten(prefix string, v any, out map[string]string) {
	switch v := v.(type) {
	case nil:
	case *tljson.Object:
		for _, f := range v.Fields {
			if volatileKeys[f.Key] || f.Key == "query" {
				continue
			}
			key := f.Key
			if prefix != "" {
				key = prefix + "." + f.Key
			}
			flatten(key, f.Value, out)
		}
	case []any:
		out[prefix+"[]"] = fmt.Sprintf("len=%d", len(v))
		if len(v) > 0 {
			if _, ok := v[0].(*tljson.Object); ok {
				flatten(prefix+"[]", v[0], out)
			}
		}
	default:
		s := fmt.Sprint(v)
		if len(s) > 32 {
			s = fmt.Sprintf("<%d chars>", len(s))
		}
		out[prefix] = s
	}
}

func (s *traceShape) addOut(t float64, items []*decode.Message) {
	var tokens, methods []string
	fields := map[string]map[string]string{}
	for _, m := range items {
		method := m.Method
		if method == "" {
			method = m.Type
		}
		methods = append(methods, method)
		s.methods[method]++
		init := false
		for _, w := range m.Wrap {
			if w == "initConnection" {
				init = true
			}
		}
		switch {
		case method == "auth.bindTempAuthKey":
			s.advance(phaseBind)
		case init:
			s.advance(phaseInit)
			s.initMsgs[m.MsgID] = true
		}
		if strings.HasPrefix(method, "auth.") && method != "auth.bindTempAuthKey" && s.cur >= phaseInit {
			s.advance(phaseAuth)
			s.authMsgs[m.MsgID] = true
			s.authAwait++
		}
		if method == "ping_delay_disconnect" {
			s.pings = append(s.pings, t)
		}
		tok := messageToken(m)
		tokens = append(tokens, tok)
		f := map[string]string{}
		flatten("", innermost(m), f)
		if init {
			flatten("initConnection", wrapperObject(m, "initConnection"), f)
		}
		fields[tok] = f
	}
	token := strings.Join(tokens, ", ")
	if len(items) > 1 || (len(items) == 1 && items[0].Container != "") {
		token = "container[" + token + "]"
	}
	p := s.phases[s.cur]
	if p.start < 0 {
		p.start = t
	}
	p.msgs = append(p.msgs, outMsg{t: t, token: token, methods: methods, fields: fields})
}

func (s *traceShape) addIn(m *decode.Message) {
	if m.Type != "rpc_result" {
		return
	}
	if s.initMsgs[m.ReqMsgID] && s.cur == phaseInit {
		s.advance(phaseBootstrap)
	}
	if s.authMsgs[m.ReqMsgID] {
		delete(s.authMsgs, m.ReqMsgID)
		s.authAwait--
		if s.authAwait <= 0 && m.Error == nil {
			s.advance(phaseBootstrap)
		}
	}
	if (m.Method == "updates.getState" || m.Method == "updates.getDifference") && m.Error == nil {
		s.advance(phaseSteady)
	}
}

func collectShape(ctx context.Context, dir, schemaFlag string, stderr io.Writer) (*traceShape, error) {
	run, err := openRun(dir)
	if err != nil {
		return nil, err
	}
	defer func() { _ = run.Close() }()
	schema, err := loadSchema(stderr, schemaFlag, run, true)
	if err != nil {
		return nil, err
	}
	s := newTraceShape()
	var (
		bad     badLines
		group   []*decode.Message
		groupT  float64
		groupID uint64
	)
	flush := func() {
		if len(group) > 0 {
			s.addOut(groupT, group)
		}
		group = nil
	}
	d := decode.New(schema, decode.Options{MaxBytes: 64})
	if err := logicalMessages(ctx, run, d, &bad, func(ev *wiretrace.Event, m *decode.Message) error {
		if m == nil {
			if ev.Kind == wiretrace.KindConnRestart {
				s.restarts = append(s.restarts, ev.Int("retry_ms"))
			}
			return nil
		}
		if m.Dir != wiretrace.DirOut {
			flush()
			s.addIn(m)
			return nil
		}
		if ev.Seq != groupID || len(group) == 0 {
			flush()
			groupID, groupT = ev.Seq, ev.T
		}
		c := *m
		group = append(group, &c)
		return nil
	}); err != nil {
		return nil, err
	}
	flush()
	bad.report(stderr)
	return s, nil
}

type diffOp struct {
	kind byte // '=', '-' (only in A), '+' (only in B)
	a, b *outMsg
}

// lcs aligns two message sequences by token.
func lcs(a, b []outMsg) []diffOp {
	n, m := len(a), len(b)
	dp := make([][]int32, n+1)
	for i := range dp {
		dp[i] = make([]int32, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			switch {
			case a[i].token == b[j].token:
				dp[i][j] = dp[i+1][j+1] + 1
			case dp[i+1][j] >= dp[i][j+1]:
				dp[i][j] = dp[i+1][j]
			default:
				dp[i][j] = dp[i][j+1]
			}
		}
	}
	var ops []diffOp
	i, j := 0, 0
	for i < n && j < m {
		switch {
		case a[i].token == b[j].token:
			ops = append(ops, diffOp{kind: '=', a: &a[i], b: &b[j]})
			i++
			j++
		case dp[i+1][j] >= dp[i][j+1]:
			ops = append(ops, diffOp{kind: '-', a: &a[i]})
			i++
		default:
			ops = append(ops, diffOp{kind: '+', b: &b[j]})
			j++
		}
	}
	for ; i < n; i++ {
		ops = append(ops, diffOp{kind: '-', a: &a[i]})
	}
	for ; j < m; j++ {
		ops = append(ops, diffOp{kind: '+', b: &b[j]})
	}
	return ops
}

// fieldDiff reports field differences of two aligned messages.
func fieldDiff(a, b *outMsg) []string {
	var out []string
	for tok, fa := range a.fields {
		fb, ok := b.fields[tok]
		if !ok {
			continue
		}
		keys := map[string]bool{}
		for k := range fa {
			keys[k] = true
		}
		for k := range fb {
			keys[k] = true
		}
		for k := range keys {
			va, inA := fa[k]
			vb, inB := fb[k]
			switch {
			case !inB:
				out = append(out, fmt.Sprintf("%s: -%s=%s", tok, k, va))
			case !inA:
				out = append(out, fmt.Sprintf("%s: +%s=%s", tok, k, vb))
			case va != vb:
				out = append(out, fmt.Sprintf("%s: ~%s %s → %s", tok, k, va, vb))
			}
		}
	}
	sort.Strings(out)
	return out
}

func median(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	s := append([]float64(nil), v...)
	sort.Float64s(s)
	return s[len(s)/2]
}

func intervals(t []float64) []float64 {
	var out []float64
	for i := 1; i < len(t); i++ {
		out = append(out, t[i]-t[i-1])
	}
	return out
}

func phaseDuration(p *phase) float64 {
	if len(p.msgs) == 0 {
		return 0
	}
	return p.msgs[len(p.msgs)-1].t - p.start
}

type diffPrinter struct {
	w      io.Writer
	timing bool
	fields bool
}

func (p *diffPrinter) phase(name string, pa, pb *phase) {
	if len(pa.msgs) == 0 && len(pb.msgs) == 0 {
		return
	}
	_, _ = fmt.Fprintf(p.w, "== phase %s: A %d msgs, B %d msgs\n", name, len(pa.msgs), len(pb.msgs))
	ma, mb := pa.msgs, pb.msgs
	if len(ma) > maxLCS || len(mb) > maxLCS {
		_, _ = fmt.Fprintf(p.w, "  (only the first %d messages of each side are compared)\n", maxLCS)
		ma, mb = ma[:min(len(ma), maxLCS)], mb[:min(len(mb), maxLCS)]
	}
	ops := lcs(ma, mb)
	onlyA, onlyB := map[string]int{}, map[string]int{}
	for k := 0; k < len(ops); k++ {
		op := ops[k]
		diffs := []string(nil)
		if op.kind == '=' && p.fields {
			diffs = fieldDiff(op.a, op.b)
		}
		// Collapse repeated identical lines.
		run := 1
		for !p.timing && len(diffs) == 0 && k+run < len(ops) && ops[k+run].kind == op.kind &&
			ops[k+run].token() == op.token() &&
			(op.kind != '=' || !p.fields || len(fieldDiff(ops[k+run].a, ops[k+run].b)) == 0) {
			run++
		}
		line := fmt.Sprintf("  %c %s", op.kind, op.token())
		if run > 1 {
			line += fmt.Sprintf(" ×%d", run)
		}
		if p.timing {
			switch op.kind {
			case '=':
				line += fmt.Sprintf("   A+%.3fs B+%.3fs", op.a.t-pa.start, op.b.t-pb.start)
			case '-':
				line += fmt.Sprintf("   A+%.3fs", op.a.t-pa.start)
			case '+':
				line += fmt.Sprintf("   B+%.3fs", op.b.t-pb.start)
			}
		}
		_, _ = fmt.Fprintln(p.w, line)
		for _, d := range diffs {
			_, _ = fmt.Fprintf(p.w, "      %s\n", d)
		}
		for r := 0; r < run; r++ {
			switch o := ops[k+r]; o.kind {
			case '-':
				for _, m := range o.a.methods {
					onlyA[m]++
				}
			case '+':
				for _, m := range o.b.methods {
					onlyB[m]++
				}
			}
		}
		k += run - 1
	}
	var moved []string
	for m := range onlyA {
		if onlyB[m] > 0 {
			moved = append(moved, m)
		}
	}
	sort.Strings(moved)
	if len(moved) > 0 {
		_, _ = fmt.Fprintf(p.w, "  reordered or reshaped: %s\n", strings.Join(moved, ", "))
	}
	if p.timing {
		_, _ = fmt.Fprintf(p.w, "  duration: A %.3fs, B %.3fs\n", phaseDuration(pa), phaseDuration(pb))
	}
}

func (op diffOp) token() string {
	if op.a != nil {
		return op.a.token
	}
	return op.b.token
}

func printOnly(w io.Writer, title string, x, y map[string]int) {
	var only []string
	for m, n := range x {
		if y[m] == 0 {
			only = append(only, fmt.Sprintf("%s×%d", m, n))
		}
	}
	if len(only) == 0 {
		return
	}
	sort.Strings(only)
	_, _ = fmt.Fprintf(w, "\n%s:\n  %s\n", title, strings.Join(only, "\n  "))
}

func printDiff(w io.Writer, a, b *traceShape, timing, fields bool) {
	p := &diffPrinter{w: w, timing: timing, fields: fields}
	for i, name := range phaseNames {
		p.phase(name, a.phases[i], b.phases[i])
	}
	printOnly(w, "methods only in A", a.methods, b.methods)
	printOnly(w, "methods only in B", b.methods, a.methods)
	if timing {
		_, _ = fmt.Fprintf(w, "\nping_delay_disconnect interval (median): A %.1fs (%d sent), B %.1fs (%d sent)\n",
			median(intervals(a.pings)), len(a.pings), median(intervals(b.pings)), len(b.pings))
		_, _ = fmt.Fprintf(w, "conn.restart retry_ms: A %v, B %v\n", a.restarts, b.restarts)
	}
}

func runDiff(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("diff", flag.ContinueOnError)
	schemaDir := fs.String("schema", "", "schema of A: directory with api.tl and mtproto.tl, or comma-separated .tl files")
	schemaB := fs.String("schema-b", "", "schema of B (default: --schema, or found next to B)")
	timing := fs.Bool("timing", false, "show offsets from the phase start, phase durations, ping cadence and backoff")
	fields := fs.Bool("fields", true, "show field differences of aligned messages (volatile fields masked)")
	if err := parseFlags(fs, args, 2, "diff [flags] A B", stderr); err != nil {
		return err
	}
	sb := *schemaB
	if sb == "" {
		sb = *schemaDir
	}
	a, err := collectShape(ctx, fs.Arg(0), *schemaDir, stderr)
	if err != nil {
		return err
	}
	b, err := collectShape(ctx, fs.Arg(1), sb, stderr)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(stdout)
	printDiff(w, a, b, *timing, *fields)
	return w.Flush()
}
