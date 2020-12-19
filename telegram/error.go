package telegram

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
	// Ignoring non-digit last part.
	last := parts[len(parts)-1]
	for _, r := range last {
		if !unicode.IsDigit(r) {
			return
		}
	}
	argument, err := strconv.Atoi(last)
	if err != nil {
		// Should be unreachable.
		return
	}

	// Argument is last underscored part, type is prefix without
	// last underscore, e.g: FLOOD_WAIT_3 -> (FLOOD_WAIT, 3).
	e.Argument = argument
	e.Type = strings.Join(parts[:len(parts)-1], "_")
}

func (e Error) Error() string {
	if e.Argument != 0 {
		return fmt.Sprintf("rpc error code %d: %s (%d)", e.Code, e.Type, e.Argument)
	}
	return fmt.Sprintf("rpc error code %d: %s", e.Code, e.Message)
}
