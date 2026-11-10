package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gotd/td/wiretrace"
)

// MethodStats are per-method request counters.
type MethodStats struct {
	Method string  `json:"method"`
	New    int     `json:"new"`
	Done   int     `json:"done"`
	Fail   int     `json:"fail"`
	Cancel int     `json:"cancel,omitempty"`
	Resend int     `json:"resend,omitempty"`
	P50    float64 `json:"p50_ms"`
	P90    float64 `json:"p90_ms"`
	P99    float64 `json:"p99_ms"`
	Max    float64 `json:"max_ms"`

	lat []float64
}

// ErrorStats counts rpc errors by method and normalized error type.
type ErrorStats struct {
	Method string `json:"method"`
	Code   int64  `json:"code"`
	Error  string `json:"error"`
	Count  int    `json:"count"`
}

// FloodStats sums flood waits per method.
type FloodStats struct {
	Method  string `json:"method"`
	Count   int    `json:"count"`
	Seconds int64  `json:"seconds"`
}

// DCStats are per shifted DC counters.
type DCStats struct {
	SDC      int64          `json:"sdc"`
	Sessions int            `json:"sessions"`
	NewIDs   int            `json:"session_new_id"`
	Keys     map[string]int `json:"keys,omitempty"`
	MsgsOut  int            `json:"msgs_out"`
	MsgsIn   int            `json:"msgs_in"`
	BytesOut int64          `json:"bytes_out"`
	BytesIn  int64          `json:"bytes_in"`
	ConnUp   int            `json:"conn_up"`
	ConnDown int            `json:"conn_down"`

	sessions map[int64]struct{}
}

// Stats is the output of the stats command.
type Stats struct {
	Events     int              `json:"events"`
	BadLines   int              `json:"bad_lines,omitempty"`
	Dropped    uint64           `json:"dropped"`
	Segments   []int            `json:"segments"`
	Duration   float64          `json:"duration_s"`
	Kinds      map[string]int   `json:"kinds"`
	Methods    []*MethodStats   `json:"methods"`
	Errors     []*ErrorStats    `json:"errors"`
	Floods     []*FloodStats    `json:"flood_waits"`
	Restarts   map[string]int   `json:"restarts"`
	Reasons    map[string]int   `json:"conn_reasons"`
	DCs        []*DCStats       `json:"dcs"`
	WireOut    map[string]int   `json:"wire_out"`
	WireIn     map[string]int   `json:"wire_in"`
	Storage    map[string]int   `json:"storage"`
	StorageLen map[string]int64 `json:"storage_bytes"`
	Redacted   map[string]int   `json:"redacted"`
	Subsystems map[string]int   `json:"callers"`
}

var (
	digitsRe = regexp.MustCompile(`\d+`)
	floodRe  = regexp.MustCompile(`^(?:FLOOD_WAIT|FLOOD_PREMIUM_WAIT|SLOWMODE_WAIT|FLOOD_TEST_PHONE_WAIT)_(\d+)$`)
)

// normalizeError masks numbers in an error type: FLOOD_WAIT_17 → FLOOD_WAIT_X.
func normalizeError(s string) string {
	return digitsRe.ReplaceAllString(s, "X")
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	i := int(math.Ceil(p*float64(len(sorted)))) - 1
	if i < 0 {
		i = 0
	}
	if i >= len(sorted) {
		i = len(sorted) - 1
	}
	return sorted[i]
}

type statsBuilder struct {
	names   *namer
	s       *Stats
	methods map[string]*MethodStats
	errors  map[string]*ErrorStats
	floods  map[string]*FloodStats
	dcs     map[int64]*DCStats
	reqType map[int64]uint32
	first   float64
	last    float64
	seen    bool
}

func newStatsBuilder(n *namer, segments []int) *statsBuilder {
	return &statsBuilder{
		names: n,
		s: &Stats{
			Segments:   segments,
			Kinds:      map[string]int{},
			Restarts:   map[string]int{},
			Reasons:    map[string]int{},
			WireOut:    map[string]int{},
			WireIn:     map[string]int{},
			Storage:    map[string]int{},
			StorageLen: map[string]int64{},
			Redacted:   map[string]int{},
			Subsystems: map[string]int{},
		},
		methods: map[string]*MethodStats{},
		errors:  map[string]*ErrorStats{},
		floods:  map[string]*FloodStats{},
		dcs:     map[int64]*DCStats{},
		reqType: map[int64]uint32{},
	}
}

func (b *statsBuilder) method(name string) *MethodStats {
	m, ok := b.methods[name]
	if !ok {
		m = &MethodStats{Method: name}
		b.methods[name] = m
	}
	return m
}

func (b *statsBuilder) dc(sdc int64) *DCStats {
	d, ok := b.dcs[sdc]
	if !ok {
		d = &DCStats{SDC: sdc, Keys: map[string]int{}, sessions: map[int64]struct{}{}}
		b.dcs[sdc] = d
	}
	return d
}

func (b *statsBuilder) reqMethod(ev *wiretrace.Event) string {
	if t := ev.Type("method"); t != 0 {
		return b.names.Name(t)
	}
	if t, ok := b.reqType[ev.Int("req")]; ok {
		return b.names.Name(t)
	}
	return "?"
}

func (b *statsBuilder) add(ev *wiretrace.Event) {
	if ev.Kind == wiretrace.KindHeader {
		return
	}
	b.s.Events++
	b.s.Kinds[ev.Kind]++
	if !b.seen || ev.T < b.first {
		b.first = ev.T
	}
	if ev.T > b.last {
		b.last = ev.T
	}
	b.seen = true
	switch ev.Kind {
	case wiretrace.KindDropped:
		b.s.Dropped += ev.ID("count")
	case wiretrace.KindReqNew:
		t := ev.Type("type")
		b.reqType[ev.Int("req")] = t
		b.method(b.names.Name(t)).New++
	case wiretrace.KindReqCaller:
		sub := ev.Str("subsystem")
		if sub == "" {
			sub = "(top level)"
		}
		b.s.Subsystems[sub]++
	case wiretrace.KindReqResend:
		if t, ok := b.reqType[ev.Int("req")]; ok {
			b.method(b.names.Name(t)).Resend++
		}
	case wiretrace.KindReqCancel:
		if t, ok := b.reqType[ev.Int("req")]; ok {
			b.method(b.names.Name(t)).Cancel++
		}
	case wiretrace.KindReqDone, wiretrace.KindReqFail:
		name := b.reqMethod(ev)
		m := b.method(name)
		if ev.Has("lat_ms") {
			m.lat = append(m.lat, ev.Float("lat_ms"))
		}
		if ev.Kind == wiretrace.KindReqDone {
			m.Done++
			break
		}
		m.Fail++
		text := ev.Str("error")
		key := name + "\x00" + normalizeError(text)
		e, ok := b.errors[key]
		if !ok {
			e = &ErrorStats{Method: name, Code: ev.Int("code"), Error: normalizeError(text)}
			b.errors[key] = e
		}
		e.Count++
		if sub := floodRe.FindStringSubmatch(text); sub != nil {
			sec, _ := strconv.ParseInt(sub[1], 10, 64)
			f, ok := b.floods[name]
			if !ok {
				f = &FloodStats{Method: name}
				b.floods[name] = f
			}
			f.Count++
			f.Seconds += sec
		}
	case wiretrace.KindConnRestart:
		b.s.Restarts[ev.Str("reason")]++
	case wiretrace.KindConnReason:
		b.s.Reasons[ev.Str("reason")]++
	case wiretrace.KindConnUp:
		b.dc(ev.Int("sdc")).ConnUp++
	case wiretrace.KindConnDown:
		b.dc(ev.Int("sdc")).ConnDown++
	case wiretrace.KindSessionNewID:
		b.dc(ev.Int("sdc")).NewIDs++
	case wiretrace.KindKeySet:
		b.dc(ev.Int("sdc")).Keys[ev.Str("key_kind")]++
	case wiretrace.KindMsg:
		b.addMsg(ev)
	}
	if strings.HasPrefix(ev.Kind, "st.") {
		k := ev.Kind + " " + storageKind(ev.Str("file"))
		b.s.Storage[k]++
		b.s.StorageLen[k] += ev.Int("len")
	}
}

func (b *statsBuilder) addMsg(ev *wiretrace.Event) {
	m, err := ev.Msg()
	if err != nil {
		return
	}
	d := b.dc(int64(m.SDC))
	d.sessions[int64(m.Session)] = struct{}{}
	for _, r := range m.Redacted {
		b.s.Redacted[r]++
	}
	wire := b.s.WireIn
	if m.Dir == wiretrace.DirOut {
		wire = b.s.WireOut
		d.MsgsOut++
		d.BytesOut += m.Len
	} else {
		d.MsgsIn++
		d.BytesIn += m.Len
	}
	name := func(typ, method, result wiretrace.TypeID) string {
		if m.Dir == wiretrace.DirIn && uint32(typ) == 0xf35c6d01 {
			return "rpc_result:" + b.names.Name(uint32(result))
		}
		return b.names.Name(innerID(method, typ))
	}
	if len(m.Inner) > 0 {
		wire["msg_container"]++
		for _, in := range m.Inner {
			wire[name(in.Type, in.Method, in.Result)]++
		}
		return
	}
	wire[name(m.Type, m.Method, m.Result)]++
}

func (b *statsBuilder) finish() *Stats {
	s := b.s
	if b.seen {
		s.Duration = b.last - b.first
	}
	for _, m := range b.methods {
		sort.Float64s(m.lat)
		m.P50 = percentile(m.lat, 0.5)
		m.P90 = percentile(m.lat, 0.9)
		m.P99 = percentile(m.lat, 0.99)
		if len(m.lat) > 0 {
			m.Max = m.lat[len(m.lat)-1]
		}
		s.Methods = append(s.Methods, m)
	}
	sort.Slice(s.Methods, func(i, j int) bool {
		a, c := s.Methods[i], s.Methods[j]
		if a.New+a.Done != c.New+c.Done {
			return a.New+a.Done > c.New+c.Done
		}
		return a.Method < c.Method
	})
	for _, e := range b.errors {
		s.Errors = append(s.Errors, e)
	}
	sort.Slice(s.Errors, func(i, j int) bool {
		if s.Errors[i].Count != s.Errors[j].Count {
			return s.Errors[i].Count > s.Errors[j].Count
		}
		return s.Errors[i].Method+s.Errors[i].Error < s.Errors[j].Method+s.Errors[j].Error
	})
	for _, f := range b.floods {
		s.Floods = append(s.Floods, f)
	}
	sort.Slice(s.Floods, func(i, j int) bool { return s.Floods[i].Seconds > s.Floods[j].Seconds })
	for _, d := range b.dcs {
		d.Sessions = len(d.sessions)
		s.DCs = append(s.DCs, d)
	}
	sort.Slice(s.DCs, func(i, j int) bool { return s.DCs[i].SDC < s.DCs[j].SDC })
	return s
}

type kv struct {
	k string
	v int64
}

func sortedCounts[V int | int64](m map[string]V) []kv {
	out := make([]kv, 0, len(m))
	for k, v := range m {
		out = append(out, kv{k: k, v: int64(v)})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].v != out[j].v {
			return out[i].v > out[j].v
		}
		return out[i].k < out[j].k
	})
	return out
}

func printCounts[V int | int64](w io.Writer, title string, m map[string]V, top int) {
	if len(m) == 0 {
		return
	}
	_, _ = fmt.Fprintf(w, "\n%s:\n", title)
	for i, e := range sortedCounts(m) {
		if top > 0 && i >= top {
			_, _ = fmt.Fprintf(w, "  … %d more\n", len(m)-top)
			break
		}
		_, _ = fmt.Fprintf(w, "  %8d  %s\n", e.v, e.k)
	}
}

func printStats(w io.Writer, s *Stats, top int) {
	_, _ = fmt.Fprintf(w, "events: %d, duration: %.1fs, segments: %v, dropped: %d", s.Events, s.Duration, s.Segments, s.Dropped)
	if s.BadLines > 0 {
		_, _ = fmt.Fprintf(w, ", unparsable lines: %d", s.BadLines)
	}
	_, _ = fmt.Fprintln(w)
	printCounts(w, "event kinds", s.Kinds, 0)

	if len(s.Methods) > 0 {
		_, _ = fmt.Fprintf(w, "\nrequests (latency from req.new to req.done/fail, ms):\n")
		_, _ = fmt.Fprintf(w, "  %6s %6s %5s %5s %5s %9s %9s %9s %9s  %s\n",
			"new", "done", "fail", "cancl", "resnd", "p50", "p90", "p99", "max", "method")
		for i, m := range s.Methods {
			if top > 0 && i >= top {
				_, _ = fmt.Fprintf(w, "  … %d more\n", len(s.Methods)-top)
				break
			}
			_, _ = fmt.Fprintf(w, "  %6d %6d %5d %5d %5d %9.1f %9.1f %9.1f %9.1f  %s\n",
				m.New, m.Done, m.Fail, m.Cancel, m.Resend, m.P50, m.P90, m.P99, m.Max, m.Method)
		}
	}
	if len(s.Errors) > 0 {
		_, _ = fmt.Fprintf(w, "\nrpc errors (numbers masked):\n")
		for _, e := range s.Errors {
			_, _ = fmt.Fprintf(w, "  %8d  %d %s  %s\n", e.Count, e.Code, e.Error, e.Method)
		}
	}
	if len(s.Floods) > 0 {
		_, _ = fmt.Fprintf(w, "\nflood waits:\n")
		for _, f := range s.Floods {
			_, _ = fmt.Fprintf(w, "  %8d waits %8ds  %s\n", f.Count, f.Seconds, f.Method)
		}
	}
	printCounts(w, "connection restarts by reason", s.Restarts, 0)
	printCounts(w, "connection reasons", s.Reasons, 0)
	if len(s.DCs) > 0 {
		_, _ = fmt.Fprintf(w, "\nper DC (shifted id):\n")
		_, _ = fmt.Fprintf(w, "  %6s %8s %6s %6s %8s %8s %10s %10s  %s\n",
			"sdc", "sessions", "new_id", "up", "out", "in", "out_bytes", "in_bytes", "keys")
		for _, d := range s.DCs {
			var keys []string
			for _, e := range sortedCounts(d.Keys) {
				keys = append(keys, fmt.Sprintf("%s×%d", e.k, e.v))
			}
			_, _ = fmt.Fprintf(w, "  %6d %8d %6d %6d %8d %8d %10d %10d  %s\n",
				d.SDC, d.Sessions, d.NewIDs, d.ConnUp, d.MsgsOut, d.MsgsIn, d.BytesOut, d.BytesIn, strings.Join(keys, " "))
		}
	}
	printCounts(w, "wire out (innermost method per message)", s.WireOut, top)
	printCounts(w, "wire in (constructor per message)", s.WireIn, top)
	printCounts(w, "request callers by subsystem", s.Subsystems, 0)
	printCounts(w, "redactions", s.Redacted, 0)
	printCounts(w, "storage events", s.Storage, top)
}

func runStats(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("stats", flag.ContinueOnError)
	schemaDir := fs.String("schema", "", "directory with api.tl and mtproto.tl, or comma-separated .tl files (optional, for names)")
	asJSON := fs.Bool("json", false, "print JSON")
	top := fs.Int("top", 40, "rows per table (0: all)")
	if err := parseFlags(fs, args, 1, "stats [flags] RUN", stderr); err != nil {
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
	b := newStatsBuilder(newNamer(schema), run.Segments)
	var bad badLines
	if err := run.Events(func(ev *wiretrace.Event) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		b.add(ev)
		return nil
	}, bad.add); err != nil {
		return err
	}
	s := b.finish()
	s.BadLines = bad.count
	w := bufio.NewWriter(stdout)
	if *asJSON {
		e := json.NewEncoder(w)
		e.SetIndent("", "  ")
		if err := e.Encode(s); err != nil {
			return err
		}
	} else {
		printStats(w, s, *top)
	}
	return w.Flush()
}
