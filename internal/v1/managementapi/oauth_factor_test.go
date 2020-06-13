package managementapi

import (
	"context"
	"testing"

	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/api/managementapi"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestListOAuthFactor(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	oAuthList, err := srv.ListOAuthFactors(ctx, &managementapi.ListOAuthFactorsRequest{
		UserId: "10",
	})
	if assert.NoError(t, err) {
		assert.Len(t, oAuthList.OauthFactors, 2)
	}
}

func TestDeleteOAuthFactor(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	oAuthList, err := srv.ListOAuthFactors(ctx, &managementapi.ListOAuthFactorsRequest{
		UserId: "10",
	})
	if assert.NoError(t, err) {
		assert.Len(t, oAuthList.OauthFactors, 2)
	}

	req1 := &managementapi.DeleteOAuthFactorRequest{
		Id: "3",
	}

	_, err = srv.DeleteOAuthFactor(ctx, req1)
	if assert.NoError(t, err) {
		oAuthList, err := srv.ListOAuthFactors(ctx, &managementapi.ListOAuthFactorsRequest{
			UserId: "10",
		})
		if assert.NoError(t, err) {
			assert.Len(t, oAuthList.OauthFactors, 1)
		}
	}

	// Remove only one Oauth factor without setting password
	req2 := &managementapi.DeleteOAuthFactorRequest{
		Id: "2",
	}

	_, err = srv.DeleteOAuthFactor(ctx, req2)
	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.InvalidArgument, status.Code())
		}
	}
}

func TestListOAuthFactorUnauthenticated(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	req := &managementapi.ListOAuthFactorsRequest{
		UserId: "10",
	}
	_, err := srv.ListOAuthFactors(context.Background(), req)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.Unauthenticated, status.Code())
		}
	}
}

func TestListOAuthFactorUnauthorized(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 3)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	_, err = srv.ListOAuthFactors(ctx, &managementapi.ListOAuthFactorsRequest{
		UserId: "10",
	})

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.PermissionDenied, status.Code())
		}
	}
}
