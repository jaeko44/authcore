package managementapi

import (
	"context"
	"testing"

	"authcore.io/authcore/internal/v1/authentication"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/api/authapi"
	"authcore.io/authcore/pkg/api/managementapi"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestChangePassword(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	clientIdentity, serverIdentity := authentication.GetIdentity()
	spake2, err := authentication.NewSPAKE2Plus()
	assert.Nil(t, err)

	salt, verifierW0, verifierL, err := authentication.GenerateSPAKE2Verifier([]byte("new_password"), clientIdentity, serverIdentity, spake2)
	assert.Nil(t, err)

	// 1. Change password
	req := &managementapi.ChangePasswordRequest{
		UserId: "1",
		PasswordVerifier: &authapi.PasswordVerifier{
			Salt:       salt,
			VerifierW0: verifierW0,
			VerifierL:  verifierL,
		},
	}
	res, err := srv.ChangePassword(ctx, req)

	if assert.Nil(t, err) {
		assert.NotNil(t, res)
	}
}

func TestChangePasswordWhenUnauthenticated(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	clientIdentity, serverIdentity := authentication.GetIdentity()
	spake2, err := authentication.NewSPAKE2Plus()
	assert.Nil(t, err)

	salt, verifierW0, verifierL, err := authentication.GenerateSPAKE2Verifier([]byte("new_password"), clientIdentity, serverIdentity, spake2)
	assert.Nil(t, err)

	// 1. Change password while the account is not authenticated
	req := &managementapi.ChangePasswordRequest{
		UserId: "1",
		PasswordVerifier: &authapi.PasswordVerifier{
			Salt:       salt,
			VerifierW0: verifierW0,
			VerifierL:  verifierL,
		},
	}
	_, err = srv.ChangePassword(context.Background(), req)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.Unauthenticated, status.Code())
		}
	}
}

func TestChangePasswordWhenUnauthorized(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 3)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	clientIdentity, serverIdentity := authentication.GetIdentity()
	spake2, err := authentication.NewSPAKE2Plus()
	assert.Nil(t, err)

	salt, verifierW0, verifierL, err := authentication.GenerateSPAKE2Verifier([]byte("new_password"), clientIdentity, serverIdentity, spake2)
	assert.Nil(t, err)

	// 1. Change password while the account is not authorized
	req := &managementapi.ChangePasswordRequest{
		UserId: "1",
		PasswordVerifier: &authapi.PasswordVerifier{
			Salt:       salt,
			VerifierW0: verifierW0,
			VerifierL:  verifierL,
		},
	}
	_, err = srv.ChangePassword(ctx, req)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.PermissionDenied, status.Code())
		}
	}
}
