package dcmanager

import (
	"context"
	"sync"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/dcmanager/mtp"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
)

type Manager struct {
	primary *mtp.Conn
	others  map[int]*mtp.Conn
	cfg     Config
	mux     sync.RWMutex

	// Immutable fields
	onMessage  func(b *bin.Buffer) error // Updates handler for primary connection
	saveConfig func(cfg Config) error    // Config saver function
	appID      int                       // For connection init
	device     DeviceConfig              // For connection init
	transport  *transport.Transport      // MTProto optional param
	network    string                    // MTProto optional param
	log        *zap.Logger               // Logger
}

func New(appID int, opts Options) (*Manager, error) {
	opts.setDefaults()

	m := &Manager{
		others:     map[int]*mtp.Conn{},
		onMessage:  opts.UpdateHandler,
		saveConfig: opts.ConfigHandler,
		appID:      appID,
		device:     opts.Device,
		transport:  opts.Transport,
		network:    opts.Network,
		log:        opts.Logger,
	}

	if opts.Config == nil {
		// 149.154.175.55
		// "2|" + telegram.AddrProduction
		if err := m.initWithoutConfig("1|149.154.175.55:443"); err != nil {
			return nil, err
		}

		return m, nil
	}

	if err := m.initWithConfig(*opts.Config); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Manager) InvokeRaw(ctx context.Context, in bin.Encoder, out bin.Decoder) error {
	m.mux.RLock()
	primary := m.primary
	m.mux.RUnlock()
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

			m.log.Info("Got migrate error: Starting migration to another DC",
				zap.String("error", rpcErr.Type), zap.Int("dc", rpcErr.Argument),
			)

			dcInfo, err := m.lookupDC(rpcErr.Argument)
			if err != nil {
				return err
			}

			// Otherwise we should change primary DC.
			if _, err := m.dc(dcInfo).AsPrimary().Connect(ctx); err != nil {
				return xerrors.Errorf("migrate to dc %d: %w", rpcErr.Argument, err)
			}

			return m.InvokeRaw(ctx, in, out)
		}

		return err
	}
	return nil
}

func (m *Manager) invokeDC(ctx context.Context, dcID int, in bin.Encoder, out bin.Decoder) (err error) {
	m.mux.RLock()
	conn, found := m.others[dcID]
	m.mux.RUnlock()
	if !found {
		dcInfo, err := m.lookupDC(dcID)
		if err != nil {
			return err
		}

		conn, err = m.dc(dcInfo).WithAuthTransfer().Connect(ctx)
		if err != nil {
			return xerrors.Errorf("dial dc %d: %w", dcID, err)
		}

		m.mux.Lock()
		m.others[dcID] = conn
		m.mux.Unlock()
	}

	return conn.InvokeRaw(ctx, &tg.InvokeWithoutUpdatesRequest{
		Query: nopDecoder{in},
	}, out)
}

type nopDecoder struct {
	bin.Encoder
}

func (n nopDecoder) Decode(b *bin.Buffer) error { return xerrors.New("not implemented") }
