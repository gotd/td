package telegram

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/pool"
	"github.com/gotd/td/rpc"
	"github.com/gotd/td/tdsync"
	"github.com/gotd/td/telegram/internal/manager"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

type migrationTestHandler func(id int64, dc int, body bin.Encoder) (bin.Encoder, error)

type migrateTestConn struct {
	testConn
	dc     int
	cfg    tg.Config
	opts   mtproto.Options
	client *Client
}

func (c *migrateTestConn) Ping(ctx context.Context) error {
	return errors.New("not implemented")
}

func (c *migrateTestConn) Run(ctx context.Context) error {
	cfg := c.cfg
	cfg.ThisDC = c.dc
	if err := c.client.onSession(cfg, mtproto.Session{
		ID:   10,
		Key:  c.opts.Key,
		Salt: 10,
	}); err != nil {
		return err
	}

	<-ctx.Done()
	return ctx.Err()
}

func newMigrationClient(t *testing.T, h migrationTestHandler) *Client {
	cfg := tg.Config{
		ThisDC: 2,
		DCOptions: []tg.DCOption{
			{
				ID:        10,
				IPAddress: "10",
			},
			{
				ID:        2,
				IPAddress: "2",
			},
		},
	}

	var client *Client
	creator := func(
		create mtproto.Dialer,
		mode manager.ConnMode,
		appID int,
		opts mtproto.Options,
		connOpts manager.ConnOptions,
	) pool.Conn {
		var engine *rpc.Engine

		ready := tdsync.NewReady()
		ready.Signal()
		engine = rpc.New(func(ctx context.Context, msgID int64, seqNo int32, in bin.Encoder) error {
			if response, err := h(msgID, connOpts.DC, in); err != nil {
				engine.NotifyError(msgID, err)
			} else {
				var b bin.Buffer
				if err := b.Encode(response); err != nil {
					return err
				}
				return engine.NotifyResult(msgID, &b)
			}
			return nil
		}, rpc.Options{})

		return &migrateTestConn{
			testConn: testConn{engine: engine, ready: ready},
			dc:       connOpts.DC,
			cfg:      cfg,
			opts:     opts,
			client:   client,
		}
	}

	client = &Client{
		log:     zaptest.NewLogger(t),
		rand:    rand.New(rand.NewSource(1)),
		appID:   TestAppID,
		appHash: TestAppHash,
		create:  creator,
		clock:   clock.System,
		session: pool.NewSyncSession(pool.Session{
			DC: 2,
		}),
		newConnBackoff:   defaultBackoff(clock.System),
		ctx:              context.Background(),
		cancel:           func() {},
		migrationTimeout: 10 * time.Second,
	}
	client.init()
	client.conn = client.createConn(0, manager.ConnModeUpdates, nil, nil)
	client.cfg.Store(cfg)
	return client
}

func TestMigration(t *testing.T) {
	codes := []int{303, 400}

	for _, code := range codes {
		t.Run(fmt.Sprintf("Code%d", code), func(t *testing.T) {
			ctx := context.Background()
			expected := &tg.BoolTrue{}
			a := require.New(t)

			client := newMigrationClient(t, func(id int64, dc int, body bin.Encoder) (bin.Encoder, error) {
				switch body.(type) {
				case *tg.UsersGetUsersRequest:
					return nil, tgerr.New(401, "AUTH_KEY_UNREGISTERED")
				case *tg.AuthLogOutRequest:
					if dc == 2 {
						return nil, tgerr.New(code, "USER_MIGRATE_10")
					}

					a.Equal(10, dc)
					return expected, nil
				default:
					return nil, errors.Errorf("unexpected body %T", body)
				}
			})

			err := client.Run(ctx, func(ctx context.Context) error {
				var result tg.BoolTrue
				return client.Invoke(ctx, &tg.AuthLogOutRequest{}, &result)
			})
			a.NoError(err)
		})
	}
}
