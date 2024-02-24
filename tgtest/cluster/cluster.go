// Package cluster contains Telegram multi-DC setup utilities.
package cluster

import (
	"io"

	"go.uber.org/zap"

	"github.com/gotd/td/exchange"
	"github.com/gotd/td/tdsync"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgtest"
	"github.com/gotd/td/tgtest/services/config"
)

type setup struct {
	srv      *tgtest.Server
	dispatch *tgtest.Dispatcher
}

// Cluster is a cluster of multiple servers, representing multiple Telegram datacenters.
type Cluster struct {
	// denotes to use websocket listener
	web bool

	setups map[int]setup
	keys   []exchange.PublicKey

	// DCs config state.
	cfg     tg.Config
	cdnCfg  tg.CDNConfig
	domains map[int]string

	// Signal for readiness.
	ready *tdsync.Ready

	// RPC dispatcher.
	common *tgtest.Dispatcher

	log      *zap.Logger
	random   io.Reader
	protocol dcs.Protocol
}

// NewCluster creates new server Cluster.
func NewCluster(opts Options) *Cluster {
	opts.setDefaults()

	q := &Cluster{
		web:      opts.Web,
		setups:   map[int]setup{},
		keys:     nil,
		cfg:      opts.Config,
		cdnCfg:   opts.CDNConfig,
		domains:  map[int]string{},
		ready:    tdsync.NewReady(),
		common:   tgtest.NewDispatcher(),
		log:      opts.Logger,
		random:   opts.Random,
		protocol: opts.Protocol,
	}
	config.NewService(&q.cfg, &q.cdnCfg).Register(q.common)
	q.common.Fallback(q.fallback())

	return q
}

// List returns DCs list.
func (c *Cluster) List() dcs.List {
	return dcs.List{
		Options: c.cfg.DCOptions,
		Domains: c.domains,
	}
}

// Resolver returns dcs.Resolver to use.
func (c *Cluster) Resolver() dcs.Resolver {
	if c.web {
		return dcs.Websocket(dcs.WebsocketOptions{})
	}

	return dcs.Plain(dcs.PlainOptions{
		Protocol: c.protocol,
	})
}

// Keys returns all servers public keys.
func (c *Cluster) Keys() []exchange.PublicKey {
	return c.keys
}

// Ready returns signal channel to await readiness.
func (c *Cluster) Ready() <-chan struct{} {
	return c.ready.Ready()
}
