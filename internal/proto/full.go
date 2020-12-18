package proto

import (
	"context"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"net"
	"sync/atomic"
	"time"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
)

// Full is full MTProto transport.
type Full struct {
	Dialer Dialer
	wSeqNo int64
	rSeqNo int64

	conn net.Conn
}

// check that Full implements Transport in compile time.
var _ Transport = &Full{}

// Dial sends protocol version.
func (i *Full) Dial(ctx context.Context, network, addr string) (err error) {
	i.conn, err = i.Dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return xerrors.Errorf("dial: %w", err)
	}

	atomic.StoreInt64(&i.wSeqNo, 0)
	atomic.StoreInt64(&i.rSeqNo, 0)
	return nil
}

// Send sends message from given buffer.
func (i *Full) Send(ctx context.Context, b *bin.Buffer) error {
	if err := i.conn.SetWriteDeadline(deadline(ctx)); err != nil {
		return xerrors.Errorf("set deadline: %w", err)
	}

	if err := writeFull(i.conn, int(atomic.LoadInt64(&i.wSeqNo)), b); err != nil {
		return xerrors.Errorf("write full: %w", err)
	}
	atomic.StoreInt64(&i.wSeqNo, i.wSeqNo+1)

	if err := i.conn.SetWriteDeadline(time.Time{}); err != nil {
		return xerrors.Errorf("reset connection deadline: %w", err)
	}

	return nil
}

// Recv fills buffer with received message.
func (i *Full) Recv(ctx context.Context, b *bin.Buffer) error {
	if err := i.conn.SetReadDeadline(deadline(ctx)); err != nil {
		return xerrors.Errorf("set deadline: %w", err)
	}

	if err := readFull(i.conn, int(atomic.LoadInt64(&i.rSeqNo)), b); err != nil {
		return xerrors.Errorf("read full: %w", err)
	}
	atomic.StoreInt64(&i.rSeqNo, i.rSeqNo+1)

	if err := i.conn.SetReadDeadline(time.Time{}); err != nil {
		return xerrors.Errorf("reset connection deadline: %w", err)
	}

	if err := checkProtocolError(b); err != nil {
		return err
	}

	return nil
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (i *Full) Close() error {
	return i.conn.Close()
}

func writeFull(w io.Writer, seqNo int, b *bin.Buffer) error {
	if b.Len() > maxMessageSize {
		return errInvalidMsgLen{n: b.Len()}
	}

	write := bin.Buffer{Buf: make([]byte, 0, 4+4+b.Len()+4)}
	// Length: length+seqno+payload+crc length encoded as 4 length bytes
	// (little endian, the length of the length field must be included, too)
	write.PutInt(4 + 4 + b.Len() + 4)
	// Seqno: the TCP sequence number for this TCP connection (different from the MTProto sequence number):
	// the first packet sent is numbered 0, the next one 1, etc.
	write.PutInt(seqNo)
	// payload: MTProto payload
	write.Put(b.Raw())
	// crc: 4 CRC32 bytes computed using length, sequence number, and payload together.
	crc := crc32.ChecksumIEEE(write.Raw())
	write.PutUint32(crc)

	if _, err := w.Write(write.Raw()); err != nil {
		return err
	}

	return nil
}

var errSeqNoMismatch = errors.New("seq_no mismatch")
var errCRCMismatch = errors.New("crc mismatch")

func readFull(r io.Reader, seqNo int, b *bin.Buffer) error {
	n, err := tryReadLength(r, b)
	if err != nil {
		return err
	}

	// Put length, because it need to count CRC.
	b.PutInt(n)
	b.Expand(n - bin.Word)
	inner := &bin.Buffer{Buf: b.Buf[bin.Word:n]}

	// Reads tail of packet to the buffer.
	// Length already read.
	if _, err := io.ReadFull(r, inner.Buf); err != nil {
		return fmt.Errorf("failed to read seqno, buffer and crc: %w", err)
	}

	serverSeqNo, err := inner.Int()
	if err != nil {
		return err
	}
	if serverSeqNo != seqNo {
		return errSeqNoMismatch
	}

	payloadLength := n - 3*bin.Word
	inner.Skip(payloadLength)

	// Cut only crc part.
	crc, err := inner.Uint32()
	if err != nil {
		return err
	}

	// Compute crc using all buffer without last 4 bytes from server.
	clientCRC := crc32.ChecksumIEEE(b.Buf[0 : n-bin.Word])
	// Compare computed and read CRCs.
	if crc != clientCRC {
		return errCRCMismatch
	}

	// 				  n
	// Length | SeqNo | payload | CRC  |
	//  Word  |  Word | ....... | Word |
	copy(b.Buf, b.Buf[2*bin.Word:n-bin.Word])
	b.Buf = b.Buf[:payloadLength]
	return nil
}
