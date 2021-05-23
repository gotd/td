package auth

import (
	"io"

	"github.com/gotd/td/tg"
)

// Client implements Telegram authentication.
type Client struct {
	api     *tg.Client
	rand    io.Reader
	appID   int
	appHash string
}

// NewClient initializes and returns Telegram authentication client.
func NewClient(
	api *tg.Client,
	rand io.Reader,
	appID int,
	appHash string,
) *Client {
	return &Client{
		api:     api,
		rand:    rand,
		appID:   appID,
		appHash: appHash,
	}
}
