package services

import "github.com/gotd/td/tgerr"

// ErrMethodNotImplemented denotes that method is not implemented.
var ErrMethodNotImplemented error = tgerr.New(400, "INPUT_METHOD_INVALID")
