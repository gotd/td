package tgmock

import (
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
)

// TestingT is simplified *testing.T interface.
type TestingT interface {
	Cleanup(cb func())
}

// Mock is a mock for tg.Invoker with testify/require support.
type Mock struct {
	calls []Handler
	*require.Assertions
}

// NewMock creates new Mock.
func NewMock(testingT TestingT, assert *require.Assertions) *Mock {
	m := &Mock{
		Assertions: assert,
	}

	testingT.Cleanup(func() {
		m.Assertions.Truef(
			m.AllWereMet(),
			"not all expected calls happen (expected yet %d)",
			len(m.calls),
		)
	})
	return m
}

// AllWereMet returns true if all expected calls happened.
func (i *Mock) AllWereMet() bool {
	return len(i.calls) == 0
}

func (i *Mock) add(h Handler) *Mock {
	i.calls = append(i.calls, h)
	return i
}

func (i *Mock) pop() (Handler, bool) {
	if len(i.calls) < 1 {
		return nil, false
	}

	h := i.calls[0]
	// Delete from SliceTricks.
	copy(i.calls, i.calls[1:])
	i.calls[len(i.calls)-1] = nil
	i.calls = i.calls[:len(i.calls)-1]
	return h, true
}

// Handler returns HandlerFunc of Mock.
func (i *Mock) Handler() HandlerFunc {
	return func(id int64, body bin.Encoder) (bin.Encoder, error) {
		h, ok := i.pop()
		i.Assertions.Truef(ok, "unexpected call")

		return h.Handle(id, body)
	}
}
