package idp

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/pkg/secret"

	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
)

const (
	// Facebook represents a Facebook OAuth provider.
	Facebook = "facebook"
)

// NewFacebookIDP returns a new IDP to authenticate using Facebook accounts.
func NewFacebookIDP() IDP {
	clientID := viper.GetString("facebook_app_id")
	appSecret := viper.Get("facebook_app_secret").(secret.String).SecretString()

	return &OAuth2Provider{
		IDString: Facebook,
		Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: appSecret,
			RedirectURL:  OauthRedirectURL(Facebook),
			Scopes:       []string{"email"},
			Endpoint:     facebook.Endpoint,
		},
		AuthCodeURLOptions: []oauth2.AuthCodeOption{
			oauth2.SetAuthURLParam("display", "popup"),
		},
		IdentityFunc: facebookFetchIdentity,
	}
}

func facebookFetchIdentity(accessToken string) (*Identity, error) {
	resp, err := http.Get("https://graph.facebook.com/me?fields=email,name,short_name&access_token=" + url.QueryEscape(accessToken))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	response, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	var data map[string]interface{}
	err = json.Unmarshal(response, &data)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	var ident = new(Identity)
	var ok bool
	ident.ID, ok = data["id"].(string)
	if !ok {
		return nil, errors.New(errors.ErrorUnknown, "invalid response when fetching identity from Facebook")
	}
	ident.Name, _ = data["name"].(string)
	ident.PreferredUsername, _ = data["short_name"].(string)
	ident.Email, _ = data["email"].(string)
	ident.EmailVerified = true
	return ident, nil
}
