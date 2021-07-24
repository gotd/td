package faketls

// RecordType represents TLS record type byte.
type RecordType uint8

const (
	// RecordTypeChangeCipherSpec is ChangeCipherSpec record type byte.
	RecordTypeChangeCipherSpec RecordType = 0x14
	// RecordTypeAlert is Alert record type byte.
	RecordTypeAlert RecordType = 0x15
	// RecordTypeHandshake is Handshake record type byte.
	RecordTypeHandshake RecordType = 0x16
	// RecordTypeApplication is Application record type byte.
	RecordTypeApplication RecordType = 0x17
	// RecordTypeHeartbeat is Heartbeat record type byte.
	RecordTypeHeartbeat RecordType = 0x18
)

// HandshakeType represents TLS handshake record type byte.
type HandshakeType uint8

const (
	// HandshakeTypeClient is client handshake message type.
	HandshakeTypeClient HandshakeType = 0x01
	// HandshakeTypeServer is server handshake message type.
	HandshakeTypeServer HandshakeType = 0x02
)

// Possible versions.
var (
	Version10Bytes = [2]byte{0x03, 0x01}
	Version11Bytes = [2]byte{0x03, 0x02}
	Version12Bytes = [2]byte{0x03, 0x03}
	Version13Bytes = [2]byte{0x03, 0x04}
)
