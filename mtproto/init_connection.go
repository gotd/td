package mtproto

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/tg"
)

type connType byte

const (
	connDefault connType = iota
	connWithoutUpdates
)

// initConnection initializes connection.
//
// Corresponding method is `initConnection#c1cd5ea9`.
func (c *Conn) initConnection(ctx context.Context, t connType) error {
	// TODO(ernado): Make versions configurable.
	const notAvailable = "n/a"

	q := proto.InitConnection{
		ID:             c.appID,
		SystemLangCode: "en",
		LangCode:       "en",
		SystemVersion:  notAvailable,
		DeviceModel:    notAvailable,
		AppVersion:     notAvailable,
		LangPack:       "",
		Query:          proto.GetConfig{},
	}
	var req bin.Object = proto.InvokeWithLayer{
		Layer: tg.Layer,
		Query: q,
	}
	if t == connWithoutUpdates {
		req = proto.InvokeWithoutUpdates{
			Query: req,
		}
	}

	var response tg.Config
	if err := c.rpcContent(ctx, req, &response); err != nil {
		return xerrors.Errorf("request: %w", err)
	}

	c.cfg = response

	c.log.Debug("Got config")
	return nil
}
