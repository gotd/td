package e2etest

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/go-faster/errors"

	"github.com/gotd/td/telegram/auth"
)

func (s *Suite) createFlow(ctx context.Context) (auth.Flow, error) {
	if s.manager != nil {
		account, err := s.manager.Acquire(ctx)
		if err != nil {
			return auth.Flow{}, errors.Wrap(err, "acquire account")
		}
		s.closers = append(s.closers, account.Close)
		return auth.NewFlow(account.UserAuthenticator, auth.SendCodeOptions{}), nil
	}

	var ua auth.UserAuthenticator
	for {
		ua = auth.Test(s.rand, s.dc)
		phone, err := ua.Phone(ctx)
		if err != nil {
			return auth.Flow{}, err
		}

		s.usedMux.Lock()
		if _, ok := s.used[phone]; !ok {
			s.used[phone] = struct{}{}
			s.usedMux.Unlock()
			break
		}
		s.usedMux.Unlock()
	}

	return auth.NewFlow(ua, auth.SendCodeOptions{}), nil
}

// Authenticate authenticates client on test server.
func (s *Suite) Authenticate(ctx context.Context, client auth.FlowClient) error {
	for {
		flow, err := s.createFlow(ctx)
		if err != nil {
			return errors.Wrap(err, "create flow")
		}

		if err := flow.Run(ctx, client); err != nil {
			if errors.Is(err, auth.ErrPasswordNotProvided) {
				continue
			}

			return errors.Wrap(err, "run flow")
		}
		return nil
	}
}

// RetryAuthenticate authenticates client on test server.
func (s *Suite) RetryAuthenticate(ctx context.Context, client auth.FlowClient) error {
	bc := backoff.NewExponentialBackOff()
	bc.MaxElapsedTime = time.Minute
	bc.MaxInterval = time.Second * 3
	bck := backoff.WithContext(bc, ctx)
	return backoff.Retry(func() error {
		return s.Authenticate(ctx, client)
	}, bck)
}
