package takeout

import (
	"context"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

type mockInvoker struct {
	initErr   error
	finishErr error
	invokeErr error
	takeoutID int64

	initCalled   bool
	finishCalled bool
	lastSuccess  bool
}

func (m *mockInvoker) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	switch input.(type) {
	case *tg.AccountInitTakeoutSessionRequest:
		m.initCalled = true
		if m.initErr != nil {
			return m.initErr
		}
		result := output.(*tg.AccountTakeout)
		result.ID = m.takeoutID
		return nil
	case *tg.AccountFinishTakeoutSessionRequest:
		m.finishCalled = true
		req := input.(*tg.AccountFinishTakeoutSessionRequest)
		m.lastSuccess = req.GetSuccess()
		if m.finishErr != nil {
			return m.finishErr
		}
		return nil
	case *tg.InvokeWithTakeoutRequest:
		return m.invokeErr
	}
	return nil
}

func TestRun(t *testing.T) {
	ctx := context.Background()
	testErr := errors.New("test error")

	t.Run("Success", func(t *testing.T) {
		mock := &mockInvoker{takeoutID: 123}

		err := Run(ctx, mock, Config{
			Contacts:     true,
			MessageUsers: true,
		}, func(ctx context.Context, client *Client) error {
			require.Equal(t, int64(123), client.ID())
			return nil
		})

		require.NoError(t, err)
		require.True(t, mock.initCalled)
		require.True(t, mock.finishCalled)
		require.True(t, mock.lastSuccess)
	})

	t.Run("FunctionError", func(t *testing.T) {
		mock := &mockInvoker{takeoutID: 123}

		err := Run(ctx, mock, Config{}, func(ctx context.Context, client *Client) error {
			return testErr
		})

		require.Error(t, err)
		require.True(t, mock.finishCalled)
		require.False(t, mock.lastSuccess)
	})

	t.Run("InitError", func(t *testing.T) {
		mock := &mockInvoker{initErr: testErr}

		err := Run(ctx, mock, Config{}, func(ctx context.Context, client *Client) error {
			t.Fatal("function should not be called")
			return nil
		})

		require.Error(t, err)
		require.Contains(t, err.Error(), "init takeout session")
	})

	t.Run("FinishError", func(t *testing.T) {
		mock := &mockInvoker{takeoutID: 123, finishErr: testErr}

		err := Run(ctx, mock, Config{}, func(ctx context.Context, client *Client) error {
			return nil
		})

		require.Error(t, err)
		require.Contains(t, err.Error(), "finish takeout session")
	})
}

func TestClient_Invoke(t *testing.T) {
	ctx := context.Background()

	invoked := false
	var capturedRequest *tg.InvokeWithTakeoutRequest

	mock := &tg.Client{}
	invoker := invokerFunc(func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
		invoked = true
		var ok bool
		capturedRequest, ok = input.(*tg.InvokeWithTakeoutRequest)
		if !ok {
			t.Fatalf("expected InvokeWithTakeoutRequest, got %T", input)
		}
		return nil
	})
	_ = mock

	client := &Client{
		id:  456,
		raw: invoker,
	}

	err := client.Invoke(ctx, &tg.MessagesGetHistoryRequest{}, &tg.MessagesMessagesBox{})
	require.NoError(t, err)
	require.True(t, invoked)
	require.NotNil(t, capturedRequest)
	require.Equal(t, int64(456), capturedRequest.TakeoutID)
}

type invokerFunc func(ctx context.Context, input bin.Encoder, output bin.Decoder) error

func (f invokerFunc) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	return f(ctx, input, output)
}

func TestConfig_request(t *testing.T) {
	cfg := Config{
		Contacts:          true,
		MessageUsers:      true,
		MessageChats:      true,
		MessageMegagroups: true,
		MessageChannels:   true,
		Files:             true,
		FileMaxSize:       1024 * 1024,
	}

	req := cfg.request()

	require.True(t, req.GetContacts())
	require.True(t, req.GetMessageUsers())
	require.True(t, req.GetMessageChats())
	require.True(t, req.GetMessageMegagroups())
	require.True(t, req.GetMessageChannels())
	require.True(t, req.GetFiles())
	size, ok := req.GetFileMaxSize()
	require.True(t, ok)
	require.Equal(t, int64(1024*1024), size)
}

func TestClient_Raw(t *testing.T) {
	mock := invokerFunc(func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
		return nil
	})

	client := &Client{
		id:  789,
		raw: mock,
	}

	require.Equal(t, int64(789), client.ID())
	require.NotNil(t, client.Raw())
}
