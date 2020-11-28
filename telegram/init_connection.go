package telegram

import (
	"context"

	"github.com/ernado/td/tg"

	"go.uber.org/zap"

	"golang.org/x/xerrors"

	"github.com/ernado/td/internal/proto"
)

// Init represents init connection.
type Init struct {
	// AppID is app api_id from `https://my.telegram.org/apps`.
	AppID int

	SystemVersion string
	AppVersion    string // like v0.1.2
	DeviceModel   string
}

const notAvailable = "n/a"

// InitConnection initializes connection.
//
// Corresponding method is `initConnection#c1cd5ea9`.
func (c *Client) InitConnection(ctx context.Context, opt Init) error {
	if opt.AppID == 0 {
		return xerrors.New("no api_id provided: ")
	}
	if opt.DeviceModel == "" {
		opt.DeviceModel = notAvailable
	}
	if opt.SystemVersion == "" {
		opt.SystemVersion = notAvailable
	}
	if opt.AppVersion == "" {
		opt.AppVersion = notAvailable
	}
	var response tg.Config
	if err := c.do(ctx, proto.InvokeWithLayer{
		Layer: proto.Layer,
		Query: proto.InitConnection{
			ID:             opt.AppID,
			SystemLangCode: "en",
			LangCode:       "en",
			SystemVersion:  opt.SystemVersion,
			DeviceModel:    opt.DeviceModel,
			AppVersion:     opt.AppVersion,
			LangPack:       "",
			Query:          proto.GetConfig{},
		},
	}, &response); err != nil {
		return xerrors.Errorf("failed to perform request: %w", err)
	}

	c.log.Debug("Got config")
	for _, dc := range response.DCOptions {
		c.log.With(
			zap.String("ip", dc.IPAddress),
			zap.Int("port", dc.Port),
		).Debug("DC option")
	}
	return nil
}
