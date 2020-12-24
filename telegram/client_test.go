package telegram

import (
	"context"
	"math/rand"
	"time"

	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram/internal/rpc"
	"github.com/gotd/td/tg"
)

type testHandler func(id int64, body bin.Encoder) (bin.Encoder, error)

func testError(err tg.Error) (bin.Encoder, error) {
	e := &Error{
		Message: err.Text,
		Code:    err.Code,
	}
	e.extractArgument()
	return nil, e
}

func newTestClient(h testHandler) *Client {
	var engine *rpc.Engine

	engine = rpc.New(func(ctx context.Context, msgID int64, seqNo int32, in bin.Encoder) error {
		if response, err := h(msgID, in); err != nil {
			engine.NotifyError(msgID, err)
		} else {
			var b bin.Buffer
			if err := b.Encode(response); err != nil {
				return err
			}
			return engine.NotifyResult(msgID, &b)
		}
		return nil
	}, rpc.Config{})

	client := &Client{
		rpc:            engine,
		log:            zap.NewNop(),
		clock:          time.Now,
		rand:           rand.New(rand.NewSource(1)),
		sessionCreated: createCondOnce(),
		appID:          TestAppID,
		appHash:        TestAppHash,
	}
	client.tg = tg.NewClient(client)
	client.sessionCreated.Done()

	return client
}
