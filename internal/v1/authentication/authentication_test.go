package authentication

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateAuthenticationState(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	// Valid - UserID=0 and it is for OAuth
	_, err := srv.CreateAuthenticationState(context.Background(), "authcore-io", 0, 1, []string{"OAUTH"}, "", "")
	assert.Nil(t, err)

	// Valid - UserID!=0 and it is not for OAuth
	_, err = srv.CreateAuthenticationState(context.Background(), "authcore-io", 1, 1, []string{"SECURE_REMOTE_PASSWORD"}, "", "")
	assert.Nil(t, err)

	// Invalid - UserID=0 but it is not for OAuth
	_, err = srv.CreateAuthenticationState(context.Background(), "authcore-io", 0, 1, []string{"SECURE_REMOTE_PASSWORD"}, "", "")
	assert.Error(t, err)

	// Invalid - UserID!=0 but it is for OAuth
	_, err = srv.CreateAuthenticationState(context.Background(), "authcore-io", 1, 1, []string{"OAUTH"}, "", "")
	assert.Error(t, err)

	// Invalid - The challenge set is invalid
	_, err = srv.CreateAuthenticationState(context.Background(), "authcore-io", 0, 1, []string{"SECURE_REMOTE_PASSWORD", "OAUTH"}, "", "")
	assert.Error(t, err)
	_, err = srv.CreateAuthenticationState(context.Background(), "authcore-io", 1, 1, []string{"SECURE_REMOTE_PASSWORD", "OAUTH"}, "", "")
	assert.Error(t, err)

	// Valid - The application can redirect to https://example.com/
	_, err = srv.CreateAuthenticationState(context.Background(), "authcore-io", 1, 1, []string{"SECURE_REMOTE_PASSWORD"}, "", "https://example.com/")
	assert.Nil(t, err)

	// Invalid - The application cannot redirect to https://evil.com/
	_, err = srv.CreateAuthenticationState(context.Background(), "authcore-io", 1, 1, []string{"SECURE_REMOTE_PASSWORD"}, "", "https://evil.com/")
	assert.Error(t, err)
}

func TestCreateUnauthenticationState(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	authState, err := srv.CreateAuthenticationState(context.Background(), "authcore-io", 1, 1, []string{"SECURE_REMOTE_PASSWORD"}, "", "")
	if assert.Nil(t, err) {
		assert.Equal(t, int64(1), authState.UserID)
		assert.NotEmpty(t, authState.TemporaryToken)
	}

	authState2, err := srv.FindAuthenticationStateByTemporaryToken(context.Background(), authState.TemporaryToken)
	if assert.Nil(t, err) {
		assert.Equal(t, int64(1), authState2.UserID)
		assert.Equal(t, authState.TemporaryToken, authState2.TemporaryToken)
		assert.Len(t, authState.Challenges, 1)
		assert.Equal(t, "SECURE_REMOTE_PASSWORD", authState.Challenges[0])
		assert.Equal(t, "", authState.PKCEChallenge)
	}

	err = srv.DeleteAuthenticationStateByTemporaryToken(context.Background(), authState.TemporaryToken)
	assert.NoError(t, err)

	_, err = srv.FindAuthenticationStateByTemporaryToken(context.Background(), authState.TemporaryToken)
	assert.Error(t, err)
}

func TestCreateAuthenticationStateWithPKCECodeChallenge(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	authState, err := srv.CreateAuthenticationState(context.Background(), "authcore-io", 1, 1, []string{"SECURE_REMOTE_PASSWORD"}, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", "")
	if assert.Nil(t, err) {
		assert.Equal(t, int64(1), authState.UserID)
		assert.NotEmpty(t, authState.TemporaryToken)
	}

	authState2, err := srv.FindAuthenticationStateByTemporaryToken(context.Background(), authState.TemporaryToken)
	if assert.Nil(t, err) {
		assert.Equal(t, int64(1), authState2.UserID)
		assert.Equal(t, authState.TemporaryToken, authState2.TemporaryToken)
		assert.Len(t, authState.Challenges, 1)
		assert.Equal(t, "SECURE_REMOTE_PASSWORD", authState.Challenges[0])
		assert.Equal(t, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", authState.PKCEChallenge)
	}

	err = srv.DeleteAuthenticationStateByTemporaryToken(context.Background(), authState.TemporaryToken)
	assert.NoError(t, err)

	_, err = srv.FindAuthenticationStateByTemporaryToken(context.Background(), authState.TemporaryToken)
	assert.Error(t, err)
}
