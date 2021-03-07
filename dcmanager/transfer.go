package dcmanager

import (
	"context"

	"github.com/gotd/td/tg"
	"golang.org/x/xerrors"
)

// transfer exports current authorization and imports it to another DC.
// See https://core.telegram.org/api/datacenter#authorization-transfer.
func (m *Manager) transfer(ctx context.Context, conn Conn, dc int) error {
	m.mux.RLock()
	var (
		from = tg.NewClient(m.primary)
		to   = tg.NewClient(conn)
	)
	m.mux.RUnlock()

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

// checkAuthResult checks that a is *tg.AuthAuthorization.
func checkAuthResult(a tg.AuthAuthorizationClass) error {
	switch v := a.(type) {
	case *tg.AuthAuthorization:
		// Ok.
		return nil
	case *tg.AuthAuthorizationSignUpRequired:
		// return &telegram.SignUpRequired{
		// 	TermsOfService: v.TermsOfService,
		// }
		_ = v
		return xerrors.Errorf("sign up required")
	default:
		return xerrors.Errorf("got unexpected response %T", a)
	}
}
