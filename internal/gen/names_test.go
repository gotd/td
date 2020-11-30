package gen

import "testing"

func TestNames(t *testing.T) {
	if pascal("Id") != "ID" {
		t.Fatal("mismatch")
	}
	if camel("user_id") != "userID" {
		t.Fatal("mismatch")
	}
}
