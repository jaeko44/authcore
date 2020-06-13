package idp

import (
	"context"
	"net/url"

	"authcore.io/authcore/internal/errors"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

// OAuth2Provider is a generic OAuth 2.0 provider.
type OAuth2Provider struct {
	IDString           string
	Config             *oauth2.Config
	AuthCodeURLOptions []oauth2.AuthCodeOption
	ExchangeOptions    []oauth2.AuthCodeOption
	UseIDToken         bool
	JWTKeyFunc         jwt.Keyfunc
	IdentityFunc       IdentityFunc
	ClientSecretFunc   ClientSecretFunc
}

// ID returns the identifier of this provider.
func (p *OAuth2Provider) ID() string {
	return p.IDString
}

// AuthorizationURL returns a third-party authorization endpoint URI used by the client to obtain
// authorization from the ID provider. This method also returns a state that is used to recover
// the state later.
func (p *OAuth2Provider) AuthorizationURL(stateToken string) (string, State, error) {
	if p.ClientSecretFunc != nil {
		var err error
		p.Config.ClientSecret, err = p.ClientSecretFunc()
		if err != nil {
			return "", "", err
		}
	}
	url := p.Config.AuthCodeURL(stateToken, p.AuthCodeURLOptions...)
	return url, "", nil
}

// Exchange converts an authorization code into tokens and the user's identity. This method
// takes that State created in CreateAuthorizationURI and an authorization code obtained from
// identity provider.
func (p *OAuth2Provider) Exchange(ctx context.Context, state State, code string) (grant *AuthorizationGrant, err error) {
	if p.ClientSecretFunc != nil {
		var err error
		p.Config.ClientSecret, err = p.ClientSecretFunc()
		if err != nil {
			return nil, err
		}
	}
	token, err := p.Config.Exchange(ctx, code, p.ExchangeOptions...)
	if err != nil {
		err = errors.Wrap(err, errors.ErrorPermissionDenied, "")
		return
	}
	grant = &AuthorizationGrant{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}
	var ident *Identity
	if p.UseIDToken {
		idToken, ok := token.Extra("id_token").(string)
		if !ok || len(idToken) == 0 {
			err = errors.New(errors.ErrorPermissionDenied, "id_token required but it is empty")
			return
		}
		ident, err = IdentityFromIDToken(idToken, p.JWTKeyFunc)
		if err != nil {
			return
		}
	} else if p.IdentityFunc != nil {
		ident, err = p.IdentityFunc(token.AccessToken)
		if err != nil {
			return
		}
	}
	grant.Identity = ident
	return
}

// IdentityFunc fetches an Identity using the given access token.
type IdentityFunc func(token string) (*Identity, error)

// ClientSecretFunc returns a new client secret. It is used when the client is not a static token
// (e.g. a JWT token with expiry).
type ClientSecretFunc func() (string, error)

// OauthRedirectURL returns an Authcore endpoint that handle a successful OAuth redirection.
func OauthRedirectURL(idp string) string {
	baseURL, err := url.Parse(viper.GetString("base_url"))
	if err != nil {
		log.Fatalf("invalid base_url: %v", err)
	}
	url, err := baseURL.Parse("/oauth/redirect")
	if err != nil {
		log.Fatalf("error building OAuth callback URL: %v", err)
	}
	return url.String()
}
