package telegram

import (
	"context"
	"testing"
	"time"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/gotd/log/logzap"

	"github.com/gotd/td/clock"
	"github.com/gotd/td/session"
	"github.com/gotd/td/tdsync"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgtest/cluster"
	"github.com/gotd/td/transport"
)

// skewedClock is the system clock shifted by skew.
type skewedClock struct{ skew time.Duration }

func (c skewedClock) Now() time.Time                      { return clock.System.Now().Add(c.skew) }
func (c skewedClock) Timer(d time.Duration) clock.Timer   { return clock.System.Timer(d) }
func (c skewedClock) Ticker(d time.Duration) clock.Ticker { return clock.System.Ticker(d) }

// A local clock 10 minutes off made every server message look too old or too
// new: the key exchange succeeded, then everything was dropped and Run hung
// (parity MT-2). The server time offset is now taken from received msg_ids.
func TestClientServerTimeOffset(t *testing.T) {
	for _, skew := range []time.Duration{10 * time.Minute, -10 * time.Minute} {
		t.Run(skew.String(), func(t *testing.T) {
			const dcID = 2
			zapLog := zaptest.NewLogger(t)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			c := cluster.NewCluster(cluster.Options{
				Logger:   logzap.New(zapLog.Named("cluster")),
				Protocol: transport.Intermediate,
			})
			c.Dispatch(dcID, "dc").Vector(tg.UsersGetUsersRequestTypeID, &tg.User{ID: 10})

			var offset time.Duration
			g := tdsync.NewCancellableGroup(ctx)
			g.Go(c.Up)
			g.Go(func(ctx context.Context) error {
				defer cancel()
				select {
				case <-c.Ready():
				case <-ctx.Done():
					return ctx.Err()
				}
				client := NewClient(1, "hash", Options{
					PublicKeys:     c.Keys(),
					Resolver:       c.Resolver(),
					DCList:         c.List(),
					DC:             dcID,
					Logger:         logzap.New(zapLog.Named("client")),
					SessionStorage: &session.StorageMemory{},
					NoUpdates:      true,
					Clock:          skewedClock{skew: skew},
				})
				return client.Run(ctx, func(ctx context.Context) error {
					if _, err := client.API().UsersGetUsers(ctx, []tg.InputUserClass{&tg.InputUserSelf{}}); err != nil {
						return errors.Wrap(err, "users.getUsers")
					}
					offset = client.ServerTimeOffset()
					return nil
				})
			})
			err := g.Wait()
			if errors.Is(err, context.Canceled) {
				err = nil
			}
			require.NoError(t, err)
			require.InDelta(t, float64(-skew), float64(offset), float64(2*time.Second),
				"offset %s for clock skew %s", offset, skew)
		})
	}
}
