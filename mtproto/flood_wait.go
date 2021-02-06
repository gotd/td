package mtproto

import (
	"time"
)

// ErrFloodWait is error type of "FLOOD_WAIT" error.
const ErrFloodWait = "FLOOD_WAIT"

// AsFloodWait returns wait duration and true boolean if err is
// the "FLOOD_WAIT" error.
//
// Client should wait for that duration before issuing new requests with
// same method.
func AsFloodWait(err error) (d time.Duration, ok bool) {
	if rpcErr, ok := AsTypeErr(err, ErrFloodWait); ok {
		return time.Second * time.Duration(rpcErr.Argument), true
	}
	return 0, false
}
