package bin

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecoder_ID(t *testing.T) {
	var b Buffer
	const id = 0x1234
	b.PutID(id)
	b.PutString("foo bar")
	b.PutBool(true)
	b.PutBytes([]byte{1, 2, 3, 4})
	b.PutInt32(-150)
	b.PutLong(100)
	b.PutDouble(1.5)
	b.PutPadding(100)
	gotID, err := b.ID()
	if err != nil {
		t.Fatal(err)
	}
	if gotID != id {
		t.Fatal("mismatch")
	}
	gotStr, err := b.String()
	if err != nil {
		t.Fatal(err)
	}
	if gotStr != "foo bar" {
		t.Fatal("mismatch")
	}
	gotBool, err := b.Bool()
	if err != nil {
		t.Fatal(err)
	}
	if !gotBool {
		t.Fatal("mismatch")
	}
	gotBytes, err := b.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotBytes, []byte{1, 2, 3, 4}) {
		t.Fatal("mismatch")
	}
	gotInt32, err := b.Int32()
	if err != nil {
		t.Fatal(err)
	}
	if gotInt32 != -150 {
		t.Fatal(gotInt32)
	}
	gotLong, err := b.Long()
	if err != nil {
		t.Fatal(err)
	}
	if gotLong != 100 {
		t.Fatal(gotLong)
	}
	gotDouble, err := b.Double()
	if err != nil {
		t.Fatal(err)
	}
	if gotDouble != 1.5 {
		t.Fatal(gotDouble)
	}
	if err := b.ConsumePadding(100); err != nil {
		t.Fatal(err)
	}
	require.Zero(t, b.Len(), "buffer should be fully consumed")
}
