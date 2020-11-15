// +build fuzz

package parser

import (
	"bytes"
	"fmt"
)

func Fuzz(data []byte) int {
	schema, err := Parse(bytes.NewReader(data))
	if err != nil {
		return 0
	}
	if schema == nil {
		panic("nil")
	}
	b := new(bytes.Buffer)
	if _, err := schema.WriteTo(b); err != nil {
		panic(err)
	}
	parsedSchema, err := Parse(bytes.NewReader(b.Bytes()))
	if err != nil {
		panic(err)
	}
	newBuf := new(bytes.Buffer)
	if _, err := parsedSchema.WriteTo(newBuf); err != nil {
		panic(err)
	}
	if !bytes.Equal(b.Bytes(), newBuf.Bytes()) {
		fmt.Printf("first cycle: %q", b)
		fmt.Printf("second cycle: %q", newBuf)
		fmt.Printf("input: %q", data)
		panic("parse-writeTo-parse-writeTo cycle deviated")
	}
	return 1
}
