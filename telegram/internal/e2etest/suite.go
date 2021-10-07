package e2etest

import (
	"io"
	"sync"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/nnqq/td/telegram"
	"github.com/nnqq/td/telegram/dcs"
)

// Suite is struct which contains external E2E test parameters.
type Suite struct {
	TB      require.TestingT
	appID   int
	appHash string
	dc      int
	logger  *zap.Logger

	rand io.Reader
	// already used phone numbers
	used    map[string]struct{}
	usedMux sync.Mutex
}

// NewSuite creates new Suite.
func NewSuite(tb require.TestingT, config TestOptions) *Suite {
	config.setDefaults()
	return &Suite{
		TB:      tb,
		appID:   config.AppID,
		appHash: config.AppHash,
		dc:      config.DC,
		logger:  config.Logger,
		rand:    config.Random,
		used:    map[string]struct{}{},
	}
}

// Client creates new *telegram.Client using this suite.
func (s *Suite) Client(logger *zap.Logger, handler telegram.UpdateHandler) *telegram.Client {
	return telegram.NewClient(s.appID, s.appHash, telegram.Options{
		DC:            s.dc,
		DCList:        dcs.Test(),
		Logger:        logger,
		UpdateHandler: handler,
	})
}
