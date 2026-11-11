package uiattr

import (
	"strings"

	"github.com/gotd/td/wiretrace"
)

// ChatView is the open chat of a window.
type ChatView struct {
	Peer     string
	PeerKind string
	Topic    int64
	Msg      int64
	Sublist  bool
}

// View is the view state at some point of a trace: what the user was
// looking at in the current main window (the window of the latest UI
// record with "w").
type View struct {
	W       int64
	Chat    *ChatView
	Section string
	// Box is the top shown box type; Layer is set while the window shows a
	// layer (ui.box with "layer":true, the type is not known).
	Box   string
	Layer bool
	// Media is the kind of the open media viewer item (photo, document,
	// story, …); MediaID its (hashed) id.
	Media   string
	MediaID string
	Window  string
	// Session is false after ui.session {"open":false}.
	Session bool
	// Call is the kind of the active call (call, group), if any.
	Call    string
	Account int64

	Locked, Idle, Inactive, ScreenLocked, Sleep bool

	// chatLater is set when the chat changed after the section.
	chatLater bool
}

// Label is a coarse label of the view: "locked", "media:document",
// "box:PeerListBox", "box:layer", "chat:user",
// "section:Info::WrapWidget/Settings" or "none".
func (v View) Label() string {
	switch {
	case v.Locked:
		return "locked"
	case v.Media != "":
		return "media:" + v.Media
	case v.Box != "":
		return "box:" + v.Box
	case v.Layer:
		return "box:layer"
	case v.Chat != nil && (v.chatLater || v.Section == ""):
		return "chat:" + v.Chat.PeerKind
	case v.Section != "":
		return "section:" + v.Section
	}
	return "none"
}

// otherWindow collects records without "w": their window is not a main
// window (media viewer, call windows, …).
const otherWindow = -1

type windowView struct {
	chat     *ChatView
	chatT    float64
	section  string
	sectionT float64
	boxes    []string
	layer    bool
	media    string
	mediaID  string
	state    string
	noSess   bool
}

type viewTracker struct {
	windows map[int64]*windowView
	cur     int64
	call    string
	account int64
	app     map[string]bool
}

func newViewTracker() viewTracker {
	return viewTracker{windows: map[int64]*windowView{}, app: map[string]bool{"active": true}}
}

func (t *viewTracker) window(w int64) *windowView {
	v, ok := t.windows[w]
	if !ok {
		v = &windowView{}
		t.windows[w] = v
	}
	return v
}

// sectionLabel is the type of a ui.section, with the info section and
// settings type for Info::WrapWidget: "Info::WrapWidget/Settings/…".
func sectionLabel(ev *wiretrace.Event) string {
	parts := []string{ev.Str("type")}
	for _, k := range []string{"info", "settings"} {
		if s := ev.Str(k); s != "" {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, "/")
}

func mediaID(ev *wiretrace.Event) string {
	for _, k := range []string{"doc", "photo", "story", "id", "peer"} {
		if s := ev.Str(k); s != "" {
			return s
		}
	}
	return ""
}

func (t *viewTracker) add(ev *wiretrace.Event) {
	idx := int64(otherWindow)
	if ev.Has("w") {
		idx = ev.Int("w")
		t.cur = idx
	}
	w := t.window(idx)
	switch ev.Kind {
	case wiretrace.KindUIChat:
		if peer := ev.Str("peer"); peer != "" {
			w.chat = &ChatView{
				Peer:     peer,
				PeerKind: ev.Str("peer_kind"),
				Topic:    ev.Int("topic"),
				Msg:      ev.Int("msg"),
				Sublist:  ev.Bool("sublist"),
			}
		} else {
			w.chat = nil
		}
		w.chatT = ev.T
	case wiretrace.KindUISection:
		w.section = sectionLabel(ev)
		w.sectionT = ev.T
	case wiretrace.KindUIBox:
		typ := ev.Str("type")
		if ev.Bool("layer") {
			w.layer = ev.Bool("shown")
			break
		}
		if ev.Bool("shown") {
			w.boxes = append(w.boxes, typ)
			break
		}
		for i := len(w.boxes) - 1; i >= 0; i-- {
			if typ == "" || w.boxes[i] == typ {
				w.boxes = append(w.boxes[:i:i], w.boxes[i+1:]...)
				break
			}
		}
	case wiretrace.KindUIMediaOpen:
		w.media = ev.Str("kind")
		if w.media == "" {
			w.media = "media"
		}
		w.mediaID = mediaID(ev)
	case wiretrace.KindUIMediaClose:
		w.media, w.mediaID = "", ""
	case wiretrace.KindAppWindow:
		w.state = ev.Str("state")
	case wiretrace.KindUISession:
		w.noSess = !ev.Bool("open")
	case wiretrace.KindUICall:
		// {"kind":"call|group","active":bool,"peer"?}.
		switch {
		case ev.Bool("active"):
			t.call = ev.Str("kind")
		case t.call == ev.Str("kind") || ev.Str("kind") == "":
			t.call = ""
		}
	case wiretrace.KindUIAccount:
		t.account = ev.Int("index")
	case wiretrace.KindAppState:
		for _, k := range []string{"active", "locked", "idle", "screen_locked", "sleep"} {
			if ev.Has(k) {
				t.app[k] = ev.Bool(k)
			}
		}
	}
}

func (t *viewTracker) snapshot() View {
	v := View{
		W:            t.cur,
		Call:         t.call,
		Session:      true,
		Account:      t.account,
		Locked:       t.app["locked"],
		Idle:         t.app["idle"],
		Inactive:     !t.app["active"],
		ScreenLocked: t.app["screen_locked"],
		Sleep:        t.app["sleep"],
	}
	w, ok := t.windows[t.cur]
	if !ok {
		return v
	}
	if w.chat != nil {
		c := *w.chat
		v.Chat = &c
	}
	v.Section = w.section
	v.chatLater = w.chatT >= w.sectionT
	if n := len(w.boxes); n > 0 {
		v.Box = w.boxes[n-1]
	}
	v.Layer = w.layer
	v.Session = !w.noSess
	v.Media, v.MediaID = w.media, w.mediaID
	v.Window = w.state
	return v
}
