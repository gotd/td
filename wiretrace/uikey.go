package wiretrace

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-faster/errors"
)

// UIKeyName is the file of the per-run HMAC key of UI ids in a run
// directory. It is never written into events.
const UIKeyName = "ui.key"

// HashPrefix starts hashed UI ids: "h:<16 hex>".
const HashPrefix = "h:"

// IDKey is the per-run key of hashed UI ids: HMAC-SHA256 over the decimal
// id, truncated to 64 bits.
type IDKey []byte

// UIKey reads the ui.key of a run directory: 64 lowercase hex chars and a
// newline, as tdesktop writes it (32 raw bytes are also accepted). The error wraps os.ErrNotExist if there is no
// key.
func UIKey(dir string) (IDKey, error) {
	data, err := os.ReadFile(filepath.Join(dir, UIKeyName))
	if err != nil {
		return nil, errors.Wrap(err, "read ui key")
	}
	if text := bytes.TrimSpace(data); len(text) == 64 {
		if k, err := hex.DecodeString(string(text)); err == nil {
			return k, nil
		}
	}
	if len(data) != 32 {
		return nil, errors.Errorf("%s: %d bytes, want 32 (or 64 hex chars)", UIKeyName, len(data))
	}
	return IDKey(data), nil
}

// Hash returns the hashed form "h:<16 hex>" of a decimal id value (for
// peers, the tdesktop PeerId.value, see PeerValue).
func (k IDKey) Hash(value uint64) string {
	m := hmac.New(sha256.New, k)
	_, _ = m.Write([]byte(strconv.FormatUint(value, 10)))
	return HashPrefix + hex.EncodeToString(m.Sum(nil)[:8])
}

// Verify reports whether h is the hashed form of value.
func (k IDKey) Verify(h string, value uint64) bool {
	return len(k) > 0 && strings.EqualFold(h, k.Hash(value))
}

// Peer kinds of PeerValue.
const (
	PeerUser    = "user"
	PeerChat    = "chat"
	PeerChannel = "channel"
)

// PeerValue returns the tdesktop PeerId.value of a bare id: the id with the
// type in bits 48+ (user 0, chat 1, channel 2).
func PeerValue(kind string, id int64) uint64 {
	v := uint64(id) & 0xFFFFFFFFFFFF // #nosec G115
	switch kind {
	case PeerChat:
		v |= 1 << 48
	case PeerChannel:
		v |= 2 << 48
	}
	return v
}
