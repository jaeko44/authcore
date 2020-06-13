package authapi

import (
	"context"
	"testing"

	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/api/authapi"

	"github.com/stretchr/testify/assert"
)

func TestDeleteOAuthFactor(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 10)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// Should delete OAuth factor as that is not the only one
	req1 := &authapi.DeleteOAuthFactorRequest{
		Id: "3",
	}

	_, err = srv.DeleteOAuthFactor(ctx, req1)
	if assert.NoError(t, err) {
		oAuthList, err := srv.ListOAuthFactors(ctx, &authapi.ListOAuthFactorsRequest{})
		if assert.NoError(t, err) {
			assert.Len(t, oAuthList.OauthFactors, 1)
		}
	}

	// Should return error as the account does not set password as there is only one OAuth factor
	req2 := &authapi.DeleteOAuthFactorRequest{
		Id: "2",
	}
	_, err = srv.DeleteOAuthFactor(ctx, req2)
	assert.Error(t, err)
}
