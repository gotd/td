package auth_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/internal/testutil"
	"github.com/nnqq/td/telegram/auth"
	"github.com/nnqq/td/tg"
)

func askCode(code string, err error) auth.CodeAuthenticatorFunc {
	return func(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
		return code, err
	}
}

func TestConstantAuth(t *testing.T) {
	a := require.New(t)
	authConst := auth.Constant("phone", "password", askCode("123", nil))
	ctx := context.Background()

	result, err := authConst.Code(ctx, nil)
	a.NoError(err)
	a.Equal("123", result)

	result, err = authConst.Phone(ctx)
	a.NoError(err)
	a.Equal("phone", result)

	result, err = authConst.Password(ctx)
	a.NoError(err)
	a.Equal("password", result)
}

func TestCodeOnlyAuth(t *testing.T) {
	a := require.New(t)
	authCodeOnly := auth.CodeOnly("phone", askCode("123", nil))
	ctx := context.Background()

	result, err := authCodeOnly.Code(ctx, nil)
	a.NoError(err)
	a.Equal("123", result)

	result, err = authCodeOnly.Phone(ctx)
	a.NoError(err)
	a.Equal("phone", result)

	_, err = authCodeOnly.Password(ctx)
	a.ErrorIs(err, auth.ErrPasswordNotProvided)
}

func TestEnvAuth(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()

	prefix := "TEST_ENV_AUTH_"
	authEnv := auth.Env(prefix, askCode("123", nil))

	result, err := authEnv.Code(ctx, nil)
	a.NoError(err)
	a.Equal("123", result)

	_, err = authEnv.Phone(ctx)
	a.Error(err)

	_, err = authEnv.Password(ctx)
	a.ErrorIs(err, auth.ErrPasswordNotProvided)

	// Set envs.
	testutil.SetEnv(t, prefix+"PHONE", "phone")
	testutil.SetEnv(t, prefix+"PASSWORD", "password")

	result, err = authEnv.Phone(ctx)
	a.NoError(err)
	a.Equal("phone", result)

	result, err = authEnv.Password(ctx)
	a.NoError(err)
	a.Equal("password", result)
}
