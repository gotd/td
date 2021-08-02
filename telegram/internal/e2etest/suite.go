package e2etest

import (
	"context"
	"io"
	"sync"

	"github.com/cenkalti/backoff/v4"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/dcs"
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
		DCList:        dcs.Staging(),
		Logger:        logger,
		UpdateHandler: handler,
	})
}

// Authenticate authenticates client on test server.
func (s *Suite) Authenticate(ctx context.Context, client *telegram.Client) error {
	var ua auth.UserAuthenticator
	for {
		ua = auth.Test(s.rand, s.dc)
		phone, err := ua.Phone(ctx)
		if err != nil {
			return err
		}

		s.usedMux.Lock()
		if _, ok := s.used[phone]; !ok {
			s.used[phone] = struct{}{}
			s.usedMux.Unlock()
			break
		}
		s.usedMux.Unlock()
	}

	return auth.NewFlow(ua, auth.SendCodeOptions{}).Run(ctx, client.Auth())
}

// RetryAuthenticate authenticates client on test server.
func (s *Suite) RetryAuthenticate(ctx context.Context, client *telegram.Client) error {
	return backoff.Retry(func() error {
		return s.Authenticate(ctx, client)
	}, backoff.WithContext(backoff.NewExponentialBackOff(), ctx))
}
