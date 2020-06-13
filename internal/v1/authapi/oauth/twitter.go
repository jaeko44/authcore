package oauth

import (
	"encoding/json"
	"io/ioutil"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/pkg/secret"

	"github.com/dghubble/oauth1"
	"github.com/dghubble/oauth1/twitter"
	"github.com/spf13/viper"
)

// TwitterOAuthFactor is a struct of Twitter OAuth factor.
type TwitterOAuthFactor struct{}

// getConfig returns the OAuth configuration for Twitter OAuth factor.
func (TwitterOAuthFactor) getConfig() (interface{}, error) {
	consumerKey := viper.GetString("twitter_consumer_key")
	consumerSecret := viper.Get("twitter_consumer_secret").(secret.String).SecretString()
	oauthRedirectURL := viper.GetString("twitter_oauth_redirect_url")
	return &oauth1.Config{
		ConsumerKey:    consumerKey,
		ConsumerSecret: consumerSecret,
		CallbackURL:    oauthRedirectURL,
		Endpoint:       twitter.AuthorizeEndpoint,
	}, nil
}

// GetUser returns a OAuth user by access token and id token (or access secret).
func (factor TwitterOAuthFactor) GetUser(accessToken, accessSecret string) (*User, error) {
	config, err := factor.getConfig()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	cfg, ok := config.(*oauth1.Config)
	if !ok {
		return nil, errors.New(errors.ErrorUnknown, "cannot get config for twitter oauth")
	}
	token := oauth1.NewToken(accessToken, accessSecret)

	httpClient := cfg.Client(oauth1.NoContext, token)

	resp, err := httpClient.Get("https://api.twitter.com/1.1/account/verify_credentials.json?include_email=true")
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
	user.ID, ok = u["id_str"].(string)
	if !ok {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "cannot get id or email address from twitter")
	}
	user.Email, ok = u["email"].(string)
	if !ok {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "cannot get id or email address from twitter")
	}
	delete(u, "id_str")
	delete(u, "email")
	user.Metadata = u
	return &user, nil
}
