package main

import (
	"bytes"
	"context"
	"testing"
)

func TestGenerate(t *testing.T) {
	var out bytes.Buffer
	if err := generate(context.Background(), &out, "testgen"); err != nil {
		t.Fatal(err)
	}
}
