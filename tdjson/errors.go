package tdjson

import (
	"fmt"

	"github.com/go-faster/errors"
)

// UnexpectedIDError means that unknown or unexpected type id was decoded.
type UnexpectedIDError struct {
	ID string
}

func (e *UnexpectedIDError) Error() string {
	return fmt.Sprintf("unexpected id %s", e.ID)
}

// NewUnexpectedID return new UnexpectedIDError.
func NewUnexpectedID(id string) error {
	return &UnexpectedIDError{ID: id}
}

// ErrTypeIDNotFound means that @type field is expected, but not found.
var ErrTypeIDNotFound = errors.New("@type field is expected, but not found")
