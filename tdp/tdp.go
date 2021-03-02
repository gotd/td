// Package tdp is td pretty-printing and formatting facilities for types from
// MTProto.
package tdp

import (
	"fmt"
	"strings"
)

// options for formatting.
type options struct {
}

// Option of formatting.
type Option interface {
	apply(o *options)
}

// Format pretty-prints v into string.
func Format(v interface{}, opts ...Option) string {
	var opt options
	for _, o := range opts {
		o.apply(&opt)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("%T", v))

	return b.String()
}
