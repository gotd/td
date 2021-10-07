// Package config contains config service implementation for tgtest server.
package config

import (
	"context"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/tg"
	"github.com/nnqq/td/tgtest"
	"github.com/nnqq/td/tgtest/services"
)

// Service is a Telegram config service.
type Service struct {
	cfg    *tg.Config
	cdnCfg *tg.CDNConfig
}

// NewService creates new Service.
func NewService(cfg *tg.Config, cdnCfg *tg.CDNConfig) *Service {
	return &Service{cfg: cfg, cdnCfg: cdnCfg}
}

func (c *Service) HelpGetCDNConfig(ctx context.Context, req *tg.HelpGetCDNConfigRequest) (*tg.CDNConfig, error) {
	cfg := c.cdnCfg
	return cfg, nil
}

func (c *Service) HelpGetConfig(ctx context.Context, dc int, req *tg.HelpGetConfigRequest) (*tg.Config, error) {
	cfg := *c.cfg
	cfg.ThisDC = dc
	return &cfg, nil
}

// OnMessage implements tgtest.Handler.
func (c *Service) OnMessage(server *tgtest.Server, req *tgtest.Request) error {
	id, err := req.Buf.PeekID()
	if err != nil {
		return err
	}

	var (
		decode bin.Decoder
		result bin.Encoder
	)
	switch id {
	case tg.HelpGetCDNConfigRequestTypeID:
		cfg := c.cdnCfg

		decode = &tg.HelpGetCDNConfigRequest{}
		result = cfg
	case tg.HelpGetConfigRequestTypeID:
		cfg := *c.cfg
		cfg.ThisDC = req.DC

		decode = &tg.HelpGetConfigRequest{}
		result = &cfg
	default:
		return services.ErrMethodNotImplemented
	}

	if err := decode.Decode(req.Buf); err != nil {
		return err
	}
	return server.SendResult(req, result)
}

// Register registers service handlers.
func (c *Service) Register(dispatcher *tgtest.Dispatcher) {
	dispatcher.HandleFunc(tg.HelpGetCDNConfigRequestTypeID, c.OnMessage)
	dispatcher.HandleFunc(tg.HelpGetConfigRequestTypeID, c.OnMessage)
}
