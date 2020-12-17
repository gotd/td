// Package e2etest contains end-to-end tests using staging Telegram server.
package e2etest

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/gotd/td/telegram"
)

// Suite is struct which contains test parameters.
type Suite struct {
	testing.TB

	appID   int
	appHash string

	dcID int
	addr string
}

// NewSuite creates new Suite.
func NewSuite(t testing.TB, appID int, appHash string, dcID int, addr string) Suite {
	return Suite{
		TB:      t,
		appID:   appID,
		appHash: appHash,
		dcID:    dcID,
		addr:    addr,
	}
}

// Client creates new *telegram.Client using this suite.
func (s Suite) Client(logger *zap.Logger, handler telegram.UpdateHandler) *telegram.Client {
	return telegram.NewClient(s.appID, s.appHash, telegram.Options{
		Addr:          s.addr,
		Logger:        logger,
		UpdateHandler: handler,
	})
}

func createLogger(name string) *zap.Logger {
	logger, _ := zap.NewDevelopment(zap.IncreaseLevel(zapcore.DebugLevel))
	logger = logger.Named(name)

	return logger
}
