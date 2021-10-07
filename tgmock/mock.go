package tgmock

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/bin"
)

// TestingT is simplified *testing.T interface.
type TestingT interface {
	require.TestingT
	assert.TestingT
	Helper()
	Cleanup(cb func())
}

// Mock is a mock for tg.Invoker with testify/require support.
type Mock struct {
	calls  []Handler
	assert assertions
}

// Option configures Mock.
type Option interface {
	apply(t TestingT, m *Mock)
}

type assertions interface {
	Truef(value bool, msg string, args ...interface{})
	Equal(expected interface{}, actual interface{}, msgAndArgs ...interface{})
}

type assertAssertions struct {
	assert *assert.Assertions
}

func (a assertAssertions) Equal(expected, actual interface{}, msgAndArgs ...interface{}) {
	a.assert.Equal(expected, actual, msgAndArgs...)
}

func (a assertAssertions) Truef(value bool, msg string, args ...interface{}) {
	a.assert.Truef(value, msg, args...)
}

type optionFunc func(t TestingT, m *Mock)

func (o optionFunc) apply(t TestingT, m *Mock) { o(t, m) }

// WithRequire configures mock to use "require" assertions.
func WithRequire() Option {
	return optionFunc(func(t TestingT, m *Mock) {
		m.assert = require.New(t)
	})
}

// NewRequire creates new Mock with "require" assertions.
func NewRequire(t TestingT) *Mock {
	return New(t, WithRequire())
}

// New creates new Mock.
func New(t TestingT, options ...Option) *Mock {
	m := &Mock{
		assert: &assertAssertions{
			assert: assert.New(t),
		},
	}
	for _, o := range options {
		o.apply(t, m)
	}

	t.Cleanup(func() {
		m.assert.Truef(
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
		i.assert.Truef(ok, "unexpected call")

		return h.Handle(id, body)
	}
}
