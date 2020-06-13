package oauth

import (
	"crypto/rsa"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/pkg/secret"

	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleOAuthFactor is a struct of Google OAuth factor.
type GoogleOAuthFactor struct{}

func (GoogleOAuthFactor) getConfig() (interface{}, error) {
	googleAppID := viper.GetString("google_app_id")
	googleAppSecret := viper.Get("google_app_secret").(secret.String).SecretString()
	oauthRedirectURL := viper.GetString("google_oauth_redirect_url")
	return &oauth2.Config{
		ClientID:     googleAppID,
		ClientSecret: googleAppSecret,
		RedirectURL:  oauthRedirectURL,
		Scopes:       []string{"email"},
		Endpoint:     google.Endpoint,
	}, nil
}

// GetUser returns a OAuth user by access token and id token (or access secret).
func (factor GoogleOAuthFactor) GetUser(accessToken, idToken string) (*User, error) {
	token, err := jwt.Parse(idToken, func(token *jwt.Token) (interface{}, error) {
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New(errors.ErrorInvalidArgument, "")
		}
		return getCertificateFromGoogle(kid)
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

func getCertificateFromGoogle(keyID string) (*rsa.PublicKey, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v1/certs")
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	response, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	var certificates map[string]string
	err = json.Unmarshal(response, &certificates)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	var googleOAuthCertificates map[string]*rsa.PublicKey
	googleOAuthCertificates = make(map[string]*rsa.PublicKey)
	for kid, certificate := range certificates {
		parsedCertificate, err := jwt.ParseRSAPublicKeyFromPEM([]byte(certificate))
		if err != nil {
			panic("cannot parse certificate from google for oauth")
		}
		googleOAuthCertificates[kid] = parsedCertificate
	}
	certificate, ok := googleOAuthCertificates[keyID]
	if !ok {
		return nil, errors.New(errors.ErrorInvalidArgument, "certificate not found")
	}
	return certificate, nil
}
