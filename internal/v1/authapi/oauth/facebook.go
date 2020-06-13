package oauth

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

// FacebookOAuthFactor is a struct of Facebook OAuth factor.
type FacebookOAuthFactor struct{}

func (FacebookOAuthFactor) getConfig() (interface{}, error) {
	facebookAppID := viper.GetString("facebook_app_id")
	facebookAppSecret := viper.Get("facebook_app_secret").(secret.String).SecretString()
	oauthRedirectURL := viper.GetString("facebook_oauth_redirect_url")
	return &oauth2.Config{
		ClientID:     facebookAppID,
		ClientSecret: facebookAppSecret,
		RedirectURL:  oauthRedirectURL,
		Scopes:       []string{"email"},
		Endpoint:     facebook.Endpoint,
	}, nil
}

// GetUser returns a OAuth user by access token and id token (or access secret).
func (factor FacebookOAuthFactor) GetUser(accessToken, idToken string) (*User, error) {
	resp, err := http.Get("https://graph.facebook.com/me?fields=email,name,picture,short_name&access_token=" + url.QueryEscape(accessToken))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	response, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	var u map[string]interface{}
	err = json.Unmarshal(response, &u)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	var user User
	var ok bool
	user.ID, ok = u["id"].(string)
	if !ok {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "cannot get id or email address from Facebook")
	}
	user.Email, ok = u["email"].(string)
	if !ok {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "cannot get id or email address from Facebook")
	}
	delete(u, "id")
	delete(u, "email")
	user.Metadata = u
	return &user, nil
}
