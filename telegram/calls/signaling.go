package calls

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"sync"

	"github.com/go-faster/errors"
)

// Signaling record type tags and sequence-number flag bits, matching the
// tgcalls EncryptedConnection wire format.
const (
	sigCustomID byte = 0x7f // 127: a custom (JSON) payload follows.
	sigAckID    byte = 0xff // -1:  an acknowledgement of a previous seq.
	sigEmptyID  byte = 0xfe // -2:  empty record, used as an ack carrier head.

	sigSingleMessageBit = uint32(1) << 31
	sigRequiresAckBit   = uint32(1) << 30
	sigCounterMask      = ^(sigSingleMessageBit | sigRequiresAckBit)
)

// signalingEncryption is the encrypted, sequenced control channel of a call,
// the Go equivalent of tgcalls' SignalingEncryption/EncryptedConnection: each
// packet is [msg_key(16) || AES-CTR(body)], where body starts with a 32-bit
// sequence number followed by typed records.
//
// The AES key/IV and msg_key are derived from the 256-byte DH shared key with a
// direction-dependent offset x (see encryptRawPacket/decryptRawPacket).
type signalingEncryption struct {
	key        []byte
	isOutgoing bool

	mu     sync.Mutex
	outSeq uint32
	seen   map[uint32]struct{}
	toAck  []uint32
}

func newSignalingEncryption(key []byte, isOutgoing bool) *signalingEncryption {
	return &signalingEncryption{
		key:        key,
		isOutgoing: isOutgoing,
		seen:       make(map[uint32]struct{}),
	}
}

func counterFromSeq(seq uint32) uint32 { return seq & sigCounterMask }

// encryptMessage seals a single JSON control message that requires an ack.
func (s *signalingEncryption) encryptMessage(payload []byte) ([]byte, error) {
	s.mu.Lock()
	s.outSeq++
	seq := s.outSeq | sigRequiresAckBit
	s.mu.Unlock()

	body := make([]byte, 0, 4+1+4+len(payload))
	body = binary.BigEndian.AppendUint32(body, seq)
	body = append(body, sigCustomID)
	body = binary.BigEndian.AppendUint32(body, uint32(len(payload)))
	body = append(body, payload...)
	return s.encryptRawPacket(body)
}

// encryptAcks seals a packet acknowledging the given received sequence numbers.
func (s *signalingEncryption) encryptAcks(seqs []uint32) ([]byte, error) {
	if len(seqs) == 0 {
		return nil, nil
	}
	s.mu.Lock()
	s.outSeq++
	head := s.outSeq
	s.mu.Unlock()

	body := make([]byte, 0, 5+len(seqs)*5)
	body = binary.BigEndian.AppendUint32(body, head)
	body = append(body, sigEmptyID)
	for _, seq := range seqs {
		body = binary.BigEndian.AppendUint32(body, seq)
		body = append(body, sigAckID)
	}
	return s.encryptRawPacket(body)
}

// drainAcks returns and clears the pending acknowledgements.
func (s *signalingEncryption) drainAcks() []uint32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := s.toAck
	s.toAck = nil
	return out
}

// decryptMessages opens a packet and returns the custom payloads it carried,
// queueing acks for those that requested them.
func (s *signalingEncryption) decryptMessages(packet []byte) ([][]byte, error) {
	body, err := s.decryptRawPacket(packet)
	if err != nil {
		return nil, err
	}

	var out [][]byte
	seq := binary.BigEndian.Uint32(body[:4])
	single := seq&sigSingleMessageBit != 0
	pos := 4
	for pos < len(body) {
		typ := body[pos]
		pos++
		switch typ {
		case sigEmptyID, sigAckID:
			// Empty/ack records carry no payload here.
		case sigCustomID:
			if pos+4 > len(body) {
				return nil, errors.New("signaling: truncated custom length")
			}
			n := int(binary.BigEndian.Uint32(body[pos : pos+4]))
			pos += 4
			if n < 0 || pos+n > len(body) {
				return nil, errors.New("signaling: truncated custom payload")
			}
			out = append(out, append([]byte(nil), body[pos:pos+n]...))
			pos += n
			if seq&sigRequiresAckBit != 0 {
				s.mu.Lock()
				s.toAck = append(s.toAck, seq)
				s.mu.Unlock()
			}
		default:
			return nil, errors.Errorf("signaling: unknown record 0x%02x", typ)
		}
		if pos >= len(body) {
			break
		}
		if single {
			return nil, errors.New("signaling: trailing data in single-message packet")
		}
		if pos+4 > len(body) {
			return nil, errors.New("signaling: truncated trailing seq")
		}
		seq = binary.BigEndian.Uint32(body[pos : pos+4])
		pos += 4
	}
	return out, nil
}

// encryptRawPacket encrypts a body, prepending the 16-byte msg_key.
func (s *signalingEncryption) encryptRawPacket(body []byte) ([]byte, error) {
	x := s.offset(true)
	msgKey := s.messageKey(body, x)
	stream, err := s.ctrStream(msgKey, x)
	if err != nil {
		return nil, err
	}
	out := make([]byte, 16+len(body))
	copy(out[:16], msgKey)
	stream.XORKeyStream(out[16:], body)
	return out, nil
}

// decryptRawPacket decrypts a packet, verifying the msg_key and rejecting replays.
func (s *signalingEncryption) decryptRawPacket(packet []byte) ([]byte, error) {
	if len(packet) < 21 {
		return nil, errors.Errorf("signaling: packet too short (%d)", len(packet))
	}
	x := s.offset(false)
	msgKey := packet[:16]
	stream, err := s.ctrStream(msgKey, x)
	if err != nil {
		return nil, err
	}
	body := make([]byte, len(packet)-16)
	stream.XORKeyStream(body, packet[16:])

	if subtle.ConstantTimeCompare(s.messageKey(body, x), msgKey) != 1 {
		return nil, errors.New("signaling: msg_key mismatch")
	}

	counter := counterFromSeq(binary.BigEndian.Uint32(body[:4]))
	s.mu.Lock()
	_, dup := s.seen[counter]
	if !dup {
		s.seen[counter] = struct{}{}
	}
	s.mu.Unlock()
	if dup {
		return nil, errors.Errorf("signaling: duplicate counter %d", counter)
	}
	return body, nil
}

// offset returns the key offset x for the given direction. For encryption the
// offset depends on whether we are the caller; for decryption it is mirrored,
// so that each side decrypts with the same x the sender encrypted with.
func (s *signalingEncryption) offset(encrypt bool) int {
	const signalingBase = 128
	outgoing := s.isOutgoing
	if !encrypt {
		outgoing = !outgoing
	}
	if outgoing {
		return signalingBase
	}
	return signalingBase + 8
}

// messageKey returns the 16-byte msg_key: bytes [8:24] of
// SHA256(key[88+x:88+x+32] || body).
func (s *signalingEncryption) messageKey(body []byte, x int) []byte {
	h := sha256.New()
	h.Write(s.key[88+x : 88+x+32])
	h.Write(body)
	return h.Sum(nil)[8:24]
}

func (s *signalingEncryption) ctrStream(msgKey []byte, x int) (cipher.Stream, error) {
	key, iv := s.prepareAESKeyIV(msgKey, x)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err, "aes cipher")
	}
	return cipher.NewCTR(block, iv), nil
}

// prepareAESKeyIV derives the AES-256 key and IV from msg_key and the shared
// key, following the MTProto-style KDF used by tgcalls (PrepareAesKeyIv).
func (s *signalingEncryption) prepareAESKeyIV(msgKey []byte, x int) (key, iv []byte) {
	a := sha256.New()
	a.Write(msgKey[:16])
	a.Write(s.key[x : x+36])
	sha256a := a.Sum(nil)

	b := sha256.New()
	b.Write(s.key[40+x : 40+x+36])
	b.Write(msgKey[:16])
	sha256b := b.Sum(nil)

	key = make([]byte, 0, 32)
	key = append(key, sha256a[0:8]...)
	key = append(key, sha256b[8:24]...)
	key = append(key, sha256a[24:32]...)

	iv = make([]byte, 0, 16)
	iv = append(iv, sha256b[0:4]...)
	iv = append(iv, sha256a[8:16]...)
	iv = append(iv, sha256b[24:28]...)
	return key, iv
}
