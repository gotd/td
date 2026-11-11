package uiattr

import (
	"fmt"
	"strings"

	"github.com/gotd/td/wiretrace"
)

// first returns the first input record of kind.
func (a *Action) first(kind string) *wiretrace.Event {
	for _, ev := range a.Events {
		if ev.Kind == kind {
			return ev
		}
	}
	return nil
}

// Widget is the widget an action is about: for scrolls the owner of the
// Ui::ScrollArea, otherwise the first path element that is not a generic
// Ui:: widget (or the receiver).
func (a *Action) Widget() string {
	p := a.Path
	if len(p) == 0 {
		return ""
	}
	if a.Kind == ActionScroll {
		for i, s := range p {
			if s == "Ui::ScrollArea" && i+1 < len(p) {
				return p[i+1]
			}
		}
	}
	for _, s := range p {
		if !strings.HasPrefix(s, "Ui::") {
			return s
		}
	}
	return p[0]
}

// verb is the kind of a click or press: "dblclick" for double clicks,
// "tap" for touch, "click:right" for other buttons.
func (a *Action) verb() string {
	ev := a.first(wiretrace.KindUIPress)
	if ev == nil {
		ev = a.first(wiretrace.KindUIClick)
	}
	if ev == nil {
		return a.Kind
	}
	switch btn := ev.Str("btn"); {
	case ev.Bool("dbl"):
		return "dbl" + a.Kind
	case btn == "touch" && a.Kind == ActionClick:
		return "tap"
	case btn == "touch":
		return a.Kind + ":touch"
	case btn != "" && btn != "left":
		return a.Kind + ":" + btn
	}
	return a.Kind
}

func (a *Action) target() string {
	if a.Handler != "" {
		return a.Handler
	}
	return a.Widget()
}

func keyName(ev *wiretrace.Event) string {
	key := ev.Str("key")
	if mods := strList(ev, "mods"); len(mods) > 0 {
		key = strings.Join(mods, "+") + "+" + key
	}
	return key
}

func (a *Action) str(kind, key string) string {
	if ev := a.first(kind); ev != nil {
		return ev.Str(key)
	}
	return ""
}

// rawStr returns a string field, or the raw JSON of other values.
func rawStr(ev *wiretrace.Event, key string) string {
	if s := ev.Str(key); s != "" {
		return s
	}
	return string(ev.Fields[key])
}

func join(parts ...string) string {
	var out []string
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, " ")
}

// Label is a short human-readable description of the action itself, e.g.
// "click Dialogs::InnerWidget", "scroll up HistoryWidget ×14 1.2s".
func (a *Action) Label() string {
	switch a.Kind {
	case ActionClick, ActionPress:
		return join(a.verb(), a.target())
	case ActionScroll:
		s := join("scroll", a.str(wiretrace.KindUIScroll, "dir"), a.Widget())
		if end := a.first(wiretrace.KindUIScrollEnd); end != nil {
			if n := end.Int("n"); n > 1 {
				s += fmt.Sprintf(" ×%d", n)
			}
			if ms := end.Float("ms"); ms > 0 {
				s += fmt.Sprintf(" %.1fs", ms/1000)
			}
		}
		return s
	case ActionType:
		s := join("type", a.Widget())
		start := a.first(wiretrace.KindUIType)
		if start != nil && start.Bool("masked") {
			s += " masked"
		}
		if start != nil && start.Bool("search") {
			s += " search"
		}
		if end := a.first(wiretrace.KindUITypeEnd); end != nil {
			if end.Has("n") {
				s += " n=" + rawStr(end, "n")
			}
			if end.Has("count") {
				s += fmt.Sprintf(" count=%d", end.Int("count"))
			}
		}
		return s
	case ActionKey:
		ev := a.first(wiretrace.KindUIKey)
		if ev == nil {
			return "key"
		}
		s := join("key", keyName(ev), a.Widget())
		if ev.Str("via") != "" {
			s += " via " + ev.Str("via")
		}
		if end := a.first(wiretrace.KindUIKeyEnd); end != nil {
			// Auto-repeats after the first press.
			s += fmt.Sprintf(" +%d repeats", end.Int("repeat"))
		}
		return s
	case ActionShortcut:
		ev := a.first(wiretrace.KindUIShortcut)
		if ev == nil {
			return "shortcut"
		}
		if cmds := strList(ev, "cmds"); len(cmds) > 1 {
			return "shortcut " + strings.Join(cmds, ",")
		}
		return join("shortcut", ev.Str("cmd"))
	case ActionMenu:
		ev := a.first(wiretrace.KindUIMenu)
		if ev == nil {
			return "menu"
		}
		if item := ev.Str("item"); item != "" {
			return "menu " + item
		}
		return fmt.Sprintf("menu ? (len %d)", ev.Int("len"))
	}
	return a.Kind
}

// Token is the alignment token of the action: the kind with its stable
// arguments plus the semantic events it caused, e.g.
// "click(Dialogs::InnerWidget)>chat(user)", "scroll(up,HistoryWidget)".
// Hashed ids, counts and times are left out, so tokens match across runs.
func (a *Action) Token() string {
	var s string
	switch a.Kind {
	case ActionClick, ActionPress:
		s = a.verb() + "(" + a.target() + ")"
	case ActionScroll:
		s = "scroll(" + a.str(wiretrace.KindUIScroll, "dir") + "," + a.Widget() + ")"
	case ActionType:
		s = "type(" + a.Widget() + ")"
	case ActionKey:
		key := ""
		if ev := a.first(wiretrace.KindUIKey); ev != nil {
			key = keyName(ev)
		}
		s = "key(" + key + "," + a.Widget() + ")"
	case ActionShortcut:
		s = "shortcut(" + a.str(wiretrace.KindUIShortcut, "cmd") + ")"
	case ActionMenu:
		s = "menu(" + a.str(wiretrace.KindUIMenu, "item") + ")"
	default:
		s = a.Kind
	}
	prev := ""
	for _, ev := range a.Semantic {
		tok := SemanticToken(ev)
		if tok == prev {
			continue
		}
		prev = tok
		s += ">" + tok
	}
	return s
}

func trimKind(kind string) string {
	return strings.TrimPrefix(kind, "ui.")
}

// SemanticToken is the alignment token of a semantic or view event:
// "chat(user)", "media.open(document)", "section(Info::WrapWidget/Settings)",
// "box(PeerListBox)".
func SemanticToken(ev *wiretrace.Event) string {
	name := trimKind(ev.Kind)
	switch ev.Kind {
	case wiretrace.KindUIChat:
		return name + "(" + ev.Str("peer_kind") + ")"
	case wiretrace.KindUIMediaOpen:
		return name + "(" + ev.Str("kind") + ")"
	case wiretrace.KindUISection:
		return name + "(" + sectionLabel(ev) + ")"
	case wiretrace.KindUIBox:
		typ := ev.Str("type")
		if ev.Bool("layer") {
			typ = "layer"
		}
		if !ev.Bool("shown") {
			return "box.hide(" + typ + ")"
		}
		return name + "(" + typ + ")"
	case wiretrace.KindUICall:
		if !ev.Bool("active") {
			return "call.end(" + ev.Str("kind") + ")"
		}
		return name + "(" + ev.Str("kind") + ")"
	case wiretrace.KindUISession:
		if !ev.Bool("open") {
			return "session(closed)"
		}
		return "session(open)"
	case wiretrace.KindUINav:
		return name + "(" + join(ev.Str("type"), ev.Str("dir")) + ")"
	case wiretrace.KindAppWindow:
		return name + "(" + ev.Str("state") + ")"
	}
	return name
}

// SemanticLabel describes a semantic or view event, with hashed peers
// resolved through peers (may be nil): "chat h:5c1e… (user)",
// "media.open document h:0b77…".
func SemanticLabel(ev *wiretrace.Event, peers map[string]string) string {
	name := trimKind(ev.Kind)
	peer := func(h string) string {
		if p, ok := peers[h]; ok {
			return h + "=" + p
		}
		return h
	}
	switch ev.Kind {
	case wiretrace.KindUIChat:
		if ev.Str("peer") == "" {
			return "chat closed"
		}
		s := join(name, peer(ev.Str("peer")), "("+ev.Str("peer_kind")+")")
		if m := ev.Int("msg"); m != 0 {
			s += " msg=" + MsgIDLabel(m)
		}
		if tp := ev.Int("topic"); tp != 0 {
			s += fmt.Sprintf(" topic=%d", tp)
		}
		return s
	case wiretrace.KindUIMediaOpen:
		return join(name, ev.Str("kind"), mediaID(ev))
	case wiretrace.KindUISection:
		return join(name, sectionLabel(ev))
	case wiretrace.KindUIBox:
		typ := ev.Str("type")
		if ev.Bool("layer") {
			typ = "layer"
		}
		if !ev.Bool("shown") {
			return join("box hidden", typ)
		}
		return join(name, typ)
	case wiretrace.KindUICall:
		state := "ended"
		if ev.Bool("active") {
			state = "active"
		}
		return join(name, ev.Str("kind"), state, peer(ev.Str("peer")))
	case wiretrace.KindUISession:
		if !ev.Bool("open") {
			return "session closed"
		}
		return "session open"
	case wiretrace.KindUISearchChat:
		return join(name, peer(ev.Str("peer")))
	case wiretrace.KindUIFolder:
		for _, k := range []string{"filter", "id"} {
			if ev.Has(k) {
				return join(name, rawStr(ev, k))
			}
		}
		return name
	case wiretrace.KindUIAccount:
		if ev.Str("user") == "" {
			return "account none"
		}
		return join(name, fmt.Sprint(ev.Int("index")), peer(ev.Str("user")))
	case wiretrace.KindAppState:
		var parts []string
		for _, k := range []string{"active", "locked", "idle", "screen_locked", "sleep"} {
			if ev.Has(k) {
				parts = append(parts, fmt.Sprintf("%s=%v", k, ev.Bool(k)))
			}
		}
		return join(name, strings.Join(parts, " "))
	}
	return SemanticToken(ev)
}

// Describe is the action label followed by its semantic events:
// "click Dialogs::InnerWidget → chat h:5c1e… (user)".
func (r *Result) Describe(a *Action) string {
	s := a.Label()
	prev := ""
	for _, ev := range a.Semantic {
		l := SemanticLabel(ev, r.Peers)
		if l == prev {
			continue
		}
		prev = l
		s += " → " + l
	}
	return s
}
