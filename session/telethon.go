package session

import (
	"encoding/base64"
	"encoding/binary"
	"net"
	"strconv"

	"github.com/go-faster/errors"

	"github.com/gotd/td/crypto"
)

// https://github.com/LonamiWebs/Telethon/blob/master/telethon/sessions/string.py#L11
const latestTelethonVersion byte = '1'

// TelethonSession decodes Telethon's StringSession string to the Data.
// Notice that Telethon does not save tg.Config and server salt.
//
// See https://docs.telethon.dev/en/latest/modules/sessions.html#telethon.sessions.string.StringSession.
//
// See https://github.com/LonamiWebs/Telethon/blob/master/telethon/sessions/string.py#L29-L46.
func TelethonSession(hx string) (*Data, error) {
	if len(hx) < 1 {
		return nil, errors.Errorf("given string too small: %d", len(hx))
	}
	version := hx[0]
	if version != latestTelethonVersion {
		return nil, errors.Errorf("unexpected version %q, latest supported is %q",
			version,
			latestTelethonVersion,
		)
	}

	data, err := base64.URLEncoding.DecodeString(hx[1:])
	if err != nil {
		return nil, errors.Wrap(err, "decode hex")
	}

	return decodeStringSession(data)
}

func decodeStringSession(data []byte) (*Data, error) {
	// Given parameter should contain version + data
	// where data encoded using pack as '>B4sH256s' or '>B16sH256s'
	// depending on IP type.
	//
	// Table:
	//
	// | Size |  Type  | Description |
	// |------|--------|-------------|
	// | 1    | byte   | DC ID       |
	// | 4/16 | bytes  | IP address  |
	// | 2    | uint16 | Port        |
	// | 256  | bytes  | Auth key    |
	var ipLength int
	switch len(data) {
	case 263:
		ipLength = 4
	case 275:
		ipLength = 16
	default:
		return nil, errors.Errorf("decoded hex has invalid length: %d", len(data))
	}

	// | 1    | byte   | DC ID       |
	dcID := data[0]

	// | 4/16 | bytes  | IP address  |
	addr := make(net.IP, 0, 16)
	addr = append(addr, data[1:1+ipLength]...)

	// | 2    | uint16 | Port        |
	port := binary.BigEndian.Uint16(data[1+ipLength : 3+ipLength])

	// | 256  | bytes  | Auth key    |
	var key crypto.Key
	copy(key[:], data[3+ipLength:])
	id := key.WithID().ID

	return &Data{
		DC:        int(dcID),
		Addr:      net.JoinHostPort(addr.String(), strconv.Itoa(int(port))),
		AuthKey:   key[:],
		AuthKeyID: id[:],
	}, nil
}
