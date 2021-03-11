package mtproto

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/gotd/td/tg"
)

// Error represents RPC error returned to request.
type Error struct {
	Code     int    // 420
	Message  string // FLOOD_WAIT_3
	Type     string // FLOOD_WAIT
	Argument int    // 3
}

// NewError creates new *Error from code and message, extracting argument
// and type.
func NewError(code int, msg string) *Error {
	e := &Error{
		Code:    code,
		Message: msg,
	}
	e.extractArgument()
	return e
}

// Is reports whether e matches target error.
func (e *Error) Is(target error) bool {
	if e == nil {
		return false
	}
	var typ tg.ErrorType
	return errors.As(target, &typ) && e.IsType(string(typ))
}

// IsType reports whether error has type t.
func (e *Error) IsType(t string) bool {
	if e == nil {
		return false
	}
	return e.Type == t
}

// IsCode reports whether error Code is equal to code.
func (e *Error) IsCode(code int) bool {
	if e == nil {
		return false
	}
	return e.Code == code
}

// IsOneOf returns true if error type is in tt.
func (e *Error) IsOneOf(tt ...string) bool {
	if e == nil {
		return false
	}
	for _, t := range tt {
		if e.IsType(t) {
			return true
		}
	}
	return false
}

// IsCodeOneOf returns true if error code is one of codes.
func (e *Error) IsCodeOneOf(codes ...int) bool {
	if e == nil {
		return false
	}
	for _, code := range codes {
		if e.IsCode(code) {
			return true
		}
	}
	return false
}

// extractArgument extracts Type and Argument from Message.
func (e *Error) extractArgument() {
	if e.Message == "" {
		return
	}

	// Defaulting Type to Message.
	e.Type = e.Message

	// Splitting by underscore.
	parts := strings.Split(e.Message, "_")
	if len(parts) < 2 {
		return
	}

	var nonDigit []string
Parts:
	for _, part := range parts {
		for _, r := range part {
			if unicode.IsDigit(r) {
				continue
			}

			// Found non-digit part, skipping.
			nonDigit = append(nonDigit, part)
			continue Parts
		}

		// Found digit-only part, using as argument.
		argument, err := strconv.Atoi(part)
		if err != nil {
			// Should be unreachable.
			return
		}
		e.Argument = argument
	}

	e.Type = strings.Join(nonDigit, "_")
}

func (e *Error) Error() string {
	if e.Argument != 0 {
		return fmt.Sprintf("rpc error code %d: %s (%d)", e.Code, e.Type, e.Argument)
	}
	return fmt.Sprintf("rpc error code %d: %s", e.Code, e.Message)
}

// AsTypeErr returns *Error from err if rpc error type is t.
func AsTypeErr(err error, t string) (rpcErr *Error, ok bool) {
	if errors.As(err, &rpcErr) && rpcErr.Type == t {
		return rpcErr, true
	}
	return nil, false
}

// AsErr extracts *Error from err if possible.
func AsErr(err error) (rpcErr *Error, ok bool) {
	if errors.As(err, &rpcErr) {
		return rpcErr, true
	}
	return nil, false
}

// IsErr returns true if err type is t.
func IsErr(err error, tt ...string) bool {
	if rpcErr, ok := AsErr(err); ok {
		return rpcErr.IsOneOf(tt...)
	}
	return false
}

// IsErrCode returns true of error code is as provided.
func IsErrCode(err error, code ...int) bool {
	if rpcErr, ok := AsErr(err); ok {
		return rpcErr.IsCodeOneOf(code...)
	}
	return false
}
