package wiretracetest

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/wiretrace"
)

// UIHeader returns header fields with "ui" (P14) set.
func UIHeader() []wiretrace.Field {
	return []wiretrace.Field{
		wiretrace.F("run", "test"),
		wiretrace.F("app", "test"),
		wiretrace.F("layer", 232),
		wiretrace.F("ui", wiretrace.HeaderUI{V: 1, IDs: "hmac64"}),
	}
}

// Stamp returns the "ui" field of an event emitted for action a.
func Stamp(a uint32, rel string, dt float64) wiretrace.Field {
	return wiretrace.F("ui", wiretrace.UIStamp{A: a, Rel: rel, DT: dt})
}

// CauseField returns the "cause" field of an event emitted while the
// response of req (0: an update) was handled.
func CauseField(req int64, typ uint32) wiretrace.Field {
	return wiretrace.F("cause", wiretrace.Cause{Req: req, Outer: 1, Type: wiretrace.TypeID(typ)})
}

// At sets the time of the next event.
func (tr *Trace) At(t float64) *Trace {
	tr.Now = t
	return tr
}

// UI writes a UI record on the main thread.
func (tr *Trace) UI(kind string, fields ...wiretrace.Field) uint64 {
	tr.t.Helper()
	return tr.Event("main", kind, -1, fields...)
}

// Click writes ui.press and ui.click of action a on window 0, ms apart.
func (tr *Trace) Click(a uint32, ms float64, path []string, extra ...wiretrace.Field) {
	tr.t.Helper()
	F := wiretrace.F
	tr.UI(wiretrace.KindUIPress, F("a", a), F("w", 0), F("btn", "left"), F("path", path))
	tr.Now += ms / 1000
	fields := []wiretrace.Field{F("a", a), F("w", 0), F("btn", "left"), F("ms", ms), F("moved", false), F("path", path)}
	tr.UI(wiretrace.KindUIClick, append(fields, extra...)...)
}

// Marker writes a ui.marker.
func (tr *Trace) Marker(n int, src, label string) uint64 {
	tr.t.Helper()
	F := wiretrace.F
	fields := []wiretrace.Field{F("n", n), F("src", src)}
	if label != "" {
		fields = append(fields, F("label", label))
	}
	return tr.UI(wiretrace.KindUIMarker, fields...)
}

// ReqNew writes a main-thread req.new of method typ.
func (tr *Trace) ReqNew(req int64, typ uint32, extra ...wiretrace.Field) uint64 {
	tr.t.Helper()
	F := wiretrace.F
	fields := []wiretrace.Field{F("req", req), F("type", wiretrace.TypeID(typ)), F("to_main", true),
		F("size", 44), F("needs_layer", false)}
	return tr.Event("main", wiretrace.KindReqNew, tr.SDC, append(fields, extra...)...)
}

// ReqDone writes req.done of req.
func (tr *Trace) ReqDone(req int64, method, typ uint32, latMS float64, length int) uint64 {
	tr.t.Helper()
	F := wiretrace.F
	return tr.Event("main", wiretrace.KindReqDone, -1, F("req", req), F("method", wiretrace.TypeID(method)),
		F("lat_ms", latMS), F("type", wiretrace.TypeID(typ)), F("outer", wiretrace.ID(1)), F("len", length))
}

// WriteUIKey writes key (32 bytes) as the hex ui.key of the run.
func (tr *Trace) WriteUIKey(key []byte) {
	tr.t.Helper()
	if err := os.WriteFile(filepath.Join(tr.Dir, wiretrace.UIKeyName), []byte(hex.EncodeToString(key)+"\n"), 0o600); err != nil {
		tr.t.Fatal(err)
	}
}

// Ids of UISample.
const (
	// UIKeyHex is the ui.key of UISample.
	UIKeyHex = "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"
	// UIUser is the user whose chat UISample opens.
	UIUser int64 = 5
	// UIChannel is the channel UISample restores at startup.
	UIChannel int64 = 7
	// UIMsgHistory is the msg_id of the getHistory message of UISample.
	UIMsgHistory uint64 = 0x6100000000000004
)

// UISample writes a synthetic tdesktop session with UI events, following
// the examples of the P14 event schema:
//
//   - marker #1 "start" (src file) and #3 (src key, labelled) at the end
//   - a1, synthetic (src app): restore the last chat → getHistory (sync)
//   - req 2 updates.getDifference: background
//   - a41 click Dialogs::InnerWidget → ui.chat (deferred) → getHistory and
//     getPeerSettings (deferred), getMessagesReactions (chain ⇐ req 310)
//   - req 320 updates.getChannelDifference: update-driven
//   - a57 click DocumentOpenClickHandler → media.open document →
//     upload.getFile (deferred) ×1 + ×2 (chain); a70 Escape → media.close
//   - a58 scroll up HistoryWidget ×14 → getHistory ×3 (sync); then
//     account.updateStatus (after, dt 1288)
//   - a59 typing, a60 Enter → messages.sendMessage (sync)
//   - a61 shortcut search, a62 menu pin → messages.updatePinnedMessage
//     (sync)
//   - a63 double click, a64 Down with 5 auto-repeats (ui.key.end), a65
//     Ctrl+K to a QShortcut (ui.key via shortcut, written when the scope
//     ends, after its sync request), a66 tap in a non-main window (no "w")
//   - a67 click that opens a modal file dialog: messages.sendMedia inside
//     the nested loop ("after", dt 4200 from the scope open)
//   - layer box, group call, session and account records
//   - app.state idle, then req 380 account.updateStatus: background
//
// The record shapes follow the tdesktop emitter (branch trace-ui). Peer
// ids are hashed with UIKeyHex, which is written as ui.key (64 lowercase hex
// chars and a newline).
func UISample(t testing.TB) *Trace {
	t.Helper()
	tr := New(t, wiretrace.WriterOptions{Header: UIHeader()})
	key, err := hex.DecodeString(UIKeyHex)
	if err != nil {
		t.Fatal(err)
	}
	tr.WriteUIKey(key)
	peer := wiretrace.IDKey(key).Hash(wiretrace.PeerValue(wiretrace.PeerUser, UIUser))
	channel := wiretrace.IDKey(key).Hash(wiretrace.PeerValue(wiretrace.PeerChannel, UIChannel))
	doc := wiretrace.IDKey(key).Hash(8812)
	F := wiretrace.F
	const (
		getHistory   = tg.MessagesGetHistoryRequestTypeID
		getSettings  = tg.MessagesGetPeerSettingsRequestTypeID
		getReactions = tg.MessagesGetMessagesReactionsRequestTypeID
		getFile      = tg.UploadGetFileRequestTypeID
	)

	tr.At(0.5).Marker(1, "file", "start")
	tr.At(0.8).UI(wiretrace.KindUIAccount, F("index", 0), F("user", wiretrace.IDKey(key).Hash(wiretrace.PeerValue(wiretrace.PeerUser, 105))))
	tr.At(0.9).UI(wiretrace.KindUISession, F("w", 0), F("open", true))
	tr.At(0.95).UI(wiretrace.KindUIChat, F("a", 1), F("src", "app"), F("w", 0), F("peer", channel), F("peer_kind", "channel"))
	tr.ReqNew(1, getHistory, Stamp(1, wiretrace.UIRelSync, 0))
	tr.At(2).ReqNew(2, tg.UpdatesGetDifferenceRequestTypeID)
	tr.ReqDone(2, tg.UpdatesGetDifferenceRequestTypeID, tg.UpdatesDifferenceEmptyTypeID, 30, 16)

	// a41: click a chat in the dialogs list.
	inner := []string{"Dialogs::InnerWidget", "Ui::ScrollArea", "Dialogs::Widget", "MainWidget"}
	tr.At(12.301533).Click(41, 97, inner)
	tr.UI(wiretrace.KindUIChat, F("w", 0), F("peer", peer), F("peer_kind", "user"), F("topic", 0),
		F("sublist", false), F("msg", 0), Stamp(41, wiretrace.UIRelDeferred, 0))
	tr.ReqNew(310, getHistory, Stamp(41, wiretrace.UIRelDeferred, 0))
	tr.Event("main", wiretrace.KindReqCaller, -1, F("req", 310), F("file", "history/history_widget.cpp"),
		F("line", 5115), F("function", "void HistoryWidget::firstLoadMessages()"), F("subsystem", "history"),
		Stamp(41, wiretrace.UIRelDeferred, 0))
	tr.Event("t1", wiretrace.KindReqMsg, tr.SDC, F("req", 310), F("was", wiretrace.ID(0)),
		F("msg_id", wiretrace.ID(UIMsgHistory)), F("seqno", 3))
	tr.Out(Item{MsgID: UIMsgHistory, SeqNo: 3, Body: Encode(t, &tg.MessagesGetHistoryRequest{
		Peer: &tg.InputPeerUser{UserID: UIUser, AccessHash: 1}, Limit: 50,
	})})
	tr.ReqNew(311, getSettings, Stamp(41, wiretrace.UIRelDeferred, 1))
	tr.At(12.46).ReqDone(310, getHistory, tg.MessagesMessagesTypeID, 61, 40000)
	tr.ReqNew(315, getReactions, CauseField(310, tg.MessagesMessagesTypeID))
	tr.ReqDone(311, getSettings, tg.MessagesPeerSettingsTypeID, 50, 100)
	tr.At(12.48).ReqDone(315, getReactions, tg.UpdatesTypeID, 20, 200)

	tr.At(13).ReqNew(320, tg.UpdatesGetChannelDifferenceRequestTypeID, CauseField(0, tg.UpdatesTypeID))

	// a57: open a document.
	list := []string{"HistoryView::ListWidget", "Ui::ScrollArea", "HistoryView::ChatWidget"}
	tr.At(19.931).Click(57, 80, list, F("handler", "DocumentOpenClickHandler"), F("entity", 0))
	tr.UI(wiretrace.KindUIMediaOpen, F("w", 0), F("kind", "document"), F("doc", doc),
		F("item", map[string]any{"peer": peer, "msg": 8812}), F("streaming", true), Stamp(57, wiretrace.UIRelDeferred, 1))
	tr.ReqNew(330, getFile, Stamp(57, wiretrace.UIRelDeferred, 1))
	tr.At(20.1).ReqDone(330, getFile, tg.UploadFileTypeID, 80, 524288)
	tr.ReqNew(331, getFile, CauseField(330, tg.UploadFileTypeID))
	tr.ReqDone(331, getFile, tg.UploadFileTypeID, 70, 524288)
	tr.ReqNew(332, getFile, CauseField(331, tg.UploadFileTypeID))
	tr.ReqDone(332, getFile, tg.UploadFileTypeID, 70, 524288)
	tr.At(22).UI(wiretrace.KindUIKey, F("a", 70), F("key", "Escape"), F("mods", []string{}),
		F("path", []string{"Media::View::OverlayWidget"}))
	tr.UI(wiretrace.KindUIMediaClose, F("w", 0), Stamp(70, wiretrace.UIRelSync, 0))

	// a58: scroll up in the history.
	hist := []string{"HistoryInner", "Ui::ScrollArea", "HistoryWidget"}
	tr.At(24.5).UI(wiretrace.KindUIScroll, F("a", 58), F("w", 0), F("dir", "up"), F("path", hist))
	for i, at := range []float64{24.81, 25.1, 25.3} {
		tr.At(at).ReqNew(int64(340+i), getHistory, Stamp(58, wiretrace.UIRelSync, 0))
	}
	tr.At(25.712).UI(wiretrace.KindUIScrollEnd, F("a", 58), F("n", 14), F("dy", -1680), F("ms", 1212))
	tr.At(27).ReqNew(350, tg.AccountUpdateStatusRequestTypeID, Stamp(58, wiretrace.UIRelAfter, 1288))

	// a59, a60: type a message and send it.
	tr.At(30.1).UI(wiretrace.KindUIType, F("a", 59), F("w", 0), F("field", []string{"Ui::InputField", "HistoryWidget"}))
	tr.At(33.9).UI(wiretrace.KindUITypeEnd, F("a", 59), F("n", "6-20"), F("ms", 3800))
	tr.At(34).UI(wiretrace.KindUIKey, F("a", 60), F("w", 0), F("key", "Enter"), F("mods", []string{}),
		F("path", []string{"Ui::InputField", "HistoryWidget"}))
	tr.ReqNew(360, tg.MessagesSendMessageRequestTypeID, Stamp(60, wiretrace.UIRelSync, 0))

	// a61, a62.
	tr.At(40).UI(wiretrace.KindUIShortcut, F("a", 61), F("cmd", "search"), Stamp(61, wiretrace.UIRelSync, 0))
	tr.At(41).UI(wiretrace.KindUIMenu, F("a", 62), F("w", 0), F("item", "lng_context_pin_msg"),
		F("path", []string{"Ui::Menu::Action", "Ui::Menu::Menu", "Ui::PopupMenu"}))
	tr.ReqNew(370, tg.MessagesUpdatePinnedMessageRequestTypeID, Stamp(62, wiretrace.UIRelSync, 0))

	// a63..a67.
	tr.At(42).UI(wiretrace.KindUIPress, F("a", 63), F("w", 0), F("btn", "left"), F("dbl", true),
		F("path", []string{"HistoryInner", "Ui::ScrollArea", "HistoryWidget"}))
	tr.At(42.05).UI(wiretrace.KindUIClick, F("a", 63), F("w", 0), F("btn", "left"), F("ms", 50), F("moved", false),
		F("path", []string{"HistoryInner", "Ui::ScrollArea", "HistoryWidget"}))
	tr.At(43).UI(wiretrace.KindUIKey, F("a", 64), F("w", 0), F("key", "Down"), F("mods", []string{}),
		F("path", []string{"HistoryInner", "Ui::ScrollArea", "HistoryWidget"}))
	tr.At(43.4).UI(wiretrace.KindUIKeyEnd, F("a", 64), F("repeat", 5), F("ms", 400))
	tr.At(44).ReqNew(385, tg.MessagesSearchRequestTypeID, Stamp(65, wiretrace.UIRelSync, 0))
	tr.UI(wiretrace.KindUIKey, F("a", 65), F("key", "K"), F("mods", []string{"ctrl"}), F("via", "shortcut"))
	tr.At(45).UI(wiretrace.KindUIBox, F("w", 0), F("layer", true), F("shown", true))
	tr.UI(wiretrace.KindUIBox, F("w", 0), F("shown", true), F("type", "PeerListBox"))
	tr.At(45.5).UI(wiretrace.KindUIBox, F("w", 0), F("shown", false), F("type", "PeerListBox"))
	tr.UI(wiretrace.KindUIBox, F("w", 0), F("layer", true), F("shown", false))
	tr.At(46).UI(wiretrace.KindUIPress, F("a", 66), F("btn", "touch"), F("path", []string{"Calls::Group::Panel"}))
	tr.At(46.1).UI(wiretrace.KindUIClick, F("a", 66), F("btn", "touch"), F("ms", 100), F("moved", false),
		F("path", []string{"Calls::Group::Panel"}))
	tr.UI(wiretrace.KindUICall, F("kind", "group"), F("active", true), F("peer", channel), Stamp(66, wiretrace.UIRelSync, 0))
	tr.At(50).UI(wiretrace.KindUICall, F("kind", "group"), F("active", false))
	attach := []string{"Ui::IconButton", "HistoryWidget"}
	tr.At(60).UI(wiretrace.KindUIPress, F("a", 67), F("w", 0), F("btn", "left"), F("path", attach))
	tr.At(60.08).UI(wiretrace.KindUIClick, F("a", 67), F("w", 0), F("btn", "left"), F("ms", 80), F("moved", false),
		F("path", attach))
	tr.At(64.28).ReqNew(390, tg.MessagesSendMediaRequestTypeID, Stamp(67, wiretrace.UIRelAfter, 4200))

	tr.At(120).UI(wiretrace.KindAppState, F("idle", true))
	tr.At(121).ReqNew(380, tg.AccountUpdateStatusRequestTypeID)
	tr.At(150).Marker(3, "key", "before switching folders")
	tr.Close()
	return tr
}
