package peers

import (
	"testing"

	"go.uber.org/zap/zaptest"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
)

func testManager(t *testing.T) (*tgmock.Mock, *Manager) {
	mock := tgmock.New(t)
	return mock, NewManager(tg.NewClient(mock), Options{
		Logger: zaptest.NewLogger(t),
	})
}
