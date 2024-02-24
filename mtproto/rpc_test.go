package mtproto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/mt"
	"github.com/gotd/td/rpc"
	"github.com/gotd/td/testutil"
)

func TestConn_dropRPC(t *testing.T) {
	dropID := int64(10)

	tests := []struct {
		name      string
		result    bin.Encoder
		resultErr error
		wantErr   bool
	}{
		{"Dropped", &mt.RPCAnswerDropped{MsgID: dropID}, nil, false},
		{"DroppedRunning", &mt.RPCAnswerDroppedRunning{}, nil, false},
		{"Unknown", &mt.RPCAnswerUnknown{}, nil, true},
		{"Error", nil, testutil.TestError(), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)

			client := newTestClient(func(msgID int64, seqNo int32, body bin.Encoder) (bin.Encoder, error) {
				req, ok := body.(*mt.RPCDropAnswerRequest)
				a.True(ok)
				if ok {
					a.Equal(dropID, req.ReqMsgID)
				}
				return tt.result, tt.resultErr
			})

			err := client.dropRPC(rpc.Request{
				MsgID: dropID,
				SeqNo: 1,
			})
			if tt.wantErr {
				a.Error(err)
			} else {
				a.NoError(err)
			}
		})
	}
}
