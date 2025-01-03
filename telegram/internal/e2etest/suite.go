package e2etest

import (
	"io"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tgacc"
)

// Suite is struct which contains external E2E test parameters.
type Suite struct {
	TB      require.TestingT
	appID   int
	appHash string
	dc      int
	logger  *zap.Logger
	manager *tgacc.TestAccountManager
	closers []func() error

	rand io.Reader
	// already used phone numbers
	used    map[string]struct{}
	usedMux sync.Mutex
}

// Close closes all resources.
func (s *Suite) Close() error {
	var err error
	for _, closer := range s.closers {
		if e := closer(); e != nil {
			err = e
		}
	}
	return err
}

// NewSuite creates new Suite.
func NewSuite(t *testing.T, config TestOptions) *Suite {
	config.setDefaults()
	manager, err := tgacc.NewTestAccountManager()
	require.NoError(t, err)
	s := &Suite{
		TB:      t,
		appID:   config.AppID,
		appHash: config.AppHash,
		dc:      config.DC,
		logger:  config.Logger,
		manager: manager,
		rand:    config.Random,
		used:    make(map[string]struct{}),
	}
	if broken, _ := strconv.ParseBool(os.Getenv("TEST_ACCOUNTS_BROKEN")); broken {
		t.Log("External test accounts are used as per TEST_ACCOUNTS_BROKEN")
	} else {
		t.Log("Normal test accounts are used")
		s.manager = nil // disable manager
	}
	if disabled, _ := strconv.ParseBool(os.Getenv("TEST_ACCOUNT_MANAGER_DISABLED")); s.manager != nil && disabled {
		t.Skip("Test account manager is disabled")
	}
	t.Cleanup(func() {
		require.NoError(t, s.Close())
	})
	return s
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
