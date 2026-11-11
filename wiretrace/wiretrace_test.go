package wiretrace

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"
)

func newTestWriter(t *testing.T, opt WriterOptions) (w *Writer, dir string) {
	t.Helper()
	dir = filepath.Join(t.TempDir(), "run")
	now := 0.0
	opt.Now = func() float64 {
		now += 0.5
		return now
	}
	if opt.Header == nil {
		opt.Header = []Field{F("run", "r1"), F("app", "test"), F("layer", 233)}
	}
	w, err := NewWriter(dir, opt)
	require.NoError(t, err)
	return w, dir
}

func TestWriterReaderRoundTrip(t *testing.T) {
	a := require.New(t)
	w, dir := newTestWriter(t, WriterOptions{InlineMax: 4})

	seq, err := w.Write("main", KindReqNew, []Field{
		F("req", 57),
		F("type", TypeID(0xa0f4cb4f)),
		F("cause", Cause{Req: 1, Outer: ID(1 << 63), Type: TypeID(0x1cb5c415)}),
	}, nil)
	a.NoError(err)
	a.Equal(uint64(1), seq)

	_, err = w.WriteMsg("mtp", &Msg{
		Dir: DirOut, SDC: 2, Session: -5521, Salt: 81, MsgID: ID(1<<63 + 5), SeqNo: 7, Len: 8,
		Type: TypeID(0x73f1f8dc), Inner: []Inner{{MsgID: 4, Type: 0xda9b0d0d, Method: 0x1fb33026}},
	}, []byte{1, 2, 3, 4, 5, 6, 7, 8})
	a.NoError(err)
	_, err = w.Write("mtp", "small", nil, []byte{9, 9})
	a.NoError(err)
	_, err = w.Write("mtp", "empty", nil, []byte{})
	a.NoError(err)
	_, err = w.WriteDropped(5)
	a.NoError(err)
	a.NoError(w.Close())

	if runtime.GOOS != "windows" { // no Unix permission bits
		info, err := os.Stat(dir)
		a.NoError(err)
		a.Equal(os.FileMode(0o700), info.Mode().Perm())
		info, err = os.Stat(filepath.Join(dir, EventsName(0)))
		a.NoError(err)
		a.Equal(os.FileMode(0o600), info.Mode().Perm())
	}

	r, err := Open(dir)
	a.NoError(err)
	defer func() { a.NoError(r.Close()) }()
	a.Equal([]int{0}, r.Segments)
	a.Equal("r1", r.Header.Run)
	a.Equal(233, r.Header.Layer)
	a.Equal(Version, r.Header.V)

	events, err := r.All()
	a.NoError(err)
	a.Len(events, 6)
	a.Equal(KindHeader, events[0].Kind)

	req := events[1]
	a.Equal(uint64(1), req.Seq)
	a.Equal(0.5, req.T)
	a.Equal("main", req.Thread)
	a.Equal(int64(57), req.Int("req"))
	a.Equal(uint32(0xa0f4cb4f), req.Type("type"))
	a.Equal(&Cause{Req: 1, Outer: ID(1 << 63), Type: TypeID(0x1cb5c415)}, req.Cause)
	a.Contains(string(req.Raw), `"outer":"9223372036854775808"`)
	a.Contains(string(req.Raw), `"type":"0xa0f4cb4f"`)
	body, err := r.Blob(req)
	a.NoError(err)
	a.Nil(body)

	msg, err := events[2].Msg()
	a.NoError(err)
	a.Equal(SignedID(-5521), msg.Session)
	a.Equal(ID(1<<63+5), msg.MsgID)
	a.Equal(uint64(1<<63+5), events[2].ID("msg_id"))
	a.Equal(TypeID(0x1fb33026), msg.Inner[0].Method)
	a.Contains(string(events[2].Raw), `"session":"-5521"`)
	a.Nil(events[2].Inline, "8 bytes are not inlined with InlineMax 4")
	body, err = r.Blob(events[2])
	a.NoError(err)
	a.Equal([]byte{1, 2, 3, 4, 5, 6, 7, 8}, body)
	a.Equal(int64(BlobFileHeaderSize+BlobRecordHeaderSize), events[2].Blob.Off)

	a.Equal([]byte{9, 9}, events[3].Inline)
	body, err = r.Blob(events[3])
	a.NoError(err)
	a.Equal([]byte{9, 9}, body)

	body, err = r.Blob(events[4])
	a.NoError(err)
	a.Empty(body)
	a.NotNil(events[4].Blob)

	a.Equal(KindDropped, events[5].Kind)
	a.Equal("trace", events[5].Thread)
	a.Equal(uint64(5), events[5].ID("count"))
}

func TestBlobVerification(t *testing.T) {
	a := require.New(t)
	w, dir := newTestWriter(t, WriterOptions{})
	_, err := w.Write("t1", KindMsg, []Field{F("dir", DirIn)}, []byte("0123456789abcdef"))
	a.NoError(err)
	a.NoError(w.Close())

	path := filepath.Join(dir, BlobsName(0))
	data, err := os.ReadFile(path)
	a.NoError(err)
	a.Equal(BlobMagic, string(data[:4]))
	a.Equal(uint16(Version), binary.LittleEndian.Uint16(data[4:]))
	a.Equal(uint64(1), binary.LittleEndian.Uint64(data[6:]))

	// Flip a body byte: crc32c fails.
	data[len(data)-1] ^= 0xff
	a.NoError(os.WriteFile(path, data, 0o600))
	r, err := Open(dir)
	a.NoError(err)
	events, err := r.All()
	a.NoError(err)
	_, err = r.Blob(events[1])
	a.ErrorIs(err, ErrBlobMismatch)
	a.NoError(r.Close())

	// Bad magic.
	data[0] = 'X'
	a.NoError(os.WriteFile(path, data, 0o600))
	r, err = Open(dir)
	a.NoError(err)
	_, err = r.Blob(events[1])
	a.Error(err)
	a.NoError(r.Close())
}

func TestVerifyBlobSHA256(t *testing.T) {
	body := []byte("body")
	var hdr [BlobRecordHeaderSize]byte
	binary.LittleEndian.PutUint64(hdr[:], 3)
	binary.LittleEndian.PutUint32(hdr[8:], uint32(len(body)))
	binary.LittleEndian.PutUint32(hdr[12:], CRC32C(body))
	ref := &BlobRef{Len: len(body), CRC32C: CRC32C(body), SHA256: "00"}
	require.ErrorIs(t, VerifyBlob(3, ref, hdr[:], body), ErrBlobMismatch)
	require.ErrorIs(t, VerifyBlob(4, &BlobRef{Len: len(body)}, hdr[:], body), ErrBlobMismatch)
	require.NoError(t, VerifyBlob(3, &BlobRef{Len: len(body)}, hdr[:], body))
}

func TestCRC32C(t *testing.T) {
	// Standard check value of CRC-32C.
	require.Equal(t, uint32(0xe3069283), CRC32C([]byte("123456789")))
}

func TestRotation(t *testing.T) {
	a := require.New(t)
	w, dir := newTestWriter(t, WriterOptions{
		SegmentEventsBytes: 200,
		SegmentBlobsBytes:  1 << 20,
		TotalBytes:         1100,
	})
	blob := make([]byte, 100)
	for i := 0; i < 20; i++ {
		_, err := w.Write("t1", "x", []Field{F("i", i)}, blob)
		a.NoError(err)
	}
	a.NoError(w.Close())

	r, err := Open(dir)
	a.NoError(err)
	defer func() { a.NoError(r.Close()) }()
	a.Equal(0, r.Segments[0], "segment 0 is never deleted")
	a.Greater(len(r.Segments), 1)
	a.Greater(r.Segments[1], 1, "oldest segments after 0 are deleted")

	events, err := r.All()
	a.NoError(err)
	var headers, last int
	for _, ev := range events {
		if ev.Kind == KindHeader {
			headers++
			h, err := ev.Header()
			a.NoError(err)
			a.Equal(ev.Segment, h.Seg)
			continue
		}
		body, err := r.Blob(ev)
		a.NoError(err)
		a.Len(body, 100)
		a.Equal(ev.Segment, ev.Blob.Seg)
		last = int(ev.Int("i"))
	}
	a.Equal(len(r.Segments), headers)
	a.Equal(19, last)
}

func TestRepairLegacyNumbers(t *testing.T) {
	for _, tc := range []struct {
		in, out string
		changed bool
	}{
		{`{"seq":1,"t":12,345678,"th":"main"}`, `{"seq":1,"t":12.345678,"th":"main"}`, true},
		{`{"lat_ms":186,900}`, `{"lat_ms":186.900}`, true},
		{`{"a":[1],"mono_t":-0,500]`, `{"a":[1],"mono_t":-0.500]`, true},
		{`{"t":1,5,"x":2,25}`, `{"t":1.5,"x":2.25}`, true},
		{`{"seq":1,"t":2}`, `{"seq":1,"t":2}`, false},
		{`{"s":"a:1,2}","t":3}`, `{"s":"a:1,2}","t":3}`, false},
		{`{"s":"q\":1,2}","t":3,4}`, `{"s":"q\":1,2}","t":3.4}`, true},
		{`{"seq":12,"x":"y"}`, `{"seq":12,"x":"y"}`, false},
	} {
		got, changed := RepairLegacyNumbers([]byte(tc.in))
		require.Equal(t, tc.out, string(got), tc.in)
		require.Equal(t, tc.changed, changed, tc.in)
	}
}

func TestParseEventLegacyLocale(t *testing.T) {
	a := require.New(t)
	line := `{"seq":901,"t":1,391002,"th":"main","ev":"req.done","req":57,"type":"0xa0f4cb4f","outer":"7449","lat_ms":186,900,` +
		`"len":4,"cause":{"req":1,"outer":"2","type":"0x1cb5c415"}}`
	ev, err := ParseEvent([]byte(line))
	a.NoError(err)
	a.Equal(1.391002, ev.T)
	a.Equal(186.9, ev.Float("lat_ms"))
	a.Equal(int64(4), ev.Int("len"))
	a.Equal(uint64(7449), ev.ID("outer"))
	a.Equal(int64(1), ev.Cause.Req)

	_, err = ParseEvent([]byte(`{"seq":1,`))
	a.Error(err)
}

func TestLegacyLocaleRun(t *testing.T) {
	a := require.New(t)
	dir := t.TempDir()
	events := `{"seq":0,"ev":"hdr","v":1,"seg":0,"layer":233,"mono_t":0,001000,"flags":[]}
{"seq":1,"t":0,005232,"th":"main","ev":"req.done","req":5,"lat_ms":12,500}
`
	a.NoError(os.WriteFile(filepath.Join(dir, EventsName(0)), []byte(events), 0o600))
	r, err := Open(dir)
	a.NoError(err)
	a.Equal(0.001, r.Header.MonoT)
	all, err := r.All()
	a.NoError(err)
	a.Len(all, 2)
	a.Equal(0.005232, all[1].T)
	a.Equal(12.5, all[1].Float("lat_ms"))
}

func TestOpenErrors(t *testing.T) {
	_, err := Open(t.TempDir())
	require.Error(t, err)

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, EventsName(0)), []byte(`{"seq":1,"ev":"msg"}`+"\n"), 0o600))
	_, err = Open(dir)
	require.Error(t, err)
}

func TestBadLines(t *testing.T) {
	a := require.New(t)
	dir := t.TempDir()
	events := `{"seq":0,"ev":"hdr","v":1,"seg":0}
not json
{"seq":1,"t":1,"th":"main","ev":"x"}
`
	a.NoError(os.WriteFile(filepath.Join(dir, EventsName(0)), []byte(events), 0o600))
	r, err := Open(dir)
	a.NoError(err)
	var bad []int
	var kinds []string
	a.NoError(r.Events(func(ev *Event) error {
		kinds = append(kinds, ev.Kind)
		return nil
	}, func(segment, line int, err error) {
		bad = append(bad, line)
	}))
	a.Equal([]int{2}, bad)
	a.Equal([]string{KindHeader, "x"}, kinds)

	stop := errors.New("stop")
	a.ErrorIs(r.Events(func(ev *Event) error { return stop }, nil), stop)
}

func TestTypeID(t *testing.T) {
	v, err := ParseTypeID("0x1cb5c415")
	require.NoError(t, err)
	require.Equal(t, "0x1cb5c415", v.String())
	_, err = ParseTypeID("zz")
	require.Error(t, err)

	var s SignedID
	require.NoError(t, s.UnmarshalJSON([]byte(`"18446744073709551615"`)))
	require.Equal(t, SignedID(-1), s)
}
