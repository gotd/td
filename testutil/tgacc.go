package testutil

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/go-faster/errors"
	"github.com/google/uuid"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/testutil/tgacc"
	"github.com/gotd/td/tg"
)

// TestAccountManager is external test account manager.
type TestAccountManager struct {
	client *tgacc.Client
}

func (t *TestAccountManager) Acquire(ctx context.Context) (*TestAccount, error) {
	jobID := os.Getenv("GITHUB_JOB_ID")
	if jobID == "" {
		return nil, errors.New("GITHUB_JOB_ID is empty")
	}
	runID, _ := strconv.ParseInt(os.Getenv("GITHUB_RUN_ID"), 10, 64)
	if runID == 0 {
		return nil, errors.New("GITHUB_RUN_ID is empty")
	}
	attempt, _ := strconv.Atoi(os.Getenv("GITHUB_RUN_ATTEMPT"))
	res, err := t.client.AcquireTelegramAccount(ctx, &tgacc.AcquireTelegramAccountReq{
		RepoOwner:  "gotd",
		RepoName:   "td",
		Job:        jobID,
		RunID:      runID,
		RunAttempt: attempt,
	})
	if err != nil {
		return nil, errors.Wrap(err, "acquire account")
	}

	phone := string(res.AccountID)

	return &TestAccount{
		Phone: phone,
		AuthFlow: &codeAuth{
			phone:  phone,
			client: t.client,
		},

		token:  res.Token,
		client: t.client,
	}, nil
}

type ghSecuritySource struct{}

func (s ghSecuritySource) TokenAuth(ctx context.Context, operationName tgacc.OperationName) (tgacc.TokenAuth, error) {
	return tgacc.TokenAuth{
		APIKey: os.Getenv("GITHUB_TOKEN"),
	}, nil
}

type TestAccount struct {
	Phone    string
	AuthFlow auth.UserAuthenticator

	token  uuid.UUID
	client *tgacc.Client
}

// Close releases telegram account.
func (t TestAccount) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	return t.client.HeartbeatTelegramAccount(ctx, tgacc.HeartbeatTelegramAccountParams{
		Token:  t.token,
		Forget: tgacc.NewOptBool(true),
	})
}

// codeAuth implements auth.UserAuthenticator prompting the external account
// manager.
type codeAuth struct {
	phone  string
	token  uuid.UUID
	client *tgacc.Client
}

func (codeAuth) SignUp(ctx context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, errors.New("not implemented")
}

func (codeAuth) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	return &auth.SignUpRequired{TermsOfService: tos}
}

func (a codeAuth) Phone(_ context.Context) (string, error) {
	return a.phone, nil
}

func (codeAuth) Password(_ context.Context) (string, error) {
	return "", errors.New("password not supported")
}

func (a codeAuth) Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = time.Minute
	bo.MaxInterval = time.Second

	return backoff.RetryWithData(func() (string, error) {
		res, err := a.client.ReceiveTelegramCode(ctx, tgacc.ReceiveTelegramCodeParams{
			Token: a.token,
		})
		if err != nil {
			return "", err
		}
		if res.Code.Value == "" {
			return "", errors.New("no code")
		}
		return res.Code.Value, err
	}, bo)
}

func NewTestAccountManager() (*TestAccountManager, error) {
	client, err := tgacc.NewClient("https://bot.gotd.dev", ghSecuritySource{})
	if err != nil {
		return nil, errors.Wrap(err, "create client")
	}
	return &TestAccountManager{client: client}, nil
}