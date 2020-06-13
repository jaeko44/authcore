package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"authcore.io/authcore/internal/session"
	"authcore.io/authcore/internal/testutil"
	"authcore.io/authcore/pkg/api/authapi"
	"authcore.io/authcore/pkg/api/managementapi"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ServerForTest *Server

func Get(url string) (res *http.Response, err error) {
	return http.Get("http://127.0.0.1:15001" + url)
}

func Post(url string) (res *http.Response, err error) {
	return http.Post("http://127.0.0.1:15001"+url, "", nil)
}

func Put(url string) (res *http.Response, err error) {
	req, err := http.NewRequest(http.MethodPut, "http://127.0.0.1:15001"+url, nil)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

func TestMain(m *testing.M) {
	viper.Set("base_path", "../..")
	viper.Set("access_token_private_key", `
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEICLBfuNNrqDL6LDLeaQFaytAGDP7hk65Q4J2c8iBumlqoAoGCCqGSM49
AwEHoUQDQgAEKY6MShC7UrSkekyczKKvZQXuxFKDRd0DEgV6r9XeDAZoYPPTvgx3
oNBTatFJjSOJ/qRrBbqvbZDiPOLpJ7vlaQ==
-----END EC PRIVATE KEY-----
`)
	viper.Set("apiv1_enabled", true)
	viper.Set("grpc_listen", "127.0.0.1:15000")
	viper.Set("http_listen", "127.0.0.1:15001")
	viper.Set("secret_key_base", "855edf399835e9c9deb61877c1a76bf14eed7c35a167e10ff1b7d43db4363268")
	testutil.MigrationsDir = "../../db/migrations"
	testutil.FixturesDir = "../../db/fixtures"
	testutil.DBSetUp()
	defer testutil.DBTearDown()

	ServerForTest = NewServer()
	go ServerForTest.Start()

	// Wait for server to start
	time.Sleep(2 * time.Second)

	os.Exit(m.Run())
}

func authServiceClient() authapi.AuthServiceClient {
	grpcServer := viper.GetString("grpc_listen")
	dialOpts := []grpc.DialOption{grpc.WithInsecure()}

	conn, err := grpc.Dial(grpcServer, dialOpts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}

	return authapi.NewAuthServiceClient(conn)
}

func managementServiceClient() managementapi.ManagementServiceClient {
	grpcServer := viper.GetString("grpc_listen")
	dialOpts := []grpc.DialOption{grpc.WithInsecure()}

	conn, err := grpc.Dial(grpcServer, dialOpts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}

	return managementapi.NewManagementServiceClient(conn)
}

func TestAPI(t *testing.T) {
	restRes, _ := Post("/api/auth/device")
	assert.NotEqual(t, 404, restRes.StatusCode)

	restRes, _ = Post("/api/auth/users")
	assert.NotEqual(t, 404, restRes.StatusCode)

	restRes, _ = Post("/api/auth/tokens")
	assert.NotEqual(t, 404, restRes.StatusCode)

	restRes, _ = Get("/api/auth/users/current")
	assert.NotEqual(t, 404, restRes.StatusCode)

	restRes, _ = Post("/api/auth/authn/password/start")
	assert.NotEqual(t, 404, restRes.StatusCode)

	restRes, _ = Post("/api/auth/authn/password/key_exchange")
	assert.NotEqual(t, 404, restRes.StatusCode)

	restRes, _ = Post("/api/auth/authn/password/finish")
	assert.NotEqual(t, 404, restRes.StatusCode)

	restRes, _ = Post("/api/auth/auth/second")
	assert.NotEqual(t, 404, restRes.StatusCode)

	restRes, _ = Get("/api/auth/users/current/password/start")
	assert.NotEqual(t, 404, restRes.StatusCode)

	restRes, _ = Post("/api/auth/users/current/password/key_exchange")
	assert.NotEqual(t, 404, restRes.StatusCode)

	restRes, _ = Post("/api/auth/users/current/password/finish")
	assert.NotEqual(t, 404, restRes.StatusCode)

	restRes, _ = Post("/api/auth/challenges/proof_of_work")
	assert.NotEqual(t, 404, restRes.StatusCode)

	restRes, _ = Get("/api/management/users")
	assert.NotEqual(t, 404, restRes.StatusCode)

	restRes, _ = Get("/api/management/users/1")
	assert.NotEqual(t, 404, restRes.StatusCode)

	restRes, _ = Put("/api/management/users/1")
	assert.NotEqual(t, 404, restRes.StatusCode)
}

// Functional test with user authorization
func TestGetCurrentUser(t *testing.T) {
	client := authServiceClient()

	ctx := context.Background()

	sess := &session.Session{ID: 1, UserID: 1}
	token, err := ServerForTest.sessionStore.GenerateAccessToken(ctx, sess, false)
	if !assert.NoError(t, err) {
		return
	}

	tokenAccess := NewTokenAccess(&authapi.AccessToken{
		AccessToken: token.AccessToken,
		TokenType:   authapi.AccessToken_BEARER,
	})
	opt := grpc.PerRPCCredentials(tokenAccess)

	res, err := client.GetCurrentUser(ctx, &empty.Empty{}, opt)

	if assert.NoError(t, err) {
		assert.Equal(t, "1", res.Id)
		assert.Equal(t, "bob", res.Username)
	}
}

func TestGetCurrentUserUnauthenticated(t *testing.T) {
	client := authServiceClient()

	ctx := context.Background()

	_, err := client.GetCurrentUser(ctx, &empty.Empty{})

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.Unauthenticated, status.Code())
		}
	}
}

func TestGetCurrentUserExpiredToken(t *testing.T) {
	client := authServiceClient()

	ctx := context.Background()

	token := authapi.AccessToken{
		AccessToken: "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NDIzNTYzMDUsImlhdCI6IjIwMTgtMTEtMTZUMDg6MTg6MjUuMjU1NTEwN1oiLCJpc3MiOiJhcGkuYXV0aGNvcmUuaW8iLCJzaWQiOiJjYzI1TGxHY0ppTVNqRVJzYlpKZ0Z1YUVfMUpFOWx4cGY1T25lbWFvdm9zIiwic3ViIjoiMSJ9.HB1Vhab7FvPic91qWrREkJ_6cgP8JcH3HieofpTEboBp_FjHfmukB5Bcd1V31xbM1a06IxHfBcuY7dlFfiROEg",
		TokenType:   authapi.AccessToken_BEARER,
	}

	tokenAccess := NewTokenAccess(&token)
	opt := grpc.PerRPCCredentials(tokenAccess)

	_, err := client.GetCurrentUser(ctx, &empty.Empty{}, opt)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.Unauthenticated, status.Code())
		}
	}
}

// Functional test for list users API with user authorization
func TestListUsers(t *testing.T) {
	client := managementServiceClient()

	ctx := context.Background()

	sess := &session.Session{ID: 1, UserID: 1}
	token, err := ServerForTest.sessionStore.GenerateAccessToken(ctx, sess, false)
	if !assert.NoError(t, err) {
		return
	}

	tokenAccess := NewTokenAccess(&authapi.AccessToken{
		AccessToken: token.AccessToken,
		TokenType:   authapi.AccessToken_BEARER,
	})
	opt := grpc.PerRPCCredentials(tokenAccess)

	res, err := client.ListUsers(ctx, &managementapi.ListUsersRequest{
		PageSize: 2,
	}, opt)
	users := res.Users

	if assert.NoError(t, err) {
		assert.Equal(t, "bob", users[0].Username)
		assert.Equal(t, "bob@example.com", users[0].PrimaryEmail)
		assert.Equal(t, "benny", users[1].Username)
		assert.Equal(t, "benny@example.com", users[1].PrimaryEmail)
	}
}

func TestListUsersUnauthorized(t *testing.T) {
	client := managementServiceClient()

	ctx := context.Background()

	_, err := client.ListUsers(ctx, &managementapi.ListUsersRequest{})

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.Unauthenticated, status.Code())
		}
	}
}

func TestListUsersExpiredToken(t *testing.T) {
	client := managementServiceClient()

	ctx := context.Background()

	token := authapi.AccessToken{
		AccessToken: "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NDIzNTYzMDUsImlhdCI6IjIwMTgtMTEtMTZUMDg6MTg6MjUuMjU1NTEwN1oiLCJpc3MiOiJhcGkuYXV0aGNvcmUuaW8iLCJzaWQiOiJjYzI1TGxHY0ppTVNqRVJzYlpKZ0Z1YUVfMUpFOWx4cGY1T25lbWFvdm9zIiwic3ViIjoiMSJ9.HB1Vhab7FvPic91qWrREkJ_6cgP8JcH3HieofpTEboBp_FjHfmukB5Bcd1V31xbM1a06IxHfBcuY7dlFfiROEg",
		TokenType:   authapi.AccessToken_BEARER,
	}

	tokenAccess := NewTokenAccess(&token)
	opt := grpc.PerRPCCredentials(tokenAccess)

	_, err := client.ListUsers(ctx, &managementapi.ListUsersRequest{}, opt)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.Unauthenticated, status.Code())
		}
	}
}
