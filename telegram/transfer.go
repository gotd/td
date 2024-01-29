package telegram

import (
	"context"
	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"

	"sync"
)

var authorizationTransferLocks = make(map[int64]*sync.Mutex)
var authorizationTransferLocksMutex = sync.Mutex{}

func (c *Client) exportAuth(ctx context.Context, dcID int) (*tg.AuthExportedAuthorization, error) {
	export, err := c.tg.AuthExportAuthorization(ctx, dcID)
	if err != nil {
		return nil, errors.Wrapf(err, "export auth to %d", dcID)
	}

	return export, nil
}

// transfer exports current authorization and imports it to another DC.
// See https://core.telegram.org/api/datacenter#authorization-transfer.
func (c *Client) transfer(ctx context.Context, to *tg.Client, dc int) (tg.AuthAuthorizationClass, error) {
	u, err := c.Self(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "get self")
	}

	authorizationTransferLocksMutex.Lock()
	mx, ok := authorizationTransferLocks[u.ID]
	if !ok {
		mx = &sync.Mutex{}
		authorizationTransferLocks[u.ID] = mx
	}
	authorizationTransferLocksMutex.Unlock()

	mx.Lock()
	defer mx.Unlock()

	auth, err := c.exportAuth(ctx, dc)
	if err != nil {
		return nil, errors.Wrapf(err, "export to %d", dc)
	}

	req := &tg.AuthImportAuthorizationRequest{}
	req.FillFrom(auth)
	r, err := to.AuthImportAuthorization(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "import from %d", dc)
	}

	return r, nil
}
