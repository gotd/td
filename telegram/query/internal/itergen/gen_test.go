package main

import (
	"bytes"
	"context"
	"testing"
)

func TestGenerate(t *testing.T) {
	var out bytes.Buffer
	if err := generate(context.Background(), &out, collectorConfig{
		ResultName: "MessagesMessagesClass",
		ElemName:   "Elem",
		Prefix:     "Messages",
		PkgName:    "messages",
	}); err != nil {
		t.Fatal(err)
	}
}
