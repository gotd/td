package auth

import (
	"crypto/rand"

	"github.com/nnqq/td/tg"
)

const (
	testAppID   = 1
	testAppHash = "hash"
)

func testClient(invoker tg.Invoker) *Client {
	return &Client{
		api:     tg.NewClient(invoker),
		rand:    rand.Reader,
		appID:   testAppID,
		appHash: testAppHash,
	}
}
