package codec

import (
	"hash/crc32"
	"io"
	"sync/atomic"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
)

// Full is full MTProto transport.
//
// See https://core.telegram.org/mtproto/mtproto-transports#full
type Full struct {
	wSeqNo int64
	rSeqNo int64
}

// WriteHeader sends protocol tag.
func (i *Full) WriteHeader(w io.Writer) (err error) {
	return nil
}

// ReadHeader reads protocol tag.
func (i *Full) ReadHeader(r io.Reader) (err error) {
	return nil
}

// Write encode to writer message from given buffer.
func (i *Full) Write(w io.Writer, b *bin.Buffer) error {
	if err := checkOutgoingMessage(b); err != nil {
		return err
	}

	if err := writeFull(w, int(atomic.AddInt64(&i.wSeqNo, 1)-1), b); err != nil {
		return errors.Wrap(err, "write full")
	}

	return nil
}

// Read fills buffer with received message.
func (i *Full) Read(r io.Reader, b *bin.Buffer) error {
	if err := readFull(r, int(atomic.AddInt64(&i.rSeqNo, 1)-1), b); err != nil {
		return errors.Wrap(err, "read full")
	}

	return checkProtocolError(b)
}

func writeFull(w io.Writer, seqNo int, b *bin.Buffer) error {
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
	n, err := readLen(r, b)
	if err != nil {
		return errors.Wrap(err, "len")
	}

	// Put length, because it need to count CRC.
	b.PutInt(n)
	b.Expand(n - bin.Word)
	inner := &bin.Buffer{Buf: b.Buf[bin.Word:n]}

	// Reads tail of packet to the buffer.
	// Length already read.
	if _, err := io.ReadFull(r, inner.Buf); err != nil {
		return errors.Wrap(err, "read seqno, buffer and crc")
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
