package mtproto

import (
	"context"

	"github.com/gotd/td/bin"
)

// InvokeRaw sens input and decodes result into output.
//
// NOTE: Assuming that call contains content message (seqno increment).
func (c *Client) InvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	return c.rpcContent(ctx, input, output)
}
