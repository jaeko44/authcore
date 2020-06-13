package idp

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"authcore.io/authcore/pkg/secret"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

const (
	// Matters represents a Matters OAuth provider.
	Matters = "matters"
)

// NewMattersIDP returns a new IDP to authenticate using Matters accounts.
func NewMattersIDP() IDP {
	clientID := viper.GetString("matters_app_id")
	clientSecret := viper.Get("matters_app_secret").(secret.String).SecretString()
	mattersURL := viper.GetString("matters_url")
	authURL := fmt.Sprintf("%s/oauth/authorize", mattersURL)
	tokenURL := fmt.Sprintf("%s/oauth/access_token", mattersURL)
	pubPEM := viper.GetString("matters_id_token_certificate")
	pub, _ := pem.Decode([]byte(pubPEM))
	key, err := x509.ParsePKIXPublicKey(pub.Bytes)
	if err != nil {
		log.Fatal("unable to decode matters_id_token_certificate")
	}
	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		log.Fatal("matters_id_token_certificate is not a rsa key")
	}

	return &OAuth2Provider{
		IDString: Matters,
		Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  OauthRedirectURL(Matters),
			Scopes:       []string{"query:viewer:info:email"},
			Endpoint: oauth2.Endpoint{
				AuthURL:   authURL,
				TokenURL:  tokenURL,
				AuthStyle: oauth2.AuthStyleInParams,
			},
		},
		AuthCodeURLOptions: []oauth2.AuthCodeOption{
			oauth2.SetAuthURLParam("prompt", "select_account"),
		},
		UseIDToken: true,
		JWTKeyFunc: func(token *jwt.Token) (interface{}, error) {
			return rsaKey, nil
		},
	}
}
