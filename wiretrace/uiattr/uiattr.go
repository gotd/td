// Package uiattr attributes requests of a wire trace to UI actions (P14):
// clicks, keys, scrolls, typing, shortcuts and menu items of the traced
// tdesktop build, and synthetic actions of the app.
//
// A first pass (Builder.Add) collects actions, markers, view state and
// requests; Builder.Finish resolves the attribution of every req.new with
// this precedence:
//
//  1. sync: the request has a "ui" stamp with rel "sync" (emitted inside the
//     input scope). Exact.
//  2. chain: the request has a "cause" with req R: it gets the attribution
//     of R (memoised, depth limited), with Parent R and Root the first
//     request of the chain. cause.req 0 (an update handler) means
//     update-driven. A chain whose root is update-driven or background
//     stays so: ViaChain always leads to an action.
//  3. deferred: a "deferred" stamp within Options.DeferredWindow.
//  4. after: an "after" stamp (or a deferred one beyond the deferred window)
//     within Options.AfterWindow. Weak. Stamps written inside a nested
//     event loop (a scope open for more than 500 ms, dt counted from the
//     scope open) are "after" stamps whose dt points back at an input
//     record of the action; they are kept beyond the window, as Nested.
//  5. none: background traffic (timers, pings, config, getDifference after
//     reconnect).
//
// Requests without a stamp and without a cause are background, unless
// Options.Infer is set: then main-thread requests are attributed by time to
// the latest action (deferred or after windows, from the last event of that
// action), marked Inferred.
//
// Traces without UI events resolve every request to update or none.
package uiattr

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gotd/td/wiretrace"
)

// Via is how a request was attributed.
type Via string

// Attribution kinds, in precedence order.
const (
	ViaSync     Via = "sync"
	ViaChain    Via = "chain"
	ViaDeferred Via = "deferred"
	ViaAfter    Via = "after"
	ViaUpdate   Via = "update"
	ViaNone     Via = "none"
)

// Vias lists attribution kinds in precedence order.
var Vias = []Via{ViaSync, ViaChain, ViaDeferred, ViaAfter, ViaUpdate, ViaNone}

// Defaults of Options.
const (
	DefaultDeferredWindow = 150 * time.Millisecond
	DefaultAfterWindow    = 3 * time.Second
	DefaultMaxChain       = 16
)

// Options configures attribution.
type Options struct {
	// DeferredWindow bounds "deferred" stamps (default 150 ms, the emitter
	// kDeferredMs).
	DeferredWindow time.Duration
	// AfterWindow bounds "after" stamps (default 3 s, --ui-window).
	AfterWindow time.Duration
	// MaxChain limits cause chains (default 16).
	MaxChain int
	// Infer attributes unstamped main-thread requests without a cause by
	// time.
	Infer bool
}

func (o *Options) setDefaults() {
	if o.DeferredWindow <= 0 {
		o.DeferredWindow = DefaultDeferredWindow
	}
	if o.AfterWindow <= 0 {
		o.AfterWindow = DefaultAfterWindow
	}
	if o.MaxChain <= 0 {
		o.MaxChain = DefaultMaxChain
	}
}

// Action kinds.
const (
	ActionPress    = "press"
	ActionClick    = "click"
	ActionKey      = "key"
	ActionScroll   = "scroll"
	ActionType     = "type"
	ActionShortcut = "shortcut"
	ActionMenu     = "menu"
	// ActionApp is a synthetic action: a semantic event with no input
	// parent ("src":"app").
	ActionApp = "app"
	// ActionUnknown is an action referenced by stamps whose input record is
	// missing (e.g. in a deleted segment).
	ActionUnknown = "unknown"
)

// Action is one user (or synthetic app) action.
type Action struct {
	A    uint32
	Kind string
	// Src is "app" for synthetic actions.
	Src string
	// T is the time of the first record of the action; End is the time of
	// its last record (including stamped events).
	T, End float64
	Seq    uint64
	W      int64
	Path   []string
	// Handler is the click handler type, if any.
	Handler string
	// Events are the input records of the action (ui.press, ui.click,
	// ui.scroll, ui.scroll.end, …).
	Events []*wiretrace.Event
	// Semantic are the semantic and view events the action caused
	// (ui.chat, ui.media.open, …), in order.
	Semantic []*wiretrace.Event
	// Requests are the requests attributed to the action, in order.
	Requests []*Request
}

// Marker is a "ui.marker" record.
type Marker struct {
	N         int64
	T         float64
	Seq       uint64
	Src       string
	Label     string
	LabelUser bool
}

// Token is the alignment token of a marker: marker(<label>), or marker(#n)
// without a label.
func (m *Marker) Token() string {
	if m.Label != "" {
		return "marker(" + m.Label + ")"
	}
	return "marker(#" + strconv.FormatInt(m.N, 10) + ")"
}

// Attribution is the resolved UI attribution of a request.
type Attribution struct {
	// Action is nil for update-driven and background requests.
	Action *Action
	Via    Via
	// Parent is the direct cause request of a chained request; Root is the
	// first request of the chain.
	Parent, Root int64
	// DT is the stamp (or inferred) delay after the action, in ms.
	DT float64
	// Inferred is set for attributions by time (Options.Infer).
	Inferred bool
	// Nested is set for "after" stamps written inside a nested event loop
	// of the action's input scope.
	Nested bool
}

// UserDriven reports whether the request is attributed to an action.
func (a Attribution) UserDriven() bool { return a.Action != nil }

// Request is one req.new with its lifecycle.
type Request struct {
	Req    int64
	Seq    uint64
	T      float64
	Thread string
	Type   uint32
	Size   int64
	Cause  *wiretrace.Cause
	Stamp  *wiretrace.UIStamp
	Caller string

	Done, Failed, Canceled bool
	DoneT                  float64
	LatMS                  float64
	RespLen                int64
	Error                  string

	// View is the view state when the request was created.
	View View
	Attr Attribution

	state int8 // 0 unresolved, 1 resolving, 2 resolved
}

// Bytes is the request size plus the response length.
func (r *Request) Bytes() int64 { return r.Size + r.RespLen }

// Result is the output of Builder.Finish.
type Result struct {
	// Actions are sorted by time.
	Actions []*Action
	ByID    map[uint32]*Action
	Markers []*Marker
	// Requests are in req.new order.
	Requests []*Request
	ByReq    map[int64]*Request
	// Views are view-state events without an action (app.state, ui.box
	// from Show, …), in order.
	Views []*wiretrace.Event
	// HasUI is set when the trace has any UI record (hdr.ui, ui.*, app.*).
	HasUI bool
	// Peers maps hashed peer ids to "user#123" style names, see JoinPeers.
	Peers map[string]string
}

// Builder collects events of one trace.
type Builder struct {
	opt  Options
	res  *Result
	view viewTracker
}

// NewBuilder creates a Builder.
func NewBuilder(opt Options) *Builder {
	opt.setDefaults()
	return &Builder{
		opt: opt,
		res: &Result{
			ByID:  map[uint32]*Action{},
			ByReq: map[int64]*Request{},
			Peers: map[string]string{},
		},
		view: newViewTracker(),
	}
}

// Build reads all events of run.
func Build(run *wiretrace.Run, opt Options) (*Result, error) {
	b := NewBuilder(opt)
	if err := run.Events(func(ev *wiretrace.Event) error {
		b.Add(ev)
		return nil
	}, nil); err != nil {
		return nil, err
	}
	return b.Finish(), nil
}

var inputKinds = map[string]string{
	wiretrace.KindUIPress:     ActionPress,
	wiretrace.KindUIClick:     ActionClick,
	wiretrace.KindUIKey:       ActionKey,
	wiretrace.KindUIKeyEnd:    ActionKey,
	wiretrace.KindUIScroll:    ActionScroll,
	wiretrace.KindUIScrollEnd: ActionScroll,
	wiretrace.KindUIType:      ActionType,
	wiretrace.KindUITypeEnd:   ActionType,
	wiretrace.KindUIShortcut:  ActionShortcut,
	wiretrace.KindUIMenu:      ActionMenu,
}

// IsUIKind reports whether kind is a UI record (ui.* or app.*).
func IsUIKind(kind string) bool {
	return strings.HasPrefix(kind, "ui.") || strings.HasPrefix(kind, "app.")
}

// action returns the action a, creating a stub if needed.
func (b *Builder) action(a uint32, t float64, seq uint64) *Action {
	act, ok := b.res.ByID[a]
	if !ok {
		act = &Action{A: a, Kind: ActionUnknown, T: t, End: t, Seq: seq}
		b.res.ByID[a] = act
		b.res.Actions = append(b.res.Actions, act)
	}
	if t < act.T {
		act.T, act.Seq = t, seq
	}
	if t > act.End {
		act.End = t
	}
	return act
}

func strList(ev *wiretrace.Event, key string) []string {
	var v []string
	if !ev.Decode(key, &v) {
		return nil
	}
	return v
}

// Add adds one event. Events must be added in trace order.
func (b *Builder) Add(ev *wiretrace.Event) {
	if ev.Kind == wiretrace.KindHeader {
		if h, err := ev.Header(); err == nil && h.UI != nil {
			b.res.HasUI = true
		}
		return
	}
	if IsUIKind(ev.Kind) {
		b.res.HasUI = true
		b.addUI(ev)
	} else if ev.UI != nil {
		b.action(ev.UI.A, ev.T, ev.Seq)
	}
	switch ev.Kind {
	case wiretrace.KindReqNew:
		r := &Request{
			Req:    ev.Int("req"),
			Seq:    ev.Seq,
			T:      ev.T,
			Thread: ev.Thread,
			Type:   ev.Type("type"),
			Size:   ev.Int("size"),
			Cause:  ev.Cause,
			Stamp:  ev.UI,
			View:   b.view.snapshot(),
		}
		if _, dup := b.res.ByReq[r.Req]; !dup {
			b.res.ByReq[r.Req] = r
			b.res.Requests = append(b.res.Requests, r)
		}
	case wiretrace.KindReqCaller:
		if r := b.res.ByReq[ev.Int("req")]; r != nil {
			r.Caller = ev.Str("file") + ":" + strconv.FormatInt(ev.Int("line"), 10)
		}
	case wiretrace.KindReqDone, wiretrace.KindReqFail:
		if r := b.res.ByReq[ev.Int("req")]; r != nil {
			r.Done = true
			r.Failed = ev.Kind == wiretrace.KindReqFail
			r.DoneT = ev.T
			r.LatMS = ev.Float("lat_ms")
			r.RespLen = ev.Int("len")
			r.Error = ev.Str("error")
		}
	case wiretrace.KindReqCancel:
		if r := b.res.ByReq[ev.Int("req")]; r != nil {
			r.Canceled = true
		}
	}
}

func (b *Builder) addUI(ev *wiretrace.Event) {
	if ev.Kind == wiretrace.KindUIMarker {
		b.res.Markers = append(b.res.Markers, &Marker{
			N:         ev.Int("n"),
			T:         ev.T,
			Seq:       ev.Seq,
			Src:       ev.Str("src"),
			Label:     ev.Str("label"),
			LabelUser: ev.Bool("label_user"),
		})
		return
	}
	b.view.add(ev)
	if kind, ok := inputKinds[ev.Kind]; ok && ev.Has("a") {
		act := b.action(uint32(ev.Int("a")), ev.T, ev.Seq) // #nosec G115
		b.fillInput(act, kind, ev)
		return
	}
	switch {
	case ev.UI != nil:
		act := b.action(ev.UI.A, ev.T, ev.Seq)
		act.Semantic = append(act.Semantic, ev)
	case ev.Has("a") && ev.Str("src") == ActionApp:
		act := b.action(uint32(ev.Int("a")), ev.T, ev.Seq) // #nosec G115
		if act.Kind == ActionUnknown {
			act.Kind, act.Src = ActionApp, ActionApp
			act.W = ev.Int("w")
		}
		act.Semantic = append(act.Semantic, ev)
	default:
		b.res.Views = append(b.res.Views, ev)
	}
}

func (b *Builder) fillInput(act *Action, kind string, ev *wiretrace.Event) {
	act.Events = append(act.Events, ev)
	// ui.click supersedes the ui.press of the same action.
	if act.Kind == ActionUnknown || act.Kind == ActionApp || (act.Kind == ActionPress && kind == ActionClick) {
		act.Kind, act.Src = kind, ""
	}
	if ev.Has("w") {
		act.W = ev.Int("w")
	}
	if p := strList(ev, "path"); len(p) > 0 {
		act.Path = p
	} else if p := strList(ev, "field"); len(p) > 0 {
		act.Path = p
	}
	if h := ev.Str("handler"); h != "" {
		act.Handler = h
	}
}

// Finish resolves attributions. The Builder must not be used after it.
func (b *Builder) Finish() *Result {
	res := b.res
	sortActions(res.Actions)
	for _, r := range res.Requests {
		b.resolve(r, 0)
	}
	sortActions(res.Actions)
	for _, r := range res.Requests {
		if act := r.Attr.Action; act != nil {
			act.Requests = append(act.Requests, r)
		}
	}
	return res
}

func sortActions(actions []*Action) {
	sort.SliceStable(actions, func(i, j int) bool {
		if actions[i].T != actions[j].T {
			return actions[i].T < actions[j].T
		}
		return actions[i].Seq < actions[j].Seq
	})
}

func (b *Builder) resolve(r *Request, depth int) Attribution {
	switch r.state {
	case 2:
		return r.Attr
	case 1:
		// A cause cycle.
		return Attribution{Via: ViaNone}
	}
	if depth > b.opt.MaxChain {
		return Attribution{Via: ViaNone}
	}
	r.state = 1
	r.Attr = b.direct(r, depth)
	r.state = 2
	return r.Attr
}

func (b *Builder) stampAction(st *wiretrace.UIStamp, r *Request) *Action {
	return b.action(st.A, r.T, r.Seq)
}

func (b *Builder) direct(r *Request, depth int) Attribution {
	st := r.Stamp
	// 1. sync.
	if st != nil && st.Rel == wiretrace.UIRelSync {
		return Attribution{Action: b.stampAction(st, r), Via: ViaSync}
	}
	// 2. chain (and update-driven).
	if c := r.Cause; c != nil {
		if c.Req == 0 {
			return Attribution{Via: ViaUpdate}
		}
		parent, ok := b.res.ByReq[c.Req]
		if !ok || parent == r {
			return Attribution{Via: ViaNone, Parent: c.Req, Root: c.Req}
		}
		pa := b.resolve(parent, depth+1)
		out := Attribution{Action: pa.Action, Via: pa.Via, Parent: c.Req, Root: pa.Root, Inferred: pa.Inferred}
		if out.Root == 0 {
			out.Root = c.Req
		}
		if out.Action != nil {
			out.Via = ViaChain
		}
		return out
	}
	deferredMS := float64(b.opt.DeferredWindow) / float64(time.Millisecond)
	afterMS := float64(b.opt.AfterWindow) / float64(time.Millisecond)
	if st != nil {
		switch st.Rel {
		case wiretrace.UIRelDeferred:
			// 3. deferred.
			if st.DT <= deferredMS {
				return Attribution{Action: b.stampAction(st, r), Via: ViaDeferred, DT: st.DT}
			}
			if st.DT <= afterMS {
				return Attribution{Action: b.stampAction(st, r), Via: ViaAfter, DT: st.DT}
			}
		case wiretrace.UIRelAfter:
			// 4. after.
			nested := b.nested(st, r, afterMS)
			if st.DT <= afterMS || nested {
				return Attribution{Action: b.stampAction(st, r), Via: ViaAfter, DT: st.DT, Nested: nested}
			}
		}
		return Attribution{Via: ViaNone}
	}
	if b.opt.Infer && r.Thread == "main" {
		if a, ok := b.infer(r, deferredMS, afterMS); ok {
			return a
		}
	}
	// 5. none.
	return Attribution{Via: ViaNone}
}

// NestedMS is the age of an input scope after which the emitter stamps
// events inside it as "after" (a nested event loop).
const NestedMS = 500

// nestedTolerance is the slack between the scope open derived from a
// nested stamp (event time - dt, dt rounded to ms) and the input record.
const nestedTolerance = 0.002

// nested reports whether an "after" stamp was written inside a nested event
// loop of the action's scope. Both kinds of "after" stamps have dt from an
// input record of the action (scope open or close), so a nested one is
// told apart only when dt > NestedMS, event time - dt is the time of an
// input record of the action, and either dt is beyond the after window or
// a newer action started before the event (input inside the modal loop;
// a plain after stamp always refers to the latest action).
func (b *Builder) nested(st *wiretrace.UIStamp, r *Request, afterMS float64) bool {
	if st.DT <= NestedMS {
		return false
	}
	act, ok := b.res.ByID[st.A]
	if !ok {
		return false
	}
	open := r.T - st.DT/1000
	match := false
	for _, ev := range act.Events {
		if d := ev.T - open; d >= -nestedTolerance && d <= nestedTolerance {
			match = true
			break
		}
	}
	if !match {
		return false
	}
	if st.DT > afterMS {
		return true
	}
	// Actions are sorted by time during resolution.
	actions := b.res.Actions
	i := sort.Search(len(actions), func(i int) bool { return actions[i].T > act.T })
	return i < len(actions) && actions[i].T < r.T
}

// infer attributes r to the latest action that started before it.
func (b *Builder) infer(r *Request, deferredMS, afterMS float64) (Attribution, bool) {
	actions := b.res.Actions
	i := sort.Search(len(actions), func(i int) bool {
		return actions[i].T > r.T || (actions[i].T == r.T && actions[i].Seq > r.Seq)
	})
	if i == 0 {
		return Attribution{}, false
	}
	last := actions[i-1]
	end := last.T
	for _, ev := range last.Events {
		if ev.T <= r.T && ev.T > end {
			end = ev.T
		}
	}
	dt := (r.T - end) * 1000
	switch {
	case dt <= deferredMS:
		return Attribution{Action: last, Via: ViaDeferred, DT: dt, Inferred: true}, true
	case dt <= afterMS:
		return Attribution{Action: last, Via: ViaAfter, DT: dt, Inferred: true}, true
	}
	return Attribution{}, false
}

// Background returns the requests not attributed to an action.
func (r *Result) Background() []*Request {
	var out []*Request
	for _, req := range r.Requests {
		if !req.Attr.UserDriven() {
			out = append(out, req)
		}
	}
	return out
}
