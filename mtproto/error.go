package mtproto

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Error represents RPC error returned to request.
type Error struct {
	Code     int    // 420
	Message  string // FLOOD_WAIT_3
	Type     string // FLOOD_WAIT
	Argument int    // 3
}

// ExtractArgument extracts Type and Argument from Message.
func (e *Error) ExtractArgument() {
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

func (e Error) Error() string {
	if e.Argument != 0 {
		return fmt.Sprintf("rpc error code %d: %s (%d)", e.Code, e.Type, e.Argument)
	}
	return fmt.Sprintf("rpc error code %d: %s", e.Code, e.Message)
}
