package idp

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"time"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/pkg/secret"

	"github.com/dgrijalva/jwt-go"
	"github.com/lestrrat-go/jwx/jwk"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

const (
	// Apple represents a Sign-in with Apple provider.
	Apple = "apple"
)

// NewAppleIDP returns a new IDP to authenticate using Sign-in with Apple.
func NewAppleIDP() IDP {
	clientID := viper.GetString("apple_app_id")
	privateKeyPEM := viper.Get("apple_app_private_key").(secret.String).SecretString()
	privateKeyBlob, _ := pem.Decode([]byte(privateKeyPEM))
	if privateKeyBlob == nil {
		log.Fatal("cannot decode apple_app_private_key")
	}
	privateKey, err := x509.ParsePKCS8PrivateKey(privateKeyBlob.Bytes)
	if err != nil {
		log.Fatalf("cannot decode apple_app_private_key: %v", err)
	}

	return &OAuth2Provider{
		IDString: Apple,
		Config: &oauth2.Config{
			ClientID:    clientID,
			RedirectURL: OauthRedirectURL(Apple),
			Scopes:      []string{"email"},
			Endpoint: oauth2.Endpoint{
				AuthURL:   "https://appleid.apple.com/auth/authorize",
				TokenURL:  "https://appleid.apple.com/auth/token",
				AuthStyle: oauth2.AuthStyleInParams,
			},
		},
		AuthCodeURLOptions: []oauth2.AuthCodeOption{
			oauth2.SetAuthURLParam("response_mode", "form_post"),
		},
		UseIDToken: true,
		JWTKeyFunc: appleJWTKeyFunc,
		ClientSecretFunc: func() (string, error) {
			return appleSignClientJWT(privateKey.(*ecdsa.PrivateKey))
		},
	}
}

func appleSignClientJWT(privateKey *ecdsa.PrivateKey) (string, error) {
	clientID := viper.GetString("apple_app_id")
	appKeyID := viper.GetString("apple_app_key_id")
	appKeyIssuer := viper.GetString("apple_app_key_issuer")
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iss": appKeyIssuer,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(10 * time.Minute).Unix(),
		"aud": "https://appleid.apple.com",
		"sub": clientID,
	})
	token.Header["kid"] = appKeyID
	clientJWT, err := token.SignedString(privateKey)
	if err != nil {
		return "", errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return clientJWT, nil
}

func appleJWTKeyFunc(token *jwt.Token) (interface{}, error) {
	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}
	return appleGetCertificate(kid)
}

func appleGetCertificate(keyID string) (*rsa.PublicKey, error) {
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
