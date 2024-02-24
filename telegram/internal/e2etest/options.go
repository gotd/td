package e2etest

import (
	"io"

	"go.uber.org/zap"

	"github.com/gotd/td/constant"
	"github.com/gotd/td/crypto"
)

// TestOptions contains some common test server settings.
type TestOptions struct {
	AppID   int
	AppHash string
	DC      int
	Random  io.Reader
	Logger  *zap.Logger
}

func (opt *TestOptions) setDefaults() {
	if opt.AppID == 0 {
		opt.AppID = constant.TestAppID
	}
	if opt.AppHash == "" {
		opt.AppHash = constant.TestAppHash
	}
	if opt.DC == 0 {
		opt.DC = 2
	}
	if opt.Random == nil {
		opt.Random = crypto.DefaultRand()
	}
	if opt.Logger == nil {
		opt.Logger = zap.NewNop()
	}
}
