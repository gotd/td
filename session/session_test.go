package session

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
)

func testStorage(storage Storage) func(t *testing.T) {
	ctx := context.Background()
	loader := Loader{
		Storage: storage,
	}

	return func(t *testing.T) {
		a := require.New(t)

		_, err := loader.Load(ctx)
		a.ErrorIs(err, ErrNotFound)

		data := &Data{
			Config:    tg.Config{},
			DC:        2,
			Addr:      "localhost:8080",
			AuthKey:   bytes.Repeat([]byte{'a'}, 256),
			AuthKeyID: []byte("gotd1337"),
			Salt:      10,
		}
		a.NoError(loader.Save(ctx, data))

		gotData, err := loader.Load(ctx)
		a.NoError(err)
		require.Equal(t, data, gotData)
	}
}
