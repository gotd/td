package jsontd

import "fmt"

// UnexpectedIDErr means that unknown or unexpected type id was decoded.
type UnexpectedIDErr struct {
	ID string
}

func (e *UnexpectedIDErr) Error() string {
	return fmt.Sprintf("unexpected id %s", e.ID)
}

// NewUnexpectedID return new UnexpectedIDErr.
func NewUnexpectedID(id string) error {
	return &UnexpectedIDErr{ID: id}
}

