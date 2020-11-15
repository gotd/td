// +build fuzz

package parser

import (
	"bytes"
)

func Fuzz(data []byte) int {
	schema, err := Parse(bytes.NewReader(data))
	if err != nil {
		return 0
	}
	if schema == nil {
		panic("nil")
	}
	return 1
}
