package peers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestReport(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	reason := &tg.InputReportReasonSpam{}
	message := "message"
	peers := []Peer{
		m.User(getTestUser()),
		m.Chat(getTestChat()),
		m.Channel(getTestChannel()),
	}

	for _, p := range peers {
		req := &tg.AccountReportPeerRequest{
			Peer:    p.InputPeer(),
			Reason:  reason,
			Message: message,
		}

		mock.ExpectCall(req).ThenRPCErr(getTestError())
		a.Error(p.Report(ctx, reason, message))

		mock.ExpectCall(req).ThenTrue()
		a.NoError(p.Report(ctx, reason, message))
	}
}
