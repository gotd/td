package telegram_test

import (
	"context"
	"encoding/binary"
	"sync"
	"testing"
	"time"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/mt"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/proto"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"github.com/gotd/td/tgtest"
	"github.com/gotd/td/transport"
)

type tracedMessage struct {
	m    mtproto.WireMessage
	body []byte
}

// wireRecorder is a mtproto.WireTracer that records everything.
type wireRecorder struct {
	mux    sync.Mutex
	msgs   []tracedMessage
	events []mtproto.WireEvent
}

func (r *wireRecorder) OnMessage(m mtproto.WireMessage, body []byte) {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.msgs = append(r.msgs, tracedMessage{m: m, body: append([]byte(nil), body...)})
}

func (r *wireRecorder) OnEvent(e mtproto.WireEvent) {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.events = append(r.events, e)
}

// rpcResults returns req_msg_id of rpc_result messages in body, unpacking a
// container.
func rpcResults(body []byte) (ids []int64, container bool) {
	if len(body) < 4 {
		return nil, false
	}
	switch binary.LittleEndian.Uint32(body) {
	case proto.ResultTypeID:
		if len(body) >= 12 {
			ids = append(ids, int64(binary.LittleEndian.Uint64(body[4:])))
		}
	case proto.MessageContainerTypeID:
		var c proto.MessageContainer
		if err := c.Decode(&bin.Buffer{Buf: body}); err != nil {
			return nil, true
		}
		for _, m := range c.Messages {
			inner, _ := rpcResults(m.Body)
			ids = append(ids, inner...)
		}
		return ids, true
	}
	return ids, false
}

func TestWireTracer(t *testing.T) {
	rec := &wireRecorder{}
	var sendMsgID int64
	testCluster(transport.Intermediate, false, func(s clusterSetup) {
		c := s.Cluster
		c.Common().Vector(tg.UsersGetUsersRequestTypeID, user)
		// DC 2 migrates the request to DC 1.
		c.Dispatch(2, "migrate").HandleFunc(tg.MessagesSendMessageRequestTypeID,
			func(server *tgtest.Server, req *tgtest.Request) error {
				return server.SendErr(req, tgerr.New(303, "NETWORK_MIGRATE_1"))
			},
		)
		// DC 1 answers with a container of the result and an ack.
		c.Dispatch(1, "server").HandleFunc(tg.MessagesSendMessageRequestTypeID,
			func(server *tgtest.Server, req *tgtest.Request) error {
				sendMsgID = req.MsgID
				var result bin.Buffer
				if err := (&tg.Updates{}).Encode(&result); err != nil {
					return err
				}
				var res, ack bin.Buffer
				if err := (&proto.Result{RequestMessageID: req.MsgID, Result: result.Raw()}).Encode(&res); err != nil {
					return err
				}
				if err := (&mt.MsgsAck{MsgIDs: []int64{req.MsgID}}).Encode(&ack); err != nil {
					return err
				}
				ids := proto.NewMessageIDGen(time.Now)
				return server.Send(req.RequestCtx, req.Session, proto.MessageServerResponse, &proto.MessageContainer{
					Messages: []proto.Message{
						{ID: ids.New(proto.MessageServerResponse), SeqNo: 1, Bytes: ack.Len(), Body: ack.Raw()},
						{ID: ids.New(proto.MessageServerResponse), SeqNo: 3, Bytes: res.Len(), Body: res.Raw()},
					},
				})
			},
		)
	}, func(ctx context.Context, c clientSetup) error {
		opts := c.Options
		opts.WireTracer = rec
		client := telegram.NewClient(1, "hash", opts)
		return client.Run(ctx, func(ctx context.Context) error {
			if err := client.SendMessage(ctx, &tg.MessagesSendMessageRequest{
				Peer:    &tg.InputPeerUser{},
				Message: "abc",
			}); err != nil {
				return errors.Wrap(err, "send")
			}
			c.Complete()
			return nil
		})
	})(t)

	a := require.New(t)
	rec.mux.Lock()
	defer rec.mux.Unlock()

	type dirKey struct {
		dc  int
		dir mtproto.WireDirection
	}
	dirs := map[dirKey]int{}
	out := map[int64]mtproto.WireMessage{}
	var (
		results   = map[int64]bool{}
		container bool
	)
	for _, m := range rec.msgs {
		dirs[dirKey{dc: m.m.Conn.DC % 10000, dir: m.m.Direction}]++
		a.NotZero(m.m.Conn.ID)
		a.NotZero(m.m.MessageID)
		a.NotEqual([8]byte{}, m.m.AuthKeyID)
		if m.m.Direction == mtproto.WireOut {
			out[m.m.MessageID] = m.m
			continue
		}
		ids, isContainer := rpcResults(m.body)
		if isContainer {
			container = true
		}
		for _, id := range ids {
			results[id] = true
		}
	}
	for _, dc := range []int{1, 2} {
		a.Positive(dirs[dirKey{dc: dc, dir: mtproto.WireOut}], "out messages of DC %d", dc)
		a.Positive(dirs[dirKey{dc: dc, dir: mtproto.WireIn}], "in messages of DC %d", dc)
	}
	a.True(container, "incoming container not observed")
	a.NotZero(sendMsgID)
	a.True(results[sendMsgID], "container result for messages.sendMessage")

	// Every rpc_result answers an outgoing message.
	for id := range results {
		_, ok := out[id]
		a.True(ok, "rpc_result for unknown msg_id %d", id)
	}

	kinds := map[mtproto.WireEventType]int{}
	requests := map[int64]mtproto.WireEvent{}
	var migrated bool
	for _, e := range rec.events {
		kinds[e.Type]++
		a.NotZero(e.Conn.ID)
		switch e.Type {
		case mtproto.WireRequest:
			requests[e.MessageID] = e
		case mtproto.WireRequestDone:
			req, ok := requests[e.MessageID]
			a.True(ok, "done before request %d", e.MessageID)
			if tgerr.Is(e.Err, "NETWORK_MIGRATE") {
				migrated = true
				a.Equal(2, e.Conn.DC%10000)
			}
			if _, rpcErr := tgerr.As(e.Err); e.Err != nil && !rpcErr {
				// Canceled on shutdown, maybe before sending.
				continue
			}
			m, ok := out[e.MessageID]
			a.True(ok, "request %d not sent", e.MessageID)
			a.Equal(req.Conn, m.Conn)
			a.Equal(req.SeqNo, m.SeqNo)
		}
	}
	a.True(migrated, "migrate error not reported")
	for _, k := range []mtproto.WireEventType{
		mtproto.WireConnStart,
		mtproto.WireConnUp,
		mtproto.WireConnDown,
		mtproto.WireKeyCreated,
		mtproto.WireKeySet,
		mtproto.WireSessionNew,
		mtproto.WireRequest,
		mtproto.WireRequestDone,
	} {
		a.Positive(kinds[k], "no %s events", k)
	}
}
