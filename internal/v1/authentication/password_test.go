package authentication

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPasswordChallengeWithUser(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	user, err := srv.UserStore.UserByID(context.Background(), 2)
	if !assert.NoError(t, err) {
		return
	}

	salt, err := base64.RawURLEncoding.DecodeString("_Jb4pAuatq5rrwdNRGRqW-PhlqzNR1pYtp1N5YWEn7s")
	assert.Nil(t, err)

	clientIdentity, serverIdentity := GetIdentity()
	spake2, err := NewSPAKE2Plus()
	assert.Nil(t, err)

	_, message, err := spake2.StartClient(clientIdentity, serverIdentity, []byte("password"), salt, nil)
	assert.Nil(t, err)

	challenge, err := srv.NewPasswordChallengeWithUser(context.Background(), user, message)

	if assert.Nil(t, err) {
		assert.NotNil(t, challenge)
		assert.NotNil(t, challenge.Token)
		assert.NotNil(t, challenge.Message)
	}
}

func TestVerifyPasswordResponseWithUser(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	user, err := srv.UserStore.UserByID(context.Background(), 2)
	if !assert.NoError(t, err) {
		return
	}

	salt, err := base64.RawURLEncoding.DecodeString("_Jb4pAuatq5rrwdNRGRqW-PhlqzNR1pYtp1N5YWEn7s")
	assert.Nil(t, err)

	clientIdentity, serverIdentity := GetIdentity()
	spake2, err := NewSPAKE2Plus()
	assert.Nil(t, err)

	// Correct password
	state, message, err := spake2.StartClient(clientIdentity, serverIdentity, []byte("password"), salt, nil)
	assert.Nil(t, err)

	challenge, err := srv.NewPasswordChallengeWithUser(context.Background(), user, message)
	assert.Nil(t, err)

	secret, err := state.Finish(challenge.Message)
	assert.Nil(t, err)

	err = srv.VerifyPasswordResponseWithUser(context.Background(), user, challenge.Token, secret.GetConfirmation())
	assert.NoError(t, err)

	// Incorrect password
	state2, message, err := spake2.StartClient(clientIdentity, serverIdentity, []byte("wrong_password"), salt, nil)
	assert.Nil(t, err)

	challenge2, err := srv.NewPasswordChallengeWithUser(context.Background(), user, message)
	assert.Nil(t, err)

	secret2, err := state2.Finish(challenge2.Message)
	assert.Nil(t, err)

	err = srv.VerifyPasswordResponseWithUser(context.Background(), user, challenge.Token, secret2.GetConfirmation())
	assert.Error(t, err)
}

func TestUpdatePasswordWithUser(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	ctx := context.Background()

	user, err := srv.UserStore.UserByID(ctx, 2)
	if !assert.NoError(t, err) {
		return
	}

	clientIdentity, serverIdentity := GetIdentity()
	spake2, err := NewSPAKE2Plus()
	assert.Nil(t, err)

	salt, verifierW0, verifierL, err := GenerateSPAKE2Verifier([]byte("password2"), clientIdentity, serverIdentity, spake2)

	err = user.SetPasswordVerifier(salt, verifierW0, verifierL)
	if !assert.NoError(t, err) {
		return
	}

	err = srv.UserStore.UpdateUser(ctx, user)
	if !assert.NoError(t, err) {
		return
	}

	// Verify password
	user, err = srv.UserStore.UserByID(context.Background(), 2)
	if !assert.NoError(t, err) {
		return
	}
	state, message, err := spake2.StartClient(clientIdentity, serverIdentity, []byte("password2"), salt, nil)
	assert.Nil(t, err)

	challenge, err := srv.NewPasswordChallengeWithUser(context.Background(), user, message)
	assert.Nil(t, err)

	secret, err := state.Finish(challenge.Message)
	assert.Nil(t, err)

	err = srv.VerifyPasswordResponseWithUser(context.Background(), user, challenge.Token, secret.GetConfirmation())
	assert.NoError(t, err)
}
