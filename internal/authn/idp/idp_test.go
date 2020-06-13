package idp

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"authcore.io/authcore/pkg/secret"
	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestIdentityFromIDToken(t *testing.T) {
	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	if !assert.NoError(t, err) {
		return
	}
	pk2, err := rsa.GenerateKey(rand.Reader, 2048)
	if !assert.NoError(t, err) {
		return
	}

	// Success
	idToken, err := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat":                   time.Now().Unix(),
		"exp":                   time.Now().Add(time.Minute).Unix(),
		"iss":                   "iss_test",
		"sub":                   "sub_test",
		"name":                  "name_test",
		"email":                 "example@example.com",
		"email_verified":        true,
		"phone_number":          "+85212345678",
		"phone_number_verified": true,
		"preferred_username":    "username_test",
	}).SignedString(pk)
	assert.NoError(t, err)

	ident, err := IdentityFromIDToken(idToken, func(token *jwt.Token) (interface{}, error) {
		return pk.Public(), nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "sub_test", ident.ID)
	assert.Equal(t, "name_test", ident.Name)
	assert.Equal(t, "example@example.com", ident.Email)
	assert.True(t, ident.EmailVerified)
	assert.Equal(t, "+85212345678", ident.PhoneNumber)
	assert.True(t, ident.PhoneNumberVerified)
	assert.Equal(t, "username_test", ident.PreferredUsername)

	// Wrong signature
	idToken2, err := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat":                   time.Now().Unix(),
		"exp":                   time.Now().Add(time.Minute).Unix(),
		"iss":                   "iss_test",
		"sub":                   "sub_test",
		"name":                  "name_test",
		"email":                 "example@example.com",
		"email_verified":        true,
		"phone_number":          "+85212345678",
		"phone_number_verified": true,
		"preferred_username":    "username_test",
	}).SignedString(pk)
	assert.NoError(t, err)

	_, err = IdentityFromIDToken(idToken2, func(token *jwt.Token) (interface{}, error) {
		return pk2.Public(), nil
	})
	assert.Error(t, err)

	// Expired
	idToken3, err := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat":                   time.Now().Unix(),
		"exp":                   time.Now().Add(-time.Hour).Unix(),
		"iss":                   "iss_test",
		"sub":                   "sub_test",
		"name":                  "name_test",
		"email":                 "example@example.com",
		"email_verified":        true,
		"phone_number":          "+85212345678",
		"phone_number_verified": true,
		"preferred_username":    "username_test",
	}).SignedString(pk)
	assert.NoError(t, err)

	_, err = IdentityFromIDToken(idToken3, func(token *jwt.Token) (interface{}, error) {
		return pk.Public(), nil
	})
	assert.Error(t, err)

	// Missing sub
	idToken4, err := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat":                   time.Now().Unix(),
		"exp":                   time.Now().Add(time.Minute).Unix(),
		"iss":                   "iss_test",
		"name":                  "name_test",
		"email":                 "example@example.com",
		"email_verified":        true,
		"phone_number":          "+85212345678",
		"phone_number_verified": true,
		"preferred_username":    "username_test",
	}).SignedString(pk)
	assert.NoError(t, err)

	_, err = IdentityFromIDToken(idToken4, func(token *jwt.Token) (interface{}, error) {
		return pk.Public(), nil
	})
	assert.Error(t, err)
}

func TestFactory(t *testing.T) {
	viper.Set("google_app_id", "testing")
	viper.Set("google_app_secret", secret.NewString("testing"))
	defer viper.Reset()

	f := NewFactory()
	f.Register(NewGoogleIDP())

	idp, err := f.IDP("google")
	assert.NoError(t, err)
	assert.Equal(t, "google", idp.ID())
}
