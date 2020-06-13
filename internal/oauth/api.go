package oauth

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/square/go-jose.v2"

	"authcore.io/authcore/internal/authn"
	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/session"
	"authcore.io/authcore/internal/user"
)

// API registers handlers for OAuth 2.0 and OIDC-compatible endpoints.
func API(userStore *user.Store, sessionStore *session.Store, tc *authn.TransactionController) func(e *echo.Echo) {
	return func(e *echo.Echo) {
		h := &handler{
			tc:           tc,
			sessionStore: sessionStore,
			userStore:    userStore,
		}
		e.GET("/oauth/authorize", h.Authorize)
		e.POST("/oauth/token", h.Token)
		e.GET("/oauth/userinfo", h.UserInfo)
		e.GET("/.well-known/openid-configuration", h.OpenIDConfiguration)
		e.GET("/.well-known/jwks.json", h.JWKS)
	}
}

type handler struct {
	tc           *authn.TransactionController
	sessionStore *session.Store
	userStore    *user.Store
}

// Authorize implements OAuth 2.0 Authorization Endpoint.
func (h *handler) Authorize(c echo.Context) error {
	r := new(AuthorizeRequest)
	if err := c.Bind(r); err != nil {
		return err
	}
	if err := c.Validate(r); err != nil {
		return err
	}

	// ValidateRedirectURI also validates client_id
	if err := authn.ValidateRedirectURI(r.ClientID, r.RedirectURI); err != nil {
		return err
	}

	if !isResponseTypeSupported(r.ResponseType) {
		return errors.New(errors.ErrorInvalidArgument, "invalid response_type")
	}

	if r.CodeChallengeMethod != "" && r.CodeChallengeMethod != "S256" {
		return errors.New(errors.ErrorInvalidArgument, "invalid code_challenge_method")
	}

	// Redirect to sign in widget
	redirectURL, err := url.Parse("/widgets/signin")
	if err != nil {
		log.Fatal(err)
	}

	q := redirectURL.Query()
	q.Add("responseType", r.ResponseType)
	q.Add("clientId", r.ClientID)
	q.Add("redirectURI", r.RedirectURI)
	q.Add("scope", r.Scope)
	q.Add("clientState", r.State)
	// directly pass as code challenge method forbades "plain" and empty if code challenge exists.
	q.Add("codeChallenge", r.CodeChallenge)
	q.Add("codeChallengeMethod", r.CodeChallengeMethod)
	redirectURL.RawQuery = q.Encode()
	c.Redirect(http.StatusFound, redirectURL.String())
	return nil
}

// Token implements OAuth 2.0 Token endpoint.
func (h *handler) Token(c echo.Context) error {
	r := new(TokenRequest)
	if err := c.Bind(r); err != nil {
		return err
	}
	if err := c.Validate(r); err != nil {
		return err
	}
	ctx := c.Request().Context()

	var sess *session.Session
	var err error
	switch strings.ToLower(r.GrantType) {
	case "authorization_code":
		sess, err = h.tc.ExchangeSession(ctx, r.ClientID, r.RedirectURI, r.Code, r.CodeVerifier)
	case "refresh_token":
		sess, err = h.sessionStore.FindSessionByRefreshToken(ctx, r.RefreshToken)
		if err != nil {
			return err
		}
		sess.Refresh(ctx, false)
		sess, err = h.sessionStore.UpdateSession(ctx, sess)
	default:
		return errors.New(errors.ErrorInvalidArgument, "unsupported grant_type")
	}
	if err != nil {
		return err
	}

	accessToken, err := h.sessionStore.GenerateAccessToken(ctx, sess, true)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &TokenResponse{
		TokenType:    "bearer",
		AccessToken:  accessToken.AccessToken,
		IDToken:      accessToken.IDToken,
		ExpiresIn:    accessToken.ExpiresIn,
		RefreshToken: sess.RefreshToken,
	})
}

// UserInfo implements OIDC UserInfo endpoint.
func (h *handler) UserInfo(c echo.Context) error {
	c.String(http.StatusOK, "")
	return nil
}

// OpenIDConfiguration implements OIDC Discovery endpoint.
func (h *handler) OpenIDConfiguration(c echo.Context) error {
	baseURL, err := url.Parse(viper.GetString("base_url"))
	if err != nil {
		log.Fatalf("configuration error: invalid base_url: %v", err)
	}
	authorizationEndpoint, err := baseURL.Parse("/oauth/authorize")
	if err != nil {
		log.Fatalf("configuration error: invalid base_url: %v", err)
	}
	tokenEndpoint, err := baseURL.Parse("/oauth/token")
	if err != nil {
		log.Fatalf("configuration error: invalid base_url: %v", err)
	}
	jwksURI, err := baseURL.Parse("/.well-known/jwks.json")
	resp := &OpenIDConfigurationResponse{
		Issuer:                 baseURL.String(),
		AuthorizationEndpoint:  authorizationEndpoint.String(),
		TokenEndpoint:          tokenEndpoint.String(),
		JWKSURI:                jwksURI.String(),
		ResponseTypesSupported: responseTypesSupported(),
	}
	c.JSON(http.StatusOK, resp)
	return nil
}

// JWKS implements OIDC JWKS endpoint.
func (h *handler) JWKS(c echo.Context) error {
	publicKey, err := h.sessionStore.AccessTokenPublicKey()
	if err != nil {
		return err
	}
	jwks := jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{*publicKey},
	}
	return c.JSON(http.StatusOK, jwks)
}

func isResponseTypeSupported(responseType string) bool {
	for _, rt := range responseTypesSupported() {
		if rt == responseType {
			return true
		}
	}
	return false
}

func responseTypesSupported() []string {
	return []string{"code", "code id_token", "id_token", "token id_token"}
}

// AuthorizeRequest is the request for Authorize.
type AuthorizeRequest struct {
	ResponseType        string `query:"response_type" validate:"required"`
	ClientID            string `query:"client_id" validate:"required"`
	RedirectURI         string `query:"redirect_uri" validate:"required"`
	Scope               string `query:"scope"`
	State               string `query:"state"`
	CodeChallenge       string `query:"code_challenge"`
	CodeChallengeMethod string `query:"code_challenge_method" validate:"required_with=CodeChallenge"` // forbade empty if code challenge exists
}

// TokenRequest is the request for Token.
type TokenRequest struct {
	ClientID     string `json:"client_id" form:"client_id"`
	ClientSecret string `json:"client_secret" form:"client_secret"`
	GrantType    string `json:"grant_type" form:"grant_type" validate:"required"`
	Code         string `json:"code" form:"code"`
	CodeVerifier string `json:"code_verifier" form:"code_verifier"`
	RefreshToken string `json:"refresh_token" form:"refresh_token"`
	RedirectURI  string `json:"redirect_uri" form:"redirect_uri"`
}

// TokenResponse is the response for Token.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	IDToken      string `json:"id_token"`
}

// OpenIDConfigurationResponse is the response for OpenIDConfiguration
type OpenIDConfigurationResponse struct {
	Issuer                 string   `json:"issuer"`
	AuthorizationEndpoint  string   `json:"authorization_endpoint"`
	TokenEndpoint          string   `json:"token_endpoint"`
	JWKSURI                string   `json:"jwks_uri"`
	ResponseTypesSupported []string `json:"response_types_supported"`
}
