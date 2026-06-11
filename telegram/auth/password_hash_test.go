package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

// recordingClient records which password method the flow used.
type recordingClient struct {
	FlowClient
	passwordCalled     bool
	passwordWithCalled bool
	hashInvoked        bool
}

func (r *recordingClient) Password(context.Context, string) (*tg.AuthAuthorization, error) {
	r.passwordCalled = true
	return &tg.AuthAuthorization{}, nil
}

func (r *recordingClient) PasswordWith(ctx context.Context, hash PasswordHashFunc) (*tg.AuthAuthorization, error) {
	r.passwordWithCalled = true
	if _, err := hash(ctx, &tg.AccountPassword{}); err == nil {
		r.hashInvoked = true
	}
	return &tg.AuthAuthorization{}, nil
}

// hashAuth is a UserAuthenticator that also provides the SRP answer directly.
type hashAuth struct {
	UserAuthenticator
}

func (hashAuth) PasswordHash(context.Context, *tg.AccountPassword) (*tg.InputCheckPasswordSRP, error) {
	return &tg.InputCheckPasswordSRP{}, nil
}

func noCode() CodeAuthenticator {
	return CodeAuthenticatorFunc(func(context.Context, *tg.AuthSentCode) (string, error) { return "", nil })
}

func TestFlowPasswordPrefersHashProvider(t *testing.T) {
	ctx := context.Background()

	t.Run("HashProvider", func(t *testing.T) {
		f := NewFlow(hashAuth{UserAuthenticator: Constant("phone", "pw", noCode())}, SendCodeOptions{})
		client := &recordingClient{}
		require.NoError(t, f.password(ctx, client))
		require.True(t, client.passwordWithCalled, "should use the hash path")
		require.True(t, client.hashInvoked, "hash callback should be invoked")
		require.False(t, client.passwordCalled, "plaintext path must be skipped")
	})

	t.Run("StringFallback", func(t *testing.T) {
		f := NewFlow(Constant("phone", "pw", noCode()), SendCodeOptions{})
		client := &recordingClient{}
		require.NoError(t, f.password(ctx, client))
		require.True(t, client.passwordCalled, "should fall back to the string path")
		require.False(t, client.passwordWithCalled)
	})
}
