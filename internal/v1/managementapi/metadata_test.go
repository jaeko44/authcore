package managementapi

import (
	"context"
	"testing"

	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/api/managementapi"

	"github.com/stretchr/testify/assert"
)

// A user should be able to get the metadata of a given user
func TestGetMetadata(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. Get metadata
	req := &managementapi.GetMetadataRequest{
		UserId: "1",
	}
	res, err := srv.GetMetadata(ctx, req)

	assert.Equal(t, `{"favourite_links":["https://github.com","https://blocksq.com"]}`, res.UserMetadata)
	assert.Equal(t, `{"kyc_status":false}`, res.AppMetadata)
	assert.NoError(t, err)
}

// A user should be able to update the metadata of a given user
func TestUpdateMetadata(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. Update metadata
	req := &managementapi.UpdateMetadataRequest{
		UserId:       "1",
		UserMetadata: `{"favourite_links":["https://github.com","https://google.com","https://blocksq.com"]}`,
		AppMetadata:  `{"kyc_status":true}`,
	}
	_, err = srv.UpdateMetadata(ctx, req)
	assert.NoError(t, err)

	// 2. Get the updated user from DB and verify
	updatedUser, err := srv.UserStore.UserByID(context.Background(), 1)
	assert.NoError(t, err)
	updatedUserMetadata, err := updatedUser.UserMetadata.String()
	assert.Equal(t, `{"favourite_links":["https://github.com","https://google.com","https://blocksq.com"]}`, updatedUserMetadata)
	assert.NoError(t, err)
	updatedAppMetadata, err := updatedUser.AppMetadata.String()
	assert.Equal(t, `{"kyc_status":true}`, updatedAppMetadata)
	assert.NoError(t, err)
}
