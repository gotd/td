package telegram

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/ernado/td/internal/proto"
)

// Client represents a MTProto client to Telegram.
type Client struct {
	conn  net.Conn
	clock func() time.Time
}

const defaultTimeout = time.Second * 5

func (c Client) startIntermediateMode(deadline time.Time) error {
	if err := c.conn.SetDeadline(deadline); err != nil {
		return fmt.Errorf("failed to set deadline: %w", err)
	}
	if _, err := c.conn.Write(proto.IntermediateClientStart); err != nil {
		return fmt.Errorf("failed to write start: %w", err)
	}
	if err := c.conn.SetDeadline(time.Time{}); err != nil {
		return fmt.Errorf("failed to reset connection deadline: %w", err)
	}
	return nil
}

type Options struct {
	Dialer  *net.Dialer
	Network string
	Addr    string
}

func Dial(ctx context.Context, opt Options) (*Client, error) {
	if opt.Dialer == nil {
		opt.Dialer = &net.Dialer{}
	}
	if opt.Network == "" {
		opt.Network = "tcp"
	}
	conn, err := opt.Dialer.DialContext(ctx, "tcp", opt.Addr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}
	client := &Client{
		conn:  conn,
		clock: time.Now,
	}
	return client, nil
}

// Connect establishes connection in intermediate mode.
func (c *Client) Connect(ctx context.Context) error {
	deadline := c.clock().Add(defaultTimeout)
	if ctxDeadline, ok := ctx.Deadline(); ok {
		deadline = ctxDeadline
	}
	if err := c.startIntermediateMode(deadline); err != nil {
		return err
	}
	return nil
}
