package telegram

import (
	"sync"
	"sync/atomic"

	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/tg"
)

type atomicConfig struct {
	atomic.Value
}

func (c *atomicConfig) Load() tg.Config {
	return c.Value.Load().(tg.Config)
}

func (c *atomicConfig) Store(cfg tg.Config) {
	c.Value.Store(cfg)
}

type sessionData struct {
	DC      int
	Addr    string
	AuthKey crypto.AuthKey
	Salt    int64
}

type syncSessionData struct {
	data sessionData
	mux  sync.RWMutex
}

func newSyncSessionData(data sessionData) *syncSessionData {
	return &syncSessionData{
		data: data,
	}
}

func (s *syncSessionData) Store(data sessionData) {
	s.mux.Lock()
	s.data = data
	s.mux.Unlock()
}

func (s *syncSessionData) Migrate(dc int, addr string) {
	s.mux.Lock()
	s.data.DC = dc
	s.data.Addr = addr
	s.data.AuthKey = crypto.AuthKey{}
	s.data.Salt = 0
	s.mux.Unlock()
}

func (s *syncSessionData) Options(opts mtproto.Options) (mtproto.Options, sessionData) {
	s.mux.RLock()
	data := s.data
	s.mux.RUnlock()

	opts.Key = data.AuthKey
	opts.Salt = data.Salt
	return opts, data
}

func (s *syncSessionData) Load() (data sessionData) {
	s.mux.RLock()
	data = s.data
	s.mux.RUnlock()

	return
}
