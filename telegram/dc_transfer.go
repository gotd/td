package telegram

import (
	"context"

	"github.com/gotd/td/tg"
	"golang.org/x/xerrors"
)

// transfer exports current authorization and imports it to another DC.
// See https://core.telegram.org/api/datacenter#authorization-transfer.
func transfer(ctx context.Context, from, to *tg.Client, dc int) error {
	auth, err := from.AuthExportAuthorization(ctx, dc)
	if err != nil {
		return xerrors.Errorf("export auth: %w", err)
	}

	result, err := to.AuthImportAuthorization(ctx, &tg.AuthImportAuthorizationRequest{
		ID:    auth.ID,
		Bytes: auth.Bytes,
	})
	if err != nil {
		return xerrors.Errorf("import to dc %d: %w", dc, err)
	}

	return checkAuthResult(result)
}
