package telegram

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConstantAuth(t *testing.T) {
	askCode := CodeAuthenticatorFunc(func(ctx context.Context) (string, error) {
		return "123", nil
	})

	a := require.New(t)
	auth := ConstantAuth("phone", "password", askCode)
	ctx := context.Background()

	result, err := auth.Code(ctx)
	a.NoError(err)
	a.Equal("123", result)

	result, err = auth.Phone(ctx)
	a.NoError(err)
	a.Equal("phone", result)

	result, err = auth.Password(ctx)
	a.NoError(err)
	a.Equal("password", result)
}
