package proto

import "github.com/ernado/td/bin"

// The Intermediate version of MTproto.
//
// See https://core.telegram.org/mtproto/mtproto-transports#intermediate

// IntermediateClientStart is starting bytes sent by client in Intermediate mode.
//
// Note that server does not respond with it.
var IntermediateClientStart = []byte{0xee, 0xee, 0xee, 0xee}

// EncodeIntermediate encodes payload to b via Intermediate protocol.
func EncodeIntermediate(b *bin.Buffer, payload []byte) {
	b.PutInt32(int32(len(payload)))
	b.Put(payload)
}
