package calls

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func testKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, keySize)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	return key
}

// TestSignalingRoundTrip verifies that a packet sealed by one peer is opened by
// the other in both directions, exercising the direction-dependent key offset.
func TestSignalingRoundTrip(t *testing.T) {
	key := testKey(t)
	caller := newSignalingEncryption(key, true)
	callee := newSignalingEncryption(key, false)

	for _, dir := range []struct {
		name     string
		from, to *signalingEncryption
	}{
		{"caller->callee", caller, callee},
		{"callee->caller", callee, caller},
	} {
		t.Run(dir.name, func(t *testing.T) {
			payload := []byte(`{"@type":"MediaState","videoState":"active"}`)
			packet, err := dir.from.encryptMessage(payload)
			if err != nil {
				t.Fatalf("encrypt: %v", err)
			}
			msgs, err := dir.to.decryptMessages(packet)
			if err != nil {
				t.Fatalf("decrypt: %v", err)
			}
			if len(msgs) != 1 || !bytes.Equal(msgs[0], payload) {
				t.Fatalf("got %q, want %q", msgs, payload)
			}
		})
	}
}

func TestSignalingRejectsReplay(t *testing.T) {
	key := testKey(t)
	caller := newSignalingEncryption(key, true)
	callee := newSignalingEncryption(key, false)

	packet, err := caller.encryptMessage([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := callee.decryptMessages(packet); err != nil {
		t.Fatalf("first decrypt: %v", err)
	}
	if _, err := callee.decryptMessages(packet); err == nil {
		t.Fatal("replay was accepted")
	}
}

func TestSignalingWrongDirectionFails(t *testing.T) {
	key := testKey(t)
	caller := newSignalingEncryption(key, true)
	// A peer that thinks it is also the caller uses the wrong key offset and
	// must fail the msg_key check.
	wrong := newSignalingEncryption(key, true)

	packet, err := caller.encryptMessage([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := wrong.decryptMessages(packet); err == nil {
		t.Fatal("packet decrypted with wrong direction")
	}
}

func TestSignalingAckQueue(t *testing.T) {
	key := testKey(t)
	caller := newSignalingEncryption(key, true)
	callee := newSignalingEncryption(key, false)

	packet, err := caller.encryptMessage([]byte("needs-ack"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := callee.decryptMessages(packet); err != nil {
		t.Fatal(err)
	}
	acks := callee.drainAcks()
	if len(acks) != 1 {
		t.Fatalf("queued acks = %d, want 1", len(acks))
	}
	if counterFromSeq(acks[0]) != 1 {
		t.Fatalf("ack counter = %d, want 1", counterFromSeq(acks[0]))
	}
	if drained := callee.drainAcks(); drained != nil {
		t.Fatal("acks not cleared after drain")
	}

	// The ack packet itself must be openable by the caller and carry no payload.
	ackPacket, err := callee.encryptAcks(acks)
	if err != nil {
		t.Fatal(err)
	}
	msgs, err := caller.decryptMessages(ackPacket)
	if err != nil {
		t.Fatalf("decrypt ack: %v", err)
	}
	if len(msgs) != 0 {
		t.Fatalf("ack packet carried %d payloads, want 0", len(msgs))
	}
}

func TestSignalingShortPacket(t *testing.T) {
	s := newSignalingEncryption(testKey(t), false)
	if _, err := s.decryptMessages(make([]byte, 10)); err == nil {
		t.Fatal("short packet accepted")
	}
}
