package oauth

import (
	"net/url"
	"strings"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/pkg/api/authapi"

	"github.com/dghubble/oauth1"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// User represents an user for OAuth
type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Metadata map[string]interface{}
}

// TODO: Cache the certificates for Google & Apple OAuth again
// https://gitlab.com/blocksq/authcore/issues/505

// GetFactorDelegate returns a FactorDelegate given authapi.service.
func GetFactorDelegate(service authapi.OAuthFactor_OAuthService) FactorDelegate {
	switch service {
	case authapi.OAuthFactor_FACEBOOK:
		return FacebookOAuthFactor{}
	case authapi.OAuthFactor_GOOGLE:
		return GoogleOAuthFactor{}
	case authapi.OAuthFactor_APPLE:
		return AppleOAuthFactor{}
	case authapi.OAuthFactor_MATTERS:
		return MattersOAuthFactor{}
	case authapi.OAuthFactor_TWITTER:
		return TwitterOAuthFactor{}
	default:
		log.Panic("cannot dispatch oauth service")
		return nil
	}
}

// FactorDelegate is an interface implementing GetEndpointURI.
type FactorDelegate interface {
	getConfig() (interface{}, error)
	GetUser(accessToken string, idToken string) (*User, error)
}

// GetOAuthEndpointURI returns the endpoint URI for Google OAuth.
func GetOAuthEndpointURI(factor FactorDelegate, state string) (string, string, error) {
	config, err := factor.getConfig()
	if err != nil {
		return "", "", err
	}
	switch cfg := config.(type) {
	case *oauth1.Config:
		return getOAuth1EndpointURI(cfg, factor, state)
	case *oauth2.Config:
		return getOAuth2EndpointURI(cfg, factor, state)
	default:
		return "", "", errors.New(errors.ErrorUnknown, "undefined config type")
	}
}

func getOAuth1EndpointURI(config *oauth1.Config, factor FactorDelegate, state string) (string, string, error) {
	requestToken, _, err := config.RequestToken()
	if err != nil {
		return "", "", errors.Wrap(err, errors.ErrorUnknown, "")
	}
	oauthEndpoint, err := config.AuthorizationURL(requestToken)
	if err != nil {
		return "", "", errors.Wrap(err, errors.ErrorUnknown, "")
	}
	oauthEndpointURI := oauthEndpoint.String()
	return oauthEndpointURI, requestToken, nil
}

func getOAuth2EndpointURI(config *oauth2.Config, factor FactorDelegate, state string) (string, string, error) {
	oauthEndpoint, err := url.Parse(config.Endpoint.AuthURL)
	if err != nil {
		return "", "", errors.Wrap(err, errors.ErrorUnknown, "")
	}
	parameters := url.Values{}
	parameters.Add("client_id", config.ClientID)
	parameters.Add("scope", strings.Join(config.Scopes, " "))
	parameters.Add("redirect_uri", config.RedirectURL)
	parameters.Add("response_type", "code")
	parameters.Add("state", state)
	parameters.Add("display", "popup")           // Facebook
	parameters.Add("response_mode", "form_post") // Apple
	switch factor.(type) {
	case GoogleOAuthFactor:
		// For Google OAuth login in, prompt the user to select account
		parameters.Add("prompt", "select_account")
	}
	oauthEndpoint.RawQuery = parameters.Encode()
	oauthEndpointURI := oauthEndpoint.String()
	return oauthEndpointURI, "", nil
}

// GetTokensByAuthorizationCode returns access and id tokens for OAuth by authorization code.
func GetTokensByAuthorizationCode(factor FactorDelegate, code string, requestToken string) (string, string, error) {
	config, err := factor.getConfig()
	if err != nil {
		return "", "", errors.Wrap(err, errors.ErrorUnknown, "")
	}
	switch cfg := config.(type) {
	case *oauth1.Config:
		return getOAuth1Token(cfg, factor, code, requestToken)
	case *oauth2.Config:
		return getOAuth2Token(cfg, factor, code)
	default:
		return "", "", errors.New(errors.ErrorUnknown, "undefined config type")
	}
}

func getOAuth1Token(config *oauth1.Config, factor FactorDelegate, verifier string, requestToken string) (string, string, error) {
	accessToken, accessSecret, err := config.AccessToken(requestToken, "", verifier)
	if err != nil {
		return "", "", err
	}
	return accessToken, accessSecret, nil
}

func getOAuth2Token(config *oauth2.Config, factor FactorDelegate, code string) (string, string, error) {
	token, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		return "", "", errors.Wrap(err, errors.ErrorUnknown, "")
	}
	accessToken := token.AccessToken
	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		idToken = ""
	}
	return accessToken, idToken, nil
}
