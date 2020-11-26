package telegram

import (
	"context"
	"runtime"

	"github.com/ernado/td/bin"
	"github.com/ernado/td/internal/proto"
)

func (c *Client) InitConnection(ctx context.Context) error {
	b := new(bin.Buffer)
	if err := c.newEncryptedMessage(proto.InvokeWithLayer{
		Layer: 121,
		Query: proto.InitConnection{
			ID:             0,
			SystemLangCode: "en",
			LangCode:       "en",
			SystemVersion:  runtime.GOOS + "/" + runtime.GOARCH,
			DeviceModel:    "PC",
			AppVersion:     "v0.0.0",
			LangPack:       "",
			Query:          proto.GetConfig{},
		},
	}, b); err != nil {
		return err
	}
	if err := proto.WriteIntermediate(c.conn, b); err != nil {
		return err
	}

	b.Reset()
	if err := proto.ReadIntermediate(c.conn, b); err != nil {
		return err
	}
	if err := c.checkProtocolError(b); err != nil {
		return err
	}

	encMessage := proto.EncryptedMessage{}
	if err := encMessage.Decode(b); err != nil {
		return err
	}

	// TODO(ernado): decode received config.

	return nil
}
