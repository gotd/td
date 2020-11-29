package telegram

import "fmt"

// Error represents RPC error returned to request.
type Error struct {
	Code    int
	Message string
}

func (e Error) Error() string {
	return fmt.Sprintf("rpc error code %d: %s", e.Code, e.Message)
}
