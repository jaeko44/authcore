package idp

import (
	"context"
	"encoding/json"
	"time"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/user"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-playground/validator/v10"
)

// IDP represents a third-party identity provider such as social login.
type IDP interface {
	// ID returns the name of the provider.
	ID() string

	// AuthorizationURL returns a third-party authorization endpoint URI used by the client to obtain
	// authorization from the ID provider. This method also returns a state that is used to recover
	// the state later.
	AuthorizationURL(stateToken string) (string, State, error)

	// Exchange converts an authorization code into tokens and the user's identity. This method
	// takes that State created in CreateAuthorizationURI and an authorization code obtained from
	// identity provider.
	Exchange(ctx context.Context, state State, code string) (*AuthorizationGrant, error)
}

// AuthorizationGrant is set of credentials representing the authorization.
type AuthorizationGrant struct {
	AccessToken  string    `json:"access_token"`
	AccessSecret string    `json:"access_secret"` // used by OAuth 1.0
	TokenType    string    `json:"token_type,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	Expiry       time.Time `json:"expiry,omitempty"`
	IDToken      string    `json:"id_token,omitempty"`
	Identity     *Identity `json:"identity,omitempty"`
}

// Identity represents a third-party user identity. It can be created from the content of an OIDC ID
// token or from a platform's user info endpoint. All fields are optional except "sub".
type Identity struct {
	ID                  string `json:"sub" validate:"required"`
	Name                string `json:"name"`
	PreferredUsername   string `json:"preferred_username"`
	Email               string `json:"email"`
	EmailVerified       bool   `json:"email_verified"`
	PhoneNumber         string `json:"phone_number"`
	PhoneNumberVerified bool   `json:"phone_number_verified"`
}

// IdentityFromIDToken converts an IDToken to Identity.
func IdentityFromIDToken(idToken string, keyFunc jwt.Keyfunc) (*Identity, error) {
	token, err := jwt.Parse(idToken, keyFunc)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}
	if !token.Valid {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}
	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}
	err = mapClaims.Valid()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}

	jsonString, err := json.Marshal(mapClaims)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	ident := new(Identity)
	err = json.Unmarshal(jsonString, ident)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	err = validate.Struct(ident)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}
	return ident, nil
}

// State represents a state of a IDP authorization flow.
type State string

// Factory creates verifier instances.
type Factory struct {
	idps map[string]IDP
}

// NewFactory returns a new Factory.
func NewFactory() *Factory {
	return &Factory{
		idps: make(map[string]IDP),
	}
}

// Register registers an IDP
func (f *Factory) Register(idp IDP) {
	f.idps[idp.ID()] = idp
}

// IDP returns an IDP with the given identifier.
func (f *Factory) IDP(identifier string) (IDP, error) {
	idp, ok := f.idps[identifier]
	if !ok {
		errors.Errorf("unknown ID provider: %v", identifier)
	}
	return idp, nil
}

// IDToOAuthService converts a string into user.OAuthService used in database models.
func IDToOAuthService(idp string) (user.OAuthService, error) {
	switch idp {
	case Google:
		return user.OAuthGoogle, nil
	case Facebook:
		return user.OAuthFacebook, nil
	case Twitter:
		return user.OAuthTwitter, nil
	case Apple:
		return user.OAuthApple, nil
	case Matters:
		return user.OAuthMatters, nil
	case "mock":
		return user.OAuthService(999), nil // for testing only
	default:
		return 0, errors.New(errors.ErrorInvalidArgument, "unknown IDP")
	}
}

// OAuthServiceToID converts a user.OAuthService to IDP identifier.
func OAuthServiceToID(service user.OAuthService) (string, error) {
	switch service {
	case user.OAuthGoogle:
		return Google, nil
	case user.OAuthFacebook:
		return Facebook, nil
	case user.OAuthTwitter:
		return Twitter, nil
	case user.OAuthApple:
		return Apple, nil
	case user.OAuthMatters:
		return Matters, nil
	default:
		return "", errors.New(errors.ErrorInvalidArgument, "unknown IDP")
	}
}

var validate = validator.New()
