package gen

import "testing"

func TestNames(t *testing.T) {
	if pascal("Id") != "ID" {
		t.Fatal("bad name")
	}
}
