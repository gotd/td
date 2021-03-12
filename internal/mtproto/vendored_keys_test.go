package mtproto

import "testing"

func TestVendoredKeys(t *testing.T) {
	keys, err := vendoredKeys()
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) == 0 {
		t.Fatal("empty keys")
	}
}
