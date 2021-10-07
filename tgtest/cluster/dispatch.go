package cluster

import (
	"crypto/rsa"

	"go.uber.org/zap"

	"github.com/nnqq/td/internal/crypto"
	"github.com/nnqq/td/internal/exchange"
	"github.com/nnqq/td/tgtest"
	"github.com/nnqq/td/tgtest/services"
	"github.com/nnqq/td/transport"
)

// Common returns common dispatcher.
func (c *Cluster) Common() *tgtest.Dispatcher {
	return c.common
}

func (c *Cluster) getCodec() (codec func() transport.Codec) {
	if !c.web {
		codec = c.protocol.Codec
	}
	return codec
}

// DC registers new server and returns it.
func (c *Cluster) DC(id int, name string) (*tgtest.Server, *tgtest.Dispatcher) {
	if s, ok := c.setups[id]; ok {
		return s.srv, s.dispatch
	}

	key, err := rsa.GenerateKey(c.random, crypto.RSAKeyBits)
	if err != nil {
		// TODO(tdakkota): Return error instead.
		panic(err)
	}
	// TODO(tdakkota): Generate new keys too.
	privateKey := exchange.PrivateKey{
		RSA:       key,
		UseRSAPad: false,
	}

	d := tgtest.NewDispatcher()
	server := tgtest.NewServer(privateKey, tgtest.UnpackInvoke(d), tgtest.ServerOptions{
		DC:     id,
		Logger: c.log.Named(name).With(zap.Int("dc_id", id)),
		Codec:  c.getCodec(),
	})
	c.setups[id] = setup{
		srv:      server,
		dispatch: d,
	}
	c.keys = append(c.keys, server.Key())

	// We set server fallback handler to dispatch request in order
	// 1) Explicit DC handler
	// 2) Explicit common handler
	// 3) Common fallback
	d.Fallback(c.Common())
	return server, d
}

// Dispatch registers new server and returns its dispatcher.
func (c *Cluster) Dispatch(id int, name string) *tgtest.Dispatcher {
	_, d := c.DC(id, name)
	return d
}

func (c *Cluster) fallback() tgtest.HandlerFunc {
	return services.NotImplemented
}
