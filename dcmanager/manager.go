package dcmanager

import (
	"context"
	"sync"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
)

type Manager struct {
	primary   Conn
	migrating bool
	callbacks []func()
	migmux    sync.RWMutex

	others map[int]Conn
	omux   sync.Mutex

	cfg    Config
	cfgMux sync.RWMutex

	g *errgroup.Group

	// Immutable fields
	fetchConfig bool                      // Indicates whether we should fetch config from server
	createConn  CreateConnFunc            // Creates connections
	onMessage   func(b *bin.Buffer) error // Updates handler for primary connection
	saveConfig  func(cfg Config) error    // Config saver function
	appID       int                       // For connection init
	device      DeviceConfig              // For connection init
	transport   *transport.Transport      // MTProto optional param
	network     string                    // MTProto optional param
	log         *zap.Logger               // Logger
}

func New(appID int, opts Options) *Manager {
	opts.setDefaults()

	m := &Manager{
		others:     map[int]Conn{},
		g:          &errgroup.Group{},
		createConn: opts.ConnCreator,
		onMessage:  opts.UpdateHandler,
		saveConfig: opts.ConfigHandler,
		appID:      appID,
		device:     opts.Device,
		transport:  opts.Transport,
		network:    opts.Network,
		log:        opts.Logger,
	}

	if opts.Config != nil {
		m.cfg = *opts.Config
	} else {
		m.fetchConfig = true
	}

	return m
}

func (m *Manager) Run(ctx context.Context, f func(context.Context) error) error {
	if m.fetchConfig {
		// 149.154.175.55
		// "2|" + telegram.AddrProduction
		m.log.Info("Fetching config from server")
		if err := m.initWithoutConfig(ctx, "1|149.154.175.55:443"); err != nil {
			return err
		}
	} else {
		m.log.Info("Using loaded config")
		if err := m.initWithConfig(ctx); err != nil {
			return err
		}
	}

	m.g.Go(func() error { return f(ctx) })
	return m.g.Wait()
}

func (m *Manager) InvokeRaw(ctx context.Context, in bin.Encoder, out bin.Decoder) error {
	m.migmux.RLock()
	primary := m.primary
	m.migmux.RUnlock()

	if err := primary.InvokeRaw(ctx, in, out); err != nil {
		// Handling datacenter migration request.
		if rpcErr, ok := mtproto.AsErr(err); ok && rpcErr.IsCode(303) {
			// If migration error is FILE_MIGRATE or STATS_MIGRATE, then the method
			// called by authorized client, so we should try to transfer auth to new DC
			// and create new connection.
			if rpcErr.IsOneOf("FILE_MIGRATE", "STATS_MIGRATE") {
				m.log.Info("Got migrate error: Creating sub-connection",
					zap.String("error", rpcErr.Type), zap.Int("dc", rpcErr.Argument),
				)
				return m.invokeDC(ctx, rpcErr.Argument, in, out)
			}

			m.log.Info("Got migrate error",
				zap.String("error", rpcErr.Type), zap.Int("dc", rpcErr.Argument),
			)

			// Prevent parallel migrations.
			cb, migrate := m.tryMigrate()
			if !migrate {
				m.log.Info("Other goroutine has already started migration (waiting for completion)")
				cb()
				m.log.Info("Other goroutine has completed the migration, re-invoking request on new DC")
				return m.InvokeRaw(ctx, in, out)
			}

			m.log.Info("Starting migration to another DC", zap.Int("dc", rpcErr.Argument))
			defer cb()
			dcInfo, err := m.lookupDC(rpcErr.Argument)
			if err != nil {
				return err
			}

			// Change primary DC.
			// TODO(ccln): change ctx
			if _, err := m.dc(dcInfo).AsPrimary().Connect(ctx); err != nil {
				return xerrors.Errorf("migrate to dc %d: %w", rpcErr.Argument, err)
			}

			m.log.Info("Migration completed, re-invoking request on new DC")
			return m.InvokeRaw(ctx, in, out)
		}
		return err
	}
	return nil
}

func (m *Manager) invokeDC(ctx context.Context, dcID int, in bin.Encoder, out bin.Decoder) (err error) {
	m.omux.Lock()
	conn, found := m.others[dcID]
	if !found {
		dcInfo, err := m.lookupDC(dcID)
		if err != nil {
			m.omux.Unlock()
			return err
		}

		// TODO(ccln): change ctx
		conn, err = m.dc(dcInfo).WithAuthTransfer().Connect(ctx)
		if err != nil {
			m.omux.Unlock()
			return xerrors.Errorf("dial dc %d: %w", dcID, err)
		}

		m.others[dcID] = conn
	}
	m.omux.Unlock()

	return conn.InvokeRaw(ctx, &tg.InvokeWithoutUpdatesRequest{
		Query: nopDecoder{in},
	}, out)
}

type nopDecoder struct {
	bin.Encoder
}

func (n nopDecoder) Decode(b *bin.Buffer) error { return xerrors.New("not implemented") }
