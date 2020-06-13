package managementapi

import (
	"context"
	"testing"

	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/api/authapi"
	"authcore.io/authcore/pkg/api/managementapi"
	"authcore.io/authcore/pkg/secret"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestListSessions(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	req := &managementapi.ListSessionsRequest{
		UserId:   "1",
		PageSize: 2,
	}
	res, err := srv.ListSessions(ctx, req)

	if assert.NoError(t, err) {
		assert.Equal(t, int32(3), res.TotalSize)
		assert.NotEmpty(t, res.Sessions)
		assert.Equal(t, "eyJkIjowLCJ2IjpbMl19", res.NextPageToken)
		assert.Len(t, res.Sessions, 2)
	}
}

func TestCreateSession(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	// JWT public key
	accessTokenPrivateKey := viper.Get("access_token_private_key").(secret.String).SecretString()
	jwtPrivateKey, err := jwt.ParseECPrivateKeyFromPEM([]byte(accessTokenPrivateKey))
	assert.NoError(t, err)
	jwtPublicKey := &jwtPrivateKey.PublicKey

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	req := &managementapi.CreateSessionRequest{
		UserId:   "1",
		DeviceId: "0",
	}
	res, err := srv.CreateSession(ctx, req)

	if assert.NoError(t, err) {
		assert.NotEmpty(t, res.AccessToken)
		assert.NotEmpty(t, res.RefreshToken)
		assert.Equal(t, authapi.AccessToken_BEARER, res.TokenType)
		assert.Equal(t, int64(28800), res.ExpiresIn)

		token, err := jwt.Parse(res.AccessToken, func(token *jwt.Token) (interface{}, error) {
			if method, ok := token.Method.(*jwt.SigningMethodECDSA); !ok || method.Alg() != "ES256" {
				return nil, errors.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return jwtPublicKey, nil
		})
		if assert.NoError(t, err) {
			claims, ok := token.Claims.(jwt.MapClaims)
			assert.True(t, ok)
			assert.True(t, token.Valid)
			assert.Equal(t, "1", claims["sub"])
			// not check sid as it is not deterministic
			// will affect by other test case (as sql id is affected)
			assert.IsType(t, "0", claims["sid"])
			assert.Equal(t, "https://authcore.localhost/", claims["iss"])
			assert.IsType(t, float64(0), claims["iat"])
			assert.IsType(t, float64(0), claims["exp"])
		}
		// Verify ID token
		idToken, err := jwt.Parse(res.IdToken, func(token *jwt.Token) (interface{}, error) {
			if method, ok := token.Method.(*jwt.SigningMethodECDSA); !ok || method.Alg() != "ES256" {
				return nil, errors.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			return jwtPublicKey, nil
		})

		if assert.NoError(t, err) {
			claims, ok := idToken.Claims.(jwt.MapClaims)
			assert.True(t, ok)
			assert.True(t, idToken.Valid)
			assert.Equal(t, "1", claims["sub"])
			assert.IsType(t, "0", claims["sid"])
			assert.Equal(t, "https://authcore.localhost/", claims["iss"])
			assert.IsType(t, float64(0), claims["iat"])
			assert.IsType(t, float64(0), claims["exp"])
			assert.Equal(t, "Bob", claims["name"])
			assert.Equal(t, "bob@example.com", claims["email"])
			assert.Equal(t, true, claims["email_verified"])
			assert.Equal(t, "+85223456789", claims["phone_number"])
			assert.Equal(t, true, claims["phone_number_verified"])
			assert.Equal(t, "bob", claims["preferred_username"])
		}
		userID, sessionID, err := srv.SessionStore.VerifyAccessToken(context.Background(), res.AccessToken)
		if assert.NoError(t, err) {
			assert.Equal(t, "1", userID)
			assert.IsType(t, "0", sessionID)
		}
	}
}

func TestDeleteSession(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	req := &managementapi.ListSessionsRequest{}
	res, err := srv.ListSessions(ctx, req)

	if assert.NoError(t, err) {
		assert.Equal(t, int32(5), res.TotalSize)
	}

	req2 := &managementapi.DeleteSessionRequest{
		SessionId: "1",
	}
	_, err = srv.DeleteSession(ctx, req2)

	assert.NoError(t, err)

	res, err = srv.ListSessions(ctx, req)

	if assert.NoError(t, err) {
		assert.Equal(t, int32(4), res.TotalSize)
	}
}
