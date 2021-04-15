package telegram

import (
	"github.com/gotd/td/telegram/internal/helpers"
)

// ErrFloodWait is error type of "FLOOD_WAIT" error.
const ErrFloodWait = helpers.ErrFloodWait

// AsFloodWait returns wait duration and true boolean if err is
// the "FLOOD_WAIT" error.
//
// Client should wait for that duration before issuing new requests with
// same method.
var AsFloodWait = helpers.AsFloodWait
