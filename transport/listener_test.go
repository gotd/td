package transport

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/internal/proto/codec"
	"github.com/nnqq/td/internal/testutil"
)

type mockListener struct {
	connData  []byte
	acceptErr error
	addr      net.Addr
}

type mockConn struct {
	reader bytes.Reader
	closed bool
}

func (m *mockConn) Read(b []byte) (int, error) {
	return m.reader.Read(b)
}

func (m *mockConn) Write(b []byte) (int, error) {
	return 0, io.ErrClosedPipe
}

func (m *mockConn) Close() error {
	m.closed = true
	return nil
}

func (m *mockConn) LocalAddr() net.Addr {
	return nil
}

func (m *mockConn) RemoteAddr() net.Addr {
	return nil
}

func (m *mockConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *mockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *mockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (m mockListener) Accept() (net.Conn, error) {
	return &mockConn{
		reader: *bytes.NewReader(m.connData),
		closed: false,
	}, m.acceptErr
}

func (m mockListener) Close() error {
	return nil
}

func (m mockListener) Addr() net.Addr {
	return m.addr
}

func TestListener_Accept(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		codec   func() Codec
		wantErr bool
	}{
		{"DetectCodec", codec.PaddedIntermediateClientStart[:], nil, false},
		{"PassCodec", codec.AbridgedClientStart[:], Abridged.Codec, false},
		{"InvalidCodec", codec.PaddedIntermediateClientStart[:], Abridged.Codec, true},
		{"FirstByteError", nil, nil, true},
		{"HeaderError", make([]byte, 3), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := require.New(t)
			m := mockListener{
				connData: tt.data,
			}

			l := ListenCodec(tt.codec, m)
			defer func() {
				a.NoError(l.Close())
			}()

			conn, err := l.Accept()
			if tt.wantErr {
				a.Error(err)
				if c, ok := conn.(*connection); ok {
					a.True(c.conn.(*mockConn).closed)
				}
			} else {
				a.NoError(err)
			}
		})
	}

	t.Run("AcceptError", func(t *testing.T) {
		e := testutil.TestError()
		m := Listener{
			listener: mockListener{
				acceptErr: e,
			},
		}

		_, err := m.Accept()
		require.ErrorIs(t, err, e)
	})
}

func TestListener_Addr(t *testing.T) {
	addr := &net.TCPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 443,
		Zone: "",
	}

	l := Listener{listener: mockListener{addr: addr}}
	require.Equal(t, addr, l.Addr())
}
