package mtproto

import (
	"bytes"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/log"
	"github.com/gotd/neo"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/proto"
	"github.com/gotd/td/testutil"
)

type tracedMessage struct {
	m    WireMessage
	body []byte
}

// wireRecorder is a WireTracer that records everything.
type wireRecorder struct {
	mux    sync.Mutex
	msgs   []tracedMessage
	events []WireEvent
}

func (r *wireRecorder) OnMessage(m WireMessage, body []byte) {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.msgs = append(r.msgs, tracedMessage{m: m, body: append([]byte(nil), body...)})
}

func (r *wireRecorder) OnEvent(e WireEvent) {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.events = append(r.events, e)
}

func newWireTestConn(t testing.TB, tracer WireTracer, compressThreshold int) *Conn {
	t.Helper()
	c := &Conn{
		dcID:              2,
		rand:              Zero{},
		log:               log.For(log.Nop),
		cipher:            crypto.NewClientCipher(Zero{}),
		clock:             neo.NewTime(testutil.Date()),
		messageIDBuf:      noopBuf{},
		compressThreshold: compressThreshold,
		sessionID:         42,
		salt:              7,
		wire:              tracer,
	}
	for i := 0; i < 256; i++ {
		c.authKey.Value[i] = byte(i)
	}
	c.authKey = c.authKey.Value.WithID()
	if tracer != nil {
		c.wireID = wireConnID.Inc()
	}
	return c
}

func TestWireTracerOutgoing(t *testing.T) {
	for _, tt := range []struct {
		name      string
		threshold int
		size      int
		gzip      bool
	}{
		{"NoCompression", -1, 2048, false},
		{"Small", 1024, 16, false},
		{"Gzip", 1024, 2048, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			a := require.New(t)
			rec := &wireRecorder{}
			c := newWireTestConn(t, rec, tt.threshold)

			payload := testPayload{Data: bytes.Repeat([]byte{1}, tt.size)}
			var want bin.Buffer
			a.NoError(payload.Encode(&want))

			var b bin.Buffer
			a.NoError(c.newEncryptedMessage(1001, 3, payload, &b))

			a.Len(rec.msgs, 1)
			m := rec.msgs[0]
			a.Equal(WireOut, m.m.Direction)
			a.Equal(int64(1001), m.m.MessageID)
			a.Equal(int32(3), m.m.SeqNo)
			a.Equal(int64(42), m.m.SessionID)
			a.Equal(int64(7), m.m.Salt)
			a.Equal(c.authKey.ID, m.m.AuthKeyID)
			a.Equal(WireConn{ID: c.wireID, DC: 2}, m.m.Conn)

			body := m.body
			if tt.gzip {
				var g proto.GZIP
				a.NoError(g.Decode(&bin.Buffer{Buf: append([]byte(nil), body...)}))
				body = g.Data
			}
			a.Equal(want.Raw(), body)

			// The traced body is exactly what is encrypted.
			server := crypto.NewServerCipher(Zero{})
			got, err := server.DecryptFromBuffer(c.authKey, &b)
			a.NoError(err)
			a.Equal(m.body, got.Data())
		})
	}
}

func TestWireTracerIncoming(t *testing.T) {
	a := require.New(t)
	rec := &wireRecorder{}
	c := newWireTestConn(t, rec, -1)
	server := crypto.NewServerCipher(Zero{})

	encrypt := func(session, msgID int64, body []byte) *bin.Buffer {
		var b bin.Buffer
		a.NoError(server.Encrypt(c.authKey, crypto.EncryptedMessageData{
			Salt:                   9,
			SessionID:              session,
			MessageID:              msgID,
			SeqNo:                  1,
			MessageDataLen:         int32(len(body)),
			MessageDataWithPadding: append([]byte(nil), body...),
		}, &b))
		return &b
	}

	var container bin.Buffer
	a.NoError((&proto.MessageContainer{Messages: []proto.Message{
		{ID: 1, SeqNo: 1, Bytes: 8, Body: []byte{1, 2, 3, 4, 5, 6, 7, 8}},
	}}).Encode(&container))

	id := proto.NewMessageID(c.clock.Now(), proto.MessageServerResponse)
	msg, err := c.decryptMessage(encrypt(42, int64(id), container.Raw()))
	a.NoError(err)
	a.Equal(container.Raw(), msg.Data())

	// Rejected: wrong session.
	_, err = c.decryptMessage(encrypt(43, int64(id)+4, []byte{1, 2, 3, 4}))
	a.ErrorIs(err, errRejected)

	// Rejected: bad message id.
	_, err = c.decryptMessage(encrypt(42, 1, []byte{1, 2, 3, 4}))
	a.ErrorIs(err, errRejected)

	// Not decryptable.
	bad := encrypt(42, int64(id)+8, []byte{1, 2, 3, 4})
	bad.Buf[len(bad.Buf)-1] ^= 0xff
	_, err = c.decryptMessage(bad)
	a.Error(err)

	a.Len(rec.msgs, 3)
	for _, m := range rec.msgs {
		a.Equal(WireIn, m.m.Direction)
		a.Equal(int64(9), m.m.Salt)
		a.Equal(c.authKey.ID, m.m.AuthKeyID)
	}
	a.Equal(container.Raw(), rec.msgs[0].body)
	a.Equal(int64(id), rec.msgs[0].m.MessageID)
	a.Equal(int64(43), rec.msgs[1].m.SessionID)

	a.Len(rec.events, 3)
	a.Equal(WireRejected, rec.events[0].Type)
	a.Equal("bad_session", rec.events[0].Reason)
	a.Equal(WireRejected, rec.events[1].Type)
	a.Equal("bad_msg_id", rec.events[1].Reason)
	a.Equal(int64(1), rec.events[1].MessageID)
	a.Equal(WireDecryptFailed, rec.events[2].Type)
	a.Error(rec.events[2].Err)
}

func TestWireTracerSaltSession(t *testing.T) {
	a := require.New(t)
	rec := &wireRecorder{}
	c := newWireTestConn(t, rec, -1)
	c.storeSalt(7) // unchanged
	c.storeSalt(8)
	a.NoError(c.newSessionID())

	a.Len(rec.events, 2)
	a.Equal(WireSaltChange, rec.events[0].Type)
	a.Equal(int64(7), rec.events[0].Old)
	a.Equal(int64(8), rec.events[0].New)
	a.Equal(WireSessionNew, rec.events[1].Type)
	a.Equal(int64(42), rec.events[1].Old)
	a.Equal(c.sessionID, rec.events[1].New)
}

func TestWireTracerTempKey(t *testing.T) {
	a := require.New(t)
	rec := &wireRecorder{}
	c := newWireTestConn(t, rec, -1)
	c.pfs = true
	c.tempKeyExpiry = 1700000000

	c.traceKey(WireKeyCreated, c.authKey, true)
	var b bin.Buffer
	a.NoError(c.newEncryptedMessage(1001, 1, testPayload{Data: []byte{1}}, &b))

	a.Len(rec.events, 1)
	a.Equal(WireKeyCreated, rec.events[0].Type)
	a.True(rec.events[0].TempKey)
	a.Equal(int64(1700000000), rec.events[0].ExpiresAt)
	a.Equal(c.authKey.ID, rec.events[0].AuthKeyID)
	a.Len(rec.msgs, 1)
	a.True(rec.msgs[0].m.TempKey)
}

func TestWireTracerDisabledAllocs(t *testing.T) {
	payload := testPayload{Data: make([]byte, 128)}
	for _, tt := range []struct {
		threshold int
		allocs    int
	}{
		// Allocations of the cipher and payload encoding, the same as
		// without the tracer hooks.
		{threshold: -1, allocs: 5},
		{threshold: 1024, allocs: 12},
	} {
		c := newWireTestConn(t, nil, tt.threshold)
		buf := &bin.Buffer{Buf: make([]byte, 0, 1024)}
		encrypt := func() {
			buf.Reset()
			if err := c.newEncryptedMessage(12345, 0, payload, buf); err != nil {
				t.Fatal(err)
			}
		}
		encrypt()
		testutil.MaxAlloc(t, tt.allocs, encrypt)
	}
}

func TestWireEventTypeString(t *testing.T) {
	a := require.New(t)
	a.Equal("request_done", WireRequestDone.String())
	a.Equal("unknown", WireEventType(0).String())
	a.Equal("unknown", WireEventType(200).String())
	a.Equal("out", WireOut.String())
	a.Equal("in", WireIn.String())
	a.Equal("unknown", WireDirection(0).String())
}
