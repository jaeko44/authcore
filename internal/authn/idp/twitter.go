package idp

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/pkg/secret"

	"github.com/dghubble/oauth1"
	"github.com/dghubble/oauth1/twitter"
	"github.com/spf13/viper"
)

const (
	// Twitter represents a Twitter OAuth provider.
	Twitter = "twitter"
)

// TwitterIDP is an ID Provider that authenticates using Twitter accounts.
type TwitterIDP struct {
	config *oauth1.Config
}

// NewTwitterIDP returns a new TwitterIDP.
func NewTwitterIDP() IDP {
	consumerKey := viper.GetString("twitter_consumer_key")
	consumerSecret := viper.Get("twitter_consumer_secret").(secret.String).SecretString()
	return &TwitterIDP{
		config: &oauth1.Config{
			ConsumerKey:    consumerKey,
			ConsumerSecret: consumerSecret,
			CallbackURL:    OauthRedirectURL(Twitter),
			Endpoint:       twitter.AuthorizeEndpoint,
		},
	}
}

// ID returns the identifier of this provider.
func (p *TwitterIDP) ID() string {
	return Twitter
}

// AuthorizationURL returns a third-party authorization endpoint URI used by the client to obtain
// authorization from the ID provider. This method also returns a state that is used to recover
// the state later.
func (p *TwitterIDP) AuthorizationURL(stateToken string) (string, State, error) {
	requestToken, _, err := p.config.RequestToken()
	if err != nil {
		return "", "", errors.Wrap(err, errors.ErrorUnknown, "")
	}
	oauthEndpoint, err := p.config.AuthorizationURL(requestToken)
	if err != nil {
		return "", "", errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return oauthEndpoint.String(), State(requestToken), nil
}

// Exchange converts an authorization code into tokens and the user's identity. This method
// takes that State created in CreateAuthorizationURI and an authorization code obtained from
// identity provider.
func (p *TwitterIDP) Exchange(ctx context.Context, state State, code string) (grant *AuthorizationGrant, err error) {
	accessToken, accessSecret, err := p.config.AccessToken(string(state), "", code)
	if err != nil {
		return nil, err
	}
	ident, err := p.fetchIdentity(accessToken, accessSecret)
	if err != nil {
		return nil, err
	}
	return &AuthorizationGrant{
		AccessToken:  accessToken,
		AccessSecret: accessSecret,
		TokenType:    "bearer",
		Identity:     ident,
	}, nil
}

func (p *TwitterIDP) fetchIdentity(accessToken, accessSecret string) (*Identity, error) {
	token := oauth1.NewToken(accessToken, accessSecret)

	httpClient := p.config.Client(oauth1.NoContext, token)

	resp, err := httpClient.Get("https://api.twitter.com/1.1/account/verify_credentials.json?include_email=true")
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
	id, ok := data["id_str"].(string)
	if !ok || len(id) == 0 {
		return nil, errors.New(errors.ErrorUnknown, "cannot get id from Twitter user")
	}
	var ident = new(Identity)
	ident.ID = id
	ident.Name, _ = data["name"].(string)
	ident.Email, _ = data["email"].(string)
	if len(ident.Email) != 0 {
		ident.EmailVerified = true
	}
	ident.PreferredUsername, _ = data["screen_name"].(string)

	return ident, nil
}
