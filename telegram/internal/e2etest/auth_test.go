package e2etest

import (
	"context"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

type mockFlow struct {
	flag bool
}

var _ auth.FlowClient = &mockFlow{}

func (m *mockFlow) SignIn(context.Context, string, string, string) (*tg.AuthAuthorization, error) {
	// Ensure retry.
	if !m.flag {
		m.flag = true
		return nil, auth.ErrPasswordAuthNeeded
	}

	return m.Password(context.Background(), "")
}

func (m *mockFlow) SendCode(context.Context, string, auth.SendCodeOptions) (tg.AuthSentCodeClass, error) {
	return &tg.AuthSentCode{
		PhoneCodeHash: "hash",
		Type:          &tg.AuthSentCodeTypeApp{},
		Timeout:       10,
	}, nil
}

func (m *mockFlow) Password(context.Context, string) (*tg.AuthAuthorization, error) {
	return &tg.AuthAuthorization{
		User: &tg.User{
			ID:       10,
			Username: "aboba",
		},
	}, nil
}

func (m *mockFlow) SignUp(context.Context, auth.SignUp) (*tg.AuthAuthorization, error) {
	return nil, errors.New("must not be called")
}

func TestSuite_Authenticate(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	s := NewSuite(t, TestOptions{
		Logger: logger,
	})
	if s.manager != nil {
		t.Skip("Not testing external manager")
	}

	flow := &mockFlow{}
	require.NoError(t, s.Authenticate(ctx, flow))
}
