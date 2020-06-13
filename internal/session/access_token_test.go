package session

import (
	"testing"
	"time"

	"authcore.io/authcore/internal/errors"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
)

const accessTokenPrivateKeyForTest string = `
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIP5b5IqmY/DZ1+74meb/U0IfD4471VFNoLVzmk93chUVoAoGCCqGSM49
AwEHoUQDQgAEHjQuqA41Mj/8B2PPb75XTeLKiacI0LQohjjQHORfvx3FsOWvABVP
8uEZGxUWflhasFeTa/wSSp264otaxOYwFQ==
-----END EC PRIVATE KEY-----
`

const serviceAccountPrivateKeyForTest string = `
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIP5b5IqmY/DZ1+74meb/U0IfD4471VFNoLVzmk93chUVoAoGCCqGSM49
AwEHoUQDQgAEHjQuqA41Mj/8B2PPb75XTeLKiacI0LQohjjQHORfvx3FsOWvABVP
8uEZGxUWflhasFeTa/wSSp264otaxOYwFQ==
-----END EC PRIVATE KEY-----
`
const serviceAccountPublicKeyForTest string = `
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEHjQuqA41Mj/8B2PPb75XTeLKiacI
0LQohjjQHORfvx3FsOWvABVP8uEZGxUWflhasFeTa/wSSp264otaxOYwFQ==
-----END PUBLIC KEY-----
`

func TestVerifyAccessTokenExpired(t *testing.T) {
	accessTokenPrivateKey, _ := jwt.ParseECPrivateKeyFromPEM([]byte(accessTokenPrivateKeyForTest))
	accessTokenPublicKey := &accessTokenPrivateKey.PublicKey
	_, _, err := verifyAccessToken(accessTokenPublicKey, nil, "eyJhbGciOiJFUzI1NiIsImtpZCI6IjhpR3NvN0NHeVZubW9jYlVFeUJVcndtUkxvbl9vUWNjOVVyQl9odzl6Z1EiLCJ0eXAiOiJKV1QifQ.eyJleHAiOjE1NDM1NTExMDcsImlhdCI6MTU0MzU1MTEwNywiaXNzIjoiYXBpLmF1dGhjb3JlLmlvIiwic2lkIjoiMSIsInN1YiI6IjEifQ.59Z_HOxb9TsPqLFD40PhJyYjQtUo9RJgxnizsCG53N-__L9n7QK-WOlbxRlutTy7CME1A-tEvlW1nNKHlrGBkg")
	assert.Error(t, err)
	assert.True(t, errors.IsKind(err, errors.ErrorUnauthenticated))
	assert.Contains(t, err.Error(), "Token is expired")
}

func TestVerifyServiceAccountAccessToken(t *testing.T) {
	_, teardown := storeForTest()
	defer teardown()

	serviceAccountPrivateKey, _ := jwt.ParseECPrivateKeyFromPEM([]byte(serviceAccountPrivateKeyForTest))
	serviceAccountPublicKey, _ := jwt.ParseECPublicKeyFromPEM([]byte(serviceAccountPublicKeyForTest))
	accessTokenPrivateKey, _ := jwt.ParseECPrivateKeyFromPEM([]byte(accessTokenPrivateKeyForTest))
	accessTokenPublicKey := &accessTokenPrivateKey.PublicKey
	serviceAccountsMap := map[string]ServiceAccount{
		"123456": ServiceAccount{
			ID:           "123456",
			PublicKeyPEM: serviceAccountPublicKeyForTest,
		},
	}

	issuer := "serviceaccount:123456"
	issuedAt := time.Now()
	expiresIn, _ := time.ParseDuration("1h")
	expireAt := issuedAt.Add(expiresIn)

	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iat": issuedAt.Unix(),
		"exp": expireAt.Unix(),
		"iss": issuer,
		"sub": issuer,
	})
	kid, err := kidFromECPublicKey(accessTokenPublicKey)
	assert.NoError(t, err)
	token.Header["kid"] = kid
	tokenString, _ := token.SignedString(serviceAccountPrivateKey)
	userID, sessionID, err := verifyAccessToken(accessTokenPublicKey, serviceAccountsMap, tokenString)
	assert.NoError(t, err)
	assert.Equal(t, "serviceaccount:123456", userID)
	assert.Equal(t, "", sessionID)

	issuer = "serviceaccount:123"
	token = jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iat": issuedAt.Unix(),
		"exp": expireAt.Unix(),
		"iss": issuer,
		"sub": issuer,
	})
	kid, err = kidFromECPublicKey(serviceAccountPublicKey)
	assert.NoError(t, err)
	token.Header["kid"] = kid
	tokenString, _ = token.SignedString(serviceAccountPrivateKey)
	userID, sessionID, err = verifyAccessToken(accessTokenPublicKey, serviceAccountsMap, tokenString)
	assert.Error(t, err)
	assert.Equal(t, "", userID)
	assert.Equal(t, "", sessionID)
}
