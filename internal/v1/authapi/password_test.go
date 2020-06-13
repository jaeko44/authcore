package authapi

import (
	"context"
	"encoding/base64"
	"testing"

	"authcore.io/authcore/internal/v1/authentication"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/api/authapi"

	"github.com/stretchr/testify/assert"
)

// Change password without previous password set.
func TestFinishChangePassword(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	salt, _ := base64.RawURLEncoding.DecodeString("_Jb4pAuatq5rrwdNRGRqW-PhlqzNR1pYtp1N5YWEn7s")
	verifierW0, _ := base64.RawURLEncoding.DecodeString("H9EeC9z9ndtqPVIz59_hWUUh8_TFdowJApvxHkbRhTZeTsrue0cxUgqUkZ_3QJShr3sjEVFbs_L5Ca3LFIHbPlpWULzMUxmbZSVDQkLSQMdxxxNP1CH9")
	verifierL, _ := base64.RawURLEncoding.DecodeString("89obToiiylZJ2bWw9neAUtD-Xvu_zhhj-HHzQveMHMUNhFZh719_tYgBRvp2LRflO6Rko9q7bUCCRgz4mSBYSibpmCo9y8GoFvWBarSUu-dBqAh2OMVT_ifCPAu2qLqdFJQZRAzM")

	req := &authapi.FinishChangePasswordRequest{
		PasswordVerifier: &authapi.PasswordVerifier{
			Salt:       salt,
			VerifierW0: verifierW0,
			VerifierL:  verifierL,
		},
	}
	res, err := srv.FinishChangePassword(ctx, req)

	if assert.Nil(t, err) {
		assert.NotNil(t, res)
	}
}

func TestChangePasswordKeyExchange(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 2)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 0. Get Password parameters
	req := &authapi.StartChangePasswordRequest{}
	res, err := srv.StartChangePassword(ctx, req)
	assert.Nil(t, err)

	// 1. Starts the create password challenge
	clientIdentity, serverIdentity := authentication.GetIdentity()
	spake2, err := authentication.NewSPAKE2Plus()
	_, message, err := spake2.StartClient(clientIdentity, serverIdentity, []byte("password"), res.Salt, nil)

	assert.Nil(t, err)

	req1 := &authapi.ChangePasswordKeyExchangeRequest{
		Message: message,
	}

	res1, err := srv.ChangePasswordKeyExchange(ctx, req1)

	if assert.Nil(t, err) {
		assert.NotNil(t, res1)
		assert.NotNil(t, res1.Token)
		assert.NotNil(t, res1.Message)
	}
}

func TestFinishChangePasswordWithOldPassword(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 2)
	if !assert.NoError(t, err) {
		return
	}

	ctx := user.NewContextWithCurrentUser(context.Background(), u)
	// 0. Get Password parameters
	req := &authapi.StartChangePasswordRequest{}
	res, err := srv.StartChangePassword(ctx, req)

	assert.Nil(t, err)

	// 1. Starts the create password challenge
	clientIdentity, serverIdentity := authentication.GetIdentity()
	spake2, err := authentication.NewSPAKE2Plus()
	state, message, err := spake2.StartClient(clientIdentity, serverIdentity, []byte("password"), res.Salt, nil)

	assert.Nil(t, err)

	req1 := &authapi.ChangePasswordKeyExchangeRequest{
		Message: message,
	}

	res1, err := srv.ChangePasswordKeyExchange(ctx, req1)
	assert.Nil(t, err)

	// 2. Change Password
	secret, err := state.Finish(res1.Message)
	assert.Nil(t, err)

	confirmation := secret.GetConfirmation()

	salt, verifierW0, verifierL, err := authentication.GenerateSPAKE2Verifier([]byte("new_password"), clientIdentity, serverIdentity, spake2)
	assert.Nil(t, err)

	req2 := &authapi.FinishChangePasswordRequest{
		PasswordVerifier: &authapi.PasswordVerifier{
			Salt:       salt,
			VerifierW0: verifierW0,
			VerifierL:  verifierL,
		},
		OldPasswordResponse: &authapi.PasswordResponse{
			Token:        res1.Token,
			Confirmation: confirmation,
		},
	}
	res2, err := srv.FinishChangePassword(ctx, req2)

	if assert.NoError(t, err) {
		assert.NotNil(t, res2)
	}
}

func TestChangePasswordMissingOldPassword(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 2)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	clientIdentity, serverIdentity := authentication.GetIdentity()
	spake2, err := authentication.NewSPAKE2Plus()
	salt, verifierW0, verifierL, err := authentication.GenerateSPAKE2Verifier([]byte("new_password"), clientIdentity, serverIdentity, spake2)
	req := &authapi.FinishChangePasswordRequest{
		PasswordVerifier: &authapi.PasswordVerifier{
			Salt:       salt,
			VerifierW0: verifierW0,
			VerifierL:  verifierL,
		},
	}
	_, err = srv.FinishChangePassword(ctx, req)

	assert.Error(t, err)
}

func TestChangePasswordIncorrectOldPassword(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 2)
	assert.Nil(t, err)

	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 0. Get Password parameters
	req := &authapi.StartChangePasswordRequest{}
	res, err := srv.StartChangePassword(ctx, req)

	assert.Nil(t, err)

	// 1. Starts the create password challenge
	clientIdentity, serverIdentity := authentication.GetIdentity()
	spake2, err := authentication.NewSPAKE2Plus()
	state, message, err := spake2.StartClient(clientIdentity, serverIdentity, []byte("wrong_password"), res.Salt, nil)

	req1 := &authapi.ChangePasswordKeyExchangeRequest{
		Message: message,
	}

	res1, err := srv.ChangePasswordKeyExchange(ctx, req1)
	assert.Nil(t, err)

	secret, err := state.Finish(res1.Message)
	assert.Nil(t, err)

	salt, verifierW0, verifierL, err := authentication.GenerateSPAKE2Verifier([]byte("new_password"), clientIdentity, serverIdentity, spake2)
	assert.Nil(t, err)

	req2 := &authapi.FinishChangePasswordRequest{
		PasswordVerifier: &authapi.PasswordVerifier{
			Salt:       salt,
			VerifierW0: verifierW0,
			VerifierL:  verifierL,
		},
		OldPasswordResponse: &authapi.PasswordResponse{
			Token:        res1.Token,
			Confirmation: secret.GetConfirmation(),
		},
	}

	_, err = srv.FinishChangePassword(ctx, req2)
	assert.Error(t, err)

	// make sure password is not changed
	u, err = srv.UserStore.UserByID(context.Background(), 2)
	if assert.NoError(t, err) {
		assert.Equal(t, "_Jb4pAuatq5rrwdNRGRqW-PhlqzNR1pYtp1N5YWEn7s", u.PasswordSaltBase64.String)
		assert.Equal(t, "H9EeC9z9ndtqPVIz59_hWUUh8_TFdowJApvxHkbRhTZeTsrue0cxUgqUkZ_3QJShr3sjEVFbs_L5Ca3LFIHbPlpWULzMUxmbZSVDQkLSQMdxxxNP1CH9", u.EncryptedPasswordVerifierW0.String)
		assert.Equal(t, "89obToiiylZJ2bWw9neAUtD-Xvu_zhhj-HHzQveMHMUNhFZh719_tYgBRvp2LRflO6Rko9q7bUCCRgz4mSBYSibpmCo9y8GoFvWBarSUu-dBqAh2OMVT_ifCPAu2qLqdFJQZRAzM", u.EncryptedPasswordVerifierL.String)
	}
}
