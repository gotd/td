package uiattr

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/wiretrace"
	"github.com/gotd/td/wiretrace/tljson"
	"github.com/gotd/td/wiretrace/wiretracetest"
)

// build parses JSON lines (seq is assigned if missing) and resolves them.
func build(t *testing.T, opt Options, lines ...string) *Result {
	t.Helper()
	b := NewBuilder(opt)
	for i, l := range lines {
		if !strings.Contains(l, `"seq"`) {
			l = fmt.Sprintf(`{"seq":%d,`, i+1) + strings.TrimPrefix(l, "{")
		}
		ev, err := wiretrace.ParseEvent([]byte(l))
		require.NoError(t, err, l)
		b.Add(ev)
	}
	return b.Finish()
}

const (
	click1  = `{"t":1.000,"th":"main","ev":"ui.press","a":1,"w":0,"btn":"left","path":["Dialogs::InnerWidget","Ui::ScrollArea"]}`
	click1r = `{"t":1.050,"th":"main","ev":"ui.click","a":1,"w":0,"btn":"left","ms":50,"path":["Dialogs::InnerWidget","Ui::ScrollArea"]}`
	key2    = `{"t":1.100,"th":"main","ev":"ui.key","a":2,"w":0,"key":"Escape","mods":[],"path":["HistoryWidget"]}`
)

func req(id int, t float64, extra string) string {
	s := fmt.Sprintf(`{"t":%g,"th":"main","ev":"req.new","req":%d,"type":"0x4423e6c5","size":44`, t, id)
	if extra != "" {
		s += "," + extra
	}
	return s + "}"
}

func stamp(a int, rel string, dt float64) string {
	return fmt.Sprintf(`"ui":{"a":%d,"rel":%q,"dt":%g}`, a, rel, dt)
}

func cause(r int) string {
	return fmt.Sprintf(`"cause":{"req":%d,"outer":"1","type":"0x1cb5c415"}`, r)
}

type want struct {
	req    int64
	via    Via
	action uint32 // 0: none
	parent int64
	root   int64
	infer  bool
	nested bool
}

func check(t *testing.T, res *Result, wants ...want) {
	t.Helper()
	for _, w := range wants {
		r := res.ByReq[w.req]
		require.NotNil(t, r, "req %d", w.req)
		got := r.Attr
		require.Equal(t, w.via, got.Via, "req %d via", w.req)
		if w.action == 0 {
			require.Nil(t, got.Action, "req %d action", w.req)
		} else {
			require.NotNil(t, got.Action, "req %d action", w.req)
			require.Equal(t, w.action, got.Action.A, "req %d action", w.req)
		}
		require.Equal(t, w.parent, got.Parent, "req %d parent", w.req)
		require.Equal(t, w.root, got.Root, "req %d root", w.req)
		require.Equal(t, w.infer, got.Inferred, "req %d inferred", w.req)
		require.Equal(t, w.nested, got.Nested, "req %d nested", w.req)
	}
}

func TestAttribution(t *testing.T) {
	for _, tt := range []struct {
		name  string
		opt   Options
		lines []string
		wants []want
	}{
		{
			name:  "sync",
			lines: []string{click1, req(10, 1.01, stamp(1, "sync", 0)), click1r},
			wants: []want{{req: 10, via: ViaSync, action: 1}},
		},
		{
			name:  "sync wins over cause",
			lines: []string{click1, req(9, 0.5, ""), req(10, 1.01, stamp(1, "sync", 0)+","+cause(9))},
			wants: []want{{req: 10, via: ViaSync, action: 1}, {req: 9, via: ViaNone}},
		},
		{
			name: "chain",
			lines: []string{
				click1, click1r,
				req(10, 1.06, stamp(1, "deferred", 10)),
				req(11, 1.2, cause(10)),
				req(12, 1.3, cause(11)),
			},
			wants: []want{
				{req: 10, via: ViaDeferred, action: 1},
				{req: 11, via: ViaChain, action: 1, parent: 10, root: 10},
				{req: 12, via: ViaChain, action: 1, parent: 11, root: 10},
			},
		},
		{
			name:  "chain wins over deferred and after stamps",
			lines: []string{click1, req(9, 0.5, ""), req(10, 1.06, stamp(1, "deferred", 0)+","+cause(9))},
			wants: []want{{req: 10, via: ViaNone, parent: 9, root: 9}},
		},
		{
			name: "update-driven",
			lines: []string{
				click1, click1r,
				req(10, 1.06, cause(0)+","+stamp(1, "deferred", 1)),
				req(11, 1.2, cause(10)),
			},
			wants: []want{
				{req: 10, via: ViaUpdate},
				{req: 11, via: ViaUpdate, parent: 10, root: 10},
			},
		},
		{
			name: "windows",
			lines: []string{
				click1, click1r,
				req(10, 1.06, stamp(1, "deferred", 150)),
				req(11, 1.2, stamp(1, "deferred", 151)),
				req(12, 2, stamp(1, "after", 3000)),
				req(13, 4.1, stamp(1, "after", 3001)),
				req(14, 4.1, stamp(1, "bogus", 1)),
			},
			wants: []want{
				{req: 10, via: ViaDeferred, action: 1},
				{req: 11, via: ViaAfter, action: 1},
				{req: 12, via: ViaAfter, action: 1},
				{req: 13, via: ViaNone},
				{req: 14, via: ViaNone},
			},
		},
		{
			name: "custom windows",
			opt:  Options{DeferredWindow: 500 * time.Millisecond, AfterWindow: 5 * time.Second},
			lines: []string{
				click1, click1r,
				req(10, 1.4, stamp(1, "deferred", 400)),
				req(11, 5, stamp(1, "after", 4000)),
			},
			wants: []want{
				{req: 10, via: ViaDeferred, action: 1},
				{req: 11, via: ViaAfter, action: 1},
			},
		},
		{
			name: "overlapping actions: the stamp wins over time",
			lines: []string{
				click1, click1r, key2,
				req(10, 1.11, stamp(1, "deferred", 60)),
				req(11, 1.11, stamp(2, "sync", 0)),
			},
			wants: []want{
				{req: 10, via: ViaDeferred, action: 1},
				{req: 11, via: ViaSync, action: 2},
			},
		},
		{
			name: "nested scopes",
			lines: []string{
				click1,
				req(10, 1.01, stamp(1, "sync", 0)),
				`{"t":1.02,"th":"main","ev":"ui.key","a":2,"w":0,"key":"Enter","mods":[],"path":["Ui::InputField","Ui::BoxContent"]}`,
				req(11, 1.03, stamp(2, "sync", 0)),
				req(12, 1.04, stamp(1, "sync", 0)),
				click1r,
			},
			wants: []want{
				{req: 10, via: ViaSync, action: 1},
				{req: 11, via: ViaSync, action: 2},
				{req: 12, via: ViaSync, action: 1},
			},
		},
		{
			name: "missing stamps are background",
			lines: []string{
				click1, click1r,
				req(10, 1.06, ""),
				req(11, 2, ""),
			},
			wants: []want{{req: 10, via: ViaNone}, {req: 11, via: ViaNone}},
		},
		{
			name: "missing stamps inferred",
			opt:  Options{Infer: true},
			lines: []string{
				req(9, 0.5, ""),
				click1, click1r, key2,
				req(10, 1.2, ""),
				req(11, 2, ""),
				req(12, 4.2, ""),
				`{"t":1.21,"th":"t1","ev":"req.new","req":13,"type":"0x4423e6c5"}`,
				req(14, 1.25, stamp(1, "deferred", 1)),
				req(15, 1.3, cause(10)),
			},
			wants: []want{
				{req: 9, via: ViaNone},
				{req: 10, via: ViaDeferred, action: 2, infer: true},
				{req: 11, via: ViaAfter, action: 2, infer: true},
				{req: 12, via: ViaNone},
				{req: 13, via: ViaNone},
				{req: 14, via: ViaDeferred, action: 1},
				{req: 15, via: ViaChain, action: 2, parent: 10, root: 10, infer: true},
			},
		},
		{
			name: "unknown parent and cycles",
			lines: []string{
				req(10, 1, cause(99)),
				req(11, 1.1, cause(12)),
				req(12, 1.2, cause(11)),
				req(13, 1.3, cause(13)),
			},
			wants: []want{
				{req: 10, via: ViaNone, parent: 99, root: 99},
				{req: 11, via: ViaNone, parent: 12, root: 11},
				{req: 12, via: ViaNone, parent: 11, root: 11},
				{req: 13, via: ViaNone, parent: 13, root: 13},
			},
		},
		{
			name: "nested event loop",
			lines: []string{
				click1, click1r,
				// A modal dialog inside the click handler: dt counts from the
				// scope open (the ui.click record at 1.05); a key inside the
				// dialog is a newer action.
				`{"t":1.3,"th":"main","ev":"ui.key","a":2,"key":"Tab","mods":[],"path":["Ui::BoxContent"]}`,
				req(10, 1.7, stamp(1, "after", 650)),
				req(11, 5.05, stamp(1, "after", 4000)),
				// The same dt, but not from an input record: a plain after
				// stamp beyond the window.
				req(12, 5.3, stamp(1, "after", 4000)),
			},
			wants: []want{
				{req: 10, via: ViaAfter, action: 1, nested: true},
				{req: 11, via: ViaAfter, action: 1, nested: true},
				{req: 12, via: ViaNone},
			},
		},
		{
			name:  "stamp of an unknown action",
			lines: []string{req(10, 5, stamp(77, "sync", 0))},
			wants: []want{{req: 10, via: ViaSync, action: 77}},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			check(t, build(t, tt.opt, tt.lines...), tt.wants...)
		})
	}
}

func TestLongChain(t *testing.T) {
	a := require.New(t)
	lines := []string{click1, req(1, 1.01, stamp(1, "sync", 0))}
	for i := 2; i <= 40; i++ {
		lines = append(lines, req(i, 1+float64(i)/100, cause(i-1)))
	}
	res := build(t, Options{}, lines...)
	check(t, res, want{req: 40, via: ViaChain, action: 1, parent: 39, root: 1})
	a.Len(res.ByID[1].Requests, 40)

	// Forward references deeper than MaxChain are cut.
	lines = []string{click1}
	for i := 1; i < 20; i++ {
		lines = append(lines, req(i, 1+float64(i)/100, cause(i+1)))
	}
	lines = append(lines, req(20, 1.3, stamp(1, "sync", 0)))
	res = build(t, Options{MaxChain: 4}, lines...)
	a.Equal(ViaNone, res.ByReq[1].Attr.Via)
	a.Equal(ViaSync, res.ByReq[20].Attr.Via)
	a.Equal(ViaChain, res.ByReq[19].Attr.Via)
}

func TestNoUI(t *testing.T) {
	a := require.New(t)
	res := build(t, Options{Infer: true},
		`{"seq":0,"ev":"hdr","v":1,"seg":0}`,
		req(1, 1, ""),
		req(2, 1.1, cause(1)),
		req(3, 1.2, cause(0)),
		`{"t":1.3,"th":"main","ev":"req.done","req":1,"method":"0x4423e6c5","lat_ms":12.5,"type":"0x1cb5c415","len":100}`,
	)
	a.False(res.HasUI)
	a.Empty(res.Actions)
	a.Empty(res.Markers)
	a.Len(res.Background(), 3)
	check(t, res,
		want{req: 1, via: ViaNone},
		want{req: 2, via: ViaNone, parent: 1, root: 1},
		want{req: 3, via: ViaUpdate},
	)
	r := res.ByReq[1]
	a.True(r.Done)
	a.Equal(12.5, r.LatMS)
	a.Equal(int64(144), r.Bytes())
	a.Equal("none", r.View.Label())

	res = build(t, Options{}, `{"seq":0,"ev":"hdr","v":1,"seg":0,"ui":{"v":1,"ids":"hmac64","raw_ids":false}}`)
	a.True(res.HasUI)
}

func TestActions(t *testing.T) {
	a := require.New(t)
	res := build(t, Options{},
		`{"t":0.95,"th":"main","ev":"ui.chat","a":100,"src":"app","w":0,"peer":"h:1","peer_kind":"channel"}`,
		req(1, 0.96, stamp(100, "sync", 0)),
		click1, click1r,
		`{"t":1.051,"th":"main","ev":"ui.chat","w":0,"peer":"h:5c1e","peer_kind":"user","topic":0,"msg":0,"ui":{"a":1,"rel":"deferred","dt":0}}`,
		req(2, 1.052, stamp(1, "deferred", 0)),
		`{"t":2,"th":"main","ev":"ui.scroll","a":3,"w":0,"dir":"up","path":["HistoryInner","Ui::ScrollArea","HistoryWidget"]}`,
		`{"t":3.2,"th":"main","ev":"ui.scroll.end","a":3,"n":14,"dy":-1680,"ms":1212}`,
		`{"t":4,"th":"main","ev":"ui.type","a":4,"w":0,"field":["Ui::InputField","HistoryWidget"],"masked":true}`,
		`{"t":5,"th":"main","ev":"ui.type.end","a":4,"masked":true,"ms":1000}`,
		`{"t":6,"th":"main","ev":"ui.key","a":5,"w":0,"key":"Enter","mods":["ctrl"],"path":["Ui::InputField","HistoryWidget"]}`,
		`{"t":6.5,"th":"main","ev":"ui.key.end","a":5,"repeat":3,"ms":120}`,
		`{"t":7,"th":"main","ev":"ui.shortcut","a":6,"cmd":"search","ui":{"a":6,"rel":"sync"}}`,
		`{"t":8,"th":"main","ev":"ui.menu","a":7,"item":"lng_context_pin_msg","path":["Ui::Menu::Action","Ui::Menu::Menu"]}`,
		`{"t":9,"th":"main","ev":"ui.menu","a":8,"item":null,"len":12,"path":["Ui::Menu::Action"]}`,
		`{"t":10,"th":"main","ev":"ui.press","a":9,"w":0,"btn":"left","path":["HistoryView::ListWidget"]}`,
		`{"t":10.1,"th":"main","ev":"ui.click","a":9,"w":0,"btn":"left","path":["HistoryView::ListWidget"],"handler":"DocumentOpenClickHandler"}`,
		`{"t":10.2,"th":"main","ev":"ui.media.open","w":0,"kind":"document","doc":"h:0b77","ui":{"a":9,"rel":"deferred","dt":1}}`,
		`{"t":10.3,"th":"main","ev":"ui.press","a":10,"w":0,"btn":"left","dbl":true,"path":["HistoryInner"]}`,
		`{"t":10.31,"th":"main","ev":"ui.click","a":10,"w":0,"btn":"left","path":["HistoryInner"]}`,
		`{"t":10.4,"th":"main","ev":"ui.press","a":11,"btn":"touch","path":["Media::View::OverlayWidget"]}`,
		`{"t":10.41,"th":"main","ev":"ui.click","a":11,"btn":"touch","path":["Media::View::OverlayWidget"]}`,
		`{"t":10.5,"th":"main","ev":"ui.press","a":12,"w":0,"btn":"right","path":["HistoryInner"]}`,
		`{"t":10.6,"th":"main","ev":"ui.key","a":13,"key":"K","mods":["ctrl"],"via":"shortcut"}`,
		`{"t":10.7,"th":"main","ev":"ui.shortcut","a":14,"cmd":"search","cmds":["search","next_chat"]}`,
		`{"t":10.8,"th":"main","ev":"ui.type","a":15,"w":0,"field":["Ui::InputField","Dialogs::Widget"],"search":true}`,
		`{"t":10.9,"th":"main","ev":"ui.type.end","a":15,"n":"2-5","search":true,"count":3,"ms":100}`,
		`{"t":11,"th":"main","ev":"ui.marker","n":3,"src":"key","label":"before switching folders","label_user":true}`,
		`{"t":11.5,"th":"main","ev":"ui.marker","n":4,"src":"key"}`,
		`{"t":12,"th":"main","ev":"app.state","idle":true}`,
	)
	a.True(res.HasUI)
	var labels, tokens []string
	for _, act := range res.Actions {
		labels = append(labels, res.Describe(act))
		tokens = append(tokens, act.Token())
	}
	a.Equal([]string{
		"app → chat h:1 (channel)",
		"click Dialogs::InnerWidget → chat h:5c1e (user)",
		"scroll up HistoryWidget ×14 1.2s",
		"type HistoryWidget masked",
		"key ctrl+Enter HistoryWidget +3 repeats",
		"shortcut search",
		"menu lng_context_pin_msg",
		"menu ? (len 12)",
		"click DocumentOpenClickHandler → media.open document h:0b77",
		"dblclick HistoryInner",
		"tap Media::View::OverlayWidget",
		"press:right HistoryInner",
		"key ctrl+K via shortcut",
		"shortcut search,next_chat",
		"type Dialogs::Widget search n=2-5 count=3",
	}, labels)
	a.Equal([]string{
		"app>chat(channel)",
		"click(Dialogs::InnerWidget)>chat(user)",
		"scroll(up,HistoryWidget)",
		"type(HistoryWidget)",
		"key(ctrl+Enter,HistoryWidget)",
		"shortcut(search)",
		"menu(lng_context_pin_msg)",
		"menu()",
		"click(DocumentOpenClickHandler)>media.open(document)",
		"dblclick(HistoryInner)",
		"tap(Media::View::OverlayWidget)",
		"press:right(HistoryInner)",
		"key(ctrl+K,)",
		"shortcut(search)",
		"type(Dialogs::Widget)",
	}, tokens)
	click := res.ByID[1]
	a.Equal(ActionClick, click.Kind)
	a.Len(click.Events, 2)
	a.Equal(1.0, click.T)
	a.Equal(1.052, click.End)
	a.Equal(ActionApp, res.ByID[100].Src)

	a.Len(res.Markers, 2)
	a.Equal("marker(before switching folders)", res.Markers[0].Token())
	a.True(res.Markers[0].LabelUser)
	a.Equal("marker(#4)", res.Markers[1].Token())
	a.Len(res.Views, 1)
	a.Equal("app.state idle=true", SemanticLabel(res.Views[0], nil))
}

func TestView(t *testing.T) {
	a := require.New(t)
	res := build(t, Options{},
		req(1, 0.1, ""),
		`{"t":1,"th":"main","ev":"ui.chat","w":0,"peer":"h:5c1e","peer_kind":"user"}`,
		req(2, 1.1, ""),
		`{"t":2,"th":"main","ev":"ui.media.open","w":0,"kind":"document","doc":"h:0b77"}`,
		req(3, 2.1, ""),
		`{"t":3,"th":"main","ev":"ui.media.close","w":0}`,
		`{"t":3.1,"th":"main","ev":"ui.box","w":0,"shown":true,"type":"PeerListBox"}`,
		`{"t":3.2,"th":"main","ev":"ui.box","w":0,"shown":true,"type":"ConfirmBox"}`,
		req(4, 3.3, ""),
		`{"t":3.4,"th":"main","ev":"ui.box","w":0,"shown":false,"type":"ConfirmBox"}`,
		req(5, 3.5, ""),
		`{"t":3.6,"th":"main","ev":"ui.box","w":0,"shown":false,"type":"PeerListBox"}`,
		`{"t":4,"th":"main","ev":"ui.section","w":0,"type":"Info::WrapWidget","info":"Settings","settings":"Settings::PrivacySecurity"}`,
		req(6, 4.1, ""),
		`{"t":5,"th":"main","ev":"ui.chat","w":0,"peer":"h:5c1e","peer_kind":"user"}`,
		req(7, 5.1, ""),
		`{"t":6,"th":"main","ev":"app.state","locked":true}`,
		req(8, 6.1, ""),
		`{"t":7,"th":"main","ev":"app.state","locked":false,"active":false}`,
		`{"t":7.5,"th":"main","ev":"ui.chat","w":1,"peer":"h:aa","peer_kind":"channel"}`,
		req(9, 7.6, ""),
		`{"t":8,"th":"main","ev":"ui.chat","w":1}`,
		req(10, 8.1, ""),
		`{"t":9,"th":"main","ev":"ui.box","w":1,"layer":true,"shown":true}`,
		`{"t":9.1,"th":"main","ev":"ui.call","kind":"group","active":true,"peer":"h:cc"}`,
		`{"t":9.2,"th":"main","ev":"ui.session","w":1,"open":false}`,
		req(11, 9.3, ""),
		`{"t":9.4,"th":"main","ev":"ui.box","w":1,"layer":true,"shown":false}`,
		`{"t":9.5,"th":"main","ev":"ui.box","shown":true,"type":"Calls::Panel"}`,
		`{"t":9.6,"th":"main","ev":"ui.call","kind":"group","active":false}`,
		req(12, 9.7, ""),
	)
	var got []string
	for _, r := range res.Requests {
		got = append(got, r.View.Label())
	}
	a.Equal([]string{
		"none", "chat:user", "media:document", "box:ConfirmBox", "box:PeerListBox",
		"section:Info::WrapWidget/Settings/Settings::PrivacySecurity", "chat:user", "locked", "chat:channel", "none",
		"box:layer", "none",
	}, got)
	a.Equal("group", res.ByReq[11].View.Call)
	a.False(res.ByReq[11].View.Session)
	a.Empty(res.ByReq[12].View.Call, "the box of a non-main window does not change the view")
	v := res.ByReq[9].View
	a.Equal(int64(1), v.W)
	a.True(v.Inactive)
	a.Equal("h:aa", v.Chat.Peer)
	a.Equal("h:0b77", res.ByReq[3].View.MediaID)
}

func TestPeers(t *testing.T) {
	a := require.New(t)
	obj := tljson.NewObject("messages.getHistory")
	peer := tljson.NewObject("inputPeerUser")
	peer.Set("user_id", tljson.Long(5))
	obj.Set("peer", peer)
	ch := tljson.NewObject("inputPeerChannel")
	ch.Set("channel_id", tljson.Long(7))
	obj.Set("list", []any{ch, int32(1)})
	peers := PeersFrom(obj)
	a.Equal([]Peer{{Kind: wiretrace.PeerUser, ID: 5}, {Kind: wiretrace.PeerChannel, ID: 7}}, peers)

	key := wiretrace.IDKey(make([]byte, 32))
	res := build(t, Options{})
	res.JoinPeers(key, peers)
	a.Equal("user#5", res.Peers[key.Hash(5)])
	a.Equal("channel#7", res.Peers[key.Hash(2<<48|7)])
	res.JoinPeers(nil, peers)
	a.Len(res.Peers, 2)

	// ui_raw_ids: decimal PeerId.value strings.
	res.JoinRawPeers(peers)
	a.Equal("channel#7", res.Peers["562949953421319"])
}

func TestUISample(t *testing.T) {
	a := require.New(t)
	tr := wiretracetest.UISample(t)
	run, err := wiretrace.Open(tr.Dir)
	a.NoError(err)
	defer func() { _ = run.Close() }()
	a.NotNil(run.Header.UI)
	res, err := Build(run, Options{})
	a.NoError(err)
	a.True(res.HasUI)
	check(t, res,
		want{req: 1, via: ViaSync, action: 1},
		want{req: 2, via: ViaNone},
		want{req: 310, via: ViaDeferred, action: 41},
		want{req: 311, via: ViaDeferred, action: 41},
		want{req: 315, via: ViaChain, action: 41, parent: 310, root: 310},
		want{req: 320, via: ViaUpdate},
		want{req: 330, via: ViaDeferred, action: 57},
		want{req: 332, via: ViaChain, action: 57, parent: 331, root: 330},
		want{req: 340, via: ViaSync, action: 58},
		want{req: 350, via: ViaAfter, action: 58},
		want{req: 360, via: ViaSync, action: 60},
		want{req: 370, via: ViaSync, action: 62},
		want{req: 380, via: ViaNone},
		want{req: 385, via: ViaSync, action: 65},
		want{req: 390, via: ViaAfter, action: 67, nested: true},
	)
	a.Equal(ActionKey, res.ByID[65].Kind, "the input record after its stamped request")
	a.Equal(44.0, res.ByID[65].T)
	a.Equal("key Down HistoryInner +5 repeats", res.ByID[64].Label())
	a.Equal("dblclick(HistoryInner)", res.ByID[63].Token())
	a.Equal("tap(Calls::Group::Panel)>call(group)", res.ByID[66].Token())
	a.Equal("history/history_widget.cpp:5115", res.ByReq[310].Caller)
	a.Equal("chat:user", res.ByReq[311].View.Label())
	a.Equal("media:document", res.ByReq[331].View.Label())
	a.True(res.ByReq[380].View.Idle)
	a.Len(res.Markers, 2)
	a.Equal("click(Dialogs::InnerWidget)>chat(user)", res.ByID[41].Token())
	a.Equal("key(Escape,Media::View::OverlayWidget)>media.close", res.ByID[70].Token())

	key, err := wiretrace.UIKey(tr.Dir)
	a.NoError(err)
	res.JoinPeers(key, []Peer{{Kind: wiretrace.PeerUser, ID: wiretracetest.UIUser}})
	a.Contains(res.Describe(res.ByID[41]), "=user#5 (user)")
}

func TestSemanticLabels(t *testing.T) {
	a := require.New(t)
	for _, tt := range []struct {
		line, label, token string
	}{
		{`{"ev":"ui.box","w":0,"layer":true,"shown":true}`, "box layer", "box(layer)"},
		{`{"ev":"ui.box","w":0,"layer":true,"shown":false}`, "box hidden layer", "box.hide(layer)"},
		{`{"ev":"ui.box","shown":true,"type":"PeerListBox"}`, "box PeerListBox", "box(PeerListBox)"},
		{`{"ev":"ui.call","kind":"call","active":true,"peer":"h:1"}`, "call call active h:1=user#1", "call(call)"},
		{`{"ev":"ui.call","kind":"group","active":false}`, "call group ended", "call.end(group)"},
		{`{"ev":"ui.session","w":0,"open":true}`, "session open", "session(open)"},
		{`{"ev":"ui.session","w":0,"open":false}`, "session closed", "session(closed)"},
		{`{"ev":"ui.account","index":1,"user":"h:2"}`, "account 1 h:2", "account"},
		{`{"ev":"ui.account","user":null}`, "account none", "account"},
		{`{"ev":"ui.folder","w":0,"filter":3}`, "folder 3", "folder"},
		{`{"ev":"ui.search.chat","w":0,"peer":null}`, "search.chat", "search.chat"},
		{`{"ev":"ui.media.open","w":0,"kind":"peer","peer":"h:3","streaming":false}`, "media.open peer h:3", "media.open(peer)"},
		{`{"ev":"app.window","w":0,"state":"minimized"}`, "app.window(minimized)", "app.window(minimized)"},
	} {
		ev, err := wiretrace.ParseEvent([]byte(tt.line))
		a.NoError(err)
		a.Equal(tt.label, SemanticLabel(ev, map[string]string{"h:1": "user#1"}), tt.line)
		a.Equal(tt.token, SemanticToken(ev), tt.line)
	}
}

// TestTDesktopHostFixture decodes the run written by the tdesktop emitter's
// host test (branch trace-ui, Telegram/SourceFiles/trace/test,
// commit 8865722a5a, TestRecords: the writer output plus ui.key), copied
// to testdata/tdesktop-host.
func TestTDesktopHostFixture(t *testing.T) {
	a := require.New(t)
	dir := "testdata/tdesktop-host"
	run, err := wiretrace.Open(dir)
	a.NoError(err)
	defer func() { _ = run.Close() }()
	a.Equal(&wiretrace.HeaderUI{V: 1, IDs: "hmac64"}, run.Header.UI)

	key, err := wiretrace.UIKey(dir)
	a.NoError(err)
	a.Len(key, 32)
	a.Equal(byte(0x1f), key[31])

	events, err := run.All()
	a.NoError(err)
	a.Len(events, 3)
	chat := events[2]
	a.Equal(wiretrace.KindUIChat, chat.Kind)
	a.Equal(&wiretrace.UIStamp{A: 41, Rel: wiretrace.UIRelDeferred}, chat.UI)
	a.True(key.Verify(chat.Str("peer"), wiretrace.PeerValue(wiretrace.PeerUser, 123456789)))
	a.Equal("quote \" and \\ and \n", chat.Str("label"))

	res, err := Build(run, Options{})
	a.NoError(err)
	a.True(res.HasUI)
	a.Len(res.Actions, 1)
	act := res.Actions[0]
	a.Equal(uint32(41), act.A)
	a.Equal(ActionPress, act.Kind)
	a.Len(act.Semantic, 1)
	a.Equal("press(Dialogs::InnerWidget)>chat(user)", act.Token())
	res.JoinPeers(key, []Peer{{Kind: wiretrace.PeerUser, ID: 123456789}})
	a.Equal("press Dialogs::InnerWidget → chat h:a2fb4a2e5d7a21b1=user#123456789 (user)", res.Describe(act))
}
