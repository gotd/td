package telegram

import (
	"context"
	"io"

	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/proto"
)

type UpdateHandler func(ctx context.Context, updates *tg.Updates) error

type Client struct {
	RPC *tg.Client
	mtp MTProto

	appID   int
	appHash string

	dh     *dataHandler
	rand   io.Reader
	ctx    context.Context
	cancel context.CancelFunc

	log *zap.Logger
}

func New(appID int, appHash string, opts Options) (*Client, error) {
	opts.setDefaults()

	ctx, cancel := context.WithCancel(context.Background())

	dh := newDataHandler(
		ctx,
		opts.Logger.Named("update_handler"),
		opts.UpdateHandler,
	)

	opts.MTProto.Handler = dh.handleData

	mtp := mtproto.NewClient(appID, appHash, opts.MTProto)
	if err := mtp.Connect(context.TODO()); err != nil {
		return nil, err
	}

	// TODO(ernado): Make versions configurable.
	const notAvailable = "n/a"

	var response tg.Config
	if err := mtp.InvokeRaw(context.TODO(), proto.InvokeWithLayer{
		Layer: tg.Layer,
		Query: proto.InitConnection{
			ID:             appID,
			SystemLangCode: "en",
			LangCode:       "en",
			SystemVersion:  notAvailable,
			DeviceModel:    notAvailable,
			AppVersion:     notAvailable,
			LangPack:       "",
			Query:          proto.GetConfig{},
		},
	}, &response); err != nil {
		return nil, xerrors.Errorf("request: %w", err)
	}

	client := &Client{
		RPC:     tg.NewClient(mtp),
		mtp:     mtp,
		appID:   appID,
		appHash: appHash,
		dh:      dh,
		rand:    opts.Random,
		ctx:     ctx,
		cancel:  cancel,
		log:     opts.Logger,
	}

	return client, nil
}

func (c *Client) Close() error {
	c.cancel()
	return c.mtp.Close()
}
