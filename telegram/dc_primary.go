package telegram

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/mtproto"
	"github.com/gotd/td/internal/mtproto/reliable"
	"github.com/gotd/td/tg"
	"golang.org/x/xerrors"
)

func (c *Client) dialPrimaryFirst(ctx context.Context) error {
	c.pmux.Lock()
	defer c.pmux.Unlock()

	// TODO(ccln): wait session?
	opts := c.opts
	opts.MessageHandler = c.onPrimaryMessage

	var once sync.Once
	gotSession := make(chan struct{})
	opts.SessionHandler = func(session mtproto.Session) error {
		once.Do(func() { close(gotSession) })
		return c.onPrimarySession(session)
	}

	conn := reliable.New(reliable.Config{
		Addr:   fmt.Sprintf("%d|%s", c.primaryDC, c.initialAddr),
		MTOpts: opts,
		OnConnected: func(conn reliable.MTConn) error {
			_, err := c.initConn(ctx, conn, false)
			return err
		},
	})

	if err := c.lf.Start(conn); err != nil {
		return err
	}

	select {
	case <-gotSession:
		break
	case <-time.After(time.Second * 10):
		return xerrors.Errorf("session timeout")
	}

	cfg, err := tg.NewClient(conn).HelpGetConfig(ctx)
	if err != nil {
		return err
	}

	c.primary = conn
	c.cfg = *cfg
	return nil
}

func (c *Client) dialPrimary(ctx context.Context) error {
	dcInfo, err := c.lookupDC(c.primaryDC)
	if err != nil {
		return err
	}

	if _, err := c.dc(dcInfo).WithCreds(c.sess.Key, c.sess.Salt).AsPrimary().Connect(ctx); err != nil {
		return err
	}

	return nil
}

func (c *Client) onPrimarySession(session mtproto.Session) error {
	c.dataMux.Lock()
	defer c.dataMux.Unlock()
	c.sess = session
	return c.storageSave()
}

func (c *Client) onPrimaryConfig(cfg tg.Config) error {
	c.dataMux.Lock()
	defer c.dataMux.Unlock()
	c.cfg = cfg
	return c.storageSave()
}

func (c *Client) onPrimaryMessage(b *bin.Buffer) error {
	updates, err := tg.DecodeUpdates(b)
	if err != nil {
		return xerrors.Errorf("decode updates: %w", err)
	}

	return c.processUpdates(updates)
}
