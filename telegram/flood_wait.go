package telegram

import (
	"github.com/gotd/td/tgerr"
)

// ErrFloodWait is error type of "FLOOD_WAIT" error.
const ErrFloodWait = tgerr.ErrFloodWait

// AsFloodWait returns wait duration and true boolean if err is
// the "FLOOD_WAIT" error.
//
// Client should wait for that duration before issuing new requests with
// same method.
var AsFloodWait = tgerr.AsFloodWait
