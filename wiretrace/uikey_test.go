package wiretrace

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUIStamp(t *testing.T) {
	a := require.New(t)
	for _, tt := range []struct {
		s    UIStamp
		want string
	}{
		{UIStamp{A: 41, Rel: UIRelSync}, `{"a":41,"rel":"sync"}`},
		{UIStamp{A: 41, Rel: UIRelDeferred}, `{"a":41,"rel":"deferred","dt":0}`},
		{UIStamp{A: 58, Rel: UIRelAfter, DT: 1288.5}, `{"a":58,"rel":"after","dt":1288.5}`},
	} {
		b, err := json.Marshal(tt.s)
		a.NoError(err)
		a.JSONEq(tt.want, string(b))
		var back UIStamp
		a.NoError(json.Unmarshal(b, &back))
		a.Equal(tt.s, back)
	}
}

func TestParseEventUI(t *testing.T) {
	a := require.New(t)
	ev, err := ParseEvent([]byte(`{"seq":9009,"t":12.398702,"th":"main","ev":"req.new","req":310,` +
		`"type":"0x4423e6c5","ui":{"a":41,"rel":"deferred","dt":0}}`))
	a.NoError(err)
	a.Equal(&UIStamp{A: 41, Rel: UIRelDeferred}, ev.UI)

	// hdr.ui is HeaderUI, not a stamp.
	ev, err = ParseEvent([]byte(`{"seq":0,"ev":"hdr","v":1,"seg":0,"ui":{"v":1,"ids":"hmac64","raw_ids":false}}`))
	a.NoError(err)
	a.Nil(ev.UI)
	h, err := ev.Header()
	a.NoError(err)
	a.Equal(&HeaderUI{V: 1, IDs: "hmac64"}, h.UI)

	// A malformed stamp does not hide the event.
	ev, err = ParseEvent([]byte(`{"seq":5,"t":1.5,"th":"main","ev":"req.new","req":1,"ui":"x"}`))
	a.NoError(err)
	a.Nil(ev.UI)
	a.Equal(KindReqNew, ev.Kind)
	a.Equal(1.5, ev.T)
	a.True(ev.Has("ui"))

	// No stamp.
	ev, err = ParseEvent([]byte(`{"seq":5,"t":1.5,"th":"main","ev":"req.new","req":1}`))
	a.NoError(err)
	a.Nil(ev.UI)
}

func TestUIKey(t *testing.T) {
	a := require.New(t)
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	dir := t.TempDir()
	_, err := UIKey(dir)
	a.ErrorIs(err, os.ErrNotExist)

	for _, data := range [][]byte{key, []byte(hex.EncodeToString(key) + "\n")} {
		a.NoError(os.WriteFile(filepath.Join(dir, UIKeyName), data, 0o600))
		k, err := UIKey(dir)
		a.NoError(err)
		a.Equal(IDKey(key), k)
	}
	a.NoError(os.WriteFile(filepath.Join(dir, UIKeyName), []byte("short"), 0o600))
	_, err = UIKey(dir)
	a.Error(err)

	m := hmac.New(sha256.New, key)
	m.Write([]byte("5"))
	want := "h:" + hex.EncodeToString(m.Sum(nil)[:8])
	k := IDKey(key)
	a.Equal(want, k.Hash(5))
	a.Len(k.Hash(5), 18)
	a.True(k.Verify(want, 5))
	a.False(k.Verify(want, 6))
	a.False(IDKey(nil).Verify(want, 5))
}

func TestPeerValue(t *testing.T) {
	a := require.New(t)
	a.Equal(uint64(5), PeerValue(PeerUser, 5))
	a.Equal(uint64(1<<48|5), PeerValue(PeerChat, 5))
	a.Equal(uint64(2<<48|1234567890), PeerValue(PeerChannel, 1234567890))
}
