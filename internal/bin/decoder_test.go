package bin

import "testing"

func TestDecoder_ID(t *testing.T) {
	var b Buffer
	const id = 0x1234
	b.PutID(id)
	d := Decoder{buf: b.buf}
	gotID, err := d.ID()
	if err != nil {
		t.Fatal(err)
	}
	if gotID != id {
		t.Fatal("mismatch")
	}
}
