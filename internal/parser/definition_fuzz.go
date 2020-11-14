// +build fuzz

package parser

import "fmt"

func FuzzDefinition(data []byte) int {
	var d Definition
	if err := d.Parse(string(data)); err != nil {
		return 0
	}

	var other Definition
	if err := other.Parse(d.String()); err != nil {
		fmt.Printf("input: %s\n", string(data))
		fmt.Printf("parsed: %#v\n", d)
		panic(err)
	}

	return 1
}
