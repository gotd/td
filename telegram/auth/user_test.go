package auth

import (
	"context"
	"encoding/hex"
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/tg"
	"github.com/nnqq/td/tgerr"
	"github.com/nnqq/td/tgmock"
)

func getHex(t testing.TB, in string) []byte {
	res, err := hex.DecodeString(in)
	if err != nil {
		t.Fatal("failed to get hex", err)
	}
	return res
}

func TestClient_AuthSignIn(t *testing.T) {
	const (
		phone    = "123123"
		code     = "1010"
		password = "secret"
		codeHash = "hash"
	)
	ctx := context.Background()
	testUser := &tg.User{ID: 1}
	invoker := tgmock.Invoker(func(body bin.Encoder) (bin.Encoder, error) {
		switch req := body.(type) {
		case *tg.AuthSendCodeRequest:
			settings := tg.CodeSettings{}
			settings.SetCurrentNumber(true)
			assert.Equal(t, &tg.AuthSendCodeRequest{
				PhoneNumber: phone,
				APIHash:     testAppHash,
				APIID:       testAppID,
				Settings:    settings,
			}, req)
			return &tg.AuthSentCode{
				Type:          &tg.AuthSentCodeTypeApp{},
				PhoneCodeHash: codeHash,
			}, nil
		case *tg.AuthSignInRequest:
			assert.Equal(t, &tg.AuthSignInRequest{
				PhoneNumber:   phone,
				PhoneCodeHash: codeHash,
				PhoneCode:     code,
			}, req)
			return nil, tgerr.New(401, "SESSION_PASSWORD_NEEDED")
		case *tg.AccountGetPasswordRequest:
			algo := &tg.PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow{
				Salt1: getHex(t, "4D11FB6BEC38F9D2546BB0F61E4F1C99A1BC0DB8F0D5F35B1291B37B213123D7ED48F3C6794D495B"),
				Salt2: getHex(t, "A1B181AAFE88188680AE32860D60BB01"),
				G:     3,
				P: getHex(t, "C71CAEB9C6B1C9048E6C522F70F13F73980D40238E3E21C14934D037563D930F"+
					"48198A0AA7C14058229493D22530F4DBFA336F6E0AC925139543AED44CCE7C37"+
					"20FD51F69458705AC68CD4FE6B6B13ABDC9746512969328454F18FAF8C595F64"+
					"2477FE96BB2A941D5BCD1D4AC8CC49880708FA9B378E3C4F3A9060BEE67CF9A4"+
					"A4A695811051907E162753B56B0F6B410DBA74D8A84B2A14B3144E0EF1284754"+
					"FD17ED950D5965B4B9DD46582DB1178D169C6BC465B0D6FF9CA3928FEF5B9AE4"+
					"E418FC15E83EBEA0F87FA9FF5EED70050DED2849F47BF959D956850CE929851F"+
					"0D8115F635B105EE2E4E15D04B2454BF6F4FADF034B10403119CD8E3B92FCC5B"),
			}
			pwd := &tg.AccountPassword{
				NewAlgo:       algo,
				NewSecureAlgo: &tg.SecurePasswordKdfAlgoPBKDF2HMACSHA512iter100000{},
			}
			pwd.SetCurrentAlgo(algo)
			return pwd, nil
		case *tg.AuthCheckPasswordRequest:
			// TODO(ernado): Check actual secure remote password here.
			switch pwd := req.Password.(type) {
			case *tg.InputCheckPasswordSRP:
				assert.NotEmpty(t, pwd.A)
				assert.NotEmpty(t, pwd.M1)
				assert.NotEqual(t, pwd.SRPID, 0)
			default:
				t.Errorf("unexpectd pwd type %T", pwd)
			}
			return &tg.AuthAuthorization{
				User: testUser,
			}, nil
		}
		return nil, xerrors.New("unexpected")
	})

	t.Run("Manual", func(t *testing.T) {
		// 1. Request code from server to device.
		client := testClient(invoker)
		sentCode, err := client.SendCode(ctx, phone, SendCodeOptions{CurrentNumber: true})
		require.NoError(t, err)
		h := sentCode.PhoneCodeHash
		require.Equal(t, codeHash, h)

		// 2. Send code from device to server.
		// Server is responding with 2FA password prompt.
		_, signInErr := client.SignIn(ctx, phone, code, h)
		require.ErrorIs(t, signInErr, ErrPasswordAuthNeeded)

		// 3. Provide 2FA password.
		result, err := client.Password(ctx, password)
		require.NoError(t, err)
		require.Equal(t, testUser, result.User)
	})
	t.Run("AuthFlow", func(t *testing.T) {
		// Using flow helper.
		u := Constant(phone, password, CodeAuthenticatorFunc(
			func(ctx context.Context, _ *tg.AuthSentCode) (string, error) {
				return code, nil
			},
		))
		require.NoError(t, NewFlow(u, SendCodeOptions{CurrentNumber: true}).Run(ctx, testClient(invoker)))
	})
}

func TestClientTestAuth(t *testing.T) {
	const (
		codeHash = "hash"
		dcID     = 2
	)
	ctx := context.Background()
	invoker := tgmock.Invoker(func(body bin.Encoder) (bin.Encoder, error) {
		switch req := body.(type) {
		case *tg.AuthSendCodeRequest:
			assert.Equal(t, &tg.AuthSendCodeRequest{
				PhoneNumber: req.PhoneNumber,
				APIHash:     testAppHash,
				APIID:       testAppID,
				Settings:    tg.CodeSettings{},
			}, req)
			return &tg.AuthSentCode{
				Type: &tg.AuthSentCodeTypeApp{
					Length: 6,
				},
				PhoneCodeHash: codeHash,
			}, nil
		case *tg.AuthSignInRequest:
			if !strings.HasPrefix(req.PhoneNumber, "99966") {
				t.Fatalf("unexpected phone number %s", req.PhoneNumber)
			}
			dcPart := req.PhoneNumber[5:6]
			assert.Equal(t, strconv.Itoa(dcID), dcPart, "dc part of phone number")
			assert.Equal(t, &tg.AuthSignInRequest{
				PhoneNumber:   req.PhoneNumber,
				PhoneCodeHash: codeHash,
				PhoneCode:     strings.Repeat(dcPart, 6),
			}, req)
			return &tg.AuthAuthorization{
				User: &tg.User{ID: 1},
			}, nil
		}
		return nil, xerrors.New("unexpected")
	})
	require.NoError(t, NewFlow(
		Test(rand.New(rand.NewSource(1)), dcID),
		SendCodeOptions{},
	).Run(ctx, testClient(invoker)))
}

func TestClientTestSignUp(t *testing.T) {
	const (
		dcID     = 2
		codeHash = "hash"
		tosID    = "foo"
	)
	ctx := context.Background()
	invoker := tgmock.Invoker(func(body bin.Encoder) (bin.Encoder, error) {
		switch req := body.(type) {
		case *tg.AuthSendCodeRequest:
			assert.Equal(t, &tg.AuthSendCodeRequest{
				PhoneNumber: req.PhoneNumber,
				APIHash:     testAppHash,
				APIID:       testAppID,
				Settings:    tg.CodeSettings{},
			}, req)
			return &tg.AuthSentCode{
				Type: &tg.AuthSentCodeTypeApp{
					Length: 6,
				},
				PhoneCodeHash: codeHash,
			}, nil
		case *tg.AuthSignUpRequest:
			assert.Equal(t, &tg.AuthSignUpRequest{
				PhoneNumber:   req.PhoneNumber,
				PhoneCodeHash: codeHash,
				FirstName:     "Test",
				LastName:      "User",
			}, req)
			return &tg.AuthAuthorization{
				User: &tg.User{ID: 1},
			}, nil
		case *tg.HelpAcceptTermsOfServiceRequest:
			return &tg.BoolTrue{}, nil
		case *tg.AuthSignInRequest:
			if !strings.HasPrefix(req.PhoneNumber, "99966") {
				t.Fatalf("unexpected phone number %s", req.PhoneNumber)
			}
			dcPart := req.PhoneNumber[5:6]
			assert.Equal(t, strconv.Itoa(dcID), dcPart, "dc part of phone number")
			assert.Equal(t, &tg.AuthSignInRequest{
				PhoneNumber:   req.PhoneNumber,
				PhoneCodeHash: codeHash,
				PhoneCode:     strings.Repeat(dcPart, 6),
			}, req)

			res := &tg.AuthAuthorizationSignUpRequired{}
			res.SetTermsOfService(tg.HelpTermsOfService{ID: tg.DataJSON{Data: tosID}})

			return res, nil
		}
		return nil, xerrors.New("unexpected")
	})
	require.NoError(t, NewFlow(
		Test(rand.New(rand.NewSource(1)), dcID),
		SendCodeOptions{},
	).Run(ctx, testClient(invoker)))
}
