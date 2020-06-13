package authapi

import (
	"context"
	"testing"

	"authcore.io/authcore/pkg/api/authapi"
	"authcore.io/authcore/pkg/secret"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateAccessToken(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	// JWT public key
	accessTokenPrivateKey := viper.Get("access_token_private_key").(secret.String).SecretString()
	jwtPrivateKey, err := jwt.ParseECPrivateKeyFromPEM([]byte(accessTokenPrivateKey))
	if !assert.NoError(t, err) {
		return
	}
	jwtPublicKey := &jwtPrivateKey.PublicKey

	req := &authapi.CreateAccessTokenRequest{
		GrantType: authapi.CreateAccessTokenRequest_REFRESH_TOKEN,
		Token:     "BOBREFRESHTOKEN1",
	}

	res, err := srv.CreateAccessToken(context.Background(), req)
	if assert.NoError(t, err) {
		assert.NotEmpty(t, res.AccessToken)
		assert.Empty(t, res.RefreshToken) // Refresh token grant should not change the token
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
			assert.Equal(t, "1", claims["sid"])
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
			assert.Equal(t, "1", claims["sid"])
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
			assert.Equal(t, "1", sessionID)
		}
	}
}

func TestCreateAccessTokenInvalidSession(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	req := &authapi.CreateAccessTokenRequest{
		GrantType: authapi.CreateAccessTokenRequest_REFRESH_TOKEN,
		Token:     "INVALID",
	}

	_, err := srv.CreateAccessToken(context.Background(), req)
	assert.Error(t, err)
}

func TestCreateAccessTokenGrantType(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	req := &authapi.CreateAccessTokenRequest{
		GrantType: authapi.CreateAccessTokenRequest_AUTHORIZATION_TOKEN,
		Token:     "BOBREFRESHTOKEN",
	}

	_, err := srv.CreateAccessToken(context.Background(), req)
	assert.Error(t, err)
}

// Users should not be able to initiate authentication if they are currently locked
func TestCreateAccessTokenForBannedUser(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	req := &authapi.CreateAccessTokenRequest{
		GrantType: authapi.CreateAccessTokenRequest_REFRESH_TOKEN,
		Token:     "BENNYREFRESHTOKEN",
	}

	_, err := srv.CreateAccessToken(context.Background(), req)
	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.PermissionDenied, status.Code())
		}
	}
}
