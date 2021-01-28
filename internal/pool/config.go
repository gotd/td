package pool

import (
	"fmt"
	"sync/atomic"

	"github.com/gotd/td/internal/tdsync"
	"github.com/gotd/td/tg"
)

type config struct {
	v   atomic.Value
	got *tdsync.Ready

	_ [0]func() // nocmp
}

func newConfig() *config {
	return &config{
		v:   atomic.Value{},
		got: tdsync.NewReady(),
	}
}

func (c *config) Store(cfg tg.Config) {
	c.v.Store(cfg)
	c.got.Signal()
}

func (c *config) Load() (tg.Config, bool) {
	v, ok := c.v.Load().(tg.Config)
	return v, ok
}

func (c *config) FindAddress(id int) (string, bool) {
	cfg, ok := c.Load()
	if !ok {
		return "", false
	}

	var addr string
	for _, dc := range cfg.DCOptions {
		if dc.ID != id {
			continue
		}
		if dc.MediaOnly || dc.Ipv6 || dc.CDN || dc.TcpoOnly {
			continue
		}
		addr = fmt.Sprintf("%s:%d", dc.IPAddress, dc.Port)
		break
	}

	if addr == "" {
		return "", false
	}

	return addr, true
}

func (c *config) Ready() <-chan struct{} {
	return c.got.Ready()
}
