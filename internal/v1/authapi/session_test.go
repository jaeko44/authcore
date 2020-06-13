package authapi

import (
	"context"
	"testing"

	"authcore.io/authcore/internal/session"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/api/authapi"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestListSessions(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)
	sess, err := srv.SessionStore.FindSessionByInternalID(ctx, 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx = session.NewContextWithCurrentSession(ctx, sess)

	req := &authapi.ListSessionsRequest{
		PageSize: 3,
	}
	res, err := srv.ListSessions(ctx, req)

	if assert.NoError(t, err) {
		assert.Equal(t, int32(3), res.TotalSize)
		assert.NotEmpty(t, res.Sessions)
		assert.Len(t, res.Sessions, 3)
		assert.True(t, res.Sessions[2].IsCurrent)
	}
}

func TestCreateMachineToken(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	req := &authapi.CreateMachineTokenRequest{}
	res, err := srv.CreateMachineToken(ctx, req)

	if assert.NoError(t, err) {
		assert.NotEmpty(t, res.MachineToken)
	}

	// Can get an access token with the machine token
	req2 := &authapi.CreateAccessTokenRequest{
		GrantType: authapi.CreateAccessTokenRequest_REFRESH_TOKEN,
		Token:     res.MachineToken,
	}
	_, err = srv.CreateAccessToken(context.Background(), req2)
	assert.NoError(t, err)
}

func TestDeleteSession(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	sess, err := srv.SessionStore.FindSessionByInternalID(ctx, 2)
	if !assert.NoError(t, err) {
		return
	}
	ctx = session.NewContextWithCurrentSession(ctx, sess)

	req := &authapi.ListSessionsRequest{}
	res, err := srv.ListSessions(ctx, req)

	if assert.NoError(t, err) {
		assert.Equal(t, int32(3), res.TotalSize)
	}

	req2 := &authapi.DeleteSessionRequest{
		SessionId: "1",
	}
	_, err = srv.DeleteSession(ctx, req2)

	assert.NoError(t, err)

	res, err = srv.ListSessions(ctx, req)

	if assert.NoError(t, err) {
		assert.Equal(t, int32(2), res.TotalSize)
	}
}

func TestDeleteCurrentSession(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	sess, err := srv.SessionStore.FindSessionByInternalID(ctx, 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx = session.NewContextWithCurrentSession(ctx, sess)

	// 1. Delete the current
	req := &authapi.DeleteCurrentSessionRequest{}
	_, err = srv.DeleteCurrentSession(ctx, req)
	assert.NoError(t, err)

	// Sessions can no longer be found in the database
	_, err = srv.SessionStore.FindSessionByInternalID(ctx, 1)
	assert.Error(t, err)
}

// A user should not be able to delete other users' session
func TestDeleteSessionOtherUser(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	req := &authapi.DeleteSessionRequest{
		SessionId: "4",
	}

	_, err = srv.DeleteSession(ctx, req)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.Unauthenticated, status.Code())
		}
	}
}
