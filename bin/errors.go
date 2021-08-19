package bin

import "fmt"

// InvalidLengthError is returned when decoder reads invalid length.
type InvalidLengthError struct {
	Length int
	Where  string
}

// Error implements error interface.
func (i *InvalidLengthError) Error() string {
	return fmt.Sprintf("invalid %s length: %d", i.Where, i.Length)
}

// UnexpectedIDErr means that unknown or unexpected type id was decoded.
type UnexpectedIDErr struct {
	ID uint32
}

// Error implements error interface.
func (e *UnexpectedIDErr) Error() string {
	return fmt.Sprintf("unexpected id %#x", e.ID)
}

// NewUnexpectedID return new UnexpectedIDErr.
func NewUnexpectedID(id uint32) error {
	return &UnexpectedIDErr{ID: id}
}

// NilError is returned when encoder/decoder is called on nil object.
type NilError struct {
	Action   string
	TypeName string
}

// Error implements error interface.
func (n *NilError) Error() string {
	return fmt.Sprintf("can't %s %s using nil value", n.Action, n.TypeName)
}

// FieldError is returned when encoder/decoder can't encode/decode FieldName of TypeName.
type FieldError struct {
	Action     string
	TypeName   string
	FieldName  string
	BareField  bool
	Underlying error
}

// Error implements error interface.
func (n *FieldError) Error() string {
	wrappedString := ""
	if n.Underlying != nil {
		wrappedString = ": " + n.Underlying.Error()
	}
	bareString := ""
	if n.BareField {
		bareString = " bare"
	}
	return fmt.Sprintf(
		"unable to %s %s:%s field %s%s",
		n.Action, n.TypeName,
		bareString, n.FieldName,
		wrappedString,
	)
}

// Unwrap returns underlying error.
func (n *FieldError) Unwrap() error {
	return n.Underlying
}

// IndexError is returned when encoder/decoder can't encode/decode vector element with Index.
type IndexError struct {
	Index      int
	Underlying error
}

// Error implements error interface.
func (n *IndexError) Error() string {
	wrappedString := ""
	if n.Underlying != nil {
		wrappedString = ": " + n.Underlying.Error()
	}
	return fmt.Sprintf("element with index %d: %s", n.Index, wrappedString)
}

// Unwrap returns underlying error.
func (n *IndexError) Unwrap() error {
	return n.Underlying
}

// DecodeError is a generic decoder error.
type DecodeError struct {
	TypeName   string
	Underlying error
}

// Error implements error interface.
func (n *DecodeError) Error() string {
	wrappedString := ""
	if n.Underlying != nil {
		wrappedString = ": " + n.Underlying.Error()
	}
	return fmt.Sprintf("unable to decode %s: %s", n.TypeName, wrappedString)
}

// Unwrap returns underlying error.
func (n *DecodeError) Unwrap() error {
	return n.Underlying
}
