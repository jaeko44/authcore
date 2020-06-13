package authapi

import (
	"context"
	"testing"

	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/api/authapi"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

// Create user in the server.
func TestCreateUser(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	req := &authapi.CreateUserRequest{
		Username:    "alice",
		Email:       "alice@example.com",
		Phone:       "+85212345678",
		DisplayName: "Alice",
	}
	ctx := context.Background()
	m := make(map[string]string)
	m["grpcgateway-origin"] = "http://0.0.0.0:8000"
	ctx = metadata.NewIncomingContext(ctx, metadata.New(m))

	res, err := srv.CreateUser(ctx, req)

	if assert.Nil(t, err) {
		assert.NotNil(t, res)
		assert.Len(t, res.RefreshToken, 43)
		u := res.User
		assert.NotEqual(t, "", u.Id)
		assert.Equal(t, "alice", u.Username)
		assert.Equal(t, "alice@example.com", u.PrimaryEmail)
		assert.Equal(t, "+85212345678", u.PrimaryPhone)
		assert.NotZero(t, u.CreatedAt)
		assert.NotZero(t, u.UpdatedAt)
	}
}

func TestGetCurrentUser(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}

	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	req := &empty.Empty{}
	res, err := srv.GetCurrentUser(ctx, req)
	if assert.NoError(t, err) {
		assert.Equal(t, u.PublicID(), res.Id)
		assert.Equal(t, u.Username.String, res.Username)
		assert.Equal(t, u.Email.String, res.PrimaryEmail)
		assert.Equal(t, u.Phone.String, res.PrimaryPhone)
		assert.Equal(t, u.DisplayNameOld, res.DisplayName)
	}
}

func TestUpdateCurrentUser(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}

	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	newUserInfo, _ := MarshalUser(u)
	newUserInfo.DisplayName = "newName"
	req := &authapi.UpdateCurrentUserRequest{
		User: newUserInfo,
	}
	res, err := srv.UpdateCurrentUser(ctx, req)
	if assert.NoError(t, err) {
		assert.Equal(t, u.PublicID(), res.Id)
		assert.Equal(t, u.Username.String, res.Username)
		assert.Equal(t, u.Email.String, res.PrimaryEmail)
		assert.Equal(t, u.Phone.String, res.PrimaryPhone)
		assert.Equal(t, newUserInfo.DisplayName, res.DisplayName)
	} else {
		return
	}

	// second time, unchange display name
	newUserInfo, _ = MarshalUser(u)
	newUserInfo.DisplayName = "newName"
	req = &authapi.UpdateCurrentUserRequest{
		User: newUserInfo,
	}
	res, err = srv.UpdateCurrentUser(ctx, req)
	if assert.NoError(t, err) {
		assert.Equal(t, u.PublicID(), res.Id)
		assert.Equal(t, u.Username.String, res.Username)
		assert.Equal(t, u.Email.String, res.PrimaryEmail)
		assert.Equal(t, u.Phone.String, res.PrimaryPhone)
		assert.Equal(t, newUserInfo.DisplayName, res.DisplayName)
	} else {
		return
	}

	// change username instead, should not return error as user is not modified.
	newUserInfo, _ = MarshalUser(u)
	newUserInfo.Username = "noUpdateUsername"
	req = &authapi.UpdateCurrentUserRequest{
		User: newUserInfo,
	}
	_, err = srv.UpdateCurrentUser(ctx, req)
	assert.NoError(t, err)
}

func TestMarshalUser(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if assert.NoError(t, err) {
		apiUser, err := MarshalUser(u)
		if assert.NoError(t, err) {
			assert.Equal(t, "1", apiUser.Id)
			assert.Equal(t, "bob@example.com", apiUser.PrimaryEmail)
			assert.Equal(t, "+85223456789", apiUser.PrimaryPhone)
		}
	}
}

// Create user that contains capital letters and dots in the email address.
// Reference: https://gitlab.com/blocksq/authcore/issues/563
func TestCreateUserWithCapitalLettersAndDotsInEmail(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	req := &authapi.CreateUserRequest{
		Username:    "",
		Email:       "Alice.Bob@example.com",
		Phone:       "",
		DisplayName: "Alice.Bob@example.com",
	}
	ctx := context.Background()
	m := make(map[string]string)
	m["grpcgateway-origin"] = "http://0.0.0.0:8000"
	ctx = metadata.NewIncomingContext(ctx, metadata.New(m))

	_, err := srv.CreateUser(ctx, req)
	assert.NoError(t, err)
}

func TestCreateUserWithInvalidEmailAddress(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	ctx := context.Background()
	m := make(map[string]string)
	m["grpcgateway-origin"] = "http://0.0.0.0:8000"
	ctx = metadata.NewIncomingContext(ctx, metadata.New(m))

	req := &authapi.CreateUserRequest{
		Username:    "",
		Email:       "IDontHaveADomain",
		Phone:       "",
		DisplayName: "IDontHaveADomain",
	}
	_, err := srv.CreateUser(ctx, req)
	assert.Error(t, err)

	req2 := &authapi.CreateUserRequest{
		Username:    "",
		Email:       "multiple_ats_are_not_allowed@@email.com",
		Phone:       "",
		DisplayName: "multiple_ats_are_not_allowed@@email.com",
	}
	_, err = srv.CreateUser(ctx, req2)
	assert.Error(t, err)
}
