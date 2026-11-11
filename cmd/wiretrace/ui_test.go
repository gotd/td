package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/wiretrace"
	"github.com/gotd/td/wiretrace/wiretracetest"
)

func TestTimelineUI(t *testing.T) {
	a := require.New(t)
	tr := wiretracetest.UISample(t)

	out, errOut, code := runCmd(t, "timeline", schemaFlag(), "--ui", tr.Dir)
	a.Equal(0, code, errOut)
	for _, want := range []string{
		`     0.500 marker #1 file "start"`,
		"    12.302 ui#41 click Dialogs::InnerWidget → chat h:",
		"=user#5 (user)  [3 req, 39 KB, 178 ms]\n" +
			"             req#310 messages.getHistory                  deferred  lat=61ms\n" +
			"             req#311 messages.getPeerSettings             deferred  lat=50ms\n" +
			"             req#315 messages.getMessagesReactions        chain ⇐ req#310  lat=20ms\n",
		"     2.000 [background]\n             req#2 updates.getDifference                background  lat=30ms",
		"             req#320 updates.getChannelDifference         update",
		"    19.931 ui#57 click DocumentOpenClickHandler → media.open document h:",
		"             req#332 upload.getFile                       chain ⇐ req#331  lat=70ms",
		"    24.500 ui#58 scroll up HistoryWidget ×14 1.2s  [4 req, 176 B, 2500 ms]",
		"             req#350 account.updateStatus                 after +1288ms",
		"    30.100 ui#59 type HistoryWidget n=6-20\n",
		"    22.000 ui#70 key Escape Media::View::OverlayWidget → media.close\n",
		"   120.000 view app.state idle=true",
		"     0.900 view session open\n",
		"    43.000 ui#64 key Down HistoryInner +5 repeats\n",
		"    44.000 ui#65 key ctrl+K via shortcut  [1 req, 44 B, 0 ms]\n" +
			"             req#385 messages.search                      sync\n",
		"    45.000 view box layer\n",
		"    46.000 ui#66 tap Calls::Group::Panel → call group active h:",
		"    50.000 view call group ended\n",
		"             req#390 messages.sendMedia                   after +4200ms nested\n",
		`   150.000 marker #3 key "before switching folders"`,
	} {
		a.Contains(out, want)
	}

	out, errOut, code = runCmd(t, "timeline", "--ui=collapse", tr.Dir)
	a.Equal(0, code, errOut)
	for _, want := range []string{
		"→ messages.getHistory (deferred), messages.getPeerSettings (deferred), messages.getMessagesReactions (chain)" +
			"  [3 req, 39 KB, 178 ms]",
		"→ upload.getFile (deferred), upload.getFile ×2 (chain)  [3 req, 1.5 MB, 173 ms]",
		"scroll up HistoryWidget ×14 1.2s → messages.getHistory ×3 (sync), account.updateStatus (after)",
		"     2.000 [background] updates.getDifference\n",
		"    13.000 [update] updates.getChannelDifference\n",
	} {
		a.Contains(out, want)
	}
	a.NotContains(out, "req#")
	a.NotContains(out, "=user#5", "no schema: no peer join")
	a.NotContains(out, "=channel#", "the channel is in no request body")

	// Unstamped requests, with inference.
	out, errOut, code = runCmd(t, "timeline", "--ui=collapse", "--ui-infer", "--ui-window=10s", tr.Dir)
	a.Equal(0, code, errOut)
	a.Contains(out, "→ messages.getHistory (sync), updates.getDifference (after)")

	_, errOut, code = runCmd(t, "timeline", "--ui=bogus", tr.Dir)
	a.Equal(2, code)
	a.Contains(errOut, "--ui=collapse")

	// The plain timeline shows stamps.
	out, errOut, code = runCmd(t, "timeline", "--kinds=req.new,ui.*", tr.Dir)
	a.Equal(0, code, errOut)
	a.Contains(out, "req#310 new messages.getHistory dc2 size=44 to_main  ⊂ ui#41 deferred +0ms")
	a.Contains(out, "req#350 new account.updateStatus dc2 size=44 to_main  ⊂ ui#58 after +1288ms")
	a.Contains(out, "ui.shortcut a=61 cmd=search  ⊂ ui#61 sync")
	a.NotContains(out, `ui={`)
}

func TestTimelineUIWithoutUI(t *testing.T) {
	a := require.New(t)
	tr := wiretracetest.Sample(t)
	out, errOut, code := runCmd(t, "timeline", "--ui=collapse", tr.Dir)
	a.Equal(0, code, errOut)
	a.NotContains(out, "ui#")
	a.Contains(out, "[background] help.getNearestDc, auth.signIn, photos.deletePhotos, help.getConfig, messages.sendMessage")
}

func TestStatsUI(t *testing.T) {
	a := require.New(t)
	tr := wiretracetest.UISample(t)
	out, errOut, code := runCmd(t, "stats", "--by=ui", tr.Dir)
	a.Equal(0, code, errOut)
	for _, want := range []string{
		"ui: 14 actions, 2 markers, 18 requests",
		"       7   38.9%      308 B  sync",
		"       3   16.7%     1.0 MB  chain",
		"       2   11.1%       88 B  after",
		"       2   11.1%      104 B  none",
		"tap(Calls::Group::Panel)>call(group)",
		"      1      4       4/4/4      176 B       0.0       0.0  scroll(up,HistoryWidget)",
		"click(Dialogs::InnerWidget)>chat(user)",
		"requests by view:",
	} {
		a.Contains(out, want)
	}

	out, errOut, code = runCmd(t, "stats", "--by=ui", "--json", tr.Dir)
	a.Equal(0, code, errOut)
	var s Stats
	a.NoError(json.Unmarshal([]byte(out), &s))
	a.NotNil(s.UI)
	a.Equal(14, s.UI.Actions)
	a.Equal(18, s.UI.Requests)
	share := map[string]int{}
	for _, sh := range s.UI.Share {
		share[sh.Via] = sh.Requests
	}
	a.Equal(map[string]int{"sync": 7, "chain": 3, "deferred": 3, "after": 2, "update": 1, "none": 2}, share)
	views := map[string]*UIViewStats{}
	for _, v := range s.UI.ByView {
		views[v.View] = v
	}
	a.Equal(3, views["media:document"].Requests)
	a.Equal(11, views["chat:user"].UserDriven)
	a.Equal(13, views["chat:user"].Requests)
	var click *UIActionStats
	for _, st := range s.UI.ByAction {
		if st.Key == "click(Dialogs::InnerWidget)>chat(user)" {
			click = st
		}
	}
	a.NotNil(click)
	a.Equal(3, click.Requests)
	a.Equal(1, click.Methods["messages.getHistory"])
	a.Equal(61.0, click.P90)

	// Default: no ui block.
	out, _, code = runCmd(t, "stats", "--json", tr.Dir)
	a.Equal(0, code)
	a.NotContains(out, `"ui"`)
	_, _, code = runCmd(t, "stats", "--by=bogus", tr.Dir)
	a.Equal(2, code)

	// No UI events: everything is background.
	out, errOut, code = runCmd(t, "stats", "--by=ui", "--json", wiretracetest.Sample(t).Dir)
	a.Equal(0, code, errOut)
	s = Stats{}
	a.NoError(json.Unmarshal([]byte(out), &s))
	a.Zero(s.UI.Actions)
	a.Empty(s.UI.ByAction)
	share = map[string]int{}
	for _, sh := range s.UI.Share {
		share[sh.Via] = sh.Requests
	}
	a.Equal(map[string]int{"sync": 0, "chain": 0, "deferred": 0, "after": 0, "update": 0, "none": 5}, share)
}

// uiVariant writes a run of the same scenario as the start of UISample,
// done differently: the click sends no getPeerSettings, the scroll makes 5
// getHistory, and there is an extra shortcut.
func uiVariant(t *testing.T) string {
	tr := wiretracetest.New(t, wiretrace.WriterOptions{Header: wiretracetest.UIHeader()})
	F := wiretrace.F
	S := wiretracetest.Stamp
	tr.At(0.4).Marker(1, "file", "start")
	tr.At(0.9).UI(wiretrace.KindUIChat, F("a", 1), F("src", "app"), F("w", 0), F("peer", "h:1"), F("peer_kind", "channel"))
	tr.ReqNew(1, tg.MessagesGetHistoryRequestTypeID, S(1, wiretrace.UIRelSync, 0))
	tr.At(2).ReqNew(2, tg.UpdatesGetDifferenceRequestTypeID)
	tr.ReqNew(3, tg.UpdatesGetDifferenceRequestTypeID)
	tr.At(5).Click(41, 90, []string{"Dialogs::InnerWidget", "Ui::ScrollArea"})
	tr.UI(wiretrace.KindUIChat, F("w", 0), F("peer", "h:2"), F("peer_kind", "user"), S(41, wiretrace.UIRelDeferred, 0))
	tr.ReqNew(10, tg.MessagesGetMessagesReactionsRequestTypeID, S(41, wiretrace.UIRelDeferred, 0))
	tr.ReqNew(11, tg.MessagesGetHistoryRequestTypeID, S(41, wiretrace.UIRelDeferred, 0))
	tr.At(8).UI(wiretrace.KindUIShortcut, F("a", 42), F("cmd", "next_chat"))
	tr.At(9).UI(wiretrace.KindUIScroll, F("a", 58), F("w", 0), F("dir", "up"),
		F("path", []string{"HistoryInner", "Ui::ScrollArea", "HistoryWidget"}))
	for i := range 5 {
		tr.ReqNew(int64(20+i), tg.MessagesGetHistoryRequestTypeID, S(58, wiretrace.UIRelSync, 0))
	}
	tr.At(50).Marker(2, "file", "before switching folders")
	tr.ReqNew(30, tg.MessagesGetDialogFiltersRequestTypeID)
	tr.Close()
	return tr.Dir
}

func TestDiffAlignUI(t *testing.T) {
	a := require.New(t)
	tr := wiretracetest.UISample(t)
	other := uiVariant(t)

	out, errOut, code := runCmd(t, "diff", "--align=ui", tr.Dir, tr.Dir)
	a.Equal(0, code, errOut)
	a.Contains(out, "== align ui: A 16 actions and markers, B 16 actions and markers")
	a.Contains(out, "  = click(Dialogs::InnerWidget)>chat(user)   A 3 req, B 3 req\n")
	a.NotContains(out, "\n  - ")
	a.NotContains(out, "\n  + ")
	a.NotContains(out, "~")

	out, errOut, code = runCmd(t, "diff", "--align=ui", tr.Dir, other)
	a.Equal(0, code, errOut)
	for _, want := range []string{
		"  = marker(start)   A 0 req, B 0 req\n",
		"  = app>chat(channel)   A 1 req, B 1 req\n",
		"  = click(Dialogs::InnerWidget)>chat(user)   A 3 req, B 2 req\n" +
			"      ~ messages.getPeerSettings: A 1, B 0\n",
		"  + shortcut(next_chat)   B 0 req\n",
		"  = scroll(up,HistoryWidget)   A 4 req, B 5 req\n" +
			"      ~ account.updateStatus: A 1, B 0\n" +
			"      ~ messages.getHistory: A 3, B 5\n",
		"  - click(DocumentOpenClickHandler)>media.open(document)   A 3 req: upload.getFile ×3\n",
		"== background (update-driven and unattributed): A 3 req, B 3 req\n" +
			"      ~ account.updateStatus: A 1, B 0\n" +
			"      ~ messages.getDialogFilters: A 0, B 1\n" +
			"      ~ updates.getChannelDifference: A 1, B 0\n" +
			"      ~ updates.getDifference: A 1, B 2\n",
	} {
		a.Contains(out, want)
	}

	// Same multiset, different order.
	b := wiretracetest.New(t, wiretrace.WriterOptions{Header: wiretracetest.UIHeader()})
	b.At(5).Click(41, 90, []string{"Dialogs::InnerWidget"})
	b.UI(wiretrace.KindUIChat, wiretrace.F("peer", "h:2"), wiretrace.F("peer_kind", "user"),
		wiretracetest.Stamp(41, wiretrace.UIRelDeferred, 0))
	for i, typ := range []uint32{
		tg.MessagesGetPeerSettingsRequestTypeID, tg.MessagesGetHistoryRequestTypeID, tg.MessagesGetMessagesReactionsRequestTypeID,
	} {
		b.ReqNew(int64(i+1), typ, wiretracetest.Stamp(41, wiretrace.UIRelDeferred, 0))
	}
	b.Close()
	out, errOut, code = runCmd(t, "diff", "--align=ui", "--timing", tr.Dir, b.Dir)
	a.Equal(0, code, errOut)
	a.Contains(out, "      order: A messages.getHistory, messages.getPeerSettings, messages.getMessagesReactions | "+
		"B messages.getPeerSettings, messages.getHistory, messages.getMessagesReactions\n")
	a.Contains(out, "  = click(Dialogs::InnerWidget)>chat(user)   A 3 req, B 3 req   A+11.802s B+0.000s")

	// Traces without UI events.
	out, errOut, code = runCmd(t, "diff", "--align=ui", wiretracetest.Sample(t).Dir, otherClient(t))
	a.Equal(0, code, errOut)
	a.Contains(out, "(no UI events in A or B)")
	a.Contains(out, "      ~ photos.deletePhotos: A 1, B 0\n")

	_, errOut, code = runCmd(t, "diff", "--align=bogus", tr.Dir, other)
	a.Equal(2, code)
	a.Contains(errOut, "--align must be")
}

func TestDiffAlignMarker(t *testing.T) {
	a := require.New(t)
	tr := wiretracetest.UISample(t)
	other := uiVariant(t)
	out, errOut, code := runCmd(t, "diff", "--align=marker", tr.Dir, other)
	a.Equal(0, code, errOut)
	for _, want := range []string{
		"== align marker: A 3 marker segments, B 3 marker segments\n",
		"  = (start)   A 0 req, B 0 req\n",
		"  = marker(start)   A 18 req, B 10 req\n",
		"      ~ messages.getHistory: A 5, B 7\n",
		"      ~ upload.getFile: A 3, B 0\n",
		"  = marker(before switching folders)   A 0 req, B 1 req\n" +
			"      ~ messages.getDialogFilters: A 0, B 1\n",
	} {
		a.Contains(out, want)
	}
	a.NotContains(out, "background")

	out, errOut, code = runCmd(t, "diff", "--align=marker", wiretracetest.Sample(t).Dir, otherClient(t))
	a.Equal(0, code, errOut)
	a.Contains(out, "(no ui.marker in A or B)")
	a.Contains(out, "  = (start)   A 5 req, B 0 req\n")
}
