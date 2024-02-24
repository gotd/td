package obfuscated2

import (
	"encoding/binary"
	"io"
)

// Metadata represents metadata received from header.
type Metadata struct {
	Protocol [4]byte
	DC       uint16
}

// Accept creates new io.ReadWriter for server-side deobfuscation.
func Accept(conn io.ReadWriter, secret []byte) (io.ReadWriter, Metadata, error) {
	var (
		buf  = make([]byte, 64)
		meta Metadata
	)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return nil, meta, err
	}

	var k keys
	if err := k.createStreams(buf, secret); err != nil {
		return nil, meta, err
	}
	// Swap to match client's streams.
	k.encrypt, k.decrypt = k.decrypt, k.encrypt

	var decrypted [64]byte
	k.decrypt.XORKeyStream(decrypted[:], buf)

	copy(meta.Protocol[:], decrypted[56:60])
	meta.DC = binary.LittleEndian.Uint16(decrypted[60:62])

	return &Obfuscated2{
		rand: nil, // Used only in Handshake.
		conn: conn,
		keys: k,
	}, meta, nil
}
