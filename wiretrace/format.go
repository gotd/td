package wiretrace

import (
	"fmt"
	"hash/crc32"
	"strconv"
)

// Format constants of a trace run directory.
const (
	// Version is the format version written in "hdr" records and in the blob
	// file header.
	Version = 1

	// BlobMagic starts every blobs-NNNN.bin file, followed by u16 Version.
	BlobMagic = "TDWT"
	// BlobFileHeaderSize is the size of the blob file header.
	BlobFileHeaderSize = 6
	// BlobRecordHeaderSize is the size of a blob record header:
	// u64 seq, u32 len, u32 crc32c (little-endian).
	BlobRecordHeaderSize = 16

	// RedactedMarker is written right after the constructor id of a redacted
	// TL object: [id][RedactedMarker][id][zeros...]. Lengths are kept.
	RedactedMarker uint32 = 0x7EDAC7ED
)

// Event kinds written by the tdesktop tracer.
const (
	KindHeader  = "hdr"
	KindDropped = "dropped"
	KindMsg     = "msg"

	KindReqNew     = "req.new"
	KindReqMsg     = "req.msg"
	KindReqResend  = "req.resend"
	KindReqDone    = "req.done"
	KindReqFail    = "req.fail"
	KindReqCancel  = "req.cancel"
	KindReqCaller  = "req.caller"
	KindReqMigrate = "req.migrate"

	KindState         = "state"
	KindSessionReset  = "session.reset"
	KindSessionNewID  = "session.new_id"
	KindSaltChange    = "salt.change"
	KindKeySet        = "key.set"
	KindKeyCreated    = "key.created"
	KindKeyDestroy    = "key.destroy_temp"
	KindConnStart     = "conn.start"
	KindConnUp        = "conn.up"
	KindConnDown      = "conn.down"
	KindConnReason    = "conn.reason"
	KindConnRestart   = "conn.restart"
	KindConfigLoaded  = "config.loaded"
	KindStorageWrite  = "st.write"
	KindStorageRaw    = "st.write.raw"
	KindStorageEnc    = "st.encrypt"
	KindStorageCommit = "st.commit"
	KindStorageRemove = "st.remove"
	KindStorageRead   = "st.read"
	KindStorageDec    = "st.decrypt"
)

// Directions of "msg" events.
const (
	DirOut = "out"
	DirIn  = "in"
)

var castagnoli = crc32.MakeTable(crc32.Castagnoli)

// CRC32C computes the crc32c (Castagnoli) checksum used by blob records.
func CRC32C(data []byte) uint32 {
	return crc32.Checksum(data, castagnoli)
}

// BlobRef references a body in a blobs-NNNN.bin segment.
//
// Off is the offset of the body itself, after the 16-byte record header.
type BlobRef struct {
	Seg    int    `json:"seg"`
	Off    int64  `json:"off"`
	Len    int    `json:"len"`
	CRC32C uint32 `json:"crc32c"`
	SHA256 string `json:"sha256,omitempty"`
}

// Cause is the causal context of an event emitted while a response or an
// update was being handled.
type Cause struct {
	// Req is the request id of the response, 0 for updates.
	Req int64 `json:"req"`
	// Outer is the outer msg_id of the response.
	Outer ID `json:"outer"`
	// Type is the constructor id of the response.
	Type TypeID `json:"type"`
}

// ID is a 64-bit identifier, encoded as a decimal JSON string.
//
// Decoding also accepts JSON numbers.
type ID uint64

// MarshalJSON implements json.Marshaler.
func (i ID) MarshalJSON() ([]byte, error) {
	return []byte(`"` + strconv.FormatUint(uint64(i), 10) + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (i *ID) UnmarshalJSON(data []byte) error {
	v, err := parseUint(data)
	if err != nil {
		return err
	}
	*i = ID(v)
	return nil
}

// SignedID is a signed 64-bit identifier (session ids), encoded as a decimal
// JSON string.
type SignedID int64

// MarshalJSON implements json.Marshaler.
func (i SignedID) MarshalJSON() ([]byte, error) {
	return []byte(`"` + strconv.FormatInt(int64(i), 10) + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (i *SignedID) UnmarshalJSON(data []byte) error {
	s := unquote(data)
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		// Some writers may emit the unsigned form.
		u, uErr := strconv.ParseUint(s, 10, 64)
		if uErr != nil {
			return fmt.Errorf("parse signed id %q: %w", s, err)
		}
		v = int64(u) // #nosec G115
	}
	*i = SignedID(v)
	return nil
}

// TypeID is a TL constructor id, encoded as "0x%08x".
type TypeID uint32

// String implements fmt.Stringer.
func (t TypeID) String() string {
	return fmt.Sprintf("0x%08x", uint32(t))
}

// MarshalJSON implements json.Marshaler.
func (t TypeID) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.String() + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *TypeID) UnmarshalJSON(data []byte) error {
	v, err := ParseTypeID(unquote(data))
	if err != nil {
		return err
	}
	*t = v
	return nil
}

// ParseTypeID parses "0x1cb5c415" (or a decimal number) into TypeID.
func ParseTypeID(s string) (TypeID, error) {
	v, err := strconv.ParseUint(s, 0, 32)
	if err != nil {
		return 0, fmt.Errorf("parse type id %q: %w", s, err)
	}
	return TypeID(v), nil
}

func unquote(data []byte) string {
	if len(data) >= 2 && data[0] == '"' && data[len(data)-1] == '"' {
		return string(data[1 : len(data)-1])
	}
	return string(data)
}

func parseUint(data []byte) (uint64, error) {
	s := unquote(data)
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse id %q: %w", s, err)
	}
	return v, nil
}

// Inner describes one message of a msg_container, as summarized by the
// writer.
type Inner struct {
	MsgID    ID     `json:"msg_id"`
	SeqNo    int64  `json:"seqno"`
	Len      int64  `json:"len"`
	Type     TypeID `json:"type"`
	Method   TypeID `json:"method,omitempty"`
	ReqMsgID ID     `json:"req_msg_id,omitempty"`
	Result   TypeID `json:"result,omitempty"`
}

// Header is the "hdr" record that starts every events segment.
type Header struct {
	V             int      `json:"v"`
	Seg           int      `json:"seg"`
	Run           string   `json:"run,omitempty"`
	App           string   `json:"app,omitempty"`
	Layer         int      `json:"layer,omitempty"`
	APISHA256     string   `json:"api_sha256,omitempty"`
	MTProtoSHA256 string   `json:"mtproto_sha256,omitempty"`
	WallUTC       string   `json:"wall_utc,omitempty"`
	MonoT         float64  `json:"mono_t,omitempty"`
	PID           int64    `json:"pid,omitempty"`
	Flags         []string `json:"flags,omitempty"`
}
