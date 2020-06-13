package oauth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"time"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/pkg/secret"

	"github.com/dgrijalva/jwt-go"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

// AppleOAuthFactor is a struct of Apple OAuth factor.
type AppleOAuthFactor struct{}

func (AppleOAuthFactor) getConfig() (interface{}, error) {
	appleAppID := viper.GetString("apple_app_id")
	privateKeyPEM := viper.Get("apple_app_private_key").(secret.String).SecretString()
	privateKeyBlob, _ := pem.Decode([]byte(privateKeyPEM))
	if privateKeyBlob == nil {
		return nil, errors.New(errors.ErrorUnknown, "cannot decode private key")
	}
	privateKey, err := x509.ParsePKCS8PrivateKey(privateKeyBlob.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	appleAppKeyID := viper.GetString("apple_app_key_id")
	appleAppKeyIssuer := viper.GetString("apple_app_key_issuer")
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iss": appleAppKeyIssuer,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(10 * time.Minute).Unix(),
		"aud": "https://appleid.apple.com",
		"sub": appleAppID,
	})
	token.Header["kid"] = appleAppKeyID
	clientSecret, err := token.SignedString(privateKey)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	oauthRedirectURL := viper.GetString("apple_oauth_redirect_url")
	return &oauth2.Config{
		ClientID:     appleAppID,
		ClientSecret: clientSecret,
		RedirectURL:  oauthRedirectURL,
		Scopes:       []string{"email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://appleid.apple.com/auth/authorize",
			TokenURL:  "https://appleid.apple.com/auth/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}, nil
}

// GetUser returns a OAuth user by access token and id token (or access secret).
func (factor AppleOAuthFactor) GetUser(accessToken, idToken string) (*User, error) {
	token, err := jwt.Parse(idToken, func(token *jwt.Token) (interface{}, error) {
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New(errors.ErrorInvalidArgument, "")
		}
		return getCertificateFromApple(kid)
	})
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}
	if !token.Valid {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}
	var user User
	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}
	id, ok := mapClaims["sub"].(string)
	if !ok {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}
	user.ID = id
	email, ok := mapClaims["email"].(string)
	if !ok {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}
	user.Email = email
	return &user, nil
}

type appleKeysResponse struct {
	Keys []jwk.RSAPublicKey `json:"keys"`
}
type appleKeyHeader struct {
	Kid string
}

func getCertificateFromApple(keyID string) (*rsa.PublicKey, error) {
	set, err := jwk.Fetch("https://appleid.apple.com/auth/keys")
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	keys := set.LookupKeyID(keyID)
	if len(keys) == 0 {
		return nil, errors.New(errors.ErrorInvalidArgument, "key not found")
	}
	key, err := keys[0].Materialize()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New(errors.ErrorUnknown, "verification key not being rsa key")
	}
	return rsaKey, nil
}
