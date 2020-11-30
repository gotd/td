package telegram

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/ernado/td/internal/proto"
	"github.com/ernado/td/tg"
)

// initConnection initializes connection.
//
// Corresponding method is `initConnection#c1cd5ea9`.
func (c *Client) initConnection(ctx context.Context) error {
	// TODO(ernado): Make versions configurable.
	const notAvailable = "n/a"

	var response tg.Config
	if err := c.rpcContent(ctx, proto.InvokeWithLayer{
		Layer: proto.Layer,
		Query: proto.InitConnection{
			ID:             c.appID,
			SystemLangCode: "en",
			LangCode:       "en",
			SystemVersion:  notAvailable,
			DeviceModel:    notAvailable,
			AppVersion:     notAvailable,
			LangPack:       "",
			Query:          proto.GetConfig{},
		},
	}, &response); err != nil {
		return xerrors.Errorf("failed to perform request: %w", err)
	}

	c.log.Debug("Got config")
	return nil
}
