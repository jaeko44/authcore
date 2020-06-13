package oauth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/pkg/secret"

	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

// MattersOAuthFactor is a struct of Matters OAuth factor.
type MattersOAuthFactor struct{}

func (MattersOAuthFactor) getConfig() (interface{}, error) {
	appID := viper.GetString("matters_app_id")
	clientSecret := viper.Get("matters_app_secret").(secret.String).SecretString()
	oauthRedirectURL := viper.GetString("matters_oauth_redirect_url")
	mattersURL := viper.GetString("matters_url")
	authURL := fmt.Sprintf("%s/oauth/authorize", mattersURL)
	tokenURL := fmt.Sprintf("%s/oauth/access_token", mattersURL)
	return &oauth2.Config{
		ClientID:     appID,
		ClientSecret: clientSecret,
		RedirectURL:  oauthRedirectURL,
		Scopes:       []string{"query:viewer:info:email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:   authURL,
			TokenURL:  tokenURL,
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}, nil
}

// GetUser returns a OAuth user by access token and id token (or access secret).
func (factor MattersOAuthFactor) GetUser(accessToken, idToken string) (*User, error) {
	token, err := jwt.Parse(idToken, func(token *jwt.Token) (interface{}, error) {
		return getCertificateFromMatters()
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
	name, ok := mapClaims["name"].(string)
	if ok {
		user.Metadata["name"] = name
	}
	return &user, nil
}

func getCertificateFromMatters() (*rsa.PublicKey, error) {
	pubPEM := viper.GetString("matters_id_token_certificate")
	pub, _ := pem.Decode([]byte(pubPEM))
	key, err := x509.ParsePKIXPublicKey(pub.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New(errors.ErrorUnknown, "public key for matters id token is not a rsa key")
	}
	return rsaKey, nil
}
