package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/go-faster/errors"

	"github.com/gotd/td/wiretrace"
	"github.com/gotd/td/wiretrace/decode"
	"github.com/gotd/td/wiretrace/tljson"
	"github.com/gotd/td/wiretrace/uiattr"
)

// uiOptions are the attribution flags shared by timeline, stats and diff.
type uiOptions struct {
	window time.Duration
	infer  bool
}

func addUIFlags(fs *flag.FlagSet) *uiOptions {
	o := &uiOptions{}
	fs.DurationVar(&o.window, "ui-window", uiattr.DefaultAfterWindow,
		`window of weak "after" UI attribution`)
	fs.BoolVar(&o.infer, "ui-infer", false,
		"attribute unstamped main-thread requests to the latest UI action by time")
	return o
}

func (o *uiOptions) attr() uiattr.Options {
	return uiattr.Options{AfterWindow: o.window, Infer: o.infer}
}

// uiMode is the value of timeline --ui[=collapse].
type uiMode string

const (
	uiOff      uiMode = ""
	uiGroup    uiMode = "group"
	uiCollapse uiMode = "collapse"
)

func (m *uiMode) String() string { return string(*m) }

// IsBoolFlag lets --ui be used without a value.
func (m *uiMode) IsBoolFlag() bool { return true }

func (m *uiMode) Set(s string) error {
	switch s {
	case "true", "group":
		*m = uiGroup
	case "false", "":
		*m = uiOff
	case "collapse":
		*m = uiCollapse
	default:
		return errors.Errorf("want --ui or --ui=collapse, got %q", s)
	}
	return nil
}

// buildUI runs the attribution pass over run. With a schema and the run's
// ui.key, hashed peers are joined to the peers of outgoing request bodies.
func buildUI(ctx context.Context, run *wiretrace.Run, schema *tljson.Schema, opt uiattr.Options,
	bad *badLines, stderr io.Writer,
) (*uiattr.Result, error) {
	b := uiattr.NewBuilder(opt)
	var onBad func(int, int, error)
	if bad != nil {
		onBad = bad.add
	}
	if err := run.Events(func(ev *wiretrace.Event) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		b.Add(ev)
		return nil
	}, onBad); err != nil {
		return nil, err
	}
	res := b.Finish()
	if !res.HasUI || schema == nil {
		return res, nil
	}
	raw := run.Header != nil && run.Header.UI != nil && run.Header.UI.RawIDs
	var key wiretrace.IDKey
	if !raw {
		k, err := wiretrace.UIKey(run.Dir)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				_, _ = fmt.Fprintf(stderr, "wiretrace: warning: %v\n", err)
			}
			return res, nil
		}
		key = k
	}
	var peers []uiattr.Peer
	d := decode.New(schema, decode.Options{MaxBytes: 64})
	if err := logicalMessages(ctx, run, d, &badLines{}, func(_ *wiretrace.Event, m *decode.Message) error {
		if m != nil {
			peers = append(peers, uiattr.PeersFrom(m.Body)...)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if raw {
		res.JoinRawPeers(peers)
	} else {
		res.JoinPeers(key, peers)
	}
	return res, nil
}

func formatBytes(n int64) string {
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%d KB", n>>10)
	}
	return fmt.Sprintf("%d B", n)
}

// actionSummary is "[5 req, 41 KB, 182 ms]": requests, bytes, and the time
// from the action start to the last response.
func actionSummary(act *uiattr.Action) string {
	if len(act.Requests) == 0 {
		return ""
	}
	var bytes int64
	last := act.T
	for _, r := range act.Requests {
		bytes += r.Bytes()
		last = max(last, r.T, r.DoneT)
	}
	return fmt.Sprintf("[%d req, %s, %.0f ms]", len(act.Requests), formatBytes(bytes), (last-act.T)*1000)
}

func viaLabel(a uiattr.Attribution) string {
	s := string(a.Via)
	switch a.Via {
	case uiattr.ViaChain:
		s += fmt.Sprintf(" ⇐ req#%d", a.Parent)
	case uiattr.ViaAfter:
		s += fmt.Sprintf(" +%.0fms", a.DT)
		if a.Nested {
			s += " nested"
		}
	case uiattr.ViaNone:
		s = "background"
		if a.Parent != 0 {
			s += fmt.Sprintf(" ⇐ req#%d", a.Parent)
		}
	case uiattr.ViaUpdate:
		if a.Parent != 0 {
			s += fmt.Sprintf(" ⇐ req#%d", a.Parent)
		}
	}
	if a.Inferred {
		s += " (inferred)"
	}
	return s
}

func requestLine(names *namer, r *uiattr.Request) string {
	s := fmt.Sprintf("%12s req#%d %-36s %s", "", r.Req, names.Name(r.Type), viaLabel(r.Attr))
	switch {
	case r.Failed:
		s += fmt.Sprintf("  lat=%.0fms FAIL %s", r.LatMS, r.Error)
	case r.Done:
		s += fmt.Sprintf("  lat=%.0fms", r.LatMS)
	case r.Canceled:
		s += "  canceled"
	}
	return s
}

// methodCounts is "messages.getHistory ×3 (sync), upload.getFile ×40
// (chain)": requests grouped by method and attribution, in order.
func methodCounts(names *namer, reqs []*uiattr.Request, withVia bool) string {
	type key struct {
		method string
		via    uiattr.Via
	}
	var order []key
	counts := map[key]int{}
	for _, r := range reqs {
		k := key{method: names.Name(r.Type)}
		if withVia {
			k.via = r.Attr.Via
		}
		if counts[k] == 0 {
			order = append(order, k)
		}
		counts[k]++
	}
	parts := make([]string, 0, len(order))
	for _, k := range order {
		s := k.method
		if n := counts[k]; n > 1 {
			s += fmt.Sprintf(" ×%d", n)
		}
		if withVia {
			s += " (" + string(k.via) + ")"
		}
		parts = append(parts, s)
	}
	return strings.Join(parts, ", ")
}

type uiItem struct {
	t      float64
	seq    uint64
	action *uiattr.Action
	marker *uiattr.Marker
	view   *wiretrace.Event
	req    *uiattr.Request
}

// printUITimeline prints actions with their requests, markers, view
// changes and background requests in time order.
func printUITimeline(w io.Writer, res *uiattr.Result, names *namer, collapse bool) {
	var items []uiItem
	for _, a := range res.Actions {
		items = append(items, uiItem{t: a.T, seq: a.Seq, action: a})
	}
	for _, m := range res.Markers {
		items = append(items, uiItem{t: m.T, seq: m.Seq, marker: m})
	}
	for _, v := range res.Views {
		items = append(items, uiItem{t: v.T, seq: v.Seq, view: v})
	}
	for _, r := range res.Background() {
		items = append(items, uiItem{t: r.T, seq: r.Seq, req: r})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].t != items[j].t {
			return items[i].t < items[j].t
		}
		return items[i].seq < items[j].seq
	})
	for i := 0; i < len(items); i++ {
		it := items[i]
		switch {
		case it.action != nil:
			a := it.action
			head := fmt.Sprintf("%10.3f ui#%d %s", a.T, a.A, res.Describe(a))
			if collapse && len(a.Requests) > 0 {
				head += " → " + methodCounts(names, a.Requests, true)
			}
			if s := actionSummary(a); s != "" {
				head += "  " + s
			}
			_, _ = fmt.Fprintln(w, head)
			if !collapse {
				for _, r := range a.Requests {
					_, _ = fmt.Fprintln(w, requestLine(names, r))
				}
			}
		case it.marker != nil:
			m := it.marker
			line := fmt.Sprintf("%10.3f marker #%d %s", m.T, m.N, m.Src)
			if m.Label != "" {
				line += fmt.Sprintf(" %q", m.Label)
			}
			_, _ = fmt.Fprintln(w, line)
		case it.view != nil:
			_, _ = fmt.Fprintf(w, "%10.3f view %s\n", it.t, uiattr.SemanticLabel(it.view, res.Peers))
		case it.req != nil:
			j := i
			var group []*uiattr.Request
			for ; j < len(items) && items[j].req != nil; j++ {
				group = append(group, items[j].req)
			}
			i = j - 1
			printBackground(w, names, group, collapse)
		}
	}
}

func printBackground(w io.Writer, names *namer, group []*uiattr.Request, collapse bool) {
	if !collapse {
		_, _ = fmt.Fprintf(w, "%10.3f [background]\n", group[0].T)
		for _, r := range group {
			_, _ = fmt.Fprintln(w, requestLine(names, r))
		}
		return
	}
	var bg, upd []*uiattr.Request
	for _, r := range group {
		if r.Attr.Via == uiattr.ViaUpdate {
			upd = append(upd, r)
		} else {
			bg = append(bg, r)
		}
	}
	for _, g := range []struct {
		label string
		reqs  []*uiattr.Request
	}{{"[background]", bg}, {"[update]", upd}} {
		if len(g.reqs) > 0 {
			_, _ = fmt.Fprintf(w, "%10.3f %s %s\n", g.reqs[0].T, g.label, methodCounts(names, g.reqs, false))
		}
	}
}

// UIShare is the share of requests per attribution kind.
type UIShare struct {
	Via      string  `json:"via"`
	Requests int     `json:"requests"`
	Bytes    int64   `json:"bytes"`
	Percent  float64 `json:"percent"`
}

// UIActionStats are per action-token counters.
type UIActionStats struct {
	Key      string         `json:"key"`
	Kind     string         `json:"kind"`
	Actions  int            `json:"actions"`
	Requests int            `json:"requests"`
	ReqMin   int            `json:"req_per_action_min"`
	ReqP50   int            `json:"req_per_action_p50"`
	ReqMax   int            `json:"req_per_action_max"`
	Bytes    int64          `json:"bytes"`
	P50      float64        `json:"p50_ms"`
	P90      float64        `json:"p90_ms"`
	Methods  map[string]int `json:"methods,omitempty"`

	perAction []int
	lat       []float64
}

// UIViewStats count requests per view.
type UIViewStats struct {
	View       string `json:"view"`
	Requests   int    `json:"requests"`
	UserDriven int    `json:"user_driven"`
	Bytes      int64  `json:"bytes"`
}

// UIStats is the "ui" block of stats --by ui.
type UIStats struct {
	Actions  int              `json:"actions"`
	Markers  int              `json:"markers"`
	Requests int              `json:"requests"`
	Share    []*UIShare       `json:"share"`
	ByAction []*UIActionStats `json:"by_action"`
	ByView   []*UIViewStats   `json:"by_view"`
}

func uiStats(res *uiattr.Result, names *namer) *UIStats {
	s := &UIStats{Actions: len(res.Actions), Markers: len(res.Markers), Requests: len(res.Requests)}
	share := map[uiattr.Via]*UIShare{}
	for _, v := range uiattr.Vias {
		sh := &UIShare{Via: string(v)}
		share[v] = sh
		s.Share = append(s.Share, sh)
	}
	views := map[string]*UIViewStats{}
	for _, r := range res.Requests {
		sh := share[r.Attr.Via]
		sh.Requests++
		sh.Bytes += r.Bytes()
		label := r.View.Label()
		v, ok := views[label]
		if !ok {
			v = &UIViewStats{View: label}
			views[label] = v
		}
		v.Requests++
		v.Bytes += r.Bytes()
		if r.Attr.UserDriven() {
			v.UserDriven++
		}
	}
	for _, sh := range s.Share {
		if s.Requests > 0 {
			sh.Percent = 100 * float64(sh.Requests) / float64(s.Requests)
		}
	}
	actions := map[string]*UIActionStats{}
	for _, a := range res.Actions {
		key := a.Token()
		st, ok := actions[key]
		if !ok {
			st = &UIActionStats{Key: key, Kind: a.Kind, Methods: map[string]int{}}
			actions[key] = st
		}
		st.Actions++
		st.Requests += len(a.Requests)
		st.perAction = append(st.perAction, len(a.Requests))
		for _, r := range a.Requests {
			st.Bytes += r.Bytes()
			st.Methods[names.Name(r.Type)]++
			if r.Done {
				st.lat = append(st.lat, r.LatMS)
			}
		}
	}
	for _, st := range actions {
		sort.Ints(st.perAction)
		st.ReqMin, st.ReqMax = st.perAction[0], st.perAction[len(st.perAction)-1]
		st.ReqP50 = st.perAction[(len(st.perAction)-1)/2]
		sort.Float64s(st.lat)
		st.P50, st.P90 = percentile(st.lat, 0.5), percentile(st.lat, 0.9)
		s.ByAction = append(s.ByAction, st)
	}
	sort.Slice(s.ByAction, func(i, j int) bool {
		a, b := s.ByAction[i], s.ByAction[j]
		if a.Requests != b.Requests {
			return a.Requests > b.Requests
		}
		return a.Key < b.Key
	})
	for _, v := range views {
		s.ByView = append(s.ByView, v)
	}
	sort.Slice(s.ByView, func(i, j int) bool {
		a, b := s.ByView[i], s.ByView[j]
		if a.Requests != b.Requests {
			return a.Requests > b.Requests
		}
		return a.View < b.View
	})
	return s
}

func printUIStats(w io.Writer, s *UIStats, top int) {
	_, _ = fmt.Fprintf(w, "\nui: %d actions, %d markers, %d requests\n", s.Actions, s.Markers, s.Requests)
	_, _ = fmt.Fprintf(w, "\nrequests by attribution:\n")
	for _, sh := range s.Share {
		_, _ = fmt.Fprintf(w, "  %8d %6.1f%% %10s  %s\n", sh.Requests, sh.Percent, formatBytes(sh.Bytes), sh.Via)
	}
	if len(s.ByAction) > 0 {
		_, _ = fmt.Fprintf(w, "\nui actions (requests per action min/p50/max, latency ms):\n")
		_, _ = fmt.Fprintf(w, "  %6s %6s %11s %10s %9s %9s  %s\n", "count", "req", "min/p50/max", "bytes", "p50", "p90", "action")
		for i, a := range s.ByAction {
			if top > 0 && i >= top {
				_, _ = fmt.Fprintf(w, "  … %d more\n", len(s.ByAction)-top)
				break
			}
			_, _ = fmt.Fprintf(w, "  %6d %6d %11s %10s %9.1f %9.1f  %s\n", a.Actions, a.Requests,
				fmt.Sprintf("%d/%d/%d", a.ReqMin, a.ReqP50, a.ReqMax), formatBytes(a.Bytes), a.P50, a.P90, a.Key)
		}
	}
	if len(s.ByView) > 0 {
		_, _ = fmt.Fprintf(w, "\nrequests by view:\n")
		_, _ = fmt.Fprintf(w, "  %8s %8s %10s  %s\n", "req", "user", "bytes", "view")
		for _, v := range s.ByView {
			_, _ = fmt.Fprintf(w, "  %8d %8d %10s  %s\n", v.Requests, v.UserDriven, formatBytes(v.Bytes), v.View)
		}
	}
}
