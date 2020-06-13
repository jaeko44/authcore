package idp

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

const (
	// Google represents a Google OAuth provider.
	Google = "google"
)

// NewGoogleIDP returns a new IDP to authenticate using Google accounts.
func NewGoogleIDP() IDP {
	clientID := viper.GetString("google_app_id")
	appSecret := viper.Get("google_app_secret").(secret.String).SecretString()

	return &OAuth2Provider{
		IDString: Google,
		Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: appSecret,
			RedirectURL:  OauthRedirectURL(Google),
			Scopes:       []string{"email"},
			Endpoint:     google.Endpoint,
		},
		AuthCodeURLOptions: []oauth2.AuthCodeOption{
			oauth2.SetAuthURLParam("prompt", "select_account"),
		},
		UseIDToken: true,
		JWTKeyFunc: googleJWTKeyFunc,
	}
}

func googleJWTKeyFunc(token *jwt.Token) (interface{}, error) {
	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}
	return googleGetCertificate(kid)
}

func googleGetCertificate(keyID string) (*rsa.PublicKey, error) {
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
