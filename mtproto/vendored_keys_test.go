package mtproto

import "testing"

func TestVendoredKeys(t *testing.T) {
	keys := vendoredKeys()
	if len(keys) == 0 {
		t.Fatal("empty keys")
	}
}
