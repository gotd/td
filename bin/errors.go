package bin

import "fmt"

// InvalidLengthError is returned when decoder reads invalid length.
type InvalidLengthError struct {
	Length int
	Where  string
}

func (i *InvalidLengthError) Error() string {
	return fmt.Sprintf("invalid %s length: %d", i.Where, i.Length)
}

// UnexpectedIDErr means that unknown or unexpected type id was decoded.
type UnexpectedIDErr struct {
	ID uint32
}

func (e *UnexpectedIDErr) Error() string {
	return fmt.Sprintf("unexpected id %#x", e.ID)
}

// NewUnexpectedID return new UnexpectedIDErr.
func NewUnexpectedID(id uint32) error {
	return &UnexpectedIDErr{ID: id}
}
