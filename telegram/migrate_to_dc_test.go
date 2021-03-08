package telegram

import (
	"context"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"go.uber.org/zap/zaptest"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/internal/pool"
	"github.com/gotd/td/internal/rpc"
	"github.com/gotd/td/internal/tdsync"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/telegram/internal/manager"
	"github.com/gotd/td/tg"
)

type migrationTestHandler func(conn, id int64, dc int, body bin.Encoder) (bin.Encoder, error)

type migrateTestConn struct {
	testConn
	dc     int
	addr   string
	cfg    tg.Config
	opts   mtproto.Options
	client *Client
}

func (c *migrateTestConn) Run(ctx context.Context) error {
	cfg := c.cfg
	cfg.ThisDC = c.dc
	if err := c.client.onSession(c.addr, cfg, mtproto.Session{
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
		id int64,
		mode manager.ConnMode,
		appID int, addr string,
		opts mtproto.Options, connOpts manager.ConnOptions,
	) pool.Conn {
		var engine *rpc.Engine

		ready := tdsync.NewReady()
		ready.Signal()
		engine = rpc.New(func(ctx context.Context, msgID int64, seqNo int32, in bin.Encoder) error {
			if response, err := h(id, msgID, connOpts.DC, in); err != nil {
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
			addr:     addr,
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
		primaryDC: *atomic.NewInt64(2),
		ctx:       context.Background(),
		cancel:    func() {},
	}
	client.init()
	client.conn = client.createConn(0, manager.ConnModeUpdates, nil)
	client.cfg.Store(cfg)
	return client
}

func TestMigration(t *testing.T) {
	ctx := context.Background()
	expected := &tg.BoolTrue{}
	a := require.New(t)

	client := newMigrationClient(t, func(conn, id int64, dc int, body bin.Encoder) (bin.Encoder, error) {
		switch body.(type) {
		case *tg.UsersGetUsersRequest:
			return nil, mtproto.NewError(401, "AUTH_KEY_UNREGISTERED")
		case *tg.AuthLogOutRequest:
			if dc == 2 {
				return nil, mtproto.NewError(303, "USER_MIGRATE_10")
			}

			a.Equal(10, dc)
			return expected, nil
		default:
			return nil, xerrors.Errorf("unexpected body %T", body)
		}
	})

	err := client.Run(ctx, func(ctx context.Context) error {
		var result tg.BoolTrue
		return client.InvokeRaw(ctx, &tg.AuthLogOutRequest{}, &result)
	})
	a.NoError(err)
}
