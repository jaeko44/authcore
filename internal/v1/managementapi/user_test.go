package managementapi

import (
	"context"
	"testing"
	"time"

	"authcore.io/authcore/internal/v1/authentication"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/api/authapi"
	"authcore.io/authcore/pkg/api/managementapi"
	"authcore.io/authcore/pkg/nulls"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestListSingleUser(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// Test query by username
	req := &managementapi.ListUsersRequest{
		UserHandle: "factor",
	}
	res, err := srv.ListUsers(ctx, req)
	if assert.NoError(t, err) {
		assert.Equal(t, "", res.NextPageToken)
		assert.Equal(t, "", res.PreviousPageToken)
		assert.Len(t, res.Users, 1)
		assert.Equal(t, "factor", res.Users[0].Username)
		assert.Equal(t, "+85233333333", res.Users[0].PrimaryPhone)
		assert.Equal(t, "factor@example.com", res.Users[0].PrimaryEmail)
		assert.NotEmpty(t, res.Users[0].LastSeenAt)
	}

	// Test query by email
	req2 := &managementapi.ListUsersRequest{
		UserHandle: "factor@example.com",
	}
	res2, err := srv.ListUsers(ctx, req2)
	if assert.NoError(t, err) {
		assert.Equal(t, "", res.NextPageToken)
		assert.Equal(t, "", res.PreviousPageToken)
		assert.Len(t, res2.Users, 1)
		assert.Equal(t, "factor", res2.Users[0].Username)
		assert.Equal(t, "+85233333333", res2.Users[0].PrimaryPhone)
		assert.Equal(t, "factor@example.com", res2.Users[0].PrimaryEmail)
		assert.NotEmpty(t, res.Users[0].LastSeenAt)
	}

	// Test query by phone
	req3 := &managementapi.ListUsersRequest{
		UserHandle: "+85233333333",
	}

	res3, err := srv.ListUsers(ctx, req3)
	if assert.NoError(t, err) {
		assert.Equal(t, "", res.NextPageToken)
		assert.Equal(t, "", res.PreviousPageToken)
		assert.Len(t, res3.Users, 1)
		assert.Equal(t, "factor", res3.Users[0].Username)
		assert.Equal(t, "+85233333333", res3.Users[0].PrimaryPhone)
		assert.Equal(t, "factor@example.com", res3.Users[0].PrimaryEmail)
		assert.NotEmpty(t, res.Users[0].LastSeenAt)
	}
}

func TestListSingleNonExistUser(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	req := &managementapi.ListUsersRequest{
		UserHandle: "non-exist-user",
	}
	res, err := srv.ListUsers(ctx, req)
	if assert.NoError(t, err) {
		assert.Equal(t, "", res.NextPageToken)
		assert.Equal(t, "", res.PreviousPageToken)
		assert.Len(t, res.Users, 0)
	}
}

func TestListUsers(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	req := &managementapi.ListUsersRequest{
		PageSize: 2,
	}
	res, err := srv.ListUsers(ctx, req)
	users := res.Users
	nextPageToken := res.NextPageToken
	prevPageToken := res.PreviousPageToken
	// totalSize := res.TotalSize

	if assert.NoError(t, err) {
		assert.Equal(t, "eyJkIjowLCJ2IjpbIjIwMTgtMTEtMTJUMDg6Mjc6NThaIiw2XX0", nextPageToken)
		assert.Equal(t, "", prevPageToken)
		// assert.Equal(t, int32(6), totalSize)
		assert.Equal(t, 2, len(users))
		assert.Equal(t, "bob", users[0].Username)
		assert.Equal(t, "bob@example.com", users[0].PrimaryEmail)
		assert.NotEmpty(t, users[0].LastSeenAt)
		assert.Equal(t, "benny", users[1].Username)
		assert.Equal(t, "benny@example.com", users[1].PrimaryEmail)
		assert.NotEmpty(t, users[1].LastSeenAt)
	}

	reqWithPageTokenOnly := &managementapi.ListUsersRequest{
		PageSize:  2,
		PageToken: nextPageToken,
	}
	res, err = srv.ListUsers(ctx, reqWithPageTokenOnly)
	users = res.Users
	nextPageToken = res.NextPageToken
	prevPageToken = res.PreviousPageToken

	if assert.NoError(t, err) {
		assert.Equal(t, "eyJkIjowLCJ2IjpbIjE5NzAtMDEtMDFUMDA6MDA6MDFaIiw5OF19", nextPageToken)
		assert.Equal(t, "eyJkIjoxLCJ2IjpbIjE5NzAtMDEtMDFUMDA6MDA6MDFaIiw5OV19", prevPageToken)
		assert.Equal(t, 2, len(users))
		assert.Equal(t, "last", users[0].Username)
		assert.Equal(t, "lastplusone", users[1].Username)
	}

	reqWithAscendingOnly := &managementapi.ListUsersRequest{
		Ascending: true,
	}
	res, err = srv.ListUsers(ctx, reqWithAscendingOnly)
	users = res.Users
	nextPageToken = res.NextPageToken
	prevPageToken = res.PreviousPageToken

	if assert.NoError(t, err) {
		assert.Equal(t, "eyJkIjowLCJ2IjpbIjE5NzAtMDEtMDFUMDA6MDA6MDFaIiw5OF19", nextPageToken)
		assert.Equal(t, "", prevPageToken)
		assert.Equal(t, 10, len(users))
		assert.Equal(t, "carol", users[0].Username)
		assert.Equal(t, "factor", users[1].Username)
		assert.Equal(t, "twofactor", users[2].Username)
		assert.Equal(t, "smith", users[3].Username)
	}

	reqWithParams := &managementapi.ListUsersRequest{
		PageSize:  2,
		PageToken: nextPageToken,
		Ascending: true,
	}

	res, err = srv.ListUsers(ctx, reqWithParams)
	users = res.Users
	nextPageToken = res.NextPageToken
	prevPageToken = res.PreviousPageToken

	if assert.NoError(t, err) {
		assert.Equal(t, "eyJkIjowLCJ2IjpbIjIwMTgtMTEtMTJUMDg6Mjc6NThaIiw2XX0", nextPageToken)
		assert.Equal(t, "eyJkIjoxLCJ2IjpbIjE5NzAtMDEtMDFUMDA6MDA6MDFaIiw5OV19", prevPageToken)
		assert.Equal(t, 2, len(users))
		assert.Equal(t, "last", users[0].Username)
		assert.Equal(t, "last@example.com", users[0].PrimaryEmail)
		assert.NotEmpty(t, users[0].LastSeenAt)
		assert.Equal(t, "benny", users[1].Username)
		assert.Equal(t, "benny@example.com", users[1].PrimaryEmail)
		assert.NotEmpty(t, users[1].LastSeenAt)
	}
}

func TestListUsersWithValidSortKeysSetting(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	req := &managementapi.ListUsersRequest{
		PageSize: 2,
		SortKey:  "",
	}
	_, err = srv.ListUsers(ctx, req)
	assert.NoError(t, err)

	req = &managementapi.ListUsersRequest{
		PageSize: 2,
		SortKey:  "invalid",
	}
	_, err = srv.ListUsers(ctx, req)
	assert.Error(t, err)

	req = &managementapi.ListUsersRequest{
		PageSize: 2,
		SortKey:  "is_locked",
	}
	_, err = srv.ListUsers(ctx, req)
	assert.NoError(t, err)

	req = &managementapi.ListUsersRequest{
		PageSize: 2,
		SortKey:  "created_at",
	}
	_, err = srv.ListUsers(ctx, req)
	assert.NoError(t, err)

	req = &managementapi.ListUsersRequest{
		PageSize: 2,
		SortKey:  "last_seen_at",
	}
	_, err = srv.ListUsers(ctx, req)
	assert.NoError(t, err)

	req = &managementapi.ListUsersRequest{
		PageSize: 2,
		SortKey:  "email",
	}
	_, err = srv.ListUsers(ctx, req)
	assert.NoError(t, err)
}

func TestListUsersQuery(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// var filterQuery = make(map[string]string)

	req := &managementapi.ListUsersRequest{
		PageSize: 2,
		SortKey:  "",
		QueryKey: "",
	}
	res, err := srv.ListUsers(ctx, req)
	users := res.Users

	if assert.NoError(t, err) {
		assert.Equal(t, "bob", users[0].Username)
	}

	req1 := &managementapi.ListUsersRequest{
		PageSize:   2,
		SortKey:    "",
		QueryKey:   "username",
		QueryValue: "carol",
	}
	res1, err := srv.ListUsers(ctx, req1)
	users = res1.Users

	if assert.NoError(t, err) {
		assert.Equal(t, "carol", users[0].Username)
	}

	req2 := &managementapi.ListUsersRequest{
		PageSize: 2,
		SortKey:  "",
		QueryKey: "invalid",
	}
	_, err = srv.ListUsers(ctx, req2)
	assert.Error(t, err)
}

func TestListUsersWhenUnauthenticated(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	// 1. List users while the account is not authenticated
	req := &managementapi.ListUsersRequest{
		PageSize: 2,
	}
	_, err := srv.ListUsers(context.Background(), req)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.Unauthenticated, status.Code())
		}
	}
}

func TestListUsersWhenUnauthorized(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 3)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. List users while the account is not authorized
	req := &managementapi.ListUsersRequest{
		PageSize: 2,
	}
	_, err = srv.ListUsers(ctx, req)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.PermissionDenied, status.Code())
		}
	}
}

func TestCreateUserAPI(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. Create user
	req := &managementapi.CreateUserRequest{
		Username:    "alice",
		Email:       "alice@example.com",
		Phone:       "+85212345678",
		DisplayName: "Alice",
	}
	res, err := srv.CreateUser(ctx, req)

	if assert.Nil(t, err) {
		assert.NotNil(t, res)
		u := res.User
		assert.NotEqual(t, "", u.Id)
		assert.Equal(t, "alice", u.Username)
		assert.Equal(t, "alice@example.com", u.PrimaryEmail)
		assert.Equal(t, "+85212345678", u.PrimaryPhone)
		assert.NotZero(t, u.CreatedAt)
		assert.NotZero(t, u.UpdatedAt)
	}
}

func TestCreateUserWithOAuth(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. Create user
	req := &managementapi.CreateUserRequest{
		Username:    "alice",
		Email:       "alice@example.com",
		Phone:       "+85212345678",
		DisplayName: "Alice",
		OauthFactors: []*managementapi.OAuthFactor{
			&managementapi.OAuthFactor{
				Service:     managementapi.OAuthFactor_GOOGLE,
				OauthUserId: "900913",
			},
		},
	}
	res, err := srv.CreateUser(ctx, req)
	if assert.Nil(t, err) {
		return
	}

	u, err = srv.UserStore.UserByPublicID(context.Background(), res.User.Id)
	if !assert.NoError(t, err) {
		return
	}
	oauthFactor, err := srv.UserStore.FindOAuthFactorByOAuthIdentity(context.Background(), user.OAuthGoogle, "900913")
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, u.ID, oauthFactor.UserID)
}

func TestCreateUserWhenUnauthenticated(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	// 1. Create user while the account is not authenticated
	req := &managementapi.CreateUserRequest{
		Username:    "alice",
		Email:       "alice@example.com",
		Phone:       "+85212345678",
		DisplayName: "Alice",
	}
	_, err := srv.CreateUser(context.Background(), req)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.Unauthenticated, status.Code())
		}
	}
}

func TestCreateUserWhenUnauthorized(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 3)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. Create user while the account is not authorized
	req := &managementapi.CreateUserRequest{
		Username:    "alice",
		Email:       "alice@example.com",
		Phone:       "+85212345678",
		DisplayName: "Alice",
	}
	_, err = srv.CreateUser(ctx, req)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.PermissionDenied, status.Code())
		}
	}
}

func TestGetUser(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	currentUser, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), currentUser)

	req := &managementapi.GetUserRequest{
		UserId: "1",
	}
	user, err := srv.GetUser(ctx, req)

	if assert.NoError(t, err) {
		assert.Equal(t, "bob", user.Username)
	}

	failReq := &managementapi.GetUserRequest{
		UserId: "0",
	}

	_, err = srv.GetUser(ctx, failReq)
	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.NotFound, status.Code())
		}
	}
}

func TestGetUserWhenUnauthenticated(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	// 1. Get user while the account is not authenticated
	req := &managementapi.GetUserRequest{
		UserId: "1",
	}
	_, err := srv.GetUser(context.Background(), req)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.Unauthenticated, status.Code())
		}
	}
}

func TestGetUserWhenUnauthorized(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 3)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. Get user while the account is not authorized
	req := &managementapi.GetUserRequest{
		UserId: "1",
	}
	_, err = srv.GetUser(ctx, req)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.PermissionDenied, status.Code())
		}
	}
}

func TestUpdateUserAPI(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	currentUser, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), currentUser)

	req := &managementapi.GetUserRequest{
		UserId: "1",
	}
	user, err := srv.GetUser(ctx, req)

	assert.NoError(t, err)

	user.DisplayName = "bob_updated"
	user.Language = "zh-HK"

	updateUserReq := &managementapi.UpdateUserRequest{
		UserId: user.Id,
		User:   user,
	}
	res, err := srv.UpdateUser(ctx, updateUserReq)

	if assert.NoError(t, err) {
		assert.Equal(t, "1", res.Id)
		assert.Equal(t, "bob_updated", res.DisplayName)
		assert.Equal(t, "zh-HK", res.Language)
	}

	// User cannot be nil
	nilUserReq := &managementapi.UpdateUserRequest{
		UserId: user.Id,
		User:   nil,
	}

	_, err = srv.UpdateUser(ctx, nilUserReq)

	assert.Error(t, err)
}

func TestUpdateUserWhenUnauthenticated(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	// 1. Update user while the account is not authenticated
	req := &managementapi.UpdateUserRequest{
		UserId: "1",
		User: &authapi.User{
			DisplayName: "Bob",
		},
	}
	_, err := srv.UpdateUser(context.Background(), req)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.Unauthenticated, status.Code())
		}
	}
}

func TestUpdateUserWhenUnauthorized(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 3)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. Update user while the account is not authorized
	req := &managementapi.UpdateUserRequest{
		UserId: "1",
		User: &authapi.User{
			DisplayName: "Bob",
		},
	}
	_, err = srv.UpdateUser(ctx, req)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.PermissionDenied, status.Code())
		}
	}
}

func TestLockUser(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. Lock user
	oneDay, err := time.ParseDuration("24h")
	assert.Nil(t, err)

	lockExpiredAt := time.Now().Add(oneDay)
	req := &managementapi.UpdateUserRequest{
		UserId: "6",
		Type:   managementapi.UpdateUserRequest_LOCK,
		User: &authapi.User{
			Locked: true,
			LockExpiredAt: &timestamp.Timestamp{
				Seconds: lockExpiredAt.Unix(),
				Nanos:   int32(lockExpiredAt.Nanosecond()),
			},
			LockDescription: "Test locking behaviour",
		},
	}
	_, err = srv.UpdateUser(ctx, req)

	assert.Nil(t, err)

	// 2. Unlock user
	req2 := &managementapi.UpdateUserRequest{
		UserId: "6",
		Type:   managementapi.UpdateUserRequest_LOCK,
		User: &authapi.User{
			Locked:          false,
			LockDescription: "Test unlocking behaviour",
		},
	}
	_, err = srv.UpdateUser(ctx, req2)

	assert.Nil(t, err)
}

func TestLockUserWhenUnauthenticated(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	// 1. Lock user while the account is not authenticated
	oneDay, err := time.ParseDuration("24h")
	assert.Nil(t, err)

	lockExpiredAt := time.Now().Add(oneDay)
	req := &managementapi.UpdateUserRequest{
		UserId: "6",
		Type:   managementapi.UpdateUserRequest_LOCK,
		User: &authapi.User{
			Locked: true,
			LockExpiredAt: &timestamp.Timestamp{
				Seconds: lockExpiredAt.Unix(),
				Nanos:   int32(lockExpiredAt.Nanosecond()),
			},
			LockDescription: "Test locking behaviour",
		},
	}
	_, err = srv.UpdateUser(context.Background(), req)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.Unauthenticated, status.Code())
		}
	}
}

func TestLockUserWhenUnauthorized(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 3)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. Lock user while the account is not authorized
	oneDay, err := time.ParseDuration("24h")
	assert.Nil(t, err)

	lockExpiredAt := time.Now().Add(oneDay)
	req := &managementapi.UpdateUserRequest{
		UserId: "6",
		Type:   managementapi.UpdateUserRequest_LOCK,
		User: &authapi.User{
			Locked: true,
			LockExpiredAt: &timestamp.Timestamp{
				Seconds: lockExpiredAt.Unix(),
				Nanos:   int32(lockExpiredAt.Nanosecond()),
			},
			LockDescription: "Test locking behaviour",
		},
	}
	_, err = srv.UpdateUser(ctx, req)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.PermissionDenied, status.Code())
		}
	}
}

func TestCreateFirstAdminUser(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	user := &user.User{
		Username:       nulls.NewString("alice"),
		Email:          nulls.NewString("alice@example.com"),
		Phone:          nulls.NewString("+85212345678"),
		DisplayNameOld: "+alice",
		Language:       nulls.NewString("en"),
	}
	_, err := srv.CreateFirstAdminUser(context.Background(), user, "password")
	assert.Error(t, err) // Should fail because the db is not empty

	// Remove all users
	_, err = srv.DB.ExecContext(context.Background(), "DELETE FROM audit_logs")
	assert.NoError(t, err)
	_, err = srv.DB.ExecContext(context.Background(), "DELETE FROM roles_users")
	assert.NoError(t, err)
	_, err = srv.DB.ExecContext(context.Background(), "DELETE FROM sessions")
	assert.NoError(t, err)
	_, err = srv.DB.ExecContext(context.Background(), "DELETE FROM contacts")
	assert.NoError(t, err)
	_, err = srv.DB.ExecContext(context.Background(), "DELETE FROM secrets")
	assert.NoError(t, err)
	_, err = srv.DB.ExecContext(context.Background(), "DELETE FROM second_factors")
	assert.NoError(t, err)
	_, err = srv.DB.ExecContext(context.Background(), "DELETE FROM oauth_factors")
	assert.NoError(t, err)
	_, err = srv.DB.ExecContext(context.Background(), "DELETE FROM users")
	assert.NoError(t, err)

	u, err := srv.CreateFirstAdminUser(context.Background(), user, "password")
	assert.NoError(t, err)
	assert.NotEqual(t, int64(0), u.ID)
	assert.Equal(t, "alice", u.Username.String)
	assert.Equal(t, "alice@example.com", u.Email.String)
	assert.Equal(t, "+85212345678", u.Phone.String)
	assert.NotZero(t, u.CreatedAt)
	assert.NotZero(t, u.UpdatedAt)

	// Verify password
	u, err = srv.UserStore.UserByID(context.Background(), u.ID)
	if !assert.NoError(t, err) {
		return
	}

	salt := u.PasswordSalt()
	clientIdentity, serverIdentity := authentication.GetIdentity()
	spake2, err := authentication.NewSPAKE2Plus()
	assert.Nil(t, err)

	state, message, err := spake2.StartClient(clientIdentity, serverIdentity, []byte("password"), salt, nil)
	assert.Nil(t, err)

	challenge, err := srv.AuthenticationService.NewPasswordChallengeWithUser(context.Background(), u, message)
	if !assert.NoError(t, err) {
		return
	}
	secret, err := state.Finish(challenge.Message)
	assert.Nil(t, err)

	err = srv.AuthenticationService.VerifyPasswordResponseWithUser(context.Background(), u, challenge.Token, secret.GetConfirmation())
	assert.NoError(t, err)
}
